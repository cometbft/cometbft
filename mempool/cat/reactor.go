package cat

import (
	"errors"
	"fmt"
	"sync"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
	"github.com/cometbft/cometbft/types"
)

const (
	// default duration to wait before considering a peer non-responsive
	// and searching for the tx from a new peer
	defaultGossipDelay = 1000 * time.Millisecond

	// Content Addressable Tx Pool gossips state based messages (SeenTx and WantTx) on a separate channel
	// for cross compatibility
	MempoolStateChannel = byte(0x31)

	// peerHeightDiff signifies the tolerance in difference in height between the peer and the height
	// the node received the tx
	peerHeightDiff = 10
)

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	mempool.BaseSyncReactor
	mempool *mempool.CListMempool

	peerIDs  sync.Map          // set of connected peers
	requests *requestScheduler // to track requested transactions

	// Thread-safe list of transactions peers have seen that we have not yet seen
	seenByPeersSet *SeenTxSet
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(config *cfg.MempoolConfig, mp *mempool.CListMempool, waitSync bool, logger log.Logger) *Reactor {
	memR := &Reactor{
		mempool:        mp,
		requests:       newRequestScheduler(defaultGossipDelay, defaultGlobalRequestTimeout),
		seenByPeersSet: NewSeenTxSet(),
	}
	memR.BaseSyncReactor = *mempool.NewBaseSyncReactor(config, waitSync)
	memR.SetLogger(logger)
	memR.mempool.SetTxRemovedCallback(func(txKey types.TxKey) {
		memR.seenByPeersSet.RemoveKey(txKey)
	})
	memR.mempool.SetNewTxReceivedCallback(func(txKey types.TxKey) {
		// If we don't find the tx in the mempool, probably it is because it was
		// invalid, so don't broadcast.
		if entry := memR.mempool.GetEntry(txKey); !entry.IsEmpty() {
			go memR.broadcastNewTx(entry)
		}
	})
	return memR
}

// InitPeer implements Reactor by creating a state for the peer.
func (memR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	memR.peerIDs.Store(peer.ID(), struct{}{})
	return peer
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l log.Logger) {
	memR.Logger = l
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	if !memR.Config.Broadcast {
		memR.Logger.Info("Tx broadcasting is disabled")
	}
	return nil
}

// OnStop implements Service
func (memR *Reactor) OnStop() {
	// stop all the timers tracking outbound requests
	memR.requests.Close()
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
func (memR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, memR.Config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	stateMsg := protomem.Message{
		Sum: &protomem.Message_SeenTx{
			SeenTx: &protomem.SeenTx{
				TxKey: make([]byte, tmhash.Size),
			},
		},
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  mempool.MempoolChannel,
			Priority:            6,
			RecvMessageCapacity: batchMsg.Size(),
			MessageType:         &protomem.Message{},
		},
		{
			ID:                  MempoolStateChannel,
			Priority:            5,
			RecvMessageCapacity: stateMsg.Size(),
			MessageType:         &protomem.Message{},
		},
	}
}

// RemovePeer implements Reactor. For all current outbound requests to this
// peer it will find a new peer to rerequest the same transactions.
func (memR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
	memR.peerIDs.Delete(peer.ID())
	// remove and rerequest all pending outbound requests to that peer since we know
	// we won't receive any responses from them.
	outboundRequests := memR.requests.ClearAllRequestsFrom(peer.ID())
	for key := range outboundRequests {
		memR.mempool.Metrics.RequestedTxs.Add(1)
		memR.findNewPeerToRequestTx(key)
	}
}

