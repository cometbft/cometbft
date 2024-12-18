package cmtdb

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestPrefixDBGet(t *testing.T) {
	pDB, err := newTestPrefixDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pDB.Close() })

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := pDB.Get(nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
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
			key         = []byte{'a'}
			prefixedKey = append(pDB.prefix, key...)
			val         = []byte{'b'}
		)
		// we are calling PebbleDB's [SetSync] directly, therefore we must prepend
		// the prefix to the key ourselves.
		if err := pDB.db.SetSync(prefixedKey, val); err != nil {
			t.Fatalf("writing to test DB: %s", err)
		}

		gotVal, err := pDB.Get(key)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !bytes.Equal(gotVal, val) {
			t.Errorf("expected value: %s, got: %v", val, gotVal)
		}
	})
}

func TestPrefixDBHas(t *testing.T) {
	pDB, err := newTestPrefixDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pDB.Close() })

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if _, err := pDB.Has(nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
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
			key         = []byte{'a'}
			prefixedKey = append(pDB.prefix, key...)
			val         = []byte{'b'}
		)
		// we are calling PebbleDB's [SetSync] directly, therefore we must prepend
		// the prefix to the key ourselves.
		if err := pDB.db.SetSync(prefixedKey, val); err != nil {
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

func TestPrefixDBSet(t *testing.T) {
	pDB, err := newTestPrefixDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pDB.Close() })

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pDB.Set(nil, nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
		}
	})

	t.Run("NilValueErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := pDB.Set(key, nil); !errors.Is(err, ErrValueNil) {
			t.Errorf("expected %s, got: %s", ErrValueNil, err)
		}
	})

	t.Run("NoErr", func(t *testing.T) {
		var (
			keys = [][]byte{{'a'}, {'b'}}
			vals = [][]byte{{0x01}, {0x02}}
		)
		if err := pDB.Set(keys[0], vals[0]); err != nil {
			t.Fatalf("unsynced Set unexpected error: %s", err)
		}

		if err := pDB.SetSync(keys[1], vals[1]); err != nil {
			t.Fatalf("synced Set unexpected error: %s", err)
		}

		for i, key := range keys {
			// we are calling PebbleDB's [Get] directly, therefore we must u
			// prepend the prefix to the key ourselves.
			prefixedKey := append(pDB.prefix, key...)
			storedVal, err := pDB.db.Get(prefixedKey)
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
	pDB, err := newTestPrefixDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pDB.Close() })

	t.Run("EmptyKeyErr", func(t *testing.T) {
		if err := pDB.Delete(nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
		}
	})

	t.Run("KeyNotExistNoErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := pDB.Delete(key); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("KeyExistNoErr", func(t *testing.T) {
		if err := deletePrefixDBHelper(pDB, false); err != nil {
			t.Errorf("unsynced Delete unexpected error: %s", err)
		}

		if err := deletePrefixDBHelper(pDB, true); err != nil {
			t.Errorf("synced Delete unexpected error: %s", err)
		}
	})
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
		if err := prefixBatch.Set(nil, nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
		}
	})

	t.Run("ValueNilErr", func(t *testing.T) {
		key := []byte{'a'}
		if err := prefixBatch.Set(key, nil); !errors.Is(err, ErrValueNil) {
			t.Errorf("expected %s, got: %s", ErrValueNil, err)
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
		if err := prefixBatch.Set(key, value); !errors.Is(err, ErrBatchClosed) {
			t.Errorf("expected %s, got: %s", ErrBatchClosed, err)
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
		if err := prefixBatch.Delete(nil); !errors.Is(err, ErrKeyEmpty) {
			t.Errorf("expected %s, got: %s", ErrKeyEmpty, err)
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
		if err := prefixBatch.Delete(key); !errors.Is(err, ErrBatchClosed) {
			t.Errorf("expected %s, got: %s", ErrBatchClosed, err)
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
			prefixDB = &prefixDB{
				db:     pebbleBatch.db,
				prefix: prefix,
			}
			prefixBatch = &prefixDBBatch{
				source: pebbleBatch,
				prefix: prefix,
			}
		)
		err = batchWriteTestHelper(prefixBatch, prefixDB, false)
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
			prefixDB = &prefixDB{
				db:     pebbleBatch.db,
				prefix: prefix,
			}
			prefixBatch = &prefixDBBatch{
				source: pebbleBatch,
				prefix: prefix,
			}
		)
		err = batchWriteTestHelper(prefixBatch, prefixDB, true)
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

		if err := prefixBatch.Write(); !errors.Is(err, ErrBatchClosed) {
			t.Errorf("expected %s, got: %s", ErrBatchClosed, err)
		}
	})
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
			wantErr: ErrKeyEmpty,
		},
		{ // empty end
			start:   start,
			end:     []byte{},
			wantErr: ErrKeyEmpty,
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
	t.Run("NoErr", func(t *testing.T) {
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
	})

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("incrementBigEndian did not panic")
			}
		}()

		incrementBigEndian([]byte{})
	})
}

// newTestPrefixDB creates an instance of a PrefixDB for testing.
// Under the hood, it wraps an in-memory instance of PebbleDB and scopes its keys
// with the prefix "test".
func newTestPrefixDB() (*prefixDB, error) {
	pebbleDB, _, err := newTestPebbleDB()
	if err != nil {
		return nil, fmt.Errorf("creating prefix-wrapped DB: %w", err)
	}

	pDB := &prefixDB{
		db:     pebbleDB,
		prefix: []byte{'t', 'e', 's', 't'},
	}

	return pDB, nil
}

// deletePrefixDBHelper is a utility function supporting TestPrefixDBDelete.
// It writes a key-value pair to the database, deletes it, then checks that the key
// is deleted.
func deletePrefixDBHelper(pDB *prefixDB, synced bool) error {
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
