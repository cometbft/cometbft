package blocksync

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	bcproto "github.com/cometbft/cometbft/api/cometbft/blocksync/v1"
	sm "github.com/cometbft/cometbft/internal/state"
	"github.com/cometbft/cometbft/internal/store"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

const (
	// BlocksyncChannel is a channel for blocks and status updates (`BlockStore` height).
	BlocksyncChannel = byte(0x40)

	trySyncIntervalMS = 10

	// stop syncing when last block's time is
	// within this much of the system time.
	// stopSyncingDurationMinutes = 10.

	// ask for best height every 10s.
	statusUpdateIntervalSeconds = 10
	// check if we should switch to consensus reactor.
	switchToConsensusIntervalSeconds = 1
)

type consensusReactor interface {
	// for when we switch from blocksync reactor and block sync to
	// the consensus machine
	SwitchToConsensus(state sm.State, skipWAL bool)
}

type mempoolReactor interface {
	// for when we finish doing block sync or state sync
	EnableInOutTxs()
}

type peerError struct {
	err    error
	peerID p2p.ID
}

func (e peerError) Error() string {
	return fmt.Sprintf("error with peer %v: %s", e.peerID, e.err.Error())
}

// Reactor handles long-term catchup syncing.
type Reactor struct {
	p2p.BaseReactor

	// immutable
	initialState sm.State

	blockExec     *sm.BlockExecutor
	store         sm.BlockStore
	pool          *BlockPool
	blockSync     bool
	poolRoutineWg sync.WaitGroup

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError

	switchToConsensusMs int

	metrics *Metrics

	// tickers
	trySyncTicker           *time.Ticker
	statusUpdateTicker      *time.Ticker
	switchToConsensusTicker *time.Ticker
}

// NewReactor returns new reactor instance.
func NewReactor(state sm.State, blockExec *sm.BlockExecutor, store *store.BlockStore,
	blockSync bool, metrics *Metrics, offlineStateSyncHeight int64,
) *Reactor {
	storeHeight := store.Height()
	if storeHeight == 0 {
		// If state sync was performed offline and the stores were bootstrapped to height H
		// the state store's lastHeight will be H while blockstore's Height and Base are still 0
		// 1. This scenario should not lead to a panic in this case, which is indicated by
		// having a OfflineStateSyncHeight > 0
		// 2. We need to instruct the blocksync reactor to start fetching blocks from H+1
		// instead of 0.
		storeHeight = offlineStateSyncHeight
	}
	if state.LastBlockHeight != storeHeight {
		panic(fmt.Sprintf("state (%v) and store (%v) height mismatch, stores were left in an inconsistent state", state.LastBlockHeight,
			storeHeight))
	}
	requestsCh := make(chan BlockRequest, maxTotalRequesters)

	const capacity = 1000                      // must be bigger than peers count
	errorsCh := make(chan peerError, capacity) // so we don't block in #Receive#pool.AddBlock

	startHeight := storeHeight + 1
	if startHeight == 1 {
		startHeight = state.InitialHeight
	}
	pool := NewBlockPool(startHeight, requestsCh, errorsCh)

	bcR := &Reactor{
		initialState: state,
		blockExec:    blockExec,
		store:        store,
		pool:         pool,
		blockSync:    blockSync,
		requestsCh:   requestsCh,
		errorsCh:     errorsCh,
		metrics:      metrics,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor("Reactor", bcR)
	return bcR
}

// SetLogger implements service.Service by setting the logger on reactor and pool.
func (bcR *Reactor) SetLogger(l log.Logger) {
	bcR.BaseService.Logger = l
	bcR.pool.Logger = l
}

// OnStart implements Reactor.
func (bcR *Reactor) OnStart() error {
	if bcR.blockSync {
		err := bcR.pool.Start()
		if err != nil {
			return err
		}
		bcR.initTickers() // Initialize tickers here
		bcR.poolRoutineWg.Add(1)
		go func() {
			defer bcR.poolRoutineWg.Done()
			bcR.poolRoutine(false)
		}()
	}
	return nil
}

// SwitchToBlockSync is called by the state sync reactor when switching to block sync.
func (bcR *Reactor) SwitchToBlockSync(state sm.State) error {
	bcR.blockSync = true
	bcR.initialState = state

	bcR.pool.height = state.LastBlockHeight + 1
	err := bcR.pool.Start()
	if err != nil {
		return err
	}
	bcR.poolRoutineWg.Add(1)
	go func() {
		defer bcR.poolRoutineWg.Done()
		bcR.poolRoutine(true)
	}()
	return nil
}

// OnStop implements Reactor.
// OnStop implements Reactor.
func (bcR *Reactor) OnStop() {
	if bcR.blockSync {
		if err := bcR.pool.Stop(); err != nil {
			bcR.Logger.Error("Error stopping pool", "err", err)
		}
		bcR.trySyncTicker.Stop()
		bcR.statusUpdateTicker.Stop()
		bcR.switchToConsensusTicker.Stop()
		bcR.poolRoutineWg.Wait()
	}
}

// GetChannels implements Reactor.
func (*Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  BlocksyncChannel,
			Priority:            5,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: MaxMsgSize,
			MessageType:         &bcproto.Message{},
		},
	}
}