// Receive implements Reactor.
// It processes one of three messages: Txs, SeenTx, WantTx.
func (memR *Reactor) Receive(e p2p.Envelope) {
	switch msg := e.Message.(type) {

	// A peer has sent us one or more transactions. This could be either because we requested them
	// or because the peer received a new transaction and is broadcasting it to us.
	// NOTE: This setup also means that we can support older mempool implementations that simply
	// flooded the network with transactions.
	case *protomem.Txs:
		protoTxs := msg.GetTxs()
		memR.Logger.Debug("Received Txs", "src", e.Src, "chId", e.ChannelID, "msg", e.Message, "len", len(protoTxs))
		if len(protoTxs) == 0 {
			memR.Logger.Error("received empty txs from peer", "src", e.Src)
			return
		}

		peerID := e.Src.ID()
		for _, txBytes := range protoTxs {
			tx := types.Tx(txBytes)
			key := tx.Key()

			// If we requested the transaction, we mark it as received.
			if memR.requests.Has(peerID, key) {
				memR.requests.MarkReceived(peerID, key)
				memR.Logger.Debug("received a response for a requested transaction", "peerID", peerID, "txKey", key)
			} else {
				// If we didn't request the transaction we simply mark the peer as having the
				// tx (we'd have already done it if we were requesting the tx).
				memR.markPeerHasTx(peerID, key)
				memR.Logger.Debug("received new transaction", "peerID", peerID, "txKey", key)
			}

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
						memR.markPeerHasTx(e.Src.ID(), tx.Key())

						// We broadcast only transactions that we deem valid and
						// actually have in our mempool.
						memR.broadcastSeenTx(key)
					}
				})
			}
		}

	// A peer has indicated to us that it has a transaction. We first verify the txKey and
	// mark that peer as having the transaction. Then we proceed with the following logic:
	//
	// 1. If we have the transaction, we do nothing.
	// 2. If we don't yet have the tx but have an outgoing request for it, we do nothing.
	// 3. If we recently evicted the tx and still don't have space for it, we do nothing.
	// 4. Else, we request the transaction from that peer.
	case *protomem.SeenTx:
		txKey, err := types.TxKeyFromBytes(msg.TxKey)
		if err != nil {
			memR.Logger.Error("peer sent SeenTx with incorrect tx key", "err", err)
			memR.Switch.StopPeerForError(e.Src, err)
			return
		}
		memR.Logger.Debug("Received SeenTx", "src", e.Src, "chId", e.ChannelID, "txKey", txKey)
		peerID := e.Src.ID()
		memR.markPeerHasTx(peerID, txKey)

		// Check if we don't already have the transaction and that it was recently rejected
		if memR.mempool.InMempool(txKey) || memR.mempool.InCache(txKey) || memR.mempool.WasRejected(txKey) {
			memR.Logger.Debug("received a seen tx for a tx we already have or is in cache or was rejected", "txKey", txKey)
			return
		}

		// If we are already requesting that tx, then we don't need to go any further.
		if _, exists := memR.requests.ForTx(txKey); exists {
			memR.Logger.Debug("received a SeenTx message for a transaction we are already requesting", "txKey", txKey)
			return
		}

		// We don't have the transaction, nor are we requesting it so we send the node
		// a want msg
		memR.requestTx(txKey, e.Src.ID())

	// A peer is requesting a transaction that we have claimed to have. Find the specified
	// transaction and broadcast it to the peer. We may no longer have the transaction
	case *protomem.WantTx:
		txKey, err := types.TxKeyFromBytes(msg.TxKey)
		if err != nil {
			memR.Logger.Error("peer sent WantTx with incorrect tx key", "err", err)
			memR.Switch.StopPeerForError(e.Src, err)
			return
		}
		memR.Logger.Debug("Received SeenTx", "src", e.Src, "chId", e.ChannelID, "txKey", txKey)
		memR.sendRequestedTx(txKey, e.Src)

	default:
		memR.Logger.Error("Received unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
		memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
		return
	}
}

func (memR *Reactor) sendRequestedTx(txKey types.TxKey, peer p2p.Peer) {
	if !memR.Config.Broadcast {
		return
	}

	if entry := memR.mempool.GetEntry(txKey); entry != nil {
		memR.Logger.Debug("sending a tx in response to a want msg", "peer", peer.ID())
		txsMsg := p2p.Envelope{
			ChannelID: mempool.MempoolChannel,
			Message:   &protomem.Txs{Txs: [][]byte{entry.Tx()}},
		}
		if peer.Send(txsMsg) {
			memR.markPeerHasTx(peer.ID(), txKey)
		} else {
			memR.Logger.Error("send Txs message with requested transactions failed", "txKey", txKey, "peerID", peer.ID())
		}
	}
}

// PeerHasTx marks that the transaction has been seen by a peer.
func (memR *Reactor) markPeerHasTx(peerID p2p.ID, txKey types.TxKey) {
	memR.Logger.Debug("mark that peer has tx", "peer", peerID, "txKey", txKey.String())
	memR.seenByPeersSet.Add(txKey, peerID)
}

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// broadcastSeenTx broadcasts a SeenTx message to all peers unless we
// know they have already seen the transaction
func (memR *Reactor) broadcastSeenTx(txKey types.TxKey) {
	if !memR.Config.Broadcast {
		return
	}

	memR.Logger.Debug("Broadcasting SeenTx...", "tx", txKey.String())

	msg := p2p.Envelope{
		ChannelID: MempoolStateChannel,
		Message:   &protomem.SeenTx{TxKey: txKey[:]},
	}

	memR.peerIDs.Range(func(key, _ interface{}) bool {
		peerID := key.(p2p.ID)
		peer := memR.Switch.Peers().Get(peerID)
		memR.Logger.Debug("Sending SeenTx...", "tx", txKey, "peer", peerID)

		if peerState, ok := peer.Get(types.PeerStateKey).(PeerState); ok {
			// make sure peer isn't too far behind. This can happen
			// if the peer is block-synching still and catching up
			// in which case we just skip sending the transaction
			if peerState.GetHeight() < memR.mempool.Height()-peerHeightDiff {
				memR.Logger.Debug("peer is too far behind us. Skipping broadcast of seen tx")
				return true
			}
		}
		// no need to send a seen tx message to a peer that already
		// has that tx.
		if memR.seenByPeersSet.Has(txKey, peerID) {
			memR.Logger.Debug("Peer has seen the transaction, not sending SeenTx", "tx", txKey, "peer", peerID)
			return true
		}

		if peer.Send(msg) {
			memR.Logger.Debug("SeenTx sent", "tx", txKey, "peer", peerID)
			return true
		}
		memR.Logger.Error("send SeenTx message failed", "txKey", txKey, "peerID", peer.ID())
		return true
	})
}

