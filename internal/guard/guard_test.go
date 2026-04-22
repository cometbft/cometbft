package guard

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuard(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("construct", func(t *testing.T) {
			g := New[string](10)
			require.NotNil(t, g)
			assert.NotNil(t, g.lru)
		})

		t.Run("zeroCapacity", func(t *testing.T) {
			assert.PanicsWithValue(t, "capacity must be greater than 0", func() { New[string](0) })
		})

		t.Run("negativeCapacity", func(t *testing.T) {
			assert.PanicsWithValue(t, "capacity must be greater than 0", func() { New[string](-1) })
		})
	})

	t.Run("Guard", func(t *testing.T) {
		t.Run("happyPath", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "item-1"

			// ACT #1
			added := g.Guard(item)

			// ASSERT #1
			assert.True(t, added)
			assert.True(t, g.Has(item))

			// ACT #2
			added2 := g.Guard(item)

			// ASSERT #2
			assert.False(t, added2, "item should exist on second call")
		})

		t.Run("differentItems", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)

			// ACT
			firstA := g.Guard("a")
			firstB := g.Guard("b")
			secondA := g.Guard("a")

			// ASSERT
			assert.True(t, firstA)
			assert.True(t, firstB)
			assert.False(t, secondA)
		})

		t.Run("eviction", func(t *testing.T) {
			// ARRANGE
			g := New[string](2)

			// ACT
			g.Guard("a")
			g.Guard("b")
			g.Guard("c")

			// ASSERT
			assert.False(t, g.Has("a"))
			assert.True(t, g.Has("b"))
			assert.True(t, g.Has("c"))
		})
	})

	t.Run("Has", func(t *testing.T) {
		t.Run("notPresent", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)

			// ACT + ASSERT
			assert.False(t, g.Has("missing"))
		})

		t.Run("present", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "present"
			g.Guard(item)

			// ACT + ASSERT
			assert.True(t, g.Has(item))
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("happyPath", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "item"
			g.Guard(item)

			// ACT
			present := g.Forget(item)

			// ASSERT
			assert.True(t, present)
			assert.False(t, g.Has(item))
		})

		t.Run("notPresent", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)

			// ACT
			present := g.Forget("missing")

			// ASSERT
			assert.False(t, present)
		})
	})

	t.Run("RemoveAfter", func(t *testing.T) {
		t.Run("zeroDuration", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "now"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, 0)

			// ASSERT
			assert.False(t, g.Has(item))
		})

		t.Run("negativeDuration", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "negative"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, -time.Second)

			// ASSERT
			assert.False(t, g.Has(item))
		})

		t.Run("asyncRemoval", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			item := "later"
			g.Guard(item)

			// ACT
			g.ForgetAfter(item, 50*time.Millisecond)

			// ASSERT #1: still present immediately after the call
			assert.True(t, g.Has(item))

			// ASSERT #2: removed after the duration elapses
			check := func() bool {
				return !g.Has(item)
			}

			require.Eventually(t, check, time.Second, 10*time.Millisecond)
		})
	})

	t.Run("Reset", func(t *testing.T) {
		t.Run("happyPath", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)
			for i := 0; i < 5; i++ {
				g.Guard("item-" + strconv.Itoa(i))
			}

			// ACT
			g.Purge()

			// ASSERT
			for i := 0; i < 5; i++ {
				assert.False(t, g.Has("item-"+strconv.Itoa(i)))
			}
		})

		t.Run("noop", func(t *testing.T) {
			// ARRANGE
			g := New[string](10)

			// ACT + ASSERT
			assert.NotPanics(t, func() { g.Purge() })
		})
	})
}