// AddPeer implements Reactor by sending our state to peer.
func (bcR *Reactor) AddPeer(peer p2p.Peer) {
	peer.Send(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.StatusResponse{
			Base:   bcR.store.Base(),
			Height: bcR.store.Height(),
		},
	})
	// it's OK if send fails. will try later in poolRoutine

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerRange
}

// RemovePeer implements Reactor by removing peer from the pool.
func (bcR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
	bcR.pool.RemovePeer(peer.ID())
}

// respondToPeer loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
func (bcR *Reactor) respondToPeer(msg *bcproto.BlockRequest, src p2p.Peer) (queued bool) {
	block, _ := bcR.store.LoadBlock(msg.Height)
	if block == nil {
		bcR.Logger.Info("Peer asking for a block we don't have", "src", src, "height", msg.Height)
		return src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message:   &bcproto.NoBlockResponse{Height: msg.Height},
		})
	}

	state, err := bcR.blockExec.Store().Load()
	if err != nil {
		bcR.Logger.Error("loading state", "err", err)
		return false
	}
	var extCommit *types.ExtendedCommit
	if state.ConsensusParams.ABCI.VoteExtensionsEnabled(msg.Height) {
		extCommit = bcR.store.LoadBlockExtendedCommit(msg.Height)
		if extCommit == nil {
			bcR.Logger.Error("found block in store with no extended commit", "block", block)
			return false
		}
	}

	bl, err := block.ToProto()
	if err != nil {
		bcR.Logger.Error("could not convert msg to protobuf", "err", err)
		return false
	}

	return src.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.BlockResponse{
			Block:     bl,
			ExtCommit: extCommit.ToProto(),
		},
	})
}

