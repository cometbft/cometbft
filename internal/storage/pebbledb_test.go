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

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

func TestGet(t *testing.T) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := pDB.Get(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistReturnsNil", func(t *testing.T) {
		val, err := pDB.Get([]byte{'a'})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if val != nil {
			t.Errorf("expected nil value, got: %s", val)
		}
	})

	t.Run("KeyExistReturnsValue", func(t *testing.T) {
		var (
			key = []byte{'a'}
			val = []byte{'b'}
		)
		if err := pDB.db.Set(key, val, pebble.Sync); err != nil {
			t.Fatalf("writing to test DB: %s", err)
		}

		gotVal, err := pDB.Get(key)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !bytes.Equal(gotVal, val) {
			t.Errorf("expected value: %s, got: %s", val, gotVal)
		}
	})
}

func TestHas(t *testing.T) {
	pDB, dbcloser, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbcloser)

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := pDB.Has(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistReturnsFalse", func(t *testing.T) {
		hasKey, err := pDB.Has([]byte{'a'})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if hasKey {
			t.Error("expected false, but got true")
		}
	})

	t.Run("KeyExistReturnsTrue", func(t *testing.T) {
		var (
			key = []byte{'a'}
			val = []byte{'b'}
		)
		if err := pDB.db.Set(key, val, pebble.Sync); err != nil {
			t.Fatalf("writing to test DB: %s", err)
		}

		hasKey, err := pDB.Has(key)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !hasKey {
			t.Error("expected true, but got false")
		}
	})
}

// Rather than having two almost identical Test* functions testing *PebbleDB.Set and
// *PebbleDB.SetSync, we have one test function that calls *PebbleDB.setWithOpts
// once with pebble.NoSync and once with pebble.Sync.
// This should be sufficient to test the Set and SetSync methods, because under the
// hood they only differ in that they call setWithOpts with pebble.NoSync and
// pebble.Sync respectively.
func TestSet(t *testing.T) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	var (
		sync   = pebble.Sync
		noSync = pebble.NoSync
	)
	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pDB.setWithOpts(nil, nil, noSync); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("NilValueErr", func(t *testing.T) {
		key := []byte{'a'}
		// called by SetSync
		if err := pDB.setWithOpts(key, nil, sync); !errors.Is(err, errValueNil) {
			t.Errorf("expected %s, got: %s", errValueNil, err)
		}
	})

	t.Run("NoErr", func(t *testing.T) {
		// called by Set
		if err := setHelper(pDB, noSync); err != nil {
			t.Fatal(err)
		}

		// called by SetSync
		if err := setHelper(pDB, sync); err != nil {
			t.Fatal(err)
		}
	})
}

