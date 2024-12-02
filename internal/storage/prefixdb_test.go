package storage

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestPrefixDBGet(t *testing.T) {
	pebbleDB, dbCloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	prefixDB := &PrefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := prefixDB.Get(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistReturnsNil", func(t *testing.T) {
		val, err := prefixDB.Get([]byte{'a'})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if val != nil {
			t.Errorf("expected nil value, got: %s", val)
		}
	})

	t.Run("KeyExistReturnsValue", func(t *testing.T) {
		var (
			key         = []byte{'a'}
			prefixedKey = append(prefixDB.prefix, key...)
			val         = []byte{'b'}
		)
		// we are calling PebbleDB's [SetSync] directly, therefore we must prepend
		// the prefix to the key ourselves.
		if err := prefixDB.db.SetSync(prefixedKey, val); err != nil {
			t.Fatalf("writing to test DB: %s", err)
		}

		gotVal, err := prefixDB.Get(key)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !bytes.Equal(gotVal, val) {
			t.Errorf("expected value: %s, got: %v", val, gotVal)
		}
	})
}

func TestPrefixDBHas(t *testing.T) {
	pebbleDB, dbcloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbcloser)

	prefixDB := &PrefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := prefixDB.Has(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistReturnsFalse", func(t *testing.T) {
		hasKey, err := prefixDB.Has([]byte{'a'})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if hasKey {
			t.Error("expected false, but got true")
		}
	})

	t.Run("KeyExistReturnsTrue", func(t *testing.T) {
		var (
			key         = []byte{'a'}
			prefixedKey = append(prefixDB.prefix, key...)
			val         = []byte{'b'}
		)
		// we are calling PebbleDB's [SetSync] directly, therefore we must prepend
		// the prefix to the key ourselves.
		if err := prefixDB.db.SetSync(prefixedKey, val); err != nil {
			t.Fatalf("writing to test DB: %s", err)
		}

		hasKey, err := prefixDB.Has(key)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !hasKey {
			t.Error("expected true, but got false")
		}
	})
}

func TestPrefixDBSet(t *testing.T) {
	pebbleDB, dbCloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	prefixDB := &PrefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := prefixDB.Set(nil, nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("NilValueErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := prefixDB.Set(key, nil); !errors.Is(err, errValueNil) {
			t.Errorf("expected %s, got: %s", errValueNil, err)
		}
	})

	t.Run("NoErr", func(t *testing.T) {
		var (
			keys = [][]byte{{'a'}, {'b'}}
			vals = [][]byte{{0x01}, {0x02}}
		)
		if err := prefixDB.Set(keys[0], vals[0]); err != nil {
			t.Fatalf("unsynced Set unexpected error: %s", err)
		}

		if err := prefixDB.SetSync(keys[1], vals[1]); err != nil {
			t.Fatalf("synced Set unexpected error: %s", err)
		}

		for i, key := range keys {
			// we are calling PebbleDB's [Get] directly, therefore we must u
			// prepend the prefix to the key ourselves.
			prefixedKey := append(prefixDB.prefix, key...)
			storedVal, err := prefixDB.db.Get(prefixedKey)
			if err != nil {
				t.Errorf("test %d: reading from test DB: %s", i, err)
			}

			wantVal := vals[i]
			if !bytes.Equal(storedVal, wantVal) {
				const format = "test %d: expected value: %v, got: %v"
				t.Errorf(format, i, wantVal, storedVal)
			}
		}
	})
}

func TestPrefixDBDelete(t *testing.T) {
	pebbleDB, dbCloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	prefixDB := &PrefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := prefixDB.Delete(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistNoErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := prefixDB.Delete(key); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("KeyExistNoErr", func(t *testing.T) {
		if err := deletePrefixDBHelper(prefixDB, false); err != nil {
			t.Errorf("unsynced Delete unexpected error: %s", err)
		}

		if err := deletePrefixDBHelper(prefixDB, true); err != nil {
			t.Errorf("synced Delete unexpected error: %s", err)
		}
	})
}

func TestPrefixDBPrint(t *testing.T) {
	pebbleDB, dbCloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	prefixDB := &PrefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	kvPairs := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	for k, v := range kvPairs {
		prefixedKey := prependPrefix(prefixDB.prefix, []byte(k))
		if err := pebbleDB.Set(prefixedKey, []byte(v)); err != nil {
			t.Fatalf("writing key %s to test DB: %s", prefixedKey, err)
		}
	}

	// Print() writes to os.Stdout, so we need to do some awkward shenanigans to
	// capture the output and check it's correct.
	r, w, err := os.Pipe()
	if err != nil {
		const format = "Error creating pipe to capture os.Stdout contents: %s"
		t.Fatalf(format, err)
	}

	// store os.Stdout and redirect it to print to the writer we just created
	stdOut := os.Stdout
	os.Stdout = w

	if err := prefixDB.Print(); err != nil {
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
			const formatStr = "\nthis line was not printed: %q\nfull print: %q"
			t.Errorf(formatStr, wantStr, outputStr)
		}
	}
}

