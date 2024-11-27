package blocksync

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	bcproto "github.com/cometbft/cometbft/api/cometbft/blocksync/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
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
	peerID nodekey.ID
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
	localAddr     crypto.Address
	poolRoutineWg sync.WaitGroup

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError

	switchToConsensusMs int

	metrics *Metrics
}

// NewReactor returns new reactor instance.
func NewReactor(state sm.State, blockExec *sm.BlockExecutor, store *store.BlockStore,
	blockSync bool, localAddr crypto.Address, metrics *Metrics, offlineStateSyncHeight int64,
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

	// It's okay to block since sendRequest is called from a separate goroutine
	// (bpRequester#requestRoutine; 1 per each peer).
	requestsCh := make(chan BlockRequest)

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
		localAddr:    localAddr,
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
		return bcR.startPool(false)
	}
	return nil
}

// SwitchToBlockSync is called by the statesync reactor when switching to blocksync.
func (bcR *Reactor) SwitchToBlockSync(state sm.State) error {
	bcR.blockSync = true

	if !state.IsEmpty() { // if we have a state, start from there
		bcR.initialState = state
		bcR.pool.height = state.LastBlockHeight + 1
		return bcR.startPool(true)
	}

	// if we don't have a state due to an error or a timeout, start from genesis.
	return bcR.startPool(false)
}

func (bcR *Reactor) startPool(stateSynced bool) error {
	err := bcR.pool.Start()
	if err != nil {
		return err
	}
	bcR.poolRoutineWg.Add(1)
	go func() {
		defer bcR.poolRoutineWg.Done()
		bcR.poolRoutine(stateSynced)
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

// StreamDescriptors implements Reactor.
func (*Reactor) StreamDescriptors() []transport.StreamDescriptor {
	return []transport.StreamDescriptor{
		tcpconn.StreamDescriptor{
			ID:                  BlocksyncChannel,
			Priority:            5,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: MaxMsgSize,
			MessageTypeI:        &bcproto.Message{},
		},
	}
}

// AddPeer implements Reactor by sending our state to peer.
func (bcR *Reactor) AddPeer(peer p2p.Peer) {
	// it's OK if send fails. will try later in poolRoutine
	_ = peer.Send(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.StatusResponse{
			Base:   bcR.store.Base(),
			Height: bcR.store.Height(),
		},
	})

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerRange
}

// RemovePeer implements Reactor by removing peer from the pool.
func (bcR *Reactor) RemovePeer(peer p2p.Peer, _ any) {
	bcR.pool.RemovePeer(peer.ID())
}

// respondToPeer loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
func (bcR *Reactor) respondToPeer(msg *bcproto.BlockRequest, src p2p.Peer) (queued bool) {
	block, _ := bcR.store.LoadBlock(msg.Height)
	if block == nil {
		bcR.Logger.Info("Peer asking for a block we don't have", "src", src, "height", msg.Height)
		err := src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message:   &bcproto.NoBlockResponse{Height: msg.Height},
		})
		return err == nil
	}

	state, err := bcR.blockExec.Store().Load()
	if err != nil {
		bcR.Logger.Error("loading state", "err", err)
		return false
	}
	var extCommit *types.ExtendedCommit
	if state.ConsensusParams.Feature.VoteExtensionsEnabled(msg.Height) {
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

	err = src.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.BlockResponse{
			Block:     bl,
			ExtCommit: extCommit.ToProto(),
		},
	})
	return err == nil
}