// Rather than having two almost identical Test* functions testing *PebbleDB.Delete
// and *PebbleDB.DeleteSync, we have one test function that calls
// *PebbleDB.deleteWithOpts once with pebble.NoSync and once with pebble.Sync.
// This should be sufficient to test the Delete and DeleteSync methods, because
// under the hood they only differ in that they call deleteWithOpts with
// pebble.NoSync and pebble.Sync respectively.
func TestDelete(t *testing.T) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	var (
		sync   = pebble.Sync
		noSync = pebble.NoSync
	)
	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pDB.deleteWithOpts(nil, noSync); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistNoErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := pDB.deleteWithOpts(key, sync); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("KeyExistNoErr", func(t *testing.T) {
		// called by Delete
		if err := deleteHelper(pDB, noSync); err != nil {
			t.Fatal(err)
		}

		// called by DeleteSync
		if err := deleteHelper(pDB, sync); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCompact(t *testing.T) {
	var (
		// make sure keys and vals are the same length.
		keys = [][]byte{
			{'a'},
			{'b'},
			{'c'},
			{'d'},
			{'e'},
			{'f'},
			{'g'},
			{'h'},
			{'i'},
			{'j'},
		}
		vals = [][]byte{
			{0x01},
			{0x02},
			{0x03},
			{0x04},
			{0x05},
			{0x06},
			{0x07},
			{0x08},
			{0x09},
			{0x0a},
		}

		sync = pebble.Sync

		createTestDB = func(t *testing.T) (*PebbleDB, func()) {
			t.Helper()

			pDB, dbCloser, err := newTestEmptyDB()
			if err != nil {
				t.Fatal(err)
			}

			for i, key := range keys {
				if err := pDB.db.Set(key, vals[i], sync); err != nil {
					t.Fatalf("writing key %s: %s", key, err)
				}
			}
			return pDB, dbCloser
		}
	)

	// The following tests will create their own DBs to test compaction, so that
	// each compaction operation works on a DB that has never been compacted
	// before.
	t.Run("NilStartNoErr", func(t *testing.T) {
		pDB, dbCloser := createTestDB(t)
		t.Cleanup(dbCloser)

		// if start is nil, compaction starts from the first key in the DB.
		end := keys[2]
		if err := pDB.Compact(nil, end); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("NilEndNoErr", func(t *testing.T) {
		pDB, dbCloser := createTestDB(t)
		t.Cleanup(dbCloser)

		// if end is nil, compaction ends at the last key in the DB.
		start := keys[0]
		if err := pDB.Compact(start, nil); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("StartEndNilNoErr", func(t *testing.T) {
		pDB, dbCloser := createTestDB(t)
		t.Cleanup(dbCloser)

		// if start and end are nil, compaction starts from the first key and ends
		// at the last key in the DB.
		if err := pDB.Compact(nil, nil); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("StartEndNoErr", func(t *testing.T) {
		pDB, dbCloser := createTestDB(t)
		t.Cleanup(dbCloser)

		var (
			start = keys[2]
			end   = keys[8]
		)
		if err := pDB.Compact(start, end); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
}

func TestPrint(t *testing.T) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	kvPairs := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	for k, v := range kvPairs {
		if err := pDB.Set([]byte(k), []byte(v)); err != nil {
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

	if err := pDB.Print(); err != nil {
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
}

func TestBatchSet(t *testing.T) {
	pBatch, dbCloser, err := newBatch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pBatch.Set(nil, nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("ValueNilErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := pBatch.Set(key, nil); !errors.Is(err, errValueNil) {
			t.Errorf("expected %s, got: %s", errValueNil, err)
		}
	})

	t.Run("BatchNilErr", func(t *testing.T) {
		var (
			pBatch = &pebbleDBBatch{
				batch: nil,
			}

			key   = []byte{'a'}
			value = []byte{'b'}
		)
		if err := pBatch.Set(key, value); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
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

			if err := pBatch.Set(key, val); err != nil {
				formatStr := "adding set (k,v)=(%s,%v) operation to batch: %s"
				t.Fatalf(formatStr, key, val, err)
			}
		}

		var (
			emptyBatch = pBatch.batch.Empty()
			nUpdates   = pBatch.batch.Count()
		)
		if emptyBatch || (nUpdates != uint32(len(keys))) {
			t.Errorf("expected %d batch updates, got %d", len(keys), nUpdates)
		}
	})
}

func TestBatchDelete(t *testing.T) {
	pBatch, dbCloser, err := newBatch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pBatch.Delete(nil); !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("BatchNilErr", func(t *testing.T) {
		var (
			pBatch = &pebbleDBBatch{
				batch: nil,
			}
			key = []byte{'a'}
		)
		if err := pBatch.Delete(key); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
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
		if err := pBatch.Delete(key); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var (
			emptyBatch = pBatch.batch.Empty()
			nUpdates   = pBatch.batch.Count()
		)
		if emptyBatch || (nUpdates != 1) {
			t.Errorf("expected %d batch updates, got %d", 1, nUpdates)
		}
	})
}

// Rather than having two almost identical TestBatch* functions testing
// *pebbleDBBatch.Write and *PebbleDBBatch.WriteSync, we have one test function that
// calls *PebbleDBBatch.commitWithOpts once with pebble.NoSync and once with
// pebble.Sync.
// This should be sufficient to test the Write and WriteSync methods, because under
// the hood they only differ in that they call commitWithOpts with pebble.NoSync and
// pebble.Sync respectively.
func TestBatchWrite(t *testing.T) {
	var (
		sync   = pebble.Sync
		noSync = pebble.NoSync
	)

	t.Run("BatchNilErr", func(t *testing.T) {
		pBatch := &pebbleDBBatch{
			batch: nil,
		}
		if err := pBatch.commitWithOpts(sync); !errors.Is(err, errBatchClosed) {
			t.Errorf("expected %s, got: %s", errBatchClosed, err)
		}
	})

	// Because Write and WriteSync close the batch after committing it, we need to
	// create a new batch for each test.
	t.Run("UnsyncedWriteNoErr", func(t *testing.T) {
		pBatch, dbCloser, err := newBatch()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(dbCloser)

		if err := batchWriteHelper(pBatch, noSync); err != nil {
			t.Error(err)
		}
	})

	t.Run("SyncedWriteNoErr", func(t *testing.T) {
		pBatch, dbCloser, err := newBatch()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(dbCloser)

		if err := batchWriteHelper(pBatch, sync); err != nil {
			t.Error(err)
		}
	})
}

func TestBatchClose(t *testing.T) {
	t.Run("BatchNilNoErr", func(t *testing.T) {
		pBatch := &pebbleDBBatch{
			batch: nil,
		}
		if err := pBatch.Close(); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("NoErr", func(t *testing.T) {
		pBatch, dbCloser, err := newBatch()
		if err != nil {
			t.Fatal(err)
		}

		if err := pBatch.Close(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if pBatch.batch != nil {
			t.Error("expected batch to be nil")
		}

		t.Cleanup(dbCloser)
	})
}

func TestNewPebbleDBIterator(t *testing.T) {
	emptyKey := []byte{}

	t.Run("EmptyStartErr", func(t *testing.T) {
		_, err := newPebbleDBIterator(nil, emptyKey, nil, false)
		if !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected error: %s\n got: %s", errKeyEmpty, err)
		}
	})

	t.Run("EmptyEndErr", func(t *testing.T) {
		_, err := newPebbleDBIterator(nil, nil, emptyKey, false)
		if !errors.Is(err, errKeyEmpty) {
			t.Errorf("expected error: %s\n got: %s", errKeyEmpty, err)
		}
	})

	pDB, dbCloser, err := newTestDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	t.Run("ForwardIteratorNoErr", func(t *testing.T) {
		var (
			// a nil start and end will default to the first and last keys in the db
			start, end []byte
			isReverse  bool
		)
		it, err := newPebbleDBIterator(pDB, start, end, isReverse)
		if err != nil {
			t.Fatalf("creating test iterator: %s", err)
		}
		t.Cleanup(func() { it.source.Close() })

		if !bytes.Equal(it.start, start) {
			formatStr := "expected iterator lower bound to be %v, got: %v"
			t.Fatalf(formatStr, start, it.start)
		}
		if !bytes.Equal(it.end, end) {
			formatStr := "expected iterator upper bound to be %v, got: %v"
			t.Fatalf(formatStr, end, it.end)
		}

		// The first key in the test db is 'a'. See [newTestDB].
		wantStart := []byte{'a'}
		if !bytes.Equal(it.source.Key(), wantStart) {
			formatStr := "expected iterator current key to be %s, got: %s"
			t.Fatalf(formatStr, wantStart, it.Key())
		}
	})

	t.Run("ReverseIteratorNoErr", func(t *testing.T) {
		var (
			// a nil start and end will default to the first and last keys in the db
			start, end []byte
			isReverse  = true
		)
		it, err := newPebbleDBIterator(pDB, start, end, isReverse)
		if err != nil {
			t.Fatalf("creating test iterator: %s", err)
		}
		t.Cleanup(func() { it.source.Close() })

		if !bytes.Equal(it.start, start) {
			formatStr := "expected iterator lower bound to be %v, got: %v"
			t.Fatalf(formatStr, start, it.start)
		}
		if !bytes.Equal(it.end, end) {
			formatStr := "expected iterator upper bound to be %v, got: %v"
			t.Fatalf(formatStr, end, it.end)
		}

		// The last key in the test db is 'd'. See [newTestDB].
		wantEnd := []byte{'d'}
		if !bytes.Equal(it.source.Key(), wantEnd) {
			formatStr := "expected iterator current key to be %s, got: %s"
			t.Fatalf(formatStr, wantEnd, it.Key())
		}
	})
}

func TestIteratorIterating(t *testing.T) {
	pDB, dbCloser, err := newTestDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dbCloser)

	var (
		a, b, c, d = []byte{'a'}, []byte{'b'}, []byte{'c'}, []byte{'d'}

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

	for i, tc := range testCases {
		it, err := newPebbleDBIterator(pDB, tc.start, tc.end, tc.reverse)
		if err != nil {
			t.Fatalf("test %d: creating test iterator: %s", i, err)
		}

		visited := make([][]byte, 0, len(tc.wantVisit))
		for it.Valid() {
			currKey := make([]byte, len(it.Key()))
			copy(currKey, it.Key())
			visited = append(visited, currKey)

			it.Next()
		}

		if err := it.Error(); err != nil {
			t.Fatalf("test %d: unexpected error: %s", i, err)
		}

		equalOrder := slices.EqualFunc(visited, tc.wantVisit, equalFunc)
		if !equalOrder {
			formatStr := "test %d:\nwant visit order: %s\ngot: %s"
			t.Fatalf(formatStr, i, tc.wantVisit, visited)
		}

		if err := it.source.Close(); err != nil {
			t.Fatalf("test %d: closing iterator: %s", i, err)
		}
	}
}

// newTestEmptyDB creates an in-memory instance of pebble for testing.
// It returns a closer function that must be called to close the database when done
// with it.
func newTestEmptyDB() (*PebbleDB, func(), error) {
	opts := &pebble.Options{FS: vfs.NewMem()}
	memDB, err := pebble.Open("", opts)
	if err != nil {
		return nil, nil, fmt.Errorf("creating test DB: %w", err)
	}

	var (
		closer = func() { memDB.Close() }
		pDB    = &PebbleDB{db: memDB}
	)
	return pDB, closer, nil
}

// newTestDB creates an in-memory instance of pebble for testing pre-populated with
// a few dummy kv pairs.
// It returns a closer function that must be called to close the database when done
// with it.
func newTestDB() (*PebbleDB, func(), error) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		return nil, nil, fmt.Errorf("creating test iterator: %w", err)
	}

	var (
		keys = [][]byte{{'a'}, {'b'}, {'c'}, {'d'}}
		vals = [][]byte{{0x01}, {0x02}, {0x03}, {0x04}}
	)
	for i, key := range keys {
		val := vals[i]
		if err := pDB.db.Set(key, val, nil); err != nil {
			formatStr := "creating test iterator: setting (k,v)=(%v,%v): %s"
			return nil, nil, fmt.Errorf(formatStr, key, val, err)
		}
	}

	return pDB, dbCloser, nil
}

// newBatch creates a new batch you can use to apply operations to a database that
// the function creates. The underlying database is an in-memory instance of pebble.
// newBatch returns a closer function that must be called to close the batch and the
// database when done with them.
func newBatch() (*pebbleDBBatch, func(), error) {
	pDB, dbCloser, err := newTestEmptyDB()
	if err != nil {
		return nil, nil, fmt.Errorf("creating test batch: %w", err)
	}

	var (
		pBatch = &pebbleDBBatch{
			db:    pDB,
			batch: pDB.db.NewBatch(),
		}
		closer = func() {
			// the Batch write methods close the batch and set it to nil, so we need
			// this check to prevent a panic.
			if pBatch.batch != nil {
				pBatch.batch.Close()
			}
			dbCloser()
		}
	)

	return pBatch, closer, nil
}

// setHelper is a utility function supporting TestSet.
// It writes a key-value pair to the database, then reads it back.
func setHelper(pDB *PebbleDB, writeOpts *pebble.WriteOptions) error {
	var (
		key = []byte{'a'}
		val = []byte{0x01}
	)
	if err := pDB.setWithOpts(key, val, writeOpts); err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	storedVal, closer, err := pDB.db.Get(key)
	if err != nil {
		return fmt.Errorf("reading from test DB: %s", err)
	}
	if !bytes.Equal(storedVal, val) {
		return fmt.Errorf("expected value: %v, got: %v", val, storedVal)
	}
	closer.Close()

	return nil
}

// deleteHelper is a utility function supporting TestDelete.
// It writes a key-value pair to the database, deletes it, then checks that the key
// is deleted.
func deleteHelper(pDB *PebbleDB, writeOpts *pebble.WriteOptions) error {
	var (
		key = []byte{'a'}
		val = []byte{0x01}
	)
	if err := pDB.db.Set(key, val, nil); err != nil {
		return fmt.Errorf("writing to test DB: %s", err)
	}

	if err := pDB.deleteWithOpts(key, writeOpts); err != nil {
		return fmt.Errorf("unsynced delete: unexpected error: %s", err)
	}

	// check key is deleted
	_, _, err := pDB.db.Get(key)
	if !errors.Is(err, pebble.ErrNotFound) {
		return fmt.Errorf("want error: %s\nbut got: %s", pebble.ErrNotFound, err)
	}

	return nil
}

// batchWriteHelper is a utility function supporting TestBatchWrite.
// It creates a batch with three sets and one delete operation, commits it, then
// reads the data back.
func batchWriteHelper(pBatch *pebbleDBBatch, writeOpts *pebble.WriteOptions) error {
	var (
		keys = [][]byte{{'a'}, {'b'}, {'c'}}
		vals = [][]byte{{0x01}, {0x02}, {0x03}}
	)
	for i, key := range keys {
		val := vals[i]

		// the nil parameter is for the write options, but pebble's own library sets
		// it to _ in the function definition, thus ignoring it.
		if err := pBatch.batch.Set(key, val, nil); err != nil {
			formatStr := "adding set (k,v)=(%s,%v) operation to batch: %s"
			return fmt.Errorf(formatStr, key, val, err)
		}
	}

	// add a Delete for good measure.
	if err := pBatch.batch.Delete(keys[0], nil); err != nil {
		formatStr := "adding delete (k)=(%s) operation to batch: %s"
		return fmt.Errorf(formatStr, keys[0], err)
	}

	if err := pBatch.commitWithOpts(writeOpts); err != nil {
		return fmt.Errorf("unexpected error: %s", err)
	}

	// check keys[0] is deleted
	_, _, err := pBatch.db.db.Get(keys[0])
	if !errors.Is(err, pebble.ErrNotFound) {
		return fmt.Errorf("want error: %s\nbut got: %s", pebble.ErrNotFound, err)
	}

	// we deleted keys[0], so we don't look for it
	for i, key := range keys[1:] {
		storedVal, closer, err := pBatch.db.db.Get(key)
		if err != nil {
			return fmt.Errorf("querying key %s: %s", key, err)
		}

		val := vals[i+1] // skip vals[0], i.e., the deleted value
		if !bytes.Equal(val, storedVal) {
			return fmt.Errorf("key %s: want val %v, but got %v", key, val, storedVal)
		}

		closer.Close()
	}

	return nil
}
