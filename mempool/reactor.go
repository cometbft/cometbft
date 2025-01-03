package mempool

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	abcicli "github.com/cometbft/cometbft/abci/client"
	protomem "github.com/cometbft/cometbft/api/cometbft/mempool/v2"
	cfg "github.com/cometbft/cometbft/config"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/types"
)

// A number in the open interval (0, 100) representing a percentage of
// config.DOGTargetRedundancy. In the DOG protocol, it defines acceptable lower
// and upper bounds for redundancy levels as a deviation from the target value.
const targetRedundancyDeltaPercent = 10

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config  *cfg.MempoolConfig
	mempool *CListMempool

	waitSync   atomic.Bool
	waitSyncCh chan struct{} // for signaling when to start receiving and sending txs

	// DOG protocol: Control enabled/disabled routes for disseminating txs.
	router            *gossipRouter
	redundancyControl *redundancyControl

	// Semaphores to keep track of how many connections to peers are active for broadcasting
	// transactions. Each semaphore has a capacity that puts an upper bound on the number of
	// connections for different groups of peers.
	activePersistentPeersSemaphore    *semaphore.Weighted
	activeNonPersistentPeersSemaphore *semaphore.Weighted
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(config *cfg.MempoolConfig, mempool *CListMempool, waitSync bool) *Reactor {
	memR := &Reactor{
		config:   config,
		mempool:  mempool,
		waitSync: atomic.Bool{},
	}
	memR.BaseReactor = *p2p.NewBaseReactor("Mempool", memR)
	if waitSync {
		memR.waitSync.Store(true)
		memR.waitSyncCh = make(chan struct{})
	}
	memR.activePersistentPeersSemaphore = semaphore.NewWeighted(int64(memR.config.ExperimentalMaxGossipConnectionsToPersistentPeers))
	memR.activeNonPersistentPeersSemaphore = semaphore.NewWeighted(int64(memR.config.ExperimentalMaxGossipConnectionsToNonPersistentPeers))

	return memR
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l log.Logger) {
	memR.Logger = l
	memR.mempool.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	if memR.WaitSync() {
		memR.Logger.Info("Starting reactor in sync mode: tx propagation will start once sync completes")
	}
	if !memR.config.Broadcast {
		memR.Logger.Info("Tx broadcasting is disabled")
	}
	if memR.config.DOGProtocolEnabled {
		memR.router = newGossipRouter()
		memR.redundancyControl = newRedundancyControl(memR.config)
		go memR.redundancyControl.controlLoop(memR)
	}
	return nil
}