func TestPrefixDBBatchSet(t *testing.T) {
	pebbleBatch, closer, err := newTestPebbleBatch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closer)

	prefixBatch := &prefixDBBatch{
		source: pebbleBatch,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := prefixBatch.Set(nil, nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("ValueNilErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := prefixBatch.Set(key, nil); !errors.Is(err, errValueNil) {
			t.Errorf("expected %s, got: %s", errValueNil, err)
		}
	})

	// Our implementation's batch isn't indexed, so we can't call Get() on it to
	// retrieve keys that we added to it. Therefore, we can't check if the call to
	// Set() added the key to the batch. To do that, we would have to commit the
	// batch and then query the database for the key. However, committing a batch is
	// what Write() and WriteSync() do—not what Set() does; we test this behavior
	// in TestBatchWrite. Therefore, here we only check that after a call to Set()
	// the batch isn't empty and contains exactly the number of updates that we set.
	t.Run("NoErr", func(t *testing.T) {
		var (
			keys = [][]byte{{'a'}, {'b'}, {'c'}}
			vals = [][]byte{{0x01}, {0x02}, {0x03}}
		)
		for i, key := range keys {
			val := vals[i]

			if err := prefixBatch.Set(key, val); err != nil {
				formatStr := "adding set (k,v)=(%s,%v) operation to batch: %s"
				t.Fatalf(formatStr, key, val, err)
			}
		}

		var (
			// we are a bit cheating here, since we don't have access to the actual
			// pebble batch from the prefixDBBatch object.
			emptyBatch = pebbleBatch.batch.Empty()
			nUpdates   = pebbleBatch.batch.Count()
		)
		if emptyBatch || (nUpdates != uint32(len(keys))) {
			t.Errorf("expected %d batch updates, got %d", len(keys), nUpdates)
		}
	})

	t.Run("BatchNilErr", func(t *testing.T) {
		if err := prefixBatch.Close(); err != nil {
			t.Fatalf("closing test batch: %s", err)
		}
		var (
			key   = []byte{'a'}
			value = []byte{'b'}
		)
		if err := prefixBatch.Set(key, value); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
		}
	})
}

func TestPrefixDBBatchDelete(t *testing.T) {
	pebbleBatch, closer, err := newTestPebbleBatch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closer)

	prefixBatch := &prefixDBBatch{
		source: pebbleBatch,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := prefixBatch.Delete(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	// Our implementation's batch isn't indexed, so we can't call Get() on it to
	// retrieve keys that we added to it. Therefore, we can't check if the call to
	// Delete() added the key to the batch. To do that, we would have to commit the
	// batch and then query the database for the key. However, committing a batch is
	// what Write() and WriteSync() do—not what Delete() does; we test this behavior
	// in TestBatchWrite. Therefore, here we only check that after a call to Delete()
	// the batch isn't empty and contains exactly one update.
	t.Run("NoErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := prefixBatch.Delete(key); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var (
			// we are a bit cheating here, since we don't have access to the actual
			// pebble batch from the prefixDBBatch object.
			emptyBatch = pebbleBatch.batch.Empty()
			nUpdates   = pebbleBatch.batch.Count()
		)
		if emptyBatch || (nUpdates != 1) {
			t.Errorf("expected %d batch updates, got %d", 1, nUpdates)
		}
	})

	t.Run("BatchNilErr", func(t *testing.T) {
		if err := prefixBatch.Close(); err != nil {
			t.Fatalf("closing test batch: %s", err)
		}

		key := []byte{'a'}
		if err := prefixBatch.Delete(key); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
		}
	})
}

