package storage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
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

func TestIteratorIteratingImpl(t *testing.T) {
	testDBs, err := newTestDBs()
	if err != nil {
		t.Fatal(err)
	}

	var (
		a, b, c, d = []byte{'a'}, []byte{'b'}, []byte{'c'}, []byte{'d'}
		keys       = [][]byte{a, b, c, d}

		testCases = []struct {
			start, end []byte
			reverse    bool

			// expected keys visited by the iterator in order.
			wantVisit [][]byte
		}{
			{start: nil, end: nil, reverse: false, wantVisit: [][]byte{a, b, c, d}},
			{start: nil, end: nil, reverse: true, wantVisit: [][]byte{d, c, b, a}},

			// Because 'end is exclusive, and because 'a' is the first key in the DB,
			// setting it as the iterator's upper bound will create an iterator over
			// an empty key range.
			{start: nil, end: a, reverse: false, wantVisit: [][]byte{}},
			{start: nil, end: a, reverse: true, wantVisit: [][]byte{}},
			{start: nil, end: b, reverse: false, wantVisit: [][]byte{a}},
			{start: nil, end: b, reverse: true, wantVisit: [][]byte{a}},
			{start: nil, end: c, reverse: false, wantVisit: [][]byte{a, b}},

			// Because 'end' is exclusive, setting 'c' as the iterator's upper bound
			// of a reverse iterator will create an iterator whose starting key
			// ('c') will be skipped.
			{start: nil, end: c, reverse: true, wantVisit: [][]byte{b, a}},
			{start: nil, end: d, reverse: false, wantVisit: [][]byte{a, b, c}},
			{start: nil, end: d, reverse: true, wantVisit: [][]byte{c, b, a}},

			{start: a, end: nil, reverse: false, wantVisit: [][]byte{a, b, c, d}},

			// 'start' is inclusive, so setting 'a' as the iterator's lower bound of
			// a reverse iterator will include 'a', even if 'a' is the last
			// effectively becomes the last key in the key range.
			{start: a, end: nil, reverse: true, wantVisit: [][]byte{d, c, b, a}},
			{start: a, end: b, reverse: false, wantVisit: [][]byte{a}},
			{start: a, end: b, reverse: true, wantVisit: [][]byte{a}},
			{start: a, end: c, reverse: false, wantVisit: [][]byte{a, b}},
			{start: a, end: c, reverse: true, wantVisit: [][]byte{b, a}},
			{start: a, end: d, reverse: false, wantVisit: [][]byte{a, b, c}},
			{start: a, end: d, reverse: true, wantVisit: [][]byte{c, b, a}},

			{start: b, end: nil, reverse: false, wantVisit: [][]byte{b, c, d}},
			{start: b, end: nil, reverse: true, wantVisit: [][]byte{d, c, b}},
			{start: b, end: c, reverse: false, wantVisit: [][]byte{b}},
			{start: b, end: c, reverse: true, wantVisit: [][]byte{b}},
			{start: b, end: d, reverse: false, wantVisit: [][]byte{b, c}},
			{start: b, end: d, reverse: true, wantVisit: [][]byte{c, b}},

			{start: c, end: nil, reverse: false, wantVisit: [][]byte{c, d}},
			{start: c, end: nil, reverse: true, wantVisit: [][]byte{d, c}},
			{start: c, end: d, reverse: false, wantVisit: [][]byte{c}},
			{start: c, end: d, reverse: true, wantVisit: [][]byte{c}},

			{start: d, end: nil, reverse: false, wantVisit: [][]byte{d}},
			{start: d, end: nil, reverse: true, wantVisit: [][]byte{d}},
		}

		equalFunc = func(a, b []byte) bool {
			return slices.Equal(a, b)
		}
	)

	for _, tDB := range testDBs {
		for i, key := range keys {
			if err := tDB.SetSync(key, []byte{byte(i)}); err != nil {
				t.Fatalf("test %d: setting key: %s", i, err)
			}
		}

		for i, tc := range testCases {
			var (
				it  Iterator
				err error
			)
			if tc.reverse {
				it, err = tDB.ReverseIterator(tc.start, tc.end)
				if err != nil {
					t.Fatalf("test %d: creating forward test iterator: %s", i, err)
				}
			} else {
				it, err = tDB.Iterator(tc.start, tc.end)
				if err != nil {
					t.Fatalf("test %d: creating reverse test iterator: %s", i, err)
				}
			}

			visited := make([][]byte, 0, len(tc.wantVisit))
			for ; it.Valid(); it.Next() {
				currKey := make([]byte, len(it.Key()))
				copy(currKey, it.Key())
				visited = append(visited, currKey)
			}

			if err := it.Error(); err != nil {
				t.Errorf("test %d: unexpected error: %s", i, err)
			}

			equalOrder := slices.EqualFunc(visited, tc.wantVisit, equalFunc)
			if !equalOrder {
				formatStr := "test %d:\nwant visit order: %s\ngot: %s"
				t.Errorf(formatStr, i, tc.wantVisit, visited)
			}

			if err := it.Close(); err != nil {
				t.Errorf("test %d: closing iterator: %s", i, err)
			}
		}

		if err := tDB.Close(); err != nil {
			t.Errorf("closing test database: %s", err)
		}
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
		it, err := IteratePrefix(pDB, prefix)
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
		it, err := IteratePrefix(pDB, prefix)
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
