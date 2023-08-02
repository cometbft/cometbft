package sendNodeList

import (
	"errors"
	"time"

	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/clist"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
	"github.com/cometbft/cometbft/types"
)

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	config          *cfg.MempoolConfig
	mempool         *mempool.CListMempool
	txsInOtherNodes *TxsInOtherNodes
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(config *cfg.MempoolConfig, mempool *mempool.CListMempool) *Reactor {
	memR := &Reactor{
		config:          config,
		mempool:         mempool,
		txsInOtherNodes: newTxsInOtherNodes(),
	}
	memR.BaseReactor = *p2p.NewBaseReactor("Mempool", memR)
	memR.mempool.SetTxRemovedCallback(func(txKey types.TxKey) {
		memR.txsInOtherNodes.removeSenders(txKey)
		memR.txsInOtherNodes.removeFromSeenNodesSet(txKey)
	})
	return memR
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
		// TODO: add max capacity for list of peers
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  mempool.MempoolChannel,
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
		go memR.broadcastTxRoutine(peer)
	}
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

		for _, txBytes := range protoTxs {
			tx := types.Tx(txBytes)
			reqRes, err := memR.mempool.CheckTx(tx)
			if errors.Is(err, mempool.ErrTxInCache) {
				memR.Logger.Debug("Tx already exists in cache", "tx", tx.String())
			} else if err != nil {
				memR.Logger.Info("Could not check tx", "tx", tx.String(), "err", err)
			} else {
				// Record the sender only when the transaction is valid and, as
				// a consequence, added to the mempool. Senders are stored until
				// the transaction is removed from the mempool. Note that it's
				// possible a tx is still in the cache but no longer in the
				// mempool. For example, after committing a block, txs are
				// removed from mempool but not the cache.
				reqRes.SetCallback(func(res *abci.Response) {
					if res.GetCheckTx().Code == abci.CodeTypeOK {
						memR.txsInOtherNodes.addSender(tx.Key(), e.Src.ID())
						memR.txsInOtherNodes.addToSeenNodesSet(tx.Key(), prefixesOf(msg.GetTxs()))
					}
				})
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
	peers := memR.Switch.Peers().List()
	peerPrefixIds := prefixesOfPeers(peers)

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
			time.Sleep(mempool.PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// If we suspect that the peer is lagging behind, at least by more than
		// one block, we don't send the transaction immediately. This code
		// reduces the mempool size and the recheck-tx rate of the receiving
		// node. See [RFC 103] for an analysis on this optimization.
		//
		// [RFC 103]: https://github.com/cometbft/cometbft/pull/735
		memTx := next.Value.(*mempool.Entry)
		if peerState.GetHeight() < memTx.Height()-1 {
			time.Sleep(mempool.PeerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		// NOTE: Transaction batching was disabled due to
		// https://github.com/tendermint/tendermint/issues/5796

		txKey := memTx.GetTxKey()
		if memR.txsInOtherNodes.shouldSendTo(txKey, peer) {
			msg := &protomem.Message{
				Sum: &protomem.Message_Txs{
					Txs: &protomem.Txs{
						Txs:    [][]byte{memTx.GetTx()},
						SeenBy: getSetElements(mergeInSet(memR.txsInOtherNodes.getSeenByNodes(txKey), peerPrefixIds)),
					},
				},
			}
			if peer.Send(p2p.Envelope{ChannelID: mempool.MempoolChannel, Message: msg}) {
				memR.Logger.Debug("sent", "tx", memTx.GetTx().String(), "peers", peers)
				memR.txsInOtherNodes.addToSeenNodesSet(txKey, peerPrefixIds)
			} else {
				time.Sleep(mempool.PeerCatchupSleepIntervalMS * time.Millisecond)
				continue
			}
		} else {
			memR.Logger.Debug("did NOT send", "tx", memTx.GetTx().String(), "peer", peer.ID())
		}

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
