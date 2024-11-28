package mempool

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	abcicli "github.com/cometbft/cometbft/abci/client"
	protomem "github.com/cometbft/cometbft/api/cometbft/mempool/v1"
	cfg "github.com/cometbft/cometbft/config"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/nodekey"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/types"
)

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
	router *gossipRouter

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
	}
	return nil
}

// StreamDescriptors implements Reactor by returning the list of channels for this
// reactor.
func (memR *Reactor) StreamDescriptors() []p2p.StreamDescriptor {
	largestTx := make([]byte, memR.config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	key := types.Tx(largestTx).Key()
	haveTxMsg := protomem.Message{
		Sum: &protomem.Message_HaveTx{HaveTx: &protomem.HaveTx{TxKey: key[:]}},
	}

	return []p2p.StreamDescriptor{
		&tcpconn.ChannelDescriptor{
			ID:                  MempoolChannel,
			Priority:            5,
			RecvMessageCapacity: batchMsg.Size(),
			MessageTypeI:        &protomem.Message{},
		},
		&tcpconn.ChannelDescriptor{
			ID:                  MempoolControlChannel,
			Priority:            10,
			RecvMessageCapacity: haveTxMsg.Size(),
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
		// Remove all routes with peer as source or target.
		memR.router.resetRoutes(peer.ID())
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
			}

		case *protomem.ResetRoute:
			memR.Logger.Debug("Received Reset", "from", senderID)
			if memR.router != nil {
				memR.router.resetRandomRouteWithTarget(senderID)
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

			success := peer.Send(p2p.Envelope{
				ChannelID: MempoolChannel,
				Message:   &protomem.Txs{Txs: [][]byte{entry.Tx()}},
			})
			if success {
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
	disabledRoutes map[nodekey.ID]map[nodekey.ID]struct{}
}

func newGossipRouter() *gossipRouter {
	return &gossipRouter{
		disabledRoutes: make(map[nodekey.ID]map[nodekey.ID]struct{}),
	}
}

// disableRoute marks the route `source -> target` as disabled.
func (r *gossipRouter) disableRoute(source, target nodekey.ID) {
	if source == noSender || target == noSender {
		// TODO: this shouldn't happen
		return
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	targets, ok := r.disabledRoutes[source]
	if !ok {
		targets = make(map[nodekey.ID]struct{})
	}
	targets[target] = struct{}{}
	r.disabledRoutes[source] = targets
}

// isRouteEnabled returns true iff the route source->target is disabled.
func (r *gossipRouter) isRouteDisabled(source, target nodekey.ID) bool {
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
func (r *gossipRouter) resetRoutes(peerID nodekey.ID) {
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
func (r *gossipRouter) resetRandomRouteWithTarget(target nodekey.ID) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	sourcesWithTarget := make([]nodekey.ID, 0)
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
