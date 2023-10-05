package mempool

import (
	"container/list"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/types"
)

// TxCache defines an interface for transaction caching.
// Currently, a TxCache does not allow direct reading or getting of transaction
// values. A TxCache is used primarily to push transactions and removing
// transactions. Pushing via Push returns a boolean telling the caller if the
// transaction already exists in the cache or not.
type TxCache[T comparable] interface {
	// Reset resets the cache to an empty state.
	Reset()

	// Push adds the given transaction key to the cache and returns true if it was
	// newly added. Otherwise, it returns false.
	Push(v T) bool

	// Remove removes the given transaction from the cache.
	Remove(v T)

	// Has reports whether tx is present in the cache. Checking for presence is
	// not treated as an access of the value.
	Has(v T) bool
}

var _ TxCache[types.TxKey] = (*LRUTxCache[types.TxKey])(nil)

// LRUTxCache maintains a thread-safe LRU cache of transaction hashes (keys).
type LRUTxCache[T comparable] struct {
	mtx      cmtsync.Mutex
	size     int
	cacheMap map[T]*list.Element
	list     *list.List
}

func NewLRUTxCache[T comparable](cacheSize int) *LRUTxCache[T] {
	return &LRUTxCache[T]{
		size:     cacheSize,
		cacheMap: make(map[T]*list.Element, cacheSize),
		list:     list.New(),
	}
}

// GetList returns the underlying linked-list that backs the LRU cache. Note,
// this should be used for testing purposes only!
func (c *LRUTxCache[T]) GetList() *list.List {
	return c.list
}

func (c *LRUTxCache[T]) Reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.cacheMap = make(map[T]*list.Element, c.size)
	c.list.Init()
}

func (c *LRUTxCache[T]) Push(v T) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	moved, ok := c.cacheMap[v]
	if ok {
		c.list.MoveToBack(moved)
		return false
	}

	if c.list.Len() >= c.size {
		front := c.list.Front()
		if front != nil {
			frontKey := front.Value.(T)
			delete(c.cacheMap, frontKey)
			c.list.Remove(front)
		}
	}

	e := c.list.PushBack(v)
	c.cacheMap[v] = e

	return true
}

func (c *LRUTxCache[T]) Remove(v T) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	e := c.cacheMap[v]
	delete(c.cacheMap, v)

	if e != nil {
		c.list.Remove(e)
	}
}

func (c *LRUTxCache[T]) Has(v T) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_, ok := c.cacheMap[v]
	return ok
}

// NopTxCache defines a no-op transaction cache.
type NopTxCache[T comparable] struct{}

var _ TxCache[types.TxKey] = (*NopTxCache[types.TxKey])(nil)

func (NopTxCache[T]) Reset()                {}
func (NopTxCache[T]) Push(types.TxKey) bool { return true }
func (NopTxCache[T]) Remove(types.TxKey)    {}
func (NopTxCache[T]) Has(types.TxKey) bool  { return false }
