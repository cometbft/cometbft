package state

import (
	"strconv"
	"testing"
)

// Fuzzing only the CalcABCIResponsesKey method, because the other methods of
// v1LegacyLayout do the same thing, just with a different prefix.
// Therefore, results will be the same.
func FuzzCalcABCIResponsesKey(f *testing.F) {
	layout := v1LegacyLayout{}

	// Add seed inputs for fuzzing.
	f.Add(int64(0))
	f.Add(int64(42))
	f.Add(int64(1245600))
	f.Add(int64(1234567890))
	f.Add(int64(9223372036854775807))

	f.Fuzz(func(t *testing.T, height int64) {
		if height < 0 {
			// height won't be < 0, so skip
			t.SkipNow()
		}

		key := layout.CalcABCIResponsesKey(height)

		const prefix = "abciResponsesKey:"
		gotPrefix := string(key[:len(prefix)])

		if len(key) < len(prefix) || gotPrefix != prefix {
			t.Fatalf("key does not start with prefix '%s': %s", prefix, key)
		}

		heightStr := string(key[len(prefix):])
		gotHeight, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			t.Fatalf("parsing height from key: %s, error: %s", key, err)
		}
		if gotHeight != height {
			t.Errorf("want height %d, but got%d", height, gotHeight)
		}
	})
}
