package mempool

import (
	"container/list"
	"sync/atomic"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/types"
)

// TxCache defines an interface for raw transaction caching in a mempool.
// Currently, a TxCache does not allow direct reading or getting of transaction
// values. A TxCache is used primarily to push transactions and removing
// transactions. Pushing via Push returns a boolean telling the caller if the
// transaction already exists in the cache or not.
type TxCache interface {
	// Reset resets the cache to an empty state.
	Reset()

	// Push adds the given raw transaction to the cache and returns true if it was
	// newly added. Otherwise, it returns false.
	Push(tx types.Tx) bool

	// Remove removes the given raw transaction from the cache.
	Remove(tx types.Tx)

	// Has reports whether tx is present in the cache. Checking for presence is
	// not treated as an access of the value.
	Has(tx types.Tx) bool
}

// CacheStats provides statistics about cache operations
type CacheStats interface {
	// Hits returns the number of cache hits
	Hits() uint64

	// Misses returns the number of cache misses
	Misses() uint64

	// Evictions returns the number of cache evictions
	Evictions() uint64

	// Size returns the current number of items in the cache
	Size() int

	// ResetStats resets all statistics counters to zero
	ResetStats()
}

var _ TxCache = (*LRUTxCache)(nil)

// LRUTxCache maintains a thread-safe LRU cache of raw transactions. The cache
// only stores the hash of the raw transaction.
type LRUTxCache struct {
	mtx      cmtsync.Mutex
	size     int
	cacheMap map[types.TxKey]*list.Element
	list     *list.List
}

// StatsLRUTxCache extends LRUTxCache with statistics tracking
type StatsLRUTxCache struct {
	LRUTxCache
	hits      atomic.Uint64
	misses    atomic.Uint64
	evictions atomic.Uint64
}

var _ TxCache = (*StatsLRUTxCache)(nil)
var _ CacheStats = (*StatsLRUTxCache)(nil)

func NewLRUTxCache(cacheSize int) *LRUTxCache {
	return &LRUTxCache{
		size:     cacheSize,
		cacheMap: make(map[types.TxKey]*list.Element, cacheSize),
		list:     list.New(),
	}
}

// NewStatsLRUTxCache creates a new LRU cache with statistics tracking
func NewStatsLRUTxCache(cacheSize int) *StatsLRUTxCache {
	return &StatsLRUTxCache{
		LRUTxCache: *NewLRUTxCache(cacheSize),
	}
}

// GetList returns the underlying linked-list that backs the LRU cache. Note,
// this should be used for testing purposes only!
func (c *LRUTxCache) GetList() *list.List {
	return c.list
}

// Reset resets the cache to an empty state.
func (c *LRUTxCache) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	clear(c.cacheMap)
	c.list.Init()
}

// Reset resets the cache to an empty state and clears statistics.
func (c *StatsLRUTxCache) Reset() {
	c.LRUTxCache.Reset()
	c.ResetStats()
}

func (c *LRUTxCache) Push(tx types.Tx) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	key := tx.Key()

	moved, ok := c.cacheMap[key]
	if ok {
		c.list.MoveToBack(moved)
		return false
	}

	if c.list.Len() >= c.size {
		front := c.list.Front()
		if front != nil {
			frontKey := front.Value.(types.TxKey)
			delete(c.cacheMap, frontKey)
			c.list.Remove(front)
		}
	}

	e := c.list.PushBack(key)
	c.cacheMap[key] = e

	return true
}

// Push adds a transaction to the cache with statistics tracking
func (c *StatsLRUTxCache) Push(tx types.Tx) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	key := tx.Key()

	moved, ok := c.cacheMap[key]
	if ok {
		c.list.MoveToBack(moved)
		c.hits.Add(1)
		return false
	}

	c.misses.Add(1)

	if c.list.Len() >= c.size {
		front := c.list.Front()
		if front != nil {
			frontKey := front.Value.(types.TxKey)
			delete(c.cacheMap, frontKey)
			c.list.Remove(front)
			c.evictions.Add(1)
		}
	}

	e := c.list.PushBack(key)
	c.cacheMap[key] = e

	return true
}

func (c *LRUTxCache) Remove(tx types.Tx) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	key := tx.Key()
	e := c.cacheMap[key]
	delete(c.cacheMap, key)

	if e != nil {
		c.list.Remove(e)
	}
}

func (c *LRUTxCache) Has(tx types.Tx) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_, ok := c.cacheMap[tx.Key()]
	return ok
}

// Has checks if a transaction is in the cache with statistics tracking
func (c *StatsLRUTxCache) Has(tx types.Tx) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_, ok := c.cacheMap[tx.Key()]
	if ok {
		c.hits.Add(1)
	} else {
		c.misses.Add(1)
	}
	return ok
}

// Hits returns the number of cache hits
func (c *StatsLRUTxCache) Hits() uint64 {
	return c.hits.Load()
}

// Misses returns the number of cache misses
func (c *StatsLRUTxCache) Misses() uint64 {
	return c.misses.Load()
}

// Evictions returns the number of cache evictions
func (c *StatsLRUTxCache) Evictions() uint64 {
	return c.evictions.Load()
}

// Size returns the current number of items in the cache
func (c *StatsLRUTxCache) Size() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.list.Len()
}

// ResetStats resets all statistics counters to zero
func (c *StatsLRUTxCache) ResetStats() {
	c.hits.Store(0)
	c.misses.Store(0)
	c.evictions.Store(0)
}

// NopTxCache defines a no-op raw transaction cache.
type NopTxCache struct{}

var _ TxCache = (*NopTxCache)(nil)

func (NopTxCache) Reset()             {}
func (NopTxCache) Push(types.Tx) bool { return true }
func (NopTxCache) Remove(types.Tx)    {}
func (NopTxCache) Has(types.Tx) bool  { return false }
