package storage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDBPrintImpl(t *testing.T) {
	testDBs, err := newTestDBs()
	if err != nil {
		t.Fatal(err)
	}

	kvPairs := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	for _, tDB := range testDBs {
		for k, v := range kvPairs {
			if err := tDB.Set([]byte(k), []byte(v)); err != nil {
				t.Fatalf("writing key %s: %s", k, err)
			}
		}

		// Print() writes to os.Stdout, so we need to do some awkward shenanigans to
		// capture the output and check it's correct.
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Error creating pipe to capture os.Stdout contents: %s", err)
		}

		// store os.Stdout and redirect it to print to the writer we just created
		stdOut := os.Stdout
		os.Stdout = w

		if err := tDB.Print(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		w.Close()

		// restore os.Stdout
		os.Stdout = stdOut

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			t.Fatalf("reading os.Stdout contents: %s", err)
		}
		r.Close()

		outputStr := buf.String()
		for k, v := range kvPairs {
			var (
				kStr = strings.ToUpper(hex.EncodeToString([]byte(k)))
				vStr = strings.ToUpper(hex.EncodeToString([]byte(v)))

				wantStr = "[" + kStr + "]:\t[" + vStr + "]\n"
			)
			if !strings.Contains(outputStr, wantStr) {
				formatStr := "this line was not printed: %q\nfull print: %q"
				t.Errorf(formatStr, wantStr, outputStr)
			}
		}

		tDB.Close()
	}
}

func TestPrefixIterator(t *testing.T) {
	// We create an an empty DB because we won't be iterating through keys.
	// This test checks whether the iterator is initialized correctly, that is,
	// whether its start and end keys are set correctly.
	pDB, closer, err := newTestPebbleDB()
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

// newTestDBs returns a slice of databases implementing the DB interface ready for
// use in testing.
func newTestDBs() ([]DB, error) {
	pDB, _, err := newTestPebbleDB()
	if err != nil {
		return nil, fmt.Errorf("creating test pebble DB: %w", err)
	}

	prefixDB, err := newTestPrefixDB()
	if err != nil {
		return nil, fmt.Errorf("creating test prefix DB: %w", err)
	}

	return []DB{pDB, prefixDB}, nil
}
