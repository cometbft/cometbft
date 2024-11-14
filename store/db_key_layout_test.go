package store

import (
	"bytes"
	"strconv"
	"testing"
)

func FuzzCalcBlockPartKey(f *testing.F) {
	layout := &v1LegacyLayout{}

	f.Add(int64(0), 0)
	f.Add(int64(1234567890123456789), 1)
	f.Add(int64(-9223372036854775808), -1)
	f.Add(int64(9223372036854775807), 2147483647)
	f.Add(int64(0), -2147483648)

	f.Fuzz(func(t *testing.T, height int64, partIndex int) {
		key := layout.CalcBlockPartKey(height, partIndex)

		// Ensure the key starts with the "P:" prefix.
		// 2 is the length of "P:".
		if len(key) < 2 || key[0] != 'P' || key[1] != ':' {
			t.Fatalf("key does not start with 'P:': %s", key)
		}

		sepIndex := bytes.LastIndexByte(key, ':')
		if sepIndex == -1 {
			t.Fatalf("key does not have ':' between height and partIndex: %s", key)
		}

		heightStr := string(key[2:sepIndex])
		gotHeight, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil {
			t.Fatalf("parsing height from key: %s, error: %s", key, err)
		}
		if gotHeight != height {
			t.Fatalf("want height %d, but got %d", height, gotHeight)
		}

		partIndexStr := string(key[sepIndex+1:])
		gotPartIndex, err := strconv.Atoi(partIndexStr)
		if err != nil {
			t.Fatalf("parsing partIndex from key: %s, error: %s", key, err)
		}
		if gotPartIndex != partIndex {
			t.Errorf("want partIndex %d, but got %d", partIndex, gotPartIndex)
		}
	})
}

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
