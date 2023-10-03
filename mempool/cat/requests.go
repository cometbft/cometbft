package cat

import (
	"sync"
	"time"

	"github.com/tendermint/tendermint/types"
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
	// to wait for any late response (after the reponseTime).
	// After this period the request is garbage collected.
	globalTimeout time.Duration

	// requestsByPeer is a lookup table of requests by peer.
	// Multiple tranasctions can be requested by a single peer at one
	requestsByPeer map[uint16]requestSet

	// requestsByTx is a lookup table for requested txs.
	// There can only be one request per tx.
	requestsByTx map[types.TxKey]uint16
}

type requestSet map[types.TxKey]*time.Timer

func newRequestScheduler(responseTime, globalTimeout time.Duration) *requestScheduler {
	return &requestScheduler{
		responseTime:   responseTime,
		globalTimeout:  globalTimeout,
		requestsByPeer: make(map[uint16]requestSet),
		requestsByTx:   make(map[types.TxKey]uint16),
	}
}

func (r *requestScheduler) Add(key types.TxKey, peer uint16, onTimeout func(key types.TxKey)) bool {
	if peer == 0 {
		return false
	}
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// not allowed to have more than one outgoing transaction at once
	if _, ok := r.requestsByTx[key]; ok {
		return false
	}

	timer := time.AfterFunc(r.responseTime, func() {
		r.mtx.Lock()
		delete(r.requestsByTx, key)
		r.mtx.Unlock()

		// trigger callback. Callback can `Add` the tx back to the scheduler
		if onTimeout != nil {
			onTimeout(key)
		}

		// We set another timeout because the peer could still send
		// a late response after the first timeout and it's important
		// to recognise that it is a transaction in response to a
		// request and not a new transaction being broadcasted to the entire
		// network. This timer cannot be stopped and is used to ensure
		// garbage collection.
		time.AfterFunc(r.globalTimeout, func() {
			r.mtx.Lock()
			defer r.mtx.Unlock()
			delete(r.requestsByPeer[peer], key)
		})
	})
	if _, ok := r.requestsByPeer[peer]; !ok {
		r.requestsByPeer[peer] = requestSet{key: timer}
	} else {
		r.requestsByPeer[peer][key] = timer
	}
	r.requestsByTx[key] = peer
	return true
}

func (r *requestScheduler) ForTx(key types.TxKey) uint16 {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.requestsByTx[key]
}

func (r *requestScheduler) Has(peer uint16, key types.TxKey) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	requestSet, ok := r.requestsByPeer[peer]
	if !ok {
		return false
	}
	_, ok = requestSet[key]
	return ok
}

func (r *requestScheduler) ClearAllRequestsFrom(peer uint16) requestSet {
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

func (r *requestScheduler) MarkReceived(peer uint16, key types.TxKey) bool {
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