// Receive implements Reactor by handling 4 types of messages (look below).
func (bcR *Reactor) Receive(e p2p.Envelope) {
	if err := ValidateMsg(e.Message); err != nil {
		bcR.Logger.Error("Peer sent us invalid msg", "peer", e.Src, "msg", e.Message, "err", err)
		bcR.Switch.StopPeerForError(e.Src, err)
		return
	}

	bcR.Logger.Debug("Receive", "e.Src", e.Src, "chID", e.ChannelID, "msg", e.Message)

	switch msg := e.Message.(type) {
	case *bcproto.BlockRequest:
		bcR.respondToPeer(msg, e.Src)
	case *bcproto.BlockResponse:
		bi, err := types.BlockFromProto(msg.Block)
		if err != nil {
			bcR.Logger.Error("Peer sent us invalid block", "peer", e.Src, "msg", e.Message, "err", err)
			bcR.Switch.StopPeerForError(e.Src, err)
			return
		}
		var extCommit *types.ExtendedCommit
		if msg.ExtCommit != nil {
			var err error
			extCommit, err = types.ExtendedCommitFromProto(msg.ExtCommit)
			if err != nil {
				bcR.Logger.Error("failed to convert extended commit from proto",
					"peer", e.Src,
					"err", err)
				bcR.Switch.StopPeerForError(e.Src, err)
				return
			}
		}

		if err := bcR.pool.AddBlock(e.Src.ID(), bi, extCommit, msg.Block.Size()); err != nil {
			bcR.Logger.Error("failed to add block", "err", err)
		}
	case *bcproto.StatusRequest:
		// Send peer our state.
		e.Src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message: &bcproto.StatusResponse{
				Height: bcR.store.Height(),
				Base:   bcR.store.Base(),
			},
		})
	case *bcproto.StatusResponse:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerRange(e.Src.ID(), msg.Base, msg.Height)
	case *bcproto.NoBlockResponse:
		bcR.Logger.Debug("Peer does not have requested block", "peer", e.Src, "height", msg.Height)
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or block sync will be slower.
func (bcR *Reactor) poolRoutine(stateSynced bool) {
	bcR.metrics.Syncing.Set(1)
	defer bcR.metrics.Syncing.Set(0)

	blocksSynced := uint64(0)

	chainID := bcR.initialState.ChainID
	state := bcR.initialState

	lastHundred := time.Now()
	lastRate := 0.0

	didProcessCh := make(chan struct{}, 1)

	// Handle requests and errors from the pool.
	go bcR.handleRequestsAndErrors()

	// FOR_LOOP: is the main loop of the block sync reactor.  It handles the following:
	// - Switching to consensus mode
	// - Processing and applying blocks
	// - Logging the sync rate
	// - Removing the processed block from the pool's request queue
	// - Handling errors during block processing
FOR_LOOP:
	for {
		select {
		case <-bcR.switchToConsensusTicker.C:
			if bcR.handleSwitchToConsensusTicker(&state, &blocksSynced, stateSynced) {
				break FOR_LOOP // exit the loop if ready to switch to consensus
			}

		case <-bcR.trySyncTicker.C: // chan time
			select {
			case didProcessCh <- struct{}{}:
			default:
			}

		case <-didProcessCh:
			// Check if there are any blocks to sync.
			first, second, extCommit := bcR.pool.PeekTwoBlocks()

			// Ensure we have two consecutive blocks for validation.
			if first == nil || second == nil {
				continue FOR_LOOP // Need two blocks for validation, continue loop.
			}

			// Sanity check: Ensure the heights of blocks are consecutive.
			if state.LastBlockHeight > 0 && state.LastBlockHeight+1 != first.Height {
				panic(fmt.Errorf("peeked first block has unexpected height; expected %d, got %d", state.LastBlockHeight+1, first.Height))
			}

			// Sanity check: Ensure the heights of blocks are consecutive.
			if first.Height+1 != second.Height {
				panic(fmt.Errorf("heights of first and second block are not consecutive; expected %d, got %d", state.LastBlockHeight, first.Height))
			}

			// Check for extended commit if required by consensus parameters.
			if extCommit == nil && state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
				panic(fmt.Errorf("peeked first block without extended commit at height %d - possible node store corruption", first.Height))
			}

			// Prepare for block processing.
			firstParts, err := first.MakePartSet(types.BlockPartSizeBytes)
			if err != nil {
				bcR.Logger.Error("failed to create part set", "height", first.Height, "err", err)
				continue FOR_LOOP // Skip processing this block on error.
			}

			// Process and apply the block. This includes validation and state update.
			if err := bcR.processAndApplyBlock(&state, first, second.LastCommit, extCommit, firstParts, chainID); err != nil {
				bcR.Logger.Error("error in processing and applying block", "height", first.Height, "err", err)

				// Handle error by potentially redoing the request from another peer.
				bcR.handleBlockProcessingError(first.Height, second.Height)
				continue FOR_LOOP
			}

			// Update metrics and increment the number of blocks synced.
			bcR.metrics.recordBlockMetrics(first)
			blocksSynced++

			// Log the sync rate every 100 blocks.
			if blocksSynced%100 == 0 {
				lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
				bcR.Logger.Info("block sync rate", "height", bcR.pool.height, "max_peer_height", bcR.pool.MaxPeerHeight(), "blocks/s", lastRate)
				lastHundred = time.Now()
			}

			// Remove the processed block from the pool's request queue.
			bcR.pool.PopRequest()

			continue FOR_LOOP

		case <-bcR.Quit():
			break FOR_LOOP
		case <-bcR.pool.Quit():
			break FOR_LOOP
		}
	}
}

// BroadcastStatusRequest broadcasts `BlockStore` base and height.
func (bcR *Reactor) BroadcastStatusRequest() {
	bcR.Switch.Broadcast(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message:   &bcproto.StatusRequest{},
	})
}

// handleBlockRequest processes a block request from the requests channel.
func (bcR *Reactor) handleBlockRequest(request BlockRequest) {
	peer := bcR.Switch.Peers().Get(request.PeerID)
	if peer == nil {
		return
	}
	queued := peer.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message:   &bcproto.BlockRequest{Height: request.Height},
	})
	if !queued {
		bcR.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
	}
}

// handlePeerError processes an error reported by a peer.
func (bcR *Reactor) handlePeerError(err peerError) {
	peer := bcR.Switch.Peers().Get(err.peerID)
	if peer != nil {
		bcR.Switch.StopPeerForError(peer, err)
	}
}

