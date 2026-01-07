package blocksync

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	bcproto "github.com/cometbft/cometbft/proto/tendermint/blocksync"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

// BlocksyncChannel is a channel for blocks and status updates (`BlockStore` height)
const BlocksyncChannel = byte(0x40)

const (
	// interval for asking other peers their base (min) and height (max) blocks
	intervalStatusUpdate = 10 * time.Second

	// interval for checking whether it's time to switch from block-sync to consensus
	intervalSwitchToConsensus = 1 * time.Second

	// interval for trying to apply a block
	intervalTrySync = 10 * time.Millisecond
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

	// indicates whether the reactor is enabled
	// eg it can be initially disabled due to state sync being performed
	enabled *atomic.Bool

	// if enabled, we suppress switching to consensus mode, keeping node perpetually in block sync
	followerMode bool

	blockExec     *sm.BlockExecutor
	store         sm.BlockStore
	pool          *BlockPool
	localAddr     crypto.Address
	poolRoutineWg sync.WaitGroup

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError

	// interval for checking whether it's time to switch from block-sync to consensus
	intervalSwitchToConsensus time.Duration

	metrics *Metrics
}

// NewReactorWithAddr returns new Reactor instance with local address
func NewReactor(
	enabled bool,
	followerMode bool,
	state sm.State,
	blockExec *sm.BlockExecutor,
	store *store.BlockStore,
	localAddr crypto.Address,
	offlineStateSyncHeight int64,
	metrics *Metrics,
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
		panic(fmt.Sprintf(
			"state (%d) and store (%d) height mismatch, stores were left in an inconsistent state",
			state.LastBlockHeight,
			storeHeight,
		))
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

	enabledFlag := &atomic.Bool{}
	enabledFlag.Store(enabled)

	r := &Reactor{
		initialState:              state,
		blockExec:                 blockExec,
		store:                     store,
		pool:                      pool,
		enabled:                   enabledFlag,
		followerMode:              followerMode,
		localAddr:                 localAddr,
		requestsCh:                requestsCh,
		errorsCh:                  errorsCh,
		metrics:                   metrics,
		intervalSwitchToConsensus: intervalSwitchToConsensus,
	}

	r.BaseReactor = *p2p.NewBaseReactor("Blocksync", r)

	return r
}

// SetLogger implements service.Service by setting the logger on reactor and pool.
func (r *Reactor) SetLogger(l log.Logger) {
	r.Logger = l
	r.pool.Logger = l
}

// OnStart implements service.Service.
func (r *Reactor) OnStart() error {
	// noop
	if !r.enabled.Load() {
		return nil
	}

	return r.runPool(false)
}

func (r *Reactor) runPool(stateSynced bool) error {
	if err := r.pool.Start(); err != nil {
		return err
	}

	r.poolRoutineWg.Add(1)
	go func() {
		defer r.poolRoutineWg.Done()
		r.poolRoutine(stateSynced)
	}()

	return nil
}

// Enable is called by the state sync reactor when switching to block sync.
func (r *Reactor) Enable(state sm.State) error {
	if !r.enabled.CompareAndSwap(false, true) {
		return ErrAlreadyEnabled
	}

	r.initialState = state
	r.pool.height = state.LastBlockHeight + 1

	return r.runPool(true)
}

// OnStop implements service.Service.
func (r *Reactor) OnStop() {
	if !r.enabled.Load() {
		return
	}

	if err := r.pool.Stop(); err != nil {
		r.Logger.Error("Error stopping pool", "err", err)
	}

	r.poolRoutineWg.Wait()
}

// GetChannels implements Reactor
func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
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
func (r *Reactor) AddPeer(peer p2p.Peer) {
	peer.Send(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.StatusResponse{
			Base:   r.store.Base(),
			Height: r.store.Height(),
		},
	})
	// it's OK if send fails. will try later in poolRoutine

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerRange
}

// RemovePeer implements Reactor by removing peer from the pool.
func (r *Reactor) RemovePeer(peer p2p.Peer, _ any) {
	r.pool.RemovePeer(peer.ID())
}

// respondToPeer loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
func (r *Reactor) respondToPeer(msg *bcproto.BlockRequest, src p2p.Peer) {
	block := r.store.LoadBlock(msg.Height)
	if block == nil {
		r.Logger.Info("Peer asking for a block we don't have", "src", src, "height", msg.Height)
		src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message:   &bcproto.NoBlockResponse{Height: msg.Height},
		})

		return
	}

	state, err := r.blockExec.Store().Load()
	if err != nil {
		r.Logger.Error("Unable to load the state", "err", err)
		return
	}

	var extCommit *types.ExtendedCommit
	if state.ConsensusParams.ABCI.VoteExtensionsEnabled(msg.Height) {
		extCommit = r.store.LoadBlockExtendedCommit(msg.Height)
		if extCommit == nil {
			r.Logger.Error("Found block in store with no extended commit", "block", block)
			return
		}
	}

	bl, err := block.ToProto()
	if err != nil {
		r.Logger.Error("Unable to convert the block to protobuf", "err", err)
		return
	}

	src.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.BlockResponse{
			Block:     bl,
			ExtCommit: extCommit.ToProto(),
		},
	})
}