func (bcR *Reactor) handlePeerResponse(msg *bcproto.BlockResponse, src p2p.Peer) {
	bi, err := types.BlockFromProto(msg.Block)
	if err != nil {
		bcR.Logger.Error("Peer sent us invalid block", "peer", src, "msg", msg, "err", err)
		bcR.Switch.StopPeerForError(src, err)
		return
	}
	var extCommit *types.ExtendedCommit
	if msg.ExtCommit != nil {
		var err error
		extCommit, err = types.ExtendedCommitFromProto(msg.ExtCommit)
		if err != nil {
			bcR.Logger.Error("failed to convert extended commit from proto",
				"peer", src,
				"err", err)
			bcR.Switch.StopPeerForError(src, err)
			return
		}
	}

	if err := bcR.pool.AddBlock(src.ID(), bi, extCommit, msg.Block.Size()); err != nil {
		bcR.Logger.Error("failed to add block", "peer", src, "err", err)
	}
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
		go bcR.handlePeerResponse(msg, e.Src)
	case *bcproto.StatusRequest:
		// Send peer our state.
		_ = e.Src.TrySend(p2p.Envelope{
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
		bcR.pool.RedoRequestFrom(msg.Height, e.Src.ID())
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (bcR *Reactor) poolRoutine(stateSynced bool) {
	bcR.metrics.Syncing.Set(1)
	defer bcR.metrics.Syncing.Set(0)

	var (
		blocksSynced = uint64(0)
		state        = bcR.initialState
		lastHundred  = time.Now()
		lastRate     = 0.0
		didProcessCh = make(chan struct{}, 1)
	)

	go bcR.handleBlockRequestsRoutine()

	if bcR.switchToConsensusMs == 0 {
		bcR.switchToConsensusMs = switchToConsensusIntervalSeconds * 1000
	}
	switchToConsensusTicker := time.NewTicker(time.Duration(bcR.switchToConsensusMs) * time.Millisecond)
	defer switchToConsensusTicker.Stop()

	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	defer trySyncTicker.Stop()

	initialCommitHasExtensions := (bcR.initialState.LastBlockHeight > 0 && bcR.store.LoadBlockExtendedCommit(bcR.initialState.LastBlockHeight) != nil)

FOR_LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			outbound, inbound, _ := bcR.Switch.NumPeers()
			bcR.Logger.Debug("Consensus ticker", "outbound", outbound, "inbound", inbound, "lastHeight", state.LastBlockHeight)

			if !initialCommitHasExtensions {
				// If require extensions, but since we don't have them yet, then we cannot switch to consensus yet.
				if bcR.isMissingExtension(state, blocksSynced) {
					continue FOR_LOOP
				}
			}

			if bcR.isCaughtUp(state, blocksSynced, stateSynced) {
				break FOR_LOOP
			}

		case <-trySyncTicker.C: // chan time
			select {
			case didProcessCh <- struct{}{}:
			default:
			}

		case <-didProcessCh:
			// NOTE: It is a subtle mistake to process more than a single block
			// at a time (e.g. 10) here, because we only TrySend 1 request per
			// loop.  The ratio mismatch can result in starving of blocks, a
			// sudden burst of requests and responses, and repeat.
			// Consequently, it is better to split these routines rather than
			// coupling them as it's written here.  TODO uncouple from request
			// routine.

			// See if there are any blocks to sync.
			first, second, extCommit := bcR.pool.PeekTwoBlocks()
			if first == nil || second == nil {
				// we need to have fetched two consecutive blocks in order to
				// perform blocksync verification
				continue FOR_LOOP
			}
			// Some sanity checks on heights
			if state.LastBlockHeight > 0 && state.LastBlockHeight+1 != first.Height {
				// Panicking because the block pool's height  MUST keep consistent with the state; the block pool is totally under our control
				panic(fmt.Errorf("peeked first block has unexpected height; expected %d, got %d", state.LastBlockHeight+1, first.Height))
			}
			if first.Height+1 != second.Height {
				// Panicking because this is an obvious bug in the block pool, which is totally under our control
				panic(fmt.Errorf("heights of first and second block are not consecutive; expected %d, got %d", state.LastBlockHeight, first.Height))
			}

			// Before priming didProcessCh for another check on the next
			// iteration, break the loop if the BlockPool or the Reactor itself
			// has quit. This avoids case ambiguity of the outer select when two
			// channels are ready.
			if !bcR.IsRunning() || !bcR.pool.IsRunning() {
				break FOR_LOOP
			}
			// Try again quickly next loop.
			didProcessCh <- struct{}{}

			firstParts, err := first.MakePartSet(types.BlockPartSizeBytes)
			if err != nil {
				bcR.Logger.Error("failed to make ",
					"height", first.Height,
					"err", err.Error())
				break FOR_LOOP
			}

			if state, err = bcR.processBlock(first, second, firstParts, state, extCommit); err != nil {
				bcR.Logger.Error("Invalid block", "height", first.Height, "err", err)
				continue FOR_LOOP
			}

			blocksSynced++

			if blocksSynced%100 == 0 {
				_, height, maxPeerHeight := bcR.pool.IsCaughtUp()
				lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
				bcR.Logger.Info("Block Sync Rate", "height", height, "max_peer_height", maxPeerHeight, "blocks/s", lastRate)
				lastHundred = time.Now()
			}

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

func (bcR *Reactor) handleBlockRequest(request BlockRequest) {
	peer := bcR.Switch.Peers().Get(request.PeerID)
	if peer == nil {
		return
	}
	err := peer.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message:   &bcproto.BlockRequest{Height: request.Height},
	})
	if err != nil {
		if e, ok := err.(p2p.SendError); ok && e.Full() {
			bcR.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
		}
	}
}

func (bcR *Reactor) handleBlockRequestsRoutine() {
	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	defer statusUpdateTicker.Stop()

	for {
		select {
		case <-bcR.Quit():
			return
		case <-bcR.pool.Quit():
			return
		case request := <-bcR.requestsCh:
			bcR.handleBlockRequest(request)
		case err := <-bcR.errorsCh:
			peer := bcR.Switch.Peers().Get(err.peerID)
			if peer != nil {
				bcR.Switch.StopPeerForError(peer, err)
			}
		case <-statusUpdateTicker.C:
			// ask for status updates
			go bcR.BroadcastStatusRequest()
		}
	}
}

func (bcR *Reactor) isMissingExtension(state sm.State, blocksSynced uint64) bool {
	// The "if" statement below is a bit confusing, so here is a breakdown
	// of its logic and purpose:
	//
	// If we are at genesis (no block in the chain), we don't need VoteExtensions
	// because the first block's LastCommit is empty anyway.
	//
	// If VoteExtensions were disabled for the previous height then we don't need
	// VoteExtensions.
	//
	// If we have sync'd at least one block, then we are guaranteed to have extensions
	// if we need them by the logic inside loop FOR_LOOP: it requires that the blocks
	// it fetches have extensions if extensions were enabled during the height.
	//
	// If we already had extensions for the initial height (e.g. we are recovering),
	// then we are guaranteed to have extensions for the last block (if required) even
	// if we did not blocksync any block.
	//
	missingExtension := true
	voteExtensionsDisabled := state.LastBlockHeight > 0 && !state.ConsensusParams.Feature.VoteExtensionsEnabled(state.LastBlockHeight)
	if state.LastBlockHeight == 0 || voteExtensionsDisabled || blocksSynced > 0 {
		missingExtension = false
	}

	if missingExtension {
		bcR.Logger.Info(
			"no extended commit yet",
			"last_block_height", state.LastBlockHeight,
			"vote_extensions_disabled", voteExtensionsDisabled,
			"blocks_synced", blocksSynced,
		)
	}

	return missingExtension
}

func (bcR *Reactor) isCaughtUp(state sm.State, blocksSynced uint64, stateSynced bool) bool {
	if isCaughtUp, height, _ := bcR.pool.IsCaughtUp(); isCaughtUp || state.Validators.ValidatorBlocksTheChain(bcR.localAddr) {
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
		// else {
		// should only happen during testing
		// }
		return true
	}
	return false
}

func (bcR *Reactor) processBlock(first, second *types.Block, firstParts *types.PartSet, state sm.State, extCommit *types.ExtendedCommit) (sm.State, error) {
	var (
		chainID            = bcR.initialState.ChainID
		firstPartSetHeader = firstParts.Header()
		firstID            = types.BlockID{Hash: first.Hash(), PartSetHeader: firstPartSetHeader}
	)

	// Finally, verify the first block using the second's commit
	// NOTE: we can probably make this more efficient, but note that calling
	// first.Hash() doesn't verify the tx contents, so MakePartSet() is
	// currently necessary.
	// TODO(sergio): Should we also validate against the extended commit?
	err := state.Validators.VerifyCommitLight(
		chainID, firstID, first.Height, second.LastCommit)

	if err == nil {
		// validate the block before we persist it
		err = bcR.blockExec.ValidateBlock(state, first)
	}

	presentExtCommit := extCommit != nil
	extensionsEnabled := state.ConsensusParams.Feature.VoteExtensionsEnabled(first.Height)
	if presentExtCommit != extensionsEnabled {
		err = fmt.Errorf("non-nil extended commit must be received iff vote extensions are enabled for its height "+
			"(height %d, non-nil extended commit %t, extensions enabled %t)",
			first.Height, presentExtCommit, extensionsEnabled,
		)
	}
	if err == nil && extensionsEnabled {
		// if vote extensions were required at this height, ensure they exist.
		err = extCommit.EnsureExtensions(true)
	}

	if err != nil {
		peerID := bcR.pool.RemovePeerAndRedoAllPeerRequests(first.Height)
		peer := bcR.Switch.Peers().Get(peerID)
		if peer != nil {
			// NOTE: we've already removed the peer's request, but we
			// still need to clean up the rest.
			bcR.Switch.StopPeerForError(peer, ErrReactorValidation{Err: err})
		}
		peerID2 := bcR.pool.RemovePeerAndRedoAllPeerRequests(second.Height)
		peer2 := bcR.Switch.Peers().Get(peerID2)
		if peer2 != nil && peer2 != peer {
			// NOTE: we've already removed the peer's request, but we
			// still need to clean up the rest.
			bcR.Switch.StopPeerForError(peer2, ErrReactorValidation{Err: err})
		}
		return state, err
	}

	// SUCCESS. Pop the block from the pool.
	bcR.pool.PopRequest()

	// TODO: batch saves so we dont persist to disk every block
	if extensionsEnabled {
		bcR.store.SaveBlockWithExtendedCommit(first, firstParts, extCommit)
	} else {
		// We use LastCommit here instead of extCommit. extCommit is not
		// guaranteed to be populated by the peer if extensions are not enabled.
		// Currently, the peer should provide an extCommit even if the vote extension data are absent
		// but this may change so using second.LastCommit is safer.
		bcR.store.SaveBlock(first, firstParts, second.LastCommit)
	}

	// TODO: same thing for app - but we would need a way to
	// get the hash without persisting the state
	state, err = bcR.blockExec.ApplyVerifiedBlock(state, firstID, first, bcR.pool.MaxPeerHeight())
	if err != nil {
		// TODO This is bad, are we zombie?
		panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
	}

	bcR.metrics.recordBlockMetrics(first)

	return state, nil
}
