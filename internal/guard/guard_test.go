package guard

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGuard(t *testing.T) {
	makeGuard := func(t *testing.T, capacity int) *Guard[string] {
		t.Helper()

		g := New[string](capacity)
		t.Cleanup(g.Close)

		return g
	}

	t.Run("New", func(t *testing.T) {
		t.Run("construct", func(t *testing.T) {
			g := makeGuard(t, 10)
			require.NotNil(t, g)

			require.PanicsWithValue(t, "capacity must be greater than 0", func() {
				makeGuard(t, 0)
			})
			require.PanicsWithValue(t, "capacity must be greater than 0", func() {
				makeGuard(t, -1)
			})
		})
	})

	t.Run("Guard", func(t *testing.T) {
		t.Run("happyPath", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)
			item := "item-1"

			// ACT #1
			added := g.Guard(item)

			// ASSERT #1
			require.True(t, added)
			require.True(t, g.Has(item))

			// ACT #2
			added2 := g.Guard(item)

			// ASSERT #2
			require.False(t, added2, "item should exist on second call")
		})

		t.Run("differentItems", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)

			// ACT
			firstA := g.Guard("a")
			firstB := g.Guard("b")
			secondA := g.Guard("a")

			// ASSERT
			require.True(t, firstA)
			require.True(t, firstB)
			require.False(t, secondA)
		})

		t.Run("eviction", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 2)

			// ACT
			g.Guard("a")
			g.Guard("b")
			g.Guard("c")

			// ASSERT
			require.False(t, g.Has("a"))
			require.True(t, g.Has("b"))
			require.True(t, g.Has("c"))
		})
	})

	t.Run("Has", func(t *testing.T) {
		t.Run("notPresent", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)

			// ACT + ASSERT
			require.False(t, g.Has("missing"))
		})

		t.Run("present", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)
			item := "present"
			g.Guard(item)

			// ACT + ASSERT
			require.True(t, g.Has(item))
		})
	})

	t.Run("ForgetAfter", func(t *testing.T) {
		t.Run("zeroDuration", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)
			item := "now"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, 0)

			// ASSERT
			require.False(t, g.Has(item))
		})

		t.Run("negativeDuration", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)
			item := "negative"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, -time.Second)

			// ASSERT
			require.False(t, g.Has(item))
		})

		t.Run("delayedRemovalAccess", func(t *testing.T) {
			// ARRANGE
			g := makeGuard(t, 10)
			item := "later"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, 100*time.Millisecond)

			// ASSERT #1: still present immediately after the call
			require.True(t, g.Has(item))

			// ASSERT #2: removed after the duration elapses
			check := func() bool {
				return !g.Has(item)
			}

			require.Eventually(t, check, time.Second, 10*time.Millisecond)
		})

		t.Run("delayedRemovalAsync", func(t *testing.T) {
			// ARRANGE
			g := NewWithInterval[string](10, 50*time.Millisecond)
			t.Cleanup(g.Close)

			item := "later"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, 10*time.Millisecond)

			// ASSERT #1: still present immediately after the call
			require.False(t, g.Guard(item))
			require.True(t, g.Has(item))

			// ASSERT #2: removed in a goroutine after the duration elapses
			time.Sleep(100 * time.Millisecond)

			g.mu.Lock()
			_, ok := g.lru.Peek(item)
			g.mu.Unlock()

			require.False(t, ok)
		})
	})
}
