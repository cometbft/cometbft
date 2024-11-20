package storage

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

func TestGet(t *testing.T) {
	pDB, err := newInMemDB()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pDB.db.Close()
	})

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
			t.Errorf("expected nil value, got: %v", val)
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
	pDB, err := newInMemDB()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pDB.db.Close()
	})

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
	pDB, err := newInMemDB()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pDB.db.Close()
	})

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
		// By changing the value, we also tests overwriting.
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
	pDB, err := newInMemDB()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pDB.db.Close()
	})

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

			pDB, err := newInMemDB()
			if err != nil {
				t.Fatal(err)
			}

			closer := func() {
				pDB.db.Close()
			}

			for i, key := range keys {
				if err := pDB.db.Set(key, vals[i], sync); err != nil {
					t.Fatalf("writing key %s: %s", key, err)
				}
			}
			return pDB, closer
		}
	)

	// The following tests will create their own DBs to test compaction, so that
	// each compaction operation works on a DB that has never been compacted
	// before.
	t.Run("NilStartNoErr", func(t *testing.T) {
		pDB, closer := createTestDB(t)
		defer closer()

		// if start is nil, compaction starts from the first key in the DB.
		end := keys[2]
		if err := pDB.Compact(nil, end); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("NilEndNoErr", func(t *testing.T) {
		pDB, closer := createTestDB(t)
		defer closer()

		// if end is nil, compaction ends at the last key in the DB.
		start := keys[0]
		if err := pDB.Compact(start, nil); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("StartEndNilNoErr", func(t *testing.T) {
		pDB, closer := createTestDB(t)
		defer closer()

		// if start and end are nil, compaction starts from the first key and ends
		// at the last key in the DB.
		if err := pDB.Compact(nil, nil); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("StartEndNoErr", func(t *testing.T) {
		pDB, closer := createTestDB(t)
		defer closer()

		var (
			start = keys[2]
			end   = keys[8]
		)
		if err := pDB.Compact(start, end); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
}

// newInMemDB is a utility function that creates an in-memory instance of pebble for
// testing.
func newInMemDB() (*PebbleDB, error) {
	opts := &pebble.Options{FS: vfs.NewMem()}
	memDB, err := pebble.Open("", opts)
	if err != nil {
		return nil, fmt.Errorf("creating test DB: %w", err)
	}

	pDB := &PebbleDB{db: memDB}

	return pDB, nil
}

// setelper is a utility function supporting TestSet.
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
		return fmt.Errorf("expected value: %s, got: %s", val, storedVal)
	}

	// better to check if it's nil before calling Close().
	// If the call to Get unexpectedly fails, closer will be nil, therefore calling
	// Close() on it will panic. If the call to Get succeeds as we expect, we are
	// good citizens and call Close().
	if closer != nil {
		closer.Close()
	}

	return nil
}

// deleteHelper is a utility function supporting TestDelete.
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
	val, closer, err := pDB.db.Get(key)
	if err == nil {
		return errors.New("expected error but got nil")
	} else if !errors.Is(err, pebble.ErrNotFound) {
		return fmt.Errorf("unexpected error: %s", err)
	}
	if val != nil {
		return fmt.Errorf("expected nil value, got: %s", val)
	}

	// better to check if it's nil before calling Close().
	// If the call to Get unexpectedly succeeds, we are good citizens and call
	// Close(). If the call to Get fails as we expect, closer will be nil,
	// therefore calling Close() on it will panic.
	if closer != nil {
		closer.Close()
	}

	return nil
}