func (r *Reactor) handlePeerResponse(msg *bcproto.BlockResponse, src p2p.Peer) {
	bi, err := types.BlockFromProto(msg.Block)
	if err != nil {
		r.Logger.Error("Peer sent us invalid block", "peer", src, "msg", msg, "err", err)
		r.Switch.StopPeerForError(src, err)
		return
	}

	var extCommit *types.ExtendedCommit
	if msg.ExtCommit != nil {
		extCommit, err = types.ExtendedCommitFromProto(msg.ExtCommit)
		if err != nil {
			r.Logger.Error("Failed to convert extended commit from proto", "peer", src, "err", err)
			r.Switch.StopPeerForError(src, err)
			return
		}
	}

	if err := r.pool.AddBlock(src.ID(), bi, extCommit, msg.Block.Size()); err != nil {
		r.Logger.Error("Failed to add block", "peer", src, "err", err)
	}
}

// Receive implements Reactor by handling 4 types of messages (look below).
func (r *Reactor) Receive(e p2p.Envelope) {
	if err := ValidateMsg(e.Message); err != nil {
		r.Logger.Error("Peer sent us invalid msg", "peer", e.Src, "msg", e.Message, "err", err)
		r.Switch.StopPeerForError(e.Src, err)
		return
	}

	r.Logger.Debug("Receive", "e.Src", e.Src, "chID", e.ChannelID, "msg", e.Message)

	switch msg := e.Message.(type) {
	case *bcproto.BlockRequest:
		// sends block response
		go r.respondToPeer(msg, e.Src)
	case *bcproto.BlockResponse:
		// adds block to the pool
		go r.handlePeerResponse(msg, e.Src)
	case *bcproto.StatusRequest:
		// Send peer our state.
		go e.Src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message: &bcproto.StatusResponse{
				Height: r.store.Height(),
				Base:   r.store.Base(),
			},
		})
	case *bcproto.StatusResponse:
		// Got a peer status. Unverified.
		r.pool.SetPeerRange(e.Src.ID(), msg.Base, msg.Height)
	case *bcproto.NoBlockResponse:
		r.Logger.Debug("Peer does not have requested block", "peer", e.Src, "height", msg.Height)
		r.pool.RedoRequestFrom(msg.Height, e.Src.ID())
	default:
		r.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

