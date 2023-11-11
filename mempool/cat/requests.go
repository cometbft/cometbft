package cat

import (
	"sync"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

const defaultGlobalRequestTimeout = 1 * time.Hour

// requestScheduler tracks the lifecycle of outbound transaction requests.
type requestScheduler struct {
	mtx sync.Mutex

	// responseTime is the time the scheduler
	// waits for a response from a peer before
	// invoking the callback
	responseTime time.Duration

	// globalTimeout represents the longest duration
	// to wait for any late response (after the responseTime).
	// After this period the request is garbage collected.
	globalTimeout time.Duration

	// requestsByPeer is a lookup table of requests by peer.
	// Multiple transactions can be requested by a single peer at one
	requestsByPeer map[p2p.ID]requestSet

	// requestsByTx is a lookup table for requested txs.
	// There can only be one request per tx.
	requestsByTx map[types.TxKey]p2p.ID
}

type requestSet map[types.TxKey]*time.Timer

func newRequestScheduler(responseTime, globalTimeout time.Duration) *requestScheduler {
	return &requestScheduler{
		responseTime:   responseTime,
		globalTimeout:  globalTimeout,
		requestsByPeer: make(map[p2p.ID]requestSet),
		requestsByTx:   make(map[types.TxKey]p2p.ID),
	}
}

// Return true iff the pair (txKey, peerID) was successfully added to the scheduler.
func (r *requestScheduler) Add(txKey types.TxKey, peerID p2p.ID, onTimeout func(key types.TxKey)) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// not allowed to have more than one outgoing transaction at once
	if _, ok := r.requestsByTx[txKey]; ok {
		return false
	}

	timer := time.AfterFunc(r.responseTime, func() {
		r.mtx.Lock()
		delete(r.requestsByTx, txKey)
		r.mtx.Unlock()

		// trigger callback. Callback can `Add` the tx back to the scheduler
		if onTimeout != nil {
			onTimeout(txKey)
		}

		// We set another timeout because the peer could still send
		// a late response after the first timeout and it's important
		// to recognize that it is a transaction in response to a
		// request and not a new transaction being broadcasted to the entire
		// network. This timer cannot be stopped and is used to ensure
		// garbage collection.
		time.AfterFunc(r.globalTimeout, func() {
			r.mtx.Lock()
			defer r.mtx.Unlock()
			delete(r.requestsByPeer[peerID], txKey)
		})
	})
	if _, ok := r.requestsByPeer[peerID]; !ok {
		r.requestsByPeer[peerID] = requestSet{txKey: timer}
	} else {
		r.requestsByPeer[peerID][txKey] = timer
	}
	r.requestsByTx[txKey] = peerID
	return true
}

func (r *requestScheduler) ForTx(key types.TxKey) (p2p.ID, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	v, ok := r.requestsByTx[key]
	return v, ok
}

func (r *requestScheduler) Has(peer p2p.ID, key types.TxKey) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	requestSet, ok := r.requestsByPeer[peer]
	if !ok {
		return false
	}
	_, ok = requestSet[key]
	return ok
}

func (r *requestScheduler) ClearAllRequestsFrom(peer p2p.ID) requestSet {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	requests, ok := r.requestsByPeer[peer]
	if !ok {
		return requestSet{}
	}
	for _, timer := range requests {
		timer.Stop()
	}
	delete(r.requestsByPeer, peer)
	return requests
}

func (r *requestScheduler) MarkReceived(peer p2p.ID, key types.TxKey) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.requestsByPeer[peer]; !ok {
		return false
	}

	if timer, ok := r.requestsByPeer[peer][key]; ok {
		timer.Stop()
	} else {
		return false
	}

	delete(r.requestsByPeer[peer], key)
	delete(r.requestsByTx, key)
	return true
}

// Close stops all timers and clears all requests.
// Add should never be called after `Close`.
func (r *requestScheduler) Close() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for _, requestSet := range r.requestsByPeer {
		for _, timer := range requestSet {
			timer.Stop()
		}
	}
}