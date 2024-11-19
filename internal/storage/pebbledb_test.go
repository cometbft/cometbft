package storage

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cockroachdb/pebble"
)

const _testDB = "test_db"

func TestGet(t *testing.T) {
	dbDirPath := t.TempDir()

	pDB, err := NewPebbleDB(_testDB, dbDirPath)
	if err != nil {
		t.Fatalf("creating test DB: %s", err)
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
	dbDirPath := t.TempDir()

	pDB, err := NewPebbleDB(_testDB, dbDirPath)
	if err != nil {
		t.Fatalf("creating test DB: %s", err)
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
	dbDirPath := t.TempDir()

	pDB, err := NewPebbleDB(_testDB, dbDirPath)
	if err != nil {
		t.Fatalf("creating test DB: %s", err)
	}

	t.Cleanup(func() {
		pDB.db.Close()
	})

	var (
		sync   = pebble.Sync
		noSync = pebble.NoSync
	)
	t.Run("EmptyKeyErr", func(t *testing.T) {
		// called by Set
		if err := pDB.setWithOpts(nil, nil, noSync); !errors.Is(err, errKeyEmpty) {
			t.Errorf("non-sync write: expected %s, got: %s", errKeyEmpty, err)
		}

		// called by SetSync
		if err := pDB.setWithOpts(nil, nil, sync); !errors.Is(err, errKeyEmpty) {
			t.Errorf("sync write: expected %s, got: %s", errKeyEmpty, err)
		}
	})

	t.Run("NilValueErr", func(t *testing.T) {
		key := []byte{'a'}

		// called by Set
		if err := pDB.setWithOpts(key, nil, noSync); !errors.Is(err, errValueNil) {
			t.Errorf("non-sync write: expected %s, got: %s", errValueNil, err)
		}

		// called by SetSync
		if err := pDB.setWithOpts(key, nil, sync); !errors.Is(err, errValueNil) {
			t.Errorf("sync write: expected %s, got: %s", errValueNil, err)
		}
	})

	t.Run("NoErr", func(t *testing.T) {
		// called by Set
		var (
			key = []byte{'a'}
			val = []byte{0x01}
		)
		if err := pDB.setWithOpts(key, val, noSync); err != nil {
			t.Fatalf("non-sync write: unexpected error: %s", err)
		}

		storedVal, closer, err := pDB.db.Get(key)
		if err != nil {
			t.Fatalf("reading from test DB: %s", err)
		}
		if !bytes.Equal(storedVal, val) {
			t.Fatalf("expected value: %s, got: %s", val, storedVal)
		}
		closer.Close()

		// called by SetSync
		// By changing the value, we also tests overwriting.
		val = []byte{0x02}
		if err := pDB.setWithOpts(key, val, sync); err != nil {
			t.Fatalf("sync write: unexpected error: %s", err)
		}

		storedVal, closer, err = pDB.db.Get(key)
		if err != nil {
			t.Fatalf("reading from test DB: %s", err)
		}
		if !bytes.Equal(storedVal, val) {
			t.Fatalf("expected value: %s, got: %s", val, storedVal)
		}
		closer.Close()
	})
}
