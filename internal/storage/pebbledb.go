package storage

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

// PebbleDB is a PebbleDB backend.
// It implements the [DB] interface.
type PebbleDB struct {
	db *pebble.DB
}

var _ DB = (*PebbleDB)(nil)

// newPebbleDB returns a new PebbleDB instance using the default options.
func newPebbleDB(name, dir string) (*PebbleDB, error) {
	opts := &pebble.Options{}

	return newPebbleDBWithOpts(name, dir, opts)
}

// newPebbleDBWithOpts returns a new PebbleDB instance using the provided options.
func newPebbleDBWithOpts(name, dir string, opts *pebble.Options) (*PebbleDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	opts.EnsureDefaults()

	db, err := pebble.Open(dbPath, opts)
	if err != nil {
		return nil, fmt.Errorf("opening pebble instance %q: %w", name, err)
	}

	pebbleDB := &PebbleDB{db: db}

	return pebbleDB, nil
}

// DB returns the underlying PebbleDB instance.
func (pDB *PebbleDB) DB() *pebble.DB {
	return pDB.db
}

// Get fetches the value of the given key, or nil if it does not exist.
// It is safe to modify the contents of key and of the returned slice after Get
// returns.
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errKeyEmpty
	}

	value, closer, err := pDB.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("fetching value for key %s: %w", key, err)
	}
	defer closer.Close()

	valueCp := make([]byte, len(value))
	copy(valueCp, value)

	return valueCp, nil
}

// Has returns true if the key exists in the database.
// It is safe to modify the contents of key after Has returns.
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, errKeyEmpty
	}

	bytesPeb, err := pDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("checking if key %s exists: %w", key, err)
	}

	return bytesPeb != nil, nil
}

// Set sets the value for the given key, overwriting it if it already exists.
// It is safe to modify the contents of the arguments after Set returns.
//
// Set does not synchronize the write to disk immediately. Instead, it may be
// cached in memory and synced to disk later during a background flush or
// compaction. Use [SetSync] to flush the write to disk immediately.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Set(key, value []byte) error {
	writeOpts := pebble.NoSync
	if err := pDB.setWithOpts(key, value, writeOpts); err != nil {
		return fmt.Errorf("unsynced write: %w", err)
	}

	return nil
}

// SetSync sets the value for the given key, overwriting it if it already exists.
// It is safe to modify the contents of the arguments after Set returns.
//
// SetSync flushes the write to disk immediately and the write operation is completed
// only after the data has been successfully written to persistent storage.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) SetSync(key, value []byte) error {
	writeOpts := pebble.Sync
	if err := pDB.setWithOpts(key, value, writeOpts); err != nil {
		return fmt.Errorf("synced write: %w", err)
	}

	return nil
}

// setWithOpts sets the value for the given key, overwriting it if it already exists.
// It is safe to modify the contents of the arguments after setWithOpts returns.
func (pDB *PebbleDB) setWithOpts(
	key, value []byte,
	writeOpts *pebble.WriteOptions,
) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}

	err := pDB.db.Set(key, value, writeOpts)
	if err != nil {
		return fmt.Errorf("setting value %s\nfor key %s: %w", value, key, err)
	}

	return nil
}

// Delete deletes the value for the given key. Deletes will succeed even if the key
// does not exist in the database.
// It is safe to modify the contents of key after Delete returns.
//
// Delete does not synchronize the delete to disk immediately. Instead, it may be
// cached in memory and synced to disk later during a background flush or
// compaction. Use [DeleteSync] to flush the delete to disk immediately.
// Delete is faster than [DeleteSync] because it does not incur the latency of disk
// I/O.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Delete(key []byte) error {
	writeOpts := pebble.NoSync
	if err := pDB.deleteWithOpts(key, writeOpts); err != nil {
		return fmt.Errorf("unsynced delete: %w", err)
	}

	return nil
}

// DeleteSync deletes the value for the given key. Deletes will succeed even if the
// key does not exist in the database.
// It is safe to modify the contents of key after Delete returns.
//
// DeleteSync flushes the delete to disk immediately and the delete operation is
// completed only after it synced with persistent storage.
// Because it incurs the latency of disk I/O, it is slower than [Delete].
//
// It implements the [DB] interface for type PebbleDB.
func (pDB PebbleDB) DeleteSync(key []byte) error {
	writeOpts := pebble.Sync
	if err := pDB.deleteWithOpts(key, writeOpts); err != nil {
		return fmt.Errorf("synced delete: %w", err)
	}

	return nil
}

