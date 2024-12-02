package storage

import (
	"bytes"
	"testing"
)

func TestPrefixIterator(t *testing.T) {
	// We create an an empty DB because we won't be iterating through keys.
	// This test checks whether the iterator is initialized correctly, that is,
	// whether its start and end keys are set correctly.
	pDB, closer, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closer)

	t.Run("EmptyPrefix", func(t *testing.T) {
		prefix := []byte{}
		it, err := PrefixIterator(pDB, prefix)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var (
			wantStart, wantEnd []byte
			gotStart, gotEnd   = it.Domain()
		)
		if !bytes.Equal(gotStart, wantStart) {
			formatStr := "expected iterator start key to be %v, got: %v"
			t.Fatalf(formatStr, wantStart, gotStart)
		}
		if !bytes.Equal(gotEnd, wantEnd) {
			formatStr := "expected iterator end key to be %v, got: %v"
			t.Fatalf(formatStr, wantEnd, gotEnd)
		}

		if err := it.Close(); err != nil {
			t.Fatalf("closing test iterator: %s", err)
		}
	})

	t.Run("Prefix", func(t *testing.T) {
		prefix := []byte("test_prefix_iterator")
		it, err := PrefixIterator(pDB, prefix)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var (
			wantStart        = prefix
			wantEnd          = incrementBigEndian(prefix)
			gotStart, gotEnd = it.Domain()
		)
		if !bytes.Equal(gotStart, wantStart) {
			formatStr := "expected iterator start key to be %v, got: %v"
			t.Fatalf(formatStr, wantStart, gotStart)
		}
		if !bytes.Equal(gotEnd, wantEnd) {
			formatStr := "expected iterator end key to be %v, got: %v"
			t.Fatalf(formatStr, wantEnd, gotEnd)
		}

		if err := it.Close(); err != nil {
			t.Fatalf("closing test iterator: %s", err)
		}
	})
}