// handleSwitchToConsensusTicker checks if the node is ready to switch to consensus mode.
func (bcR *Reactor) handleSwitchToConsensusTicker(state *sm.State, blocksSynced *uint64, stateSynced bool) bool {
	height, numPending, lenRequesters := bcR.pool.GetStatus()
	outbound, inbound, _ := bcR.Switch.NumPeers()
	bcR.Logger.Debug("Consensus ticker", "numPending", numPending, "total", lenRequesters,
		"outbound", outbound, "inbound", inbound, "lastHeight", state.LastBlockHeight)

	// Check if we should be using vote extensions.
	missingExtension := true
	if state.LastBlockHeight == 0 ||
		!state.ConsensusParams.ABCI.VoteExtensionsEnabled(state.LastBlockHeight) ||
		*blocksSynced > 0 ||
		(bcR.initialState.LastBlockHeight > 0 && bcR.store.LoadBlockExtendedCommit(bcR.initialState.LastBlockHeight) != nil) {
		missingExtension = false
	}

	// Check if we should switch to consensus mode.
	if !missingExtension && bcR.pool.IsCaughtUp() {
		bcR.Logger.Info("Switching to consensus mode!", "height", height)
		if err := bcR.pool.Stop(); err != nil {
			bcR.Logger.Error("Error stopping pool", "err", err)
		}
		if memR, ok := bcR.Switch.Reactor("MEMPOOL").(mempoolReactor); ok {
			memR.EnableInOutTxs()
		}
		if conR, ok := bcR.Switch.Reactor("CONSENSUS").(consensusReactor); ok {
			conR.SwitchToConsensus(*state, *blocksSynced > 0 || stateSynced)
		}
		return true // indicate that the switch to consensus is ready
	}

	return false // indicate that the switch to consensus is not ready
}

// processAndApplyBlock validates and applies a block to the state.
func (bcR *Reactor) processAndApplyBlock(state *sm.State, first *types.Block, second *types.Commit, extCommit *types.ExtendedCommit, firstParts *types.PartSet, chainID string) error {
	firstID := types.BlockID{Hash: first.Hash(), PartSetHeader: firstParts.Header()}

	// Verify the first block using the second's commit
	if err := state.Validators.VerifyCommitLight(chainID, firstID, first.Height, second); err != nil {
		return err
	}

	// Validate the block before persisting it
	if err := bcR.blockExec.ValidateBlock(*state, first); err != nil {
		return err
	}

	// Ensure vote extensions if required
	if state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
		if err := extCommit.EnsureExtensions(true); err != nil {
			return err
		}
	} else if extCommit != nil {
		return fmt.Errorf("received non-nil extCommit for height %d (extensions disabled)", first.Height)
	}

	// Save and apply the block
	if state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
		bcR.store.SaveBlockWithExtendedCommit(first, firstParts, extCommit)
	} else {
		bcR.store.SaveBlock(first, firstParts, second)
	}

	newState, err := bcR.blockExec.ApplyBlock(*state, firstID, first)
	if err != nil {
		return fmt.Errorf("failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err)
	}

	*state = newState // Update the state reference
	return nil
}

// handleBlockProcessingError handles errors during block processing by redoing requests from peers.
func (bcR *Reactor) handleBlockProcessingError(firstBlockHeight, secondBlockHeight int64) {
	// Redo the request for the first block from another peer.
	peerID1 := bcR.pool.RedoRequest(firstBlockHeight)
	peer1 := bcR.Switch.Peers().Get(peerID1)
	if peer1 != nil {
		// Stop the peer for error.
		bcR.Switch.StopPeerForError(peer1, ErrReactorValidation{Err: fmt.Errorf("error in block validation at height %d", firstBlockHeight)})
	}

	// Redo the request for the second block from another peer.
	peerID2 := bcR.pool.RedoRequest(secondBlockHeight)
	// Check if the second peer is different from the first peer to avoid redundancy.
	if peerID2 != peerID1 {
		peer2 := bcR.Switch.Peers().Get(peerID2)
		if peer2 != nil {
			// Stop the peer for error.
			bcR.Switch.StopPeerForError(peer2, ErrReactorValidation{Err: fmt.Errorf("error in block validation at height %d", secondBlockHeight)})
		}
	}
}

// initTickers initializes all tickers used in the Reactor.
func (bcR *Reactor) initTickers() {
	bcR.trySyncTicker = time.NewTicker(trySyncIntervalMS * time.Millisecond)
	bcR.statusUpdateTicker = time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	if bcR.switchToConsensusMs == 0 {
		bcR.switchToConsensusMs = switchToConsensusIntervalSeconds * 1000
	}
	bcR.switchToConsensusTicker = time.NewTicker(time.Duration(bcR.switchToConsensusMs) * time.Millisecond)
}

// handleRequestsAndErrors handles requests and errors from the pool.
func (bcR *Reactor) handleRequestsAndErrors() {
	for {
		select {
		case <-bcR.Quit():
			return
		case <-bcR.pool.Quit():
			return
		case request := <-bcR.requestsCh:
			bcR.handleBlockRequest(request)
		case err := <-bcR.errorsCh:
			bcR.handlePeerError(err)

		case <-bcR.statusUpdateTicker.C:
			// ask for status updates
			go bcR.BroadcastStatusRequest()
		}
	}
}