// deleteWithOpts deletes the value for the given key. Deletes will succeed even if
// the key does not exist in the database.
// It is safe to modify the contents of the arguments after deleteWithOpts returns.
func (pDB *PebbleDB) deleteWithOpts(
	key []byte,
	writeOpts *pebble.WriteOptions,
) error {
	if len(key) == 0 {
		return errKeyEmpty
	}

	if err := pDB.db.Delete(key, writeOpts); err != nil {
		return fmt.Errorf("deleting key %s: %w", key, err)
	}

	return nil
}

// Compact compacts the specified range of keys in the database.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Compact(start, end []byte) error {
	// Currently nil,nil is an invalid range in Pebble.
	// This was taken from https://github.com/cockroachdb/pebble/issues/1474
	// If start==end pebbleDB will throw an error.
	if start != nil && end != nil {
		if err := pDB.db.Compact(start, end, true /* parallelize */); err != nil {
			return fmt.Errorf("compacting range [%s, %s]: %w", start, end, err)
		}
	}

	iter, err := pDB.db.NewIter(nil)
	if err != nil {
		return fmt.Errorf("creating compaction iterator: %w", err)
	}

	// iter.First() moves the iterator to the first key/value pair and returns true
	// if it is pointing to a valid entry.
	if start == nil && iter.First() {
		start = append(start, iter.Key()...)
	}

	// iter.Last() moves the iterator to the last key/value pair and returns true
	// if it is pointing to a valid entry.
	if end == nil && iter.Last() {
		end = append(end, iter.Key()...)
	}

	if err := pDB.db.Compact(start, end, true /* parallelize */); err != nil {
		compactErr := fmt.Errorf("compacting range [%s, %s]: %w", start, end, err)

		if err := iter.Close(); err != nil {
			itCloseErr := fmt.Errorf("closing compaction iterator: %w", err)

			formatStr := "multiple errors during compaction:\n%w\n%w"
			return fmt.Errorf(formatStr, compactErr, itCloseErr)
		}

		return compactErr
	}

	if err := iter.Close(); err != nil {
		formatStr := "closing iterator after successful compaction: %w"
		return fmt.Errorf(formatStr, err)
	}

	return nil
}

// Close closes the database connection.
// It is not safe to close a DB until all outstanding iterators are closed
// or to call Close concurrently with any other DB method. It is not valid
// to call any of a DB's methods after the DB has been closed.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Close() error {
	if err := pDB.db.Close(); err != nil {
		return fmt.Errorf("closing database: %w", err)
	}

	return nil
}

// Print prints all the key/value pairs in the database for debugging purposes.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Print() error {
	itr, err := pDB.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("creating iterator for debug printing: %w", err)
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}

	return nil
}

// Stats implements the [DB] interface.
//
// It implements the [DB] interface for type PebbleDB.
func (*PebbleDB) Stats() map[string]string {
	return nil
}

// NewBatch creates a batch for atomic database updates.
// The caller is responsible for calling Batch.Close() once done.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) NewBatch() Batch {
	return newPebbleDBBatch(pDB)
}

// Iterator returns an iterator over a domain of keys, in ascending order.
// The caller must call [Close] when done. End is exclusive, and start must be
// less than end. A nil start iterates from the first key, and a nil end
// iterates to the last key (inclusive). Empty keys are not valid.
// No writes may happen within a domain while an iterator exists over it.
//
// It is unsafe to modify the contents of the arguments while the returned
// iterator is in use.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) Iterator(start, end []byte) (Iterator, error) {
	it, err := newPebbleDBIterator(pDB, start, end, false /* reverse */)
	if err != nil {
		return nil, fmt.Errorf("creating new forward iterator: %w", err)
	}

	return it, nil
}

// ReverseIterator returns an iterator over a domain of keys, in descending
// order. The caller must call Close when done. End is exclusive, and start must
// be less than end. A nil end iterates from the last key (inclusive), and a nil
// start iterates to the first key (inclusive). Empty keys are not valid.
// No writes may happen within a domain of keys while an iterator exists over it.
//
// It is unsafe to modify the contents of the arguments while the returned
// iterator is in use.
//
// It implements the [DB] interface for type PebbleDB.
func (pDB *PebbleDB) ReverseIterator(start, end []byte) (Iterator, error) {
	it, err := newPebbleDBIterator(pDB, start, end, true /* reverse */)
	if err != nil {
		return nil, fmt.Errorf("creating new reverse iterator: %w", err)
	}

	return it, nil
}

