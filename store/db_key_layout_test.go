package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"
)

// run: go test -fuzz=FuzzCalcBlockPartKey -fuzztime 30s
func FuzzCalcBlockPartKey(f *testing.F) {
	layout := &v1LegacyLayout{}

	f.Add(int64(0), 0)
	f.Add(int64(141241), 980)
	f.Add(int64(1234567890), 12345678901)
	f.Add(int64(9223372036854775807), 2147483647)
	f.Add(int64(42), 2147483648)

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

// run: go test -fuzz=FuzzCalcBlockHashKey -fuzztime 30s
func FuzzCalcBlockHashKey(f *testing.F) {
	layout := &v1LegacyLayout{}

	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0xFF})
	f.Add([]byte{0x01, 0x02, 0x03, 0x04})
	f.Add(make([]byte, 32))

	// empty hash: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	f.Add(sha256.New().Sum(nil))

	f.Fuzz(func(t *testing.T, hash []byte) {
		key := layout.CalcBlockHashKey(hash)

		if len(key) < 3 || key[0] != 'B' || key[1] != 'H' || key[2] != ':' {
			t.Fatalf("key does not start with 'BH:': %s", key)
		}

		var (
			hashHex = key[3:]
			gotHash = make([]byte, len(hash))
		)
		_, err := hex.Decode(gotHash, hashHex)
		if err != nil {
			t.Fatalf("decoding hash from key: %s, error: %s", key, err)
		}

		// Ensure the decoded hash matches the input hash.
		if !bytes.Equal(gotHash, hash) {
			t.Fatalf("want hash %x\ngot: %x\n", hash, gotHash)
		}
	})
}