func (r *Reactor) localNodeBlocksTheChain(state sm.State) bool {
	_, val := state.Validators.GetByAddress(r.localAddr)
	if val == nil {
		return false
	}
	total := state.Validators.TotalVotingPower()
	return val.VotingPower >= total/3
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (r *Reactor) poolRoutine(stateSynced bool) {
	r.metrics.Syncing.Set(1)
	defer r.metrics.Syncing.Set(0)

	trySyncTicker := time.NewTicker(intervalTrySync)
	defer trySyncTicker.Stop()

	statusUpdateTicker := time.NewTicker(intervalStatusUpdate)
	defer statusUpdateTicker.Stop()

	switchToConsensusTicker := time.NewTicker(r.intervalSwitchToConsensus)
	defer switchToConsensusTicker.Stop()

	go r.poolEventsRoutine(statusUpdateTicker)

	var (
		chainID                    = r.initialState.ChainID
		state                      = r.initialState
		initialCommitHasExtensions = state.LastBlockHeight > 0 &&
			r.store.LoadBlockExtendedCommit(state.LastBlockHeight) != nil

		didProcessCh = make(chan struct{}, 1)

		// metrics tracking
		blocksSynced = 0
		lastHundred  = time.Now()
		lastRate     = 0.0
	)

FOR_LOOP:
	for {
		select {
		case <-r.Quit():
			break FOR_LOOP
		case <-r.pool.Quit():
			break FOR_LOOP
		case <-switchToConsensusTicker.C:
			height, numPending, lenRequesters := r.pool.GetStatus()
			outbound, inbound, _ := r.Switch.NumPeers()

			r.Logger.Debug(
				"Consensus ticker",
				"numPending", numPending,
				"total", lenRequesters,
				"outbound", outbound,
				"inbound", inbound,
				"lastHeight", state.LastBlockHeight,
			)

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
			if state.LastBlockHeight == 0 ||
				!state.ConsensusParams.ABCI.VoteExtensionsEnabled(state.LastBlockHeight) ||
				blocksSynced > 0 ||
				initialCommitHasExtensions {
				missingExtension = false
			}

			// If require extensions, but since we don't have them yet, then we cannot switch to consensus yet.
			if missingExtension {
				r.Logger.Info(
					"No extended commit yet",
					"height", height,
					"last_block_height", state.LastBlockHeight,
					"initial_height", state.InitialHeight,
					"max_peer_height", r.pool.MaxPeerHeight(),
				)
				continue FOR_LOOP
			}

			// keep syncing
			if !r.pool.IsCaughtUp() && !r.localNodeBlocksTheChain(state) {
				continue FOR_LOOP
			}

			if r.followerMode {
				r.Logger.Debug("Follower mode is enabled, continuing to sync", "height", state.LastBlockHeight)
				continue FOR_LOOP
			}

			r.Logger.Info("Time to switch to consensus mode!", "height", height)
			if err := r.pool.Stop(); err != nil {
				r.Logger.Error("Error stopping pool", "err", err)
			}

			memR, exists := r.Switch.Reactor("MEMPOOL")
			if exists {
				if memR, ok := memR.(mempoolReactor); ok {
					memR.EnableInOutTxs()
				}
			}

			conR, exists := r.Switch.Reactor("CONSENSUS")
			if exists {
				if conR, ok := conR.(consensusReactor); ok {
					conR.SwitchToConsensus(state, blocksSynced > 0 || stateSynced)
				}
			}

			break FOR_LOOP
		case <-trySyncTicker.C:
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
			first, second, extCommit := r.pool.PeekTwoBlocks()
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
			if !r.IsRunning() || !r.pool.IsRunning() {
				break FOR_LOOP
			}
			// Try again quickly next loop.
			didProcessCh <- struct{}{}

			firstParts, err := first.MakePartSet(types.BlockPartSizeBytes)
			if err != nil {
				r.Logger.Error("Failed to make part set", "height", first.Height, "err", err.Error())
				break FOR_LOOP
			}

			firstPartSetHeader := firstParts.Header()
			firstID := types.BlockID{Hash: first.Hash(), PartSetHeader: firstPartSetHeader}

			// Finally, verify the first block using the second's commit
			// NOTE: we can probably make this more efficient, but note that calling
			// first.Hash() doesn't verify the tx contents, so MakePartSet() is
			// currently necessary.
			// TODO(sergio): Should we also validate against the extended commit?
			err = state.Validators.VerifyCommitLight(chainID, firstID, first.Height, second.LastCommit)

			if err == nil {
				// validate the block before we persist it
				err = r.blockExec.ValidateBlock(state, first)
			}
			presentExtCommit := extCommit != nil
			extensionsEnabled := state.ConsensusParams.ABCI.VoteExtensionsEnabled(first.Height)
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
				r.Logger.Error("Error in validation", "err", err)
				peerID := r.pool.RemovePeerAndRedoAllPeerRequests(first.Height)
				peer := r.Switch.Peers().Get(peerID)
				if peer != nil {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					r.Switch.StopPeerForError(peer, ErrReactorValidation{Err: err})
				}
				peerID2 := r.pool.RemovePeerAndRedoAllPeerRequests(second.Height)
				peer2 := r.Switch.Peers().Get(peerID2)
				if peer2 != nil && peer2 != peer {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					r.Switch.StopPeerForError(peer2, ErrReactorValidation{Err: err})
				}
				continue FOR_LOOP
			}

			r.pool.PopRequest()

			// TODO: batch saves so we don't persist to disk every block
			if extensionsEnabled {
				r.store.SaveBlockWithExtendedCommit(first, firstParts, extCommit)
			} else {
				// We use LastCommit here instead of extCommit. extCommit is not
				// guaranteed to be populated by the peer if extensions are not enabled.
				// Currently, the peer should provide an extCommit even if the vote extension data are absent
				// but this may change so using second.LastCommit is safer.
				r.store.SaveBlock(first, firstParts, second.LastCommit)
			}

			// TODO: same thing for app - but we would need a way to
			// get the hash without persisting the state
			state, err = r.blockExec.ApplyVerifiedBlock(state, firstID, first)
			if err != nil {
				// TODO This is bad, are we zombie?
				panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
			}

			r.metrics.recordBlockMetrics(first)
			blocksSynced++

			if blocksSynced%100 == 0 {
				lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
				lastHundred = time.Now()
				r.Logger.Info(
					"Block Sync Rate",
					"height", r.pool.height,
					"max_peer_height", r.pool.MaxPeerHeight(),
					"blocks/s", lastRate,
				)
			}

			continue FOR_LOOP
		}
	}
}

func (r *Reactor) poolEventsRoutine(statusUpdateTicker *time.Ticker) {
	for {
		select {
		case <-r.Quit():
			return
		case <-r.pool.Quit():
			return
		case request := <-r.requestsCh:
			// request is pushed to the requestsCh by the pool internally.
			peer := r.Switch.Peers().Get(request.PeerID)
			if peer == nil {
				continue
			}

			queued := peer.TrySend(p2p.Envelope{
				ChannelID: BlocksyncChannel,
				Message:   &bcproto.BlockRequest{Height: request.Height},
			})

			if !queued {
				r.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
			}
		case err := <-r.errorsCh:
			// error is pushed to the errorsCh by the pool internally.
			peer := r.Switch.Peers().Get(err.peerID)
			if peer != nil {
				r.Switch.StopPeerForError(peer, err)
			}

		case <-statusUpdateTicker.C:
			// ask other peers for status updates
			r.Switch.BroadcastAsync(p2p.Envelope{
				ChannelID: BlocksyncChannel,
				Message:   &bcproto.StatusRequest{},
			})
		}
	}
}