var _ Batch = (*pebbleDBBatch)(nil)

// pebbleDBBatch is a sequence of database operations that are applied atomically.
// A batch is not safe for concurrent use; callers should use a batch per goroutine
// or provide their own synchronization methods.
//
// It implements the [Batch] interface.
type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

// newPebbleDBBatch returns a new batch to be used for atomic database updates.
func newPebbleDBBatch(pDB *PebbleDB) *pebbleDBBatch {
	return &pebbleDBBatch{
		// Because newPebbleDBBatch can only be called by the exported method
		// PebbleDB.NewBatch, pDB is going to be non-nil; Therefore we don't need to
		// initialize the DB here.
		// This is set to enable general DB operations like compaction
		// (e.g., a call do pebbleDBBatch.db.Compact() would throw a nil pointer
		// exception)
		db:    pDB,
		batch: pDB.db.NewBatch(),
	}
}

// Set adds a set update to the batch that sets the key to map to the value.
// It is safe to modify the contents of the arguments after Set returns.
//
// It implements the [Batch] interface for type pebbleDBBatch.
func (b *pebbleDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	if b.batch == nil {
		return errBatchClosed
	}

	// the nil parameter is for the write options, but pebble's own library sets it
	// to _ in the function definition, thus ignoring it.
	if err := b.batch.Set(key, value, nil); err != nil {
		formatStr := "adding set update (k,v)=(%s,%s) to batch: %w"
		return fmt.Errorf(formatStr, key, value, err)
	}

	return nil
}

// Delete adds a delete update to the batch that deletes database the entry for
// key. It is safe to modify the contents of the arguments after Delete returns.
//
// It implements the [Batch] interface for type pebbleDBBatch.
func (b *pebbleDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if b.batch == nil {
		return errBatchClosed
	}

	// the nil parameter is for the write options. pebble's own library sets it
	// to _ in the function definition, thus ignoring it.
	if err := b.batch.Delete(key, nil); err != nil {
		formatStr := "adding delete update (k)=(%s) to batch: %w"
		return fmt.Errorf(formatStr, key, err)
	}

	return nil
}

// Write applies the batch to the database. Write does not guarantees that the batch
// is persisted to disk before returning.
//
// It implements the [Batch] interface for type pebbleDBBatch.
func (b *pebbleDBBatch) Write() error {
	writeOpts := pebble.NoSync
	if err := b.commitWithOpts(writeOpts); err != nil {
		return fmt.Errorf("unsynced batch write: %w", err)
	}

	return nil
}

// WriteSync applies the batch to the database. WriteSync guarantees that the batch
// is persisted to disk before returning.
//
// It implements the [Batch] interface for type pebbleDBBatch.
func (b *pebbleDBBatch) WriteSync() error {
	writeOpts := pebble.Sync
	if err := b.commitWithOpts(writeOpts); err != nil {
		return fmt.Errorf("synced batch write: %w", err)
	}

	return nil
}

// commitWithOpts applies the batch to the database.
func (b *pebbleDBBatch) commitWithOpts(writeOpts *pebble.WriteOptions) error {
	if b.batch == nil {
		return errBatchClosed
	}

	if err := b.batch.Commit(writeOpts); err != nil {
		return fmt.Errorf("writing batch to database: %w", err)
	}

	// Make sure batch cannot be used afterwards.
	// Callers should still call Close(), on it.
	if err := b.Close(); err != nil {
		return fmt.Errorf("batch post-write routine: %w", err)
	}

	return nil
}

// Close closes the batch without committing it. Close is idempotent, but calling
// other methods on the batch after closing it will return an error.
//
// It implements the [Batch] interface for type pebbleDBBatch.
func (b *pebbleDBBatch) Close() error {
	if b.batch == nil {
		return nil
	}

	if err := b.batch.Close(); err != nil {
		return fmt.Errorf("closing batch: %w", err)
	}

	b.batch = nil

	return nil
}

