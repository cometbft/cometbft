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

// OnStart implements service.Service.
func (bcR *Reactor) OnStart() error {
	if bcR.blockSync {
		err := bcR.pool.Start()
		if err != nil {
			return err
		}
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

// OnStop implements service.Service.
func (bcR *Reactor) OnStop() {
	if bcR.blockSync {
		if err := bcR.pool.Stop(); err != nil {
			bcR.Logger.Error("Error stopping pool", "err", err)
		}
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

func (bcR *Reactor) handleConsensusSwitch(state sm.State, blocksSynced uint64, stateSynced bool) bool {
	height, numPending, lenRequesters := bcR.pool.GetStatus()
	outbound, inbound, _ := bcR.Switch.NumPeers()
	bcR.Logger.Debug("Consensus ticker", "numPending", numPending, "total", lenRequesters,
		"outbound", outbound, "inbound", inbound, "lastHeight", state.LastBlockHeight)

	missingExtension := true
	if state.LastBlockHeight == 0 ||
		!state.ConsensusParams.ABCI.VoteExtensionsEnabled(state.LastBlockHeight) ||
		blocksSynced > 0 ||
		(bcR.initialState.LastBlockHeight > 0 && bcR.store.LoadBlockExtendedCommit(bcR.initialState.LastBlockHeight) != nil) {
		missingExtension = false
	}

	if missingExtension {
		bcR.Logger.Info(
			"no extended commit yet",
			"height", height,
			"last_block_height", state.LastBlockHeight,
			"initial_height", state.InitialHeight,
			"max_peer_height", bcR.pool.MaxPeerHeight(),
		)
		return false
	}
	if bcR.pool.IsCaughtUp() {
		bcR.Logger.Info("Time to switch to consensus mode!", "height", height)
		if err := bcR.pool.Stop(); err != nil {
			bcR.Logger.Error("Error stopping pool", "err", err)
		}
		if memR, ok := bcR.Switch.Reactor("MEMPOOL").(mempoolReactor); ok {
			memR.EnableInOutTxs()
		}
		if conR, ok := bcR.Switch.Reactor("CONSENSUS").(consensusReactor); ok {
			conR.SwitchToConsensus(state, blocksSynced > 0 || stateSynced)
		}
		return true
	}
	return false
}

func (bcR *Reactor) handleBlockSync(state sm.State, didProcessCh chan struct{}) {
	first, second, extCommit := bcR.pool.PeekTwoBlocks()
	if first == nil || second == nil {
		return
	}

	// Log the heights of the blocks we are about to process
	bcR.Logger.Info("Processing blocks", "firstHeight", first.Height, "secondHeight", second.Height)

	// Ensure the blocks are in the correct order
	if first.Height >= second.Height {
		bcR.Logger.Error("received blocks out of order", "firstHeight", first.Height, "secondHeight", second.Height)
		return
	}

	// Additional checks to ensure we have the correct first block
	if state.LastBlockHeight > 0 && state.LastBlockHeight+1 != first.Height {
		bcR.Logger.Error("received unexpected first block height", "expected", state.LastBlockHeight+1, "received", first.Height)
		return
	}
	if state.LastBlockHeight > 0 && state.LastBlockHeight+1 != first.Height {
		panic(fmt.Errorf("peeked first block has unexpected height; expected %d, got %d", state.LastBlockHeight+1, first.Height))
	}
	if first.Height+1 != second.Height {
		panic(fmt.Errorf("heights of first and second block are not consecutive; expected %d, got %d", state.LastBlockHeight, first.Height))
	}
	if extCommit == nil && state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
		panic(fmt.Errorf("peeked first block without extended commit at height %d - possible node store corruption", first.Height))
	}

	if !bcR.IsRunning() || !bcR.pool.IsRunning() {
		return
	}
	didProcessCh <- struct{}{}

	firstParts, err := first.MakePartSet(types.BlockPartSizeBytes)
	if err != nil {
		bcR.Logger.Error("failed to make part set", "height", first.Height, "err", err)
		return
	}
	firstPartSetHeader := firstParts.Header()
	firstID := types.BlockID{Hash: first.Hash(), PartSetHeader: firstPartSetHeader}
	err = state.Validators.VerifyCommitLight(state.ChainID, firstID, first.Height, second.LastCommit)
	if err == nil {
		err = bcR.blockExec.ValidateBlock(state, first)
	}
	if err == nil {
		if state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
			err = extCommit.EnsureExtensions(true)
		} else if extCommit != nil {
			err = fmt.Errorf("received non-nil extCommit for height %d (extensions disabled)", first.Height)
		}
		if err != nil {
			bcR.Logger.Error("Error in validation", "err", err)
			peerID := bcR.pool.RedoRequest(first.Height)
			peer := bcR.Switch.Peers().Get(peerID)
			if peer != nil {
				bcR.Switch.StopPeerForError(peer, ErrReactorValidation{Err: err})
			}
			return
		}
	}

	bcR.pool.PopRequest()

	if state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height) {
		bcR.store.SaveBlockWithExtendedCommit(first, firstParts, extCommit)
	} else {
		bcR.store.SaveBlock(first, firstParts, second.LastCommit)
	}

	state, err = bcR.blockExec.ApplyBlock(state, firstID, first)
	if err != nil {
		panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
	}
	bcR.metrics.recordBlockMetrics(first)
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (bcR *Reactor) poolRoutine(stateSynced bool) {
	bcR.metrics.Syncing.Set(1)
	defer bcR.metrics.Syncing.Set(0)

	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	defer trySyncTicker.Stop()

	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	defer statusUpdateTicker.Stop()

	if bcR.switchToConsensusMs == 0 {
		bcR.switchToConsensusMs = switchToConsensusIntervalSeconds * 1000
	}
	switchToConsensusTicker := time.NewTicker(time.Duration(bcR.switchToConsensusMs) * time.Millisecond)
	defer switchToConsensusTicker.Stop()

	blocksSynced := uint64(0)
	state := bcR.initialState
	didProcessCh := make(chan struct{}, 1)

	go bcR.handleRequests(didProcessCh)
	go bcR.handleStatusUpdates(statusUpdateTicker)

FOR_LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			if bcR.handleConsensusSwitch(state, blocksSynced, stateSynced) {
				break FOR_LOOP
			}
		case <-trySyncTicker.C:
			// Send a request to try syncing blocks
			select {
			case didProcessCh <- struct{}{}:
			default:
			}
		case <-didProcessCh:
			bcR.handleBlockSync(state, didProcessCh)
		case <-bcR.Quit():
			// Reactor is quitting
			break FOR_LOOP
		case <-bcR.pool.Quit():
			// Block pool has quit
			break FOR_LOOP
		}
	}
}

func (bcR *Reactor) handleRequests(didProcessCh chan struct{}) {
	for {
		select {
		case <-bcR.Quit():
			return
		case request := <-bcR.requestsCh:
			peer := bcR.Switch.Peers().Get(request.PeerID)
			if peer == nil {
				continue
			}
			queued := peer.TrySend(p2p.Envelope{
				ChannelID: BlocksyncChannel,
				Message:   &bcproto.BlockRequest{Height: request.Height},
			})
			if !queued {
				bcR.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
			}
		}
	}
}

func (bcR *Reactor) handleStatusUpdates(statusUpdateTicker *time.Ticker) {
	for {
		select {
		case <-bcR.Quit():
			return
		case <-statusUpdateTicker.C:
			bcR.BroadcastStatusRequest()
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
