package guard

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Guard[K comparable] struct {
	lru *lru.Cache[K, bool]
}

func New[K comparable](capacity int) *Guard[K] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	cache, _ := lru.New[K, bool](capacity)

	return &Guard[K]{
		lru: cache,
	}
}

// Guard guards the item against duplicates.
// Returns true if the item was added to the guard, false if it was already present.
func (g *Guard[K]) Guard(key K) (added bool) {
	found, _ := g.lru.ContainsOrAdd(key, true)

	return !found
}

func (g *Guard[K]) Has(key K) bool {
	return g.lru.Contains(key)
}

func (g *Guard[K]) Forget(key K) (present bool) {
	return g.lru.Remove(key)
}

func (g *Guard[K]) ForgetAfter(key K, duration time.Duration) {
	if duration <= 0 {
		g.Forget(key)
		return
	}

	// todo async way w/o many goroutines
	go func() {
		time.Sleep(duration)
		g.Forget(key)
	}()
}

func (g *Guard[K]) Purge() {
	g.lru.Purge()
}