// pebbleDBIterator is an Iterator iterating over a database's key/value pairs in
// key order. It is the caller's responsibility to call [Close] on it when done.
// pebbleDBIterator is not safe for concurrent use, but it is safe to use multiple
// iterators concurrently.
// Callers cannot write to the underlying database whilethere exists an iterator
// over it.
// If there is an error during any operation, it is stored in the Iterator and can
// be retrieved via the [Error] method.
//
// It implements the [Iterator] interface.
type pebbleDBIterator struct {
	source     *pebble.Iterator
	start, end []byte // end is exclusive.
	isReverse  bool
	isInvalid  bool
}

var _ Iterator = (*pebbleDBIterator)(nil)

// newPebbleDBIterator returns a new pebbleDBIterator to iterate over a range of
// database key/value pairs of the given instance of PebbleDB.
func newPebbleDBIterator(
	pDB *PebbleDB,
	start, end []byte,
	isReverse bool,
) (*pebbleDBIterator, error) {
	if start != nil && len(start) == 0 {
		return nil, fmt.Errorf("iterator's lower bound: %w", errKeyEmpty)
	}

	if end != nil && len(end) == 0 {
		return nil, fmt.Errorf("iterator's upper bound: %w", errKeyEmpty)
	}

	o := pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	}
	it, err := pDB.db.NewIter(&o)
	if err != nil {
		formatStr := "iterator with bounds [%X, %X]: %w"
		return nil, fmt.Errorf(formatStr, start, end, err)
	}

	if isReverse {
		it.Last()
	} else {
		it.First()
	}

	pbIt := &pebbleDBIterator{
		source:    it,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}

	return pbIt, nil
}

// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
// Callers must not modify the returned slices. Instead, they should make a copy if
// they need to modify them.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Valid returns whether the current iterator is valid. Once invalid, the
// Iterator remains invalid forever.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Valid() bool {
	if itr.isInvalid {
		return false
	}

	if err := itr.source.Error(); err != nil {
		itr.isInvalid = true
		return false
	}

	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If the key of the current key/value pair is either before the start or after
	// the end, the iterator is invalid.
	var (
		start = itr.start
		end   = itr.end
		key   = itr.source.Key()

		// If 'start' is nil, the iterator's lower bound is the first key in the
		// database, and therefore no key can be before it. Therefore, the
		// 'start != nil' check is a shortcut to avoid a useless call to
		// bytes.Compare(non_nil_key, nil), which would return 1 anyway because any
		// non-empty slice is considered greater than nil.
		itrBeforeStart = start != nil && bytes.Compare(key, start) < 0

		// We check if 'end != nil' because if 'end' is nil, the iterator's upper
		// bound is the last key in the database, and therefore no key can be after
		// it. Additionally, bytes.Compare(non_nil_key, nil) returns 1 because any
		// non-empty slice is considered greater than nil. Therefore, without
		// checking 'end != nil', we would incorrectly set `itrAfterEnd` to true.
		itrAfterEnd = end != nil && bytes.Compare(key, end) >= 0
	)
	if (itr.isReverse && itrBeforeStart) || (!itr.isReverse && itrAfterEnd) {
		itr.isInvalid = true
		return false
	}

	return true
}

// Key returns the key at the current position. Panics if the iterator is invalid.
// The caller should not modify the contents of the returned slice and its contents
// may change on the next call to [Next].
// Therefore, callers should make a copy of the returned slice if they need to
// modify it.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Key() []byte {
	itr.assertIsValid()
	return itr.source.Key()
}

// Value returns the value of the current key/value pair. It panics if the
// iterator is invalid.
// The caller should not modify the contents of the returned slice, and
// its contents may change on the next call to Next.
// Therefore, callers should make a copy of the returned slice if they need to
// modify it.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Value() []byte {
	itr.assertIsValid()
	return itr.source.Value()
}

// Next moves the iterator to the next key in the database, as defined by order
// of iteration. It panics if the iterator is invalid.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr pebbleDBIterator) Next() {
	itr.assertIsValid()

	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Error returns the last error encountered by the iterator, if any.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Error() error {
	return itr.source.Error()
}

// Close closes the iterator, releasing any allocated resources.
//
// It implements the [Iterator] interface for type pebbleDBIterator.
func (itr *pebbleDBIterator) Close() error {
	if err := itr.source.Close(); err != nil {
		return fmt.Errorf("closing iterator: %w", err)
	}

	return nil
}

// assertIsValid panics if the iterator is invalid.
func (itr *pebbleDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