// broadcastNewTx broadcast new transaction to all peers unless we are already
// sure they have seen the tx.
func (memR *Reactor) broadcastNewTx(entry *mempool.CListEntry) {
	if !memR.Config.Broadcast {
		return
	}

	tx := entry.Tx()
	txKey := tx.Key()
	memR.Logger.Debug("Broadcasting new transaction...", "tx", txKey.String())

	msg := p2p.Envelope{
		ChannelID: mempool.MempoolChannel,
		Message:   &protomem.Txs{Txs: [][]byte{tx}},
	}

	memR.peerIDs.Range(func(key, _ interface{}) bool {
		peerID := key.(p2p.ID)
		peer := memR.Switch.Peers().Get(peerID)
		memR.Logger.Debug("Sending new transaction to...", "tx", txKey, "peer", peerID)

		if peerState, ok := peer.Get(types.PeerStateKey).(PeerState); ok {
			// make sure peer isn't too far behind. This can happen
			// if the peer is blocksyncing still and catching up
			// in which case we just skip sending the transaction
			if peerState.GetHeight() < entry.Height()-peerHeightDiff {
				memR.Logger.Debug("Peer is too far behind us, don't send new tx")
				return true
			}
		}

		if memR.seenByPeersSet.Has(txKey, peerID) {
			memR.Logger.Debug("Peer has seen the transaction, not sending it", "tx", txKey, "peer", peerID)
			return true
		}

		if peer.Send(msg) {
			memR.Logger.Debug("New transaction sent", "tx", txKey, "peer", peerID)
			memR.markPeerHasTx(peerID, txKey)
			return true
		}

		memR.Logger.Error("send Txs message with new transactions failed", "txKey", txKey, "peerID", peerID)
		return true
	})
}

// requestTx requests a transaction from a peer and tracks it,
// requesting it from another peer if the first peer does not respond.
func (memR *Reactor) requestTx(txKey types.TxKey, peerID p2p.ID) {
	if !memR.Config.Broadcast {
		return
	}

	if !memR.Switch.Peers().Has(peerID) {
		// we have disconnected from the peer
		return
	}

	memR.Logger.Debug("requesting tx", "txKey", txKey, "peerID", peerID)
	peer := memR.Switch.Peers().Get(peerID)
	msg := p2p.Envelope{
		ChannelID: MempoolStateChannel,
		Message:   &protomem.WantTx{TxKey: txKey[:]},
	}
	if peer.Send(msg) {
		memR.mempool.Metrics.RequestedTxs.Add(1)
		added := memR.requests.Add(txKey, peerID, memR.findNewPeerToRequestTx)
		if !added {
			memR.Logger.Error("have already marked a tx as requested", "txKey", txKey, "peerID", peerID)
		}
	} else {
		memR.Logger.Error("send WantTx message failed", "txKey", txKey, "peerID", peerID)
	}
}

// findNewPeerToSendTx finds a new peer that has already seen the transaction to
// request a transaction from.
func (memR *Reactor) findNewPeerToRequestTx(txKey types.TxKey) {
	// ensure that we are connected to peers
	if memR.Switch.Peers().Size() == 0 {
		return
	}

	// pop the next peer in the list of remaining peers that have seen the tx
	// and does not already have an outbound request for that tx
	seenMap := memR.seenByPeersSet.Get(txKey)
	var peerID *p2p.ID
	for possiblePeer := range seenMap {
		possiblePeer := possiblePeer
		if !memR.requests.Has(possiblePeer, txKey) {
			peerID = &possiblePeer
			break
		}
	}

	if peerID == nil {
		// No other free peer has the transaction we are looking for.
		// We give up 🤷‍♂️ and hope either a peer responds late or the tx
		// is gossiped again
		memR.mempool.Metrics.NoPeerForTx.Add(1)
		memR.Logger.Info("no other peer has the tx we are looking for", "txKey", txKey)
		return
	}

	if !memR.Switch.Peers().Has(*peerID) {
		// we disconnected from that peer, retry again until we exhaust the list
		memR.findNewPeerToRequestTx(txKey)
	} else {
		memR.mempool.Metrics.RerequestedTxs.Add(1)
		memR.requestTx(txKey, *peerID)
	}
}