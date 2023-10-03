package cat

import (
	"fmt"

	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/p2p"
)

const firstPeerID = mempool.UnknownPeerID + 1

// mempoolIDs is a thread-safe map of peer IDs to shorter uint16 IDs used by the Reactor for tracking peer
// messages and peer state such as what transactions peers have seen
type mempoolIDs struct {
	mtx       tmsync.RWMutex
	peerMap   map[p2p.ID]uint16   // quick lookup table for peer ID to short ID
	nextID    uint16              // assumes that a node will never have over 65536 active peers
	activeIDs map[uint16]p2p.Peer // used to check if a given peerID key is used, the value doesn't matter
}

func newMempoolIDs() *mempoolIDs {
	return &mempoolIDs{
		peerMap:   make(map[p2p.ID]uint16),
		activeIDs: make(map[uint16]p2p.Peer),
		nextID:    firstPeerID, // reserve unknownPeerID(0) for mempoolReactor.BroadcastTx
	}
}

// ReserveForPeer searches for the next unused ID and assigns it to the
// peer.
func (ids *mempoolIDs) ReserveForPeer(peer p2p.Peer) {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	if _, ok := ids.peerMap[peer.ID()]; ok {
		panic("duplicate peer added to mempool")
	}

	curID := ids.nextPeerID()
	ids.peerMap[peer.ID()] = curID
	ids.activeIDs[curID] = peer
}

// nextPeerID returns the next unused peer ID to use.
// This assumes that ids's mutex is already locked.
func (ids *mempoolIDs) nextPeerID() uint16 {
	if len(ids.activeIDs) == mempool.MaxActiveIDs {
		panic(fmt.Sprintf("node has maximum %d active IDs and wanted to get one more", mempool.MaxActiveIDs))
	}

	_, idExists := ids.activeIDs[ids.nextID]
	for idExists {
		ids.nextID++
		_, idExists = ids.activeIDs[ids.nextID]
	}
	curID := ids.nextID
	ids.nextID++
	return curID
}

// Reclaim returns the ID reserved for the peer back to unused pool.
func (ids *mempoolIDs) Reclaim(peerID p2p.ID) uint16 {
	ids.mtx.Lock()
	defer ids.mtx.Unlock()

	removedID, ok := ids.peerMap[peerID]
	if ok {
		delete(ids.activeIDs, removedID)
		delete(ids.peerMap, peerID)
		return removedID
	}
	return 0
}

// GetIDForPeer returns the shorthand ID reserved for the peer.
func (ids *mempoolIDs) GetIDForPeer(peerID p2p.ID) uint16 {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	id, exists := ids.peerMap[peerID]
	if !exists {
		return 0
	}
	return id
}

// GetPeer returns the peer for the given shorthand ID.
func (ids *mempoolIDs) GetPeer(id uint16) p2p.Peer {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return ids.activeIDs[id]
}

// GetAll returns all active peers.
func (ids *mempoolIDs) GetAll() map[uint16]p2p.Peer {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	// make a copy of the map.
	peers := make(map[uint16]p2p.Peer, len(ids.activeIDs))
	for id, peer := range ids.activeIDs {
		peers[id] = peer
	}
	return peers
}

// Len returns the number of active peers.
func (ids *mempoolIDs) Len() int {
	ids.mtx.RLock()
	defer ids.mtx.RUnlock()

	return len(ids.activeIDs)
}
