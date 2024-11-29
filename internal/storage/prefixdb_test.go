package storage

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestPrefixDBGet(t *testing.T) {
	pebbleDB, dbCloser, err := newTestEmptyDB()
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
	pebbleDB, dbcloser, err := newTestEmptyDB()
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
	pebbleDB, dbCloser, err := newTestEmptyDB()
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
	pebbleDB, dbCloser, err := newTestEmptyDB()
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
