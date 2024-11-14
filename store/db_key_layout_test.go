package store

import (
	"strconv"
	"testing"
)

func FuzzBuildKey(f *testing.F) {
	layout := &v1LegacyLayout{}

	f.Add([]byte("prefix:"), int64(42))
	f.Add([]byte("H:"), int64(42))
	f.Add([]byte("p"), int64(1234567890123456789))
	f.Add([]byte{}, int64(-1))
	f.Add([]byte("anotherPrefix"), int64(-9223372036854775808))
	f.Add([]byte(":"), int64(9223372036854775807))

	f.Fuzz(func(t *testing.T, prefix []byte, height int64) {
		key := layout.buildKey(prefix, height)

		gotPrefix := string(key[:len(prefix)])
		if len(key) < len(prefix) || gotPrefix != string(prefix) {
			t.Fatalf("key does not start with prefix: %s, got: %s", prefix, key)
		}

		heightStr := string(key[len(prefix):])
		gotHeight, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			t.Fatalf("parsing height from key: %s, error: %s", key, err)
		}
		if gotHeight != height {
			t.Fatalf("want height %d, but got %d", height, gotHeight)
		}
	})
}