func TestPrefixDBBatchWrite(t *testing.T) {
	// Because Write and WriteSync close the batch after committing it, we need to
	// create a new batch for each test.
	t.Run("UnsyncedWriteNoErr", func(t *testing.T) {
		pebbleBatch, closer, err := newTestPebbleBatch()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(closer)

		var (
			prefix   = []byte{'t', 'e', 's', 't'}
			prefixDB = &PrefixDB{
				db:     pebbleBatch.db,
				prefix: prefix,
			}
			prefixBatch = &prefixDBBatch{
				source: pebbleBatch,
				prefix: prefix,
			}
		)
		err = pebbleBatchWriteTestHelper(prefixBatch, prefixDB, false)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("SyncedWriteNoErr", func(t *testing.T) {
		pebbleBatch, closer, err := newTestPebbleBatch()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(closer)

		var (
			prefix   = []byte{'t', 'e', 's', 't'}
			prefixDB = &PrefixDB{
				db:     pebbleBatch.db,
				prefix: prefix,
			}
			prefixBatch = &prefixDBBatch{
				source: pebbleBatch,
				prefix: prefix,
			}
		)
		err = pebbleBatchWriteTestHelper(prefixBatch, prefixDB, true)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("BatchNilErr", func(t *testing.T) {
		pebbleBatch, closer, err := newTestPebbleBatch()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(closer)

		prefixBatch := &prefixDBBatch{
			source: pebbleBatch,
			prefix: []byte{'t', 'e', 's', 't'},
		}
		if err := prefixBatch.Close(); err != nil {
			t.Fatalf("closing test batch: %s", err)
		}
		if err := prefixBatch.Write(); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
		}
	})
}

func TestPrefixIteratorIterating(t *testing.T) {
	pebbleDB, dbCloser, err := newTestPebbleDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	var (
		pDB = &PrefixDB{
			db:     pebbleDB,
			prefix: []byte{'t', 'e', 's', 't'},
		}

		a, b, c, d = []byte{'a'}, []byte{'b'}, []byte{'c'}, []byte{'d'}
		keys       = [][]byte{a, b, c, d}

		testCases = []struct { //nolint:dupl
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

	for i, key := range keys {
		if err := pDB.SetSync(key, []byte{byte(i)}); err != nil {
			t.Fatalf("test %d: setting key: %s", i, err)
		}
	}

	for i, tc := range testCases {
		var it Iterator
		if tc.reverse {
			it, err = pDB.ReverseIterator(tc.start, tc.end)
			if err != nil {
				t.Fatalf("test %d: creating test iterator: %s", i, err)
			}
		} else {
			it, err = pDB.Iterator(tc.start, tc.end)
			if err != nil {
				t.Fatalf("test %d: creating test iterator: %s", i, err)
			}
		}

		visited := make([][]byte, 0, len(tc.wantVisit))
		for it.Valid() {
			currKey := make([]byte, len(it.Key()))
			copy(currKey, it.Key())
			visited = append(visited, currKey)

			it.Next()
		}

		if err := it.Error(); err != nil {
			t.Errorf("test %d: unexpected error: %s", i, err)
			continue
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
}

func TestPrependPrefix(t *testing.T) {
	var (
		prefix = []byte{'t', 'e', 's', 't'}
		key    = []byte{'k', 'e', 'y'}

		testCases = []struct {
			prefix []byte
			key    []byte
			want   []byte
		}{
			{ // non-nil prefix and key
				prefix: prefix,
				key:    key,
				want:   append(prefix, key...),
			},
			{ // nil key
				prefix: prefix,
				key:    nil,
				want:   prefix,
			},
			{ // nil prefix
				prefix: nil,
				key:    key,
				want:   key,
			},
			{ //  nil prefix and key
				prefix: nil,
				key:    nil,
				want:   []byte{},
			},
		}
	)

	for i, tc := range testCases {
		got := prependPrefix(tc.prefix, tc.key)
		if !bytes.Equal(got, tc.want) {
			t.Errorf("test %d: want %s, but got %s", i, tc.want, got)
		}
	}
}

func TestPrefixedIteratorBounds(t *testing.T) {
	var (
		prefix        = []byte{'t', 'e', 's', 't'}
		start         = []byte{'k', 'e', 'y', '1'}
		end           = []byte{'k', 'e', 'y', '2'}
		prefixedStart = append(prefix, start...)
		prefixedEnd   = append(prefix, end...)

		successCases = []struct {
			prefix    []byte
			start     []byte
			end       []byte
			wantStart []byte
			wantEnd   []byte
			wantErr   error
		}{
			{ // valid prefix and bounds
				prefix:    prefix,
				start:     start,
				end:       end,
				wantStart: prefixedStart,
				wantEnd:   prefixedEnd,
				wantErr:   nil,
			},
			{ // nil end (upper bound is incremented prefix)
				prefix:    prefix,
				start:     start,
				end:       nil,
				wantStart: prefixedStart,
				wantEnd:   incrementBigEndian(prefix),
				wantErr:   nil,
			},
			{ // nil start and end
				prefix:    prefix,
				start:     nil,
				end:       nil,
				wantStart: prefix,
				wantEnd:   incrementBigEndian(prefix),
				wantErr:   nil,
			},
			{ // nil prefix
				prefix:    nil,
				start:     start,
				end:       end,
				wantStart: start,
				wantEnd:   end,
				wantErr:   nil,
			},
		}
	)
	for i, tc := range successCases {
		start, end, err := prefixedIteratorBounds(tc.prefix, tc.start, tc.end)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", i, err)
		}

		if !bytes.Equal(start, tc.wantStart) {
			t.Errorf("unexpected start bound: got %s, want %s", start, tc.wantStart)
		}
		if !bytes.Equal(end, tc.wantEnd) {
			t.Errorf("unexpected end bound: got %s, want %s", end, tc.wantEnd)
		}
	}

	failureCases := []struct {
		start   []byte
		end     []byte
		wantErr error
	}{
		{ // empty start
			start:   []byte{},
			end:     end,
			wantErr: errKeyEmpty,
		},
		{ // empty end
			start:   start,
			end:     []byte{},
			wantErr: errKeyEmpty,
		},
	}
	for i, tc := range failureCases {
		_, _, err := prefixedIteratorBounds(prefix, tc.start, tc.end)
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("test %d: want error %q, but got %q", i, tc.wantErr, err)
		}
	}
}

func TestIncrementBigEndian(t *testing.T) {
	testCases := []struct {
		input      []byte
		wantResult []byte
	}{
		{[]byte{0xFE}, []byte{0xFF}},             // simple increment
		{[]byte{0xFF}, nil},                      // overflow
		{[]byte{0x00, 0x01}, []byte{0x00, 0x02}}, // simple increment
		{[]byte{0x00, 0xFF}, []byte{0x01, 0x00}}, // carry over
		{[]byte{0xFF, 0xFF}, nil},                // overflow
	}

	for i, tc := range testCases {
		gotResult := incrementBigEndian(tc.input)
		if !bytes.Equal(gotResult, tc.wantResult) {
			t.Errorf("test %d: want: %v, got: %v", i, tc.wantResult, gotResult)
		}
	}
}

// deletePrefixDBHelper is a utility function supporting TestPrefixDBDelete.
// It writes a key-value pair to the database, deletes it, then checks that the key
// is deleted.
func deletePrefixDBHelper(pDB *PrefixDB, synced bool) error {
	var (
		key = []byte{'a'}
		val = []byte{0x01}
		// we are calling the underlying DB's methods directly, therefore we must
		// append the prefix to the key ourselves.
		prefixedKey = append(pDB.prefix, key...)
	)
	if err := pDB.db.SetSync(prefixedKey, val); err != nil {
		return fmt.Errorf("writing to test DB: %s", err)
	}

	if !synced {
		if err := pDB.Delete(key); err != nil {
			return fmt.Errorf("unsynced Delete unexpected error: %s", err)
		}
	} else {
		if err := pDB.DeleteSync(key); err != nil {
			return fmt.Errorf("synced Delete unexpected error: %s", err)
		}
	}

	// check key is deleted
	gotVal, err := pDB.db.Get(prefixedKey)
	if err != nil {
		return fmt.Errorf("reading form test DB: %s", err)
	}
	if len(gotVal) > 0 {
		// our implementation of PebbleDB does not return an error if a key
		// is not found. Instead, it returns a nil error and a nil value.
		// Therefore, to check if the deletion was successful we must check
		// that the value has 0 length (len(nil_slice)==0).
		return fmt.Errorf("expected nil value, got: %v", gotVal)
	}

	return nil
}
