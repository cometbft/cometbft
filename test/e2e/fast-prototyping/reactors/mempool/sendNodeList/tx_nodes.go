package sendNodeList

import (
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

type NodeIdPrefix = string

const PrefixLength = 8

func prefixOf(id p2p.ID) NodeIdPrefix {
	return NodeIdPrefix(id[PrefixLength:])
}

// Convert a list peer ids into a list of id prefixes
func prefixesOf(ids [][]byte) []NodeIdPrefix {
	ps := make([]NodeIdPrefix, len(ids))
	for i, id := range ids {
		ps[i] = prefixOf(p2p.ID(id))
	}
	return ps
}

// Convert a list peers into a list of id prefixes
func prefixesOfPeers(peers []p2p.Peer) []NodeIdPrefix {
	ps := make([]NodeIdPrefix, len(peers))
	for i, peer := range peers {
		ps[i] = prefixOf(peer.ID())
	}
	return ps
}

// For keeping track of which node have send and which nodes have seen each
// transaction.
type TxsInOtherNodes struct {
	// `txSenders` maps every received transaction to the set of peer IDs that
	// have sent the transaction to this node. Sender IDs are used during
	// transaction propagation to avoid sending a transaction to a peer that
	// already has it. A sender ID is the internal peer ID used in the mempool
	// to identify the sender, storing two bytes with each transaction instead
	// of 20 bytes for the types.NodeID.
	txSenders map[types.TxKey]map[p2p.ID]bool

	// For each transaction, we keep track of the nodes that have received the
	// transaction.
	txSeenByNodes map[types.TxKey]map[NodeIdPrefix]struct{}

	mtx cmtsync.Mutex
}

func newTxsInOtherNodes() *TxsInOtherNodes {
	return &TxsInOtherNodes{
		txSenders:     make(map[types.TxKey]map[p2p.ID]bool),
		txSeenByNodes: make(map[types.TxKey]map[NodeIdPrefix]struct{}),
	}
}

func (ts *TxsInOtherNodes) isSender(txKey types.TxKey, peerID p2p.ID) bool {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	sendersSet, ok := ts.txSenders[txKey]
	return ok && sendersSet[peerID]
}

func (ts *TxsInOtherNodes) addSender(txKey types.TxKey, senderID p2p.ID) bool {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if sendersSet, ok := ts.txSenders[txKey]; ok {
		sendersSet[senderID] = true
		return false
	}
	ts.txSenders[txKey] = map[p2p.ID]bool{senderID: true}
	return true
}

func (ts *TxsInOtherNodes) removeSenders(txKey types.TxKey) {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if ts.txSenders != nil {
		delete(ts.txSenders, txKey)
	}
}

func (ts *TxsInOtherNodes) getSeenByNodes(txKey types.TxKey) []NodeIdPrefix {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if nodesSet, ok := ts.txSeenByNodes[txKey]; !ok {
		return nil
	} else {
		return getSetElements(nodesSet)
	}
}

func (ts *TxsInOtherNodes) wasSeenBy(txKey types.TxKey, peer p2p.Peer) bool {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if nodesSet, ok := ts.txSeenByNodes[txKey]; ok {
		_, ok := nodesSet[NodeIdPrefix(peer.ID()[PrefixLength:])]
		return ok
	} else {
		return false
	}
}

func (ts *TxsInOtherNodes) addToSeenNodesSet(txKey types.TxKey, peers []NodeIdPrefix) {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if nodesSet, ok := ts.txSeenByNodes[txKey]; ok {
		ts.txSeenByNodes[txKey] = mergeInSet(getSetElements(nodesSet), peers)
	} else {
		ts.txSeenByNodes[txKey] = toSet(peers)
	}
}

func (ts *TxsInOtherNodes) removeFromSeenNodesSet(txKey types.TxKey) {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()

	if ts.txSeenByNodes != nil {
		delete(ts.txSeenByNodes, txKey)
	}
}

func (ts *TxsInOtherNodes) shouldSendTo(txKey types.TxKey, peer p2p.Peer) bool {
	peer_is_sender := ts.isSender(txKey, peer.ID())
	tx_was_sent := ts.wasSeenBy(txKey, peer)
	return !peer_is_sender && !tx_was_sent
}
