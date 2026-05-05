// Package guard provides a thread-safe dedup utility designed to prevent
// repeated processing of already "observed" items. Use-case example:
// avoiding handling already seen mempool txs.
package guard

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"
)

// Guard is a thread-safe LRU dedup cache with optional TTL before removal.
type Guard[K comparable] struct {
	lru         *simplelru.LRU[K, *entry]
	mu          sync.Mutex
	ttlRemovals map[K]struct{}
	doneCh      chan struct{}
	doneOnce    sync.Once
}

type entry struct {
	expiresAt time.Time
}

const defaultCleanInterval = 2 * time.Second

// New Guard constructor.
func New[K comparable](capacity int) *Guard[K] {
	return NewWithInterval[K](capacity, defaultCleanInterval)
}

// New Guard constructor with custom check interval.
func NewWithInterval[K comparable](capacity int, cleanInterval time.Duration) *Guard[K] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	if cleanInterval <= 0 {
		panic("cleanInterval must be greater than 0")
	}

	ttlRemovals := map[K]struct{}{}

	lru, _ := simplelru.NewLRU(
		capacity,
		func(key K, _ *entry) { delete(ttlRemovals, key) },
	)

	g := &Guard[K]{
		lru:         lru,
		doneCh:      make(chan struct{}),
		ttlRemovals: ttlRemovals,
	}

	go func() {
		ticker := time.NewTicker(cleanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				g.removeExpired()
			case <-g.doneCh:
				return
			}
		}
	}()

	return g
}

// Close closes the guard.
func (g *Guard[K]) Close() {
	g.doneOnce.Do(func() {
		g.mu.Lock()
		defer g.mu.Unlock()

		close(g.doneCh)
	})
}

// Guard guards the item against duplicates.
// Doesn't have a TTL besides LRU eviction.
// Returns true if the key was added, false if it was already present.
func (g *Guard[K]) Guard(key K) (added bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if e, exists := g.lru.Peek(key); exists {
		if !e.expired() {
			// exists and not expired -> guard!
			return false
		}
		// lazy removal
		g.lru.Remove(key)
	}

	// add new entry
	g.lru.Add(key, &entry{})

	return true
}

// Has checks if the key is already guarded.
func (g *Guard[K]) Has(key K) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	e, exists := g.lru.Peek(key)
	if !exists {
		return false
	}

	// lazy removal
	if e.expired() {
		g.lru.Remove(key)
		return false
	}

	return true
}

// ForgetAfter removes the key from the guard after a duration.
func (g *Guard[K]) ForgetAfter(key K, duration time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if duration <= 0 {
		g.lru.Remove(key)
		return
	}

	if e, ok := g.lru.Peek(key); ok {
		// access by pointer to avoid .Add call
		e.expiresAt = time.Now().Add(duration)
		g.ttlRemovals[key] = struct{}{}
	}
}

func (g *Guard[K]) removeExpired() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for key := range g.ttlRemovals {
		e, exists := g.lru.Peek(key)
		if !exists {
			continue
		}

		if e.expired() {
			g.lru.Remove(key)
		}
	}
}

func (e *entry) expired() bool {
	if e.expiresAt.IsZero() {
		return false
	}

	return time.Now().After(e.expiresAt)
}
