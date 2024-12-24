package mempool

import (
	"context"
	"errors"
	"time"

	"fmt"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
	"github.com/cometbft/cometbft/types"
	"golang.org/x/sync/semaphore"
)

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config  *cfg.MempoolConfig
	mempool *CListMempool
	ids     *mempoolIDs

	// Semaphores to keep track of how many connections to peers are active for broadcasting
	// transactions. Each semaphore has a capacity that puts an upper bound on the number of
	// connections for different groups of peers.
	activePersistentPeersSemaphore    *semaphore.Weighted
	activeNonPersistentPeersSemaphore *semaphore.Weighted
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(config *cfg.MempoolConfig, mempool *CListMempool) *Reactor {
	memR := &Reactor{
		config:  config,
		mempool: mempool,
		ids:     newMempoolIDs(),
	}
	memR.BaseReactor = *p2p.NewBaseReactor("Mempool", memR)
	memR.activePersistentPeersSemaphore = semaphore.NewWeighted(int64(memR.config.ExperimentalMaxGossipConnectionsToPersistentPeers))
	memR.activeNonPersistentPeersSemaphore = semaphore.NewWeighted(int64(memR.config.ExperimentalMaxGossipConnectionsToNonPersistentPeers))

	return memR
}

// InitPeer implements Reactor by creating a state for the peer.
func (memR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	memR.ids.ReserveForPeer(peer)
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l log.Logger) {
	memR.Logger = l
	memR.mempool.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	if !memR.config.Broadcast {
		memR.Logger.Info("Tx broadcasting is disabled")
	}
	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (memR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, memR.config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  MempoolChannel,
			Priority:            5,
			RecvMessageCapacity: batchMsg.Size(),
			MessageType:         &protomem.Message{},
		},
	}
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
func (memR *Reactor) AddPeer(peer p2p.Peer) {
	if memR.config.Broadcast {
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

// RemovePeer implements Reactor.
func (memR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
	memR.ids.Reclaim(peer)
	// broadcast routine checks if peer is gone and returns
}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (memR *Reactor) Receive(e p2p.Envelope) {
	memR.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
	switch msg := e.Message.(type) {
	case *protomem.Txs:
		protoTxs := msg.GetTxs()
		if len(protoTxs) == 0 {
			memR.Logger.Error("received empty txs from peer", "src", e.Src)
			return
		}
		txInfo := TxInfo{SenderID: memR.ids.GetForPeer(e.Src)}
		if e.Src != nil {
			txInfo.SenderP2PID = e.Src.ID()
		}

		var err error
		for _, tx := range protoTxs {
			ntx := types.Tx(tx)
			err = memR.mempool.CheckTx(ntx, nil, txInfo)
			if err != nil {
				switch {
				case errors.Is(err, ErrTxInCache):
					memR.Logger.Debug("Tx already exists in cache", "tx", ntx.String())
				case errors.As(err, &ErrMempoolIsFull{}):
					// using debug level to avoid flooding when traffic is high
					memR.Logger.Debug(err.Error())
				default:
					memR.Logger.Info("Could not check tx", "tx", ntx.String(), "err", err)
				}
			}
		}
	default:
		memR.Logger.Error("unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
		return
	}

	// broadcasting happens from go routines per peer
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// Send new mempool txs to peer.
func (memR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
	peerID := memR.ids.GetForPeer(peer)
	var next *clist.CElement

	for {
		// In case of both next.NextWaitChan() and peer.Quit() are variable at the same time
		if !memR.IsRunning() || !peer.IsRunning() {
			return
		}

		// This happens because the CElement we were looking at got garbage
		// collected (removed). That is, .NextWait() returned nil. Go ahead and
		// start from the beginning.
		if next == nil {
			select {
			case <-memR.mempool.TxsWaitChan(): // Wait until a tx is available
				if next = memR.mempool.TxsFront(); next == nil {
					continue
				}
			case <-peer.Quit():
				return
			case <-memR.Quit():
				return
			}
		}

		// Make sure the peer is up to date.
		peerState, ok := peer.Get(types.PeerStateKey).(PeerState)
		if !ok {
			// Peer does not have a state yet. We set it in the consensus reactor, but
			// when we add peer in Switch, the order we call reactors#AddPeer is
			// different every time due to us using a map. Sometimes other reactors
			// will be initialized before the consensus reactor. We should wait a few
			// milliseconds and retry.
			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// Allow for a lag of 1 block.
		memTx := next.Value.(*mempoolTx)
		if peerState.GetHeight() < memTx.Height()-1 {
			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// NOTE: Transaction batching was disabled due to
		// https://github.com/tendermint/tendermint/issues/5796

		if !memTx.isSender(peerID) {
			success := peer.Send(p2p.Envelope{
				ChannelID: MempoolChannel,
				Message:   &protomem.Txs{Txs: [][]byte{memTx.tx}},
			})
			if !success {
				time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
				continue
			}
		}

<<<<<<< HEAD
=======
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
}

func newGossipRouter() *gossipRouter {
	return &gossipRouter{
		disabledRoutes: make(map[p2p.ID]map[p2p.ID]struct{}),
	}
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
func (r *gossipRouter) resetRoutes(peerID p2p.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Remove peer as source.
	delete(r.disabledRoutes, peerID)

	// Remove peer as target.
	for _, targets := range r.disabledRoutes {
		delete(targets, peerID)
	}
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
		randomPeer := memR.Switch.Peers().Random()
		if randomPeer != nil {
			err := randomPeer.Send(p2p.Envelope{ChannelID: MempoolControlChannel, Message: &protomem.ResetRoute{}})
			if err != nil {
				memR.Logger.Error("Failed to send Reset message", "peer", randomPeer.ID(), "err", err)
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
>>>>>>> be894a0e (perf(mempool): use atomic instead of mutex to improve perf (#4701))
		select {
		case <-next.NextWaitChan():
			// see the start of the for loop for nil check
			next = next.Next()
		case <-peer.Quit():
			return
		case <-memR.Quit():
			return
		}
	}
}

<<<<<<< HEAD
// TxsMessage is a Message containing transactions.
type TxsMessage struct {
	Txs []types.Tx
}

// String returns a string representation of the TxsMessage.
func (m *TxsMessage) String() string {
	return fmt.Sprintf("[TxsMessage %v]", m.Txs)
=======
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
>>>>>>> be894a0e (perf(mempool): use atomic instead of mutex to improve perf (#4701))
}
