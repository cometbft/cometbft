package cat

import (
	"sync"
	"time"

	"github.com/tendermint/tendermint/types"
)

// simple, thread-safe in memory store for transactions
type store struct {
	mtx   sync.RWMutex
	bytes int64
	txs   map[types.TxKey]*wrappedTx
}

func newStore() *store {
	return &store{
		bytes: 0,
		txs:   make(map[types.TxKey]*wrappedTx),
	}
}

func (s *store) set(wtx *wrappedTx) bool {
	if wtx == nil {
		return false
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if tx, exists := s.txs[wtx.key]; !exists || tx.height == -1 {
		s.txs[wtx.key] = wtx
		s.bytes += wtx.size()
		return true
	}
	return false
}

func (s *store) get(txKey types.TxKey) *wrappedTx {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.txs[txKey]
}

func (s *store) has(txKey types.TxKey) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, has := s.txs[txKey]
	return has
}

func (s *store) remove(txKey types.TxKey) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	tx, exists := s.txs[txKey]
	if !exists {
		return false
	}
	s.bytes -= tx.size()
	delete(s.txs, txKey)
	return true
}

// reserve adds an empty placeholder for the specified key to prevent
// a transaction with the same key from being added
func (s *store) reserve(txKey types.TxKey) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, has := s.txs[txKey]
	if !has {
		s.txs[txKey] = &wrappedTx{height: -1}
		return true
	}
	return false
}

// release is called when a pending transaction failed
// to enter the mempool. The empty element and key is removed.
func (s *store) release(txKey types.TxKey) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	value, ok := s.txs[txKey]
	if ok && value.height == -1 {
		delete(s.txs, txKey)
	}
}

func (s *store) size() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.txs)
}

func (s *store) totalBytes() int64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.bytes
}

func (s *store) getAllKeys() []types.TxKey {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	keys := make([]types.TxKey, len(s.txs))
	idx := 0
	for key := range s.txs {
		keys[idx] = key
		idx++
	}
	return keys
}

func (s *store) getAllTxs() []*wrappedTx {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	txs := make([]*wrappedTx, len(s.txs))
	idx := 0
	for _, tx := range s.txs {
		txs[idx] = tx
		idx++
	}
	return txs
}

func (s *store) getTxsBelowPriority(priority int64) ([]*wrappedTx, int64) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	txs := make([]*wrappedTx, 0, len(s.txs))
	bytes := int64(0)
	for _, tx := range s.txs {
		if tx.priority < priority {
			txs = append(txs, tx)
			bytes += tx.size()
		}
	}
	return txs, bytes
}

// purgeExpiredTxs removes all transactions that are older than the given height
// and time. Returns the amount of transactions that were removed
func (s *store) purgeExpiredTxs(expirationHeight int64, expirationAge time.Time) int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	counter := 0
	for key, tx := range s.txs {
		if tx.height < expirationHeight || tx.timestamp.Before(expirationAge) {
			s.bytes -= tx.size()
			delete(s.txs, key)
			counter++
		}
	}
	return counter
}

func (s *store) reset() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.bytes = 0
	s.txs = make(map[types.TxKey]*wrappedTx)
}