// StreamDescriptors implements Reactor by returning the list of channels for this
// reactor.
func (memR *Reactor) StreamDescriptors() []p2p.StreamDescriptor {
	var (
		batchMsgSize  int
		haveTxMsgSize int
	)

	// Calculate max message size for batchMsg and haveTxMsg,
	// and free the memory immediately after.
	{
		largestTx := make([]byte, memR.config.MaxTxBytes)
		batchMsg := protomem.Message{
			Sum: &protomem.Message_Txs{
				Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
			},
		}
		batchMsgSize = batchMsg.Size()

		key := types.Tx(largestTx).Key()
		haveTxMsg := protomem.Message{
			Sum: &protomem.Message_HaveTx{HaveTx: &protomem.HaveTx{TxKey: key[:]}},
		}
		haveTxMsgSize = haveTxMsg.Size()
	}

	return []p2p.StreamDescriptor{
		tcpconn.StreamDescriptor{
			ID:                  MempoolChannel,
			Priority:            5,
			RecvMessageCapacity: batchMsgSize,
			MessageTypeI:        &protomem.Message{},
		},
		tcpconn.StreamDescriptor{
			ID:                  MempoolControlChannel,
			Priority:            10,
			RecvMessageCapacity: haveTxMsgSize,
			MessageTypeI:        &protomem.Message{},
		},
	}
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
func (memR *Reactor) AddPeer(peer p2p.Peer) {
	if memR.config.Broadcast && peer.HasChannel(MempoolChannel) {
		go func() {
			// Always forward transactions to unconditional peers.
			if !memR.Switch.IsPeerUnconditional(peer.ID()) {
				// Depending on the type of peer, we choose a semaphore to limit the gossiping peers.
				var peerSemaphore *semaphore.Weighted
				if peer.IsPersistent() && memR.config.ExperimentalMaxGossipConnectionsToPersistentPeers > 0 {
					peerSemaphore = memR.activePersistentPeersSemaphore
				} else if !peer.IsPersistent() && memR.config.ExperimentalMaxGossipConnectionsToNonPersistentPeers > 0 {
					peerSemaphore = memR.activeNonPersistentPeersSemaphore
				}

				if peerSemaphore != nil {
					for peer.IsRunning() {
						// Block on the semaphore until a slot is available to start gossiping with this peer.
						// Do not block indefinitely, in case the peer is disconnected before gossiping starts.
						ctxTimeout, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
						// Block sending transactions to peer until one of the connections become
						// available in the semaphore.
						err := peerSemaphore.Acquire(ctxTimeout, 1)
						cancel()

						if err != nil {
							continue
						}

						// Release semaphore to allow other peer to start sending transactions.
						defer peerSemaphore.Release(1)
						break
					}
				}
			}

			memR.mempool.metrics.ActiveOutboundConnections.Add(1)
			defer memR.mempool.metrics.ActiveOutboundConnections.Add(-1)
			memR.broadcastTxRoutine(peer)
		}()
	}
}

func (memR *Reactor) RemovePeer(peer p2p.Peer, _ any) {
	if memR.router != nil {
		// Remove all routes with peer as source or target and immediately
		// adjust redundancy.
		memR.router.resetRoutes(peer.ID())
		memR.redundancyControl.triggerAdjustment(memR)
		memR.mempool.metrics.DisabledRoutes.Set(float64(memR.router.numRoutes()))
	}
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (memR *Reactor) Receive(e p2p.Envelope) {
	if memR.WaitSync() {
		memR.Logger.Debug("Ignore message received while syncing", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		return
	}

	senderID := e.Src.ID()

	switch e.ChannelID {
	case MempoolControlChannel:
		switch msg := e.Message.(type) {
		case *protomem.HaveTx:
			txKey := types.TxKey(msg.GetTxKey())
			if len(txKey) == 0 {
				memR.Logger.Error("Received empty HaveTx message from peer", "src", e.Src.ID())
				return
			}
			memR.Logger.Debug("Received HaveTx", "from", senderID, "txKey", txKey)

			if memR.router != nil {
				// Get tx's list of senders.
				senders, err := memR.mempool.GetSenders(txKey)
				if err != nil || len(senders) == 0 || senders[0] == noSender {
					// It is possible that tx got removed from the mempool.
					memR.Logger.Debug("Received HaveTx but failed to get sender", "tx", txKey.Hash(), "err", err)
					return
				}
				// Disable route with tx's first sender as source and peer as target.
				memR.router.disableRoute(senders[0], senderID)

				memR.Logger.Debug("Disable route", "source", senders[0], "target", senderID)
				memR.mempool.metrics.DisabledRoutes.Set(float64(memR.router.numRoutes()))
			}

		case *protomem.ResetRoute:
			memR.Logger.Debug("Received Reset", "from", senderID)
			if memR.router != nil {
				memR.router.resetRandomRouteWithTarget(senderID)
				memR.mempool.metrics.DisabledRoutes.Set(float64(memR.router.numRoutes()))
			}

		default:
			memR.Logger.Error("Unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
			memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
		}

	case MempoolChannel:
		switch msg := e.Message.(type) {
		case *protomem.Txs:
			protoTxs := msg.GetTxs()
			if len(protoTxs) == 0 {
				memR.Logger.Error("Received empty Txs message from peer", "src", e.Src.ID())
				return
			}

			memR.Logger.Debug("Received Txs", "from", senderID, "msg", e.Message)
			for _, txBytes := range protoTxs {
				_, _ = memR.TryAddTx(types.Tx(txBytes), e.Src)
			}

		default:
			memR.Logger.Error("Unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
			memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
			return
		}

	default:
		memR.Logger.Error("Unknown channel", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message on channel: %T", e.Message))
	}

	// broadcasting happens from go routines per peer
}

// TryAddTx attempts to add an incoming transaction to the mempool.
// When the sender is nil, it means the transaction comes from an RPC endpoint.
func (memR *Reactor) TryAddTx(tx types.Tx, sender p2p.Peer) (*abcicli.ReqRes, error) {
	senderID := noSender
	if sender != nil {
		senderID = sender.ID()
	}

	reqRes, err := memR.mempool.CheckTx(tx, senderID)
	if err != nil {
		txKey := tx.Key()
		switch {
		case errors.Is(err, ErrTxInCache):
			memR.Logger.Debug("Tx already exists in cache", "tx", txKey.Hash(), "sender", senderID)
			if memR.redundancyControl != nil {
				memR.redundancyControl.incDuplicateTxs()
				if memR.redundancyControl.isHaveTxBlocked() {
					return nil, err
				}
				err := sender.Send(p2p.Envelope{ChannelID: MempoolControlChannel, Message: &protomem.HaveTx{TxKey: txKey[:]}})
				if err != nil {
					memR.Logger.Error("Failed to send HaveTx message", "peer", senderID, "txKey", txKey, "err", err)
				} else {
					memR.Logger.Debug("Sent HaveTx message", "tx", txKey.Hash(), "peer", senderID)
					memR.router.insertRequestedPartialGossip(sender.ID())
					// Block HaveTx and restart timer, during which time, sending HaveTx is not allowed.
					memR.redundancyControl.blockHaveTx()
				}
			}
			return nil, err

		case errors.As(err, &ErrMempoolIsFull{}):
			// using debug level to avoid flooding when traffic is high
			memR.Logger.Debug(err.Error())
			return nil, err

		default:
			memR.Logger.Info("Could not check tx", "tx", txKey.Hash(), "sender", senderID, "err", err)
			return nil, err
		}
	}

	if memR.redundancyControl != nil {
		memR.redundancyControl.incFirstTimeTxs()
	}

	return reqRes, nil
}

func (memR *Reactor) EnableInOutTxs() {
	memR.Logger.Info("Enabling inbound and outbound transactions")
	if !memR.waitSync.CompareAndSwap(true, false) {
		return
	}

	// Releases all the blocked broadcastTxRoutine instances.
	if memR.config.Broadcast {
		close(memR.waitSyncCh)
	}
}

func (memR *Reactor) WaitSync() bool {
	return memR.waitSync.Load()
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// Send new mempool txs to peer.
func (memR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
	// If the node is catching up, don't start this routine immediately.
	if memR.WaitSync() {
		select {
		case <-memR.waitSyncCh:
			// EnableInOutTxs() has set WaitSync() to false.
		case <-memR.Quit():
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-peer.Quit():
			cancel()
		case <-memR.Quit():
			cancel()
		}
	}()

	iter := NewBlockingIterator(ctx, memR.mempool, string(peer.ID()))
	for {
		// In case of both next.NextWaitChan() and peer.Quit() are variable at the same time
		if !memR.IsRunning() || !peer.IsRunning() {
			return
		}

		entry := <-iter.WaitNextCh()
		// If the entry we were looking at got garbage collected (removed), try again.
		if entry == nil {
			continue
		}

		// If we suspect that the peer is lagging behind, at least by more than
		// one block, we don't send the transaction immediately. This code
		// reduces the mempool size and the recheck-tx rate of the receiving
		// node. See [RFC 103] for an analysis on this optimization.
		//
		// [RFC 103]: https://github.com/CometBFT/cometbft/blob/main/docs/references/rfc/rfc-103-incoming-txs-when-catching-up.md
		for {
			// Make sure the peer's state is up to date. The peer may not have a
			// state yet. We set it in the consensus reactor, but when we add
			// peer in Switch, the order we call reactors#AddPeer is different
			// every time due to us using a map. Sometimes other reactors will
			// be initialized before the consensus reactor. We should wait a few
			// milliseconds and retry.
			peerState, ok := peer.Get(types.PeerStateKey).(PeerState)
			if ok && peerState.GetHeight()+1 >= entry.Height() {
				break
			}
			select {
			case <-time.After(PeerCatchupSleepIntervalMS * time.Millisecond):
			case <-peer.Quit():
				return
			case <-memR.Quit():
				return
			}
		}

		// NOTE: Transaction batching was disabled due to
		// https://github.com/tendermint/tendermint/issues/5796

		// We are paying the cost of computing the transaction hash in
		// any case, even when logger level > debug. So it only once.
		// See: https://github.com/cometbft/cometbft/issues/4167
		txKey := entry.Tx().Key()
		txHash := txKey.Hash()

		if entry.IsSender(peer.ID()) {
			// Do not send this transaction if we receive it from peer.
			memR.Logger.Debug("Skipping transaction, peer is sender",
				"tx", txHash, "peer", peer.ID())
			continue
		}

		if memR.router != nil {
			// Do not send if the route from the first sender to peer is disabled.
			senders := entry.Senders()
			if len(senders) > 0 && memR.router.isRouteDisabled(senders[0], peer.ID()) {
				memR.Logger.Debug("Disabled route: do not send transaction to peer",
					"tx", txHash, "peer", peer.ID(), "senders", senders)
				continue
			}
		}

		for {
			// The entry may have been removed from the mempool since it was
			// chosen at the beginning of the loop. Skip it if that's the case.
			if !memR.mempool.Contains(txKey) {
				break
			}

			memR.Logger.Debug("Sending transaction to peer",
				"tx", txHash, "peer", peer.ID())

			err := peer.Send(p2p.Envelope{
				ChannelID: MempoolChannel,
				Message:   &protomem.Txs{Txs: [][]byte{entry.Tx()}},
			})
			if err == nil {
				break
			}

			memR.Logger.Debug("Failed sending transaction to peer",
				"tx", txHash, "peer", peer.ID())

			select {
			case <-time.After(PeerCatchupSleepIntervalMS * time.Millisecond):
			case <-peer.Quit():
				return
			case <-memR.Quit():
				return
			}
		}
	}
}

type gossipRouter struct {
	mtx cmtsync.RWMutex
	// A set of `source -> target` routes that are disabled for disseminating
	// transactions, where source and target are node IDs.
	disabledRoutes map[p2p.ID]map[p2p.ID]struct{}
	// Tracks the peer IDs it has been asked to stop gossiping from some sources.
	requestedPartialGossip map[p2p.ID]struct{}
}

func newGossipRouter() *gossipRouter {
	return &gossipRouter{
		disabledRoutes:         make(map[p2p.ID]map[p2p.ID]struct{}),
		requestedPartialGossip: make(map[p2p.ID]struct{}),
	}
}

func (r *gossipRouter) insertRequestedPartialGossip(id p2p.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.requestedPartialGossip[id] = struct{}{}
}

func (r *gossipRouter) removeRequestedPartialGossip(id p2p.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	delete(r.requestedPartialGossip, id)
}

func (r *gossipRouter) getRequestedPartialGossipList() []p2p.ID {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	list := make([]p2p.ID, 0, len(r.requestedPartialGossip))
	for id := range r.requestedPartialGossip {
		list = append(list, id)
	}
	return list
}

// disableRoute marks the route `source -> target` as disabled.
func (r *gossipRouter) disableRoute(source, target p2p.ID) {
	if source == noSender || target == noSender {
		// TODO: this shouldn't happen
		return
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	targets, ok := r.disabledRoutes[source]
	if !ok {
		targets = make(map[p2p.ID]struct{})
	}
	targets[target] = struct{}{}
	r.disabledRoutes[source] = targets
}

// isRouteEnabled returns true iff the route source->target is disabled.
func (r *gossipRouter) isRouteDisabled(source, target p2p.ID) bool {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	if targets, ok := r.disabledRoutes[source]; ok {
		if _, ok := targets[target]; ok {
			return true
		}
	}
	return false
}

// resetRoutes removes all disabled routes with peerID as source or target.
// It clears peerID from requestedPartialGossip tracker.
func (r *gossipRouter) resetRoutes(peerID p2p.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Remove peer as source.
	delete(r.disabledRoutes, peerID)

	// Remove peer as target.
	for _, targets := range r.disabledRoutes {
		delete(targets, peerID)
	}

	// Remove peer from the requestedPartialGossip list.
	delete(r.requestedPartialGossip, peerID)
}

// resetRandomRouteWithTarget removes a random disabled route that has the given
// target.
func (r *gossipRouter) resetRandomRouteWithTarget(target p2p.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	sourcesWithTarget := make([]p2p.ID, 0)
	for s, targets := range r.disabledRoutes {
		if _, ok := targets[target]; ok {
			sourcesWithTarget = append(sourcesWithTarget, s)
		}
	}

	if len(sourcesWithTarget) > 0 {
		randomSource := sourcesWithTarget[cmtrand.Intn(len(sourcesWithTarget))]
		delete(r.disabledRoutes[randomSource], target)
	}
}

// numRoutes returns the number of disabled routes in this node. Used for
// metrics.
func (r *gossipRouter) numRoutes() int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	count := 0
	for _, targets := range r.disabledRoutes {
		count += len(targets)
	}
	return count
}

type redundancyControl struct {
	// Pre-computed upper and lower bounds of accepted redundancy.
	lowerBound float64
	upperBound float64

	// Timer to adjust redundancy periodically.
	adjustTicker   *time.Ticker
	adjustInterval time.Duration

	// Counters for calculating the redundancy level.
	firstTimeTxs atomic.Int64 // number of transactions received for the first time
	duplicateTxs atomic.Int64 // number of duplicate transactions

	// If true, do not send HaveTx messages.
	haveTxBlocked atomic.Bool
}

func newRedundancyControl(config *cfg.MempoolConfig) *redundancyControl {
	adjustInterval := config.DOGAdjustInterval
	targetRedundancyDeltaAbs := config.DOGTargetRedundancy * targetRedundancyDeltaPercent / 100
	return &redundancyControl{
		lowerBound:     config.DOGTargetRedundancy - targetRedundancyDeltaAbs,
		upperBound:     config.DOGTargetRedundancy + targetRedundancyDeltaAbs,
		adjustTicker:   time.NewTicker(adjustInterval),
		adjustInterval: adjustInterval,
	}
}

func (rc *redundancyControl) adjustRedundancy(memR *Reactor) {
	// Compute current redundancy level and reset transaction counters.
	redundancy := rc.currentRedundancy()
	if redundancy < 0 {
		// There were no transactions during the last iteration. Do not adjust.
		return
	}

	// If redundancy level is low, ask a random peer for more transactions.
	if redundancy < rc.lowerBound {
		memR.Logger.Debug("TX redundancy BELOW lower limit: increase it (send Reset)", "redundancy", redundancy)
		// Send Reset message to random peer.
		partialGossipPeers := memR.router.getRequestedPartialGossipList()
		randomPeer, ok := memR.Switch.Peers().RandomFrom(partialGossipPeers)
		if ok && randomPeer != nil {
			err := randomPeer.Send(p2p.Envelope{ChannelID: MempoolControlChannel, Message: &protomem.ResetRoute{}})
			if err != nil {
				memR.Logger.Error("Failed to send Reset message", "peer", randomPeer.ID(), "err", err)
			} else {
				memR.router.removeRequestedPartialGossip(randomPeer.ID())
			}
		}
	}

	// If redundancy level is high, ask peers for less txs.
	if redundancy >= rc.upperBound {
		memR.Logger.Debug("TX redundancy ABOVE upper limit: decrease it (unblock HaveTx)", "redundancy", redundancy)
		// Unblock HaveTx messages.
		rc.haveTxBlocked.Store(false)
	}

	// Update metrics.
	memR.mempool.metrics.Redundancy.Set(redundancy)
}

func (rc *redundancyControl) controlLoop(memR *Reactor) {
	for {
		select {
		case <-rc.adjustTicker.C:
			rc.adjustRedundancy(memR)
		case <-memR.Quit():
			return
		}
	}
}

// currentRedundancy returns the current redundancy level and resets the
// counters. If there are no transactions, return -1. If firstTimeTxs is 0,
// return upperBound. If duplicateTxs is 0, return 0.
func (rc *redundancyControl) currentRedundancy() float64 {
	firstTimeTxs := rc.firstTimeTxs.Load()
	duplicateTxs := rc.duplicateTxs.Load()

	if firstTimeTxs+duplicateTxs == 0 {
		return -1
	}

	redundancy := rc.upperBound
	if firstTimeTxs != 0 {
		redundancy = float64(duplicateTxs) / float64(firstTimeTxs)
	}

	// Reset counters atomically
	rc.firstTimeTxs.Store(0)
	rc.duplicateTxs.Store(0)
	return redundancy
}

func (rc *redundancyControl) incDuplicateTxs() {
	rc.duplicateTxs.Add(1)
}

func (rc *redundancyControl) incFirstTimeTxs() {
	rc.firstTimeTxs.Add(1)
}

func (rc *redundancyControl) isHaveTxBlocked() bool {
	return rc.haveTxBlocked.Load()
}

// blockHaveTx blocks sending HaveTx messages and restarts the timer that
// adjusts redundancy.
func (rc *redundancyControl) blockHaveTx() {
	rc.haveTxBlocked.Store(true)
	// Wait until next adjustment to check if HaveTx messages should be unblocked.
	rc.adjustTicker.Reset(rc.adjustInterval)
}

func (rc *redundancyControl) triggerAdjustment(memR *Reactor) {
	rc.adjustRedundancy(memR)
	rc.adjustTicker.Reset(rc.adjustInterval)
}
