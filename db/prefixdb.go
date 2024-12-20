package db

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
)

// prefixDB provides a logical database by wrapping a namespace of another database.
// It allows for the creation of multiple logical databases on top of a single
// physical database by automatically prefixing keys with a specified namespace.
//
// It wraps a [DB] interface, which represents the underlying database, and
// delegates operations to the wrapped database after prepending the prefix to the
// key. That is, prefixDB ensures that all keys written to or read from the
// underlying database are scoped to the provided prefix.
//
// Concurrent access to a prefixDB is safe as long as the underlying [DB]
// implementation supports concurrent operations.
//
// Example usage:
//
//	db := newPrefixDB(baseDB, []byte("namespace:"))
//	err := db.Set([]byte("key"), []byte("value"))
//	if err != nil {
//		// handle error
//	}
//
// In this example, the key "key" will be stored in 'baseDB' with the actual key
// being "namespace:key".
// The responsibility for closing the underlying [DB] is on the code that created
// the prefixDB instance after it is no longer needed.
type prefixDB struct {
	db     DB
	prefix []byte
}

var _ DB = (*prefixDB)(nil)

// newPrefixDB returns a new PrefixDB wrapping the given database and scoping it to
// the given prefix.
func newPrefixDB(db DB, prefix []byte) (*prefixDB, error) {
	if len(prefix) == 0 {
		return nil, errors.New("trying to create a prefixed DB namespace with an empty prefix")
	}

	pDB := &prefixDB{
		prefix: prefix,
		db:     db,
	}

	return pDB, nil
}

// Get fetches the value of the given key, or nil if it does not exist.
// It is safe to modify the contents of key and of the returned slice after Get
// returns.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeyEmpty
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	value, err := pDB.db.Get(prefixedKey)
	if err != nil {
		return nil, fmt.Errorf("prefixed DB namespace get: %w", err)
	}
	return value, nil
}

// Has returns true if the key exists in the database.
// It is safe to modify the contents of key after Has returns.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, ErrKeyEmpty
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	ok, err := pDB.db.Has(prefixedKey)
	if err != nil {
		return ok, fmt.Errorf("prefixed DB namespace key lookup: %w", err)
	}

	return ok, nil
}

// Set sets the value for the given key, overwriting it if it already exists.
// It is safe to modify the contents of the arguments after Set returns.
//
// Set does not synchronize the write to disk immediately. Instead, it may be
// cached in memory and synced to disk later during a background flush or
// compaction. Use [SetSync] to flush the write to disk immediately.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueNil
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	if err := pDB.db.Set(prefixedKey, value); err != nil {
		return fmt.Errorf("prefixed DB namespace unsynced write: %w", err)
	}

	return nil
}

// SetSync sets the value for the given key, overwriting it if it already exists.
// It is safe to modify the contents of the arguments after Set returns.
//
// SetSync flushes the write to disk immediately and the write operation is completed
// only after the data has been successfully written to persistent storage.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) SetSync(key []byte, value []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueNil
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	if err := pDB.db.SetSync(prefixedKey, value); err != nil {
		return fmt.Errorf("prefixed DB namespace synced write: %w", err)
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
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	if err := pDB.db.Delete(prefixedKey); err != nil {
		return fmt.Errorf("prefixed DB namespace unsynced delete: %w", err)
	}

	return nil
}

// DeleteSync deletes the value for the given key. Deletes will succeed even if the
// key does not exist in the database.
// It is safe to modify the contents of key after Delete returns.
//
// DeleteSync flushes the delete to disk immediately and the delete operation is
// completed only after it synced with persistent storage.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) DeleteSync(key []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}

	prefixedKey := prependPrefix(pDB.prefix, key)
	if err := pDB.db.DeleteSync(prefixedKey); err != nil {
		return fmt.Errorf("prefixed DB namespace synced delete: %w", err)
	}

	return nil
}

// Iterator returns an iterator over a domain of keys, in ascending order.
// The caller must call [Close] when done. End is exclusive, and start must be
// less than end. A nil start iterates from the first key, and a nil end
// iterates to the last key (inclusive). Empty keys are not valid.
// No writes may happen within a domain while an iterator exists over it.
//
// Do not modify the contents of the arguments while the returned iterator is in use.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Iterator(start, end []byte) (Iterator, error) {
	itStart, itEnd, err := prefixedIteratorBounds(pDB.prefix, start, end)
	if err != nil {
		return nil, fmt.Errorf("prefixed DB namespace reverse iterator: %w", err)
	}

	it, err := pDB.db.Iterator(itStart, itEnd)
	if err != nil {
		return nil, fmt.Errorf("prefixed DB namespace iterator: %w", err)
	}

	return newPrefixDBIterator(pDB.prefix, start, end, it)
}

// ReverseIterator returns an iterator over a domain of keys, in descending
// order. The caller must call Close when done. End is exclusive, and start must
// be less than end. A nil end iterates from the last key (inclusive), and a nil
// start iterates to the first key (inclusive). Empty keys are not valid.
// No writes may happen within a domain of keys while an iterator exists over it.
//
// Do not modify the contents of the arguments while the returned iterator is in use.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) ReverseIterator(start, end []byte) (Iterator, error) {
	itStart, itEnd, err := prefixedIteratorBounds(pDB.prefix, start, end)
	if err != nil {
		return nil, fmt.Errorf("prefixed DB namespace reverse iterator: %w", err)
	}

	it, err := pDB.db.ReverseIterator(itStart, itEnd)
	if err != nil {
		return nil, fmt.Errorf("prefixed DB namespace reverse iterator: %w", err)
	}

	return newPrefixDBIterator(pDB.prefix, start, end, it)
}

// NewBatch creates a batch for atomic database updates.
// The caller is responsible for calling Batch.Close() once done.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) NewBatch() Batch {
	return newPrefixDBBatch(pDB.prefix, pDB.db.NewBatch())
}

// Compact compacts the specified range of keys in the database.
// Note that, as with all operations of a [prefixDB], the start and end keys are
// prefixed with the prefix given to [newPrefixDB].
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Compact(start, end []byte) error {
	itStart, itEnd, err := prefixedIteratorBounds(pDB.prefix, start, end)
	if err != nil {
		return fmt.Errorf("prefixed DB namespace compaction: %w", err)
	}

	if err := pDB.db.Compact(itStart, itEnd); err != nil {
		return fmt.Errorf("prefixed DB namespace compaction: %w", err)
	}
	return nil
}

// Close does nothing for PrefixDB because it's a logical wrapper around
// an underlying database. If multiple PrefixDB instances share the same
// underlying database, calling Close on one PrefixDB will close the
// underlying database for all of them, which may be unexpected.
// Therefore, the responsibility for closing the underlying database is on the code
// that created it, after all PrefixDB instances are no longer needed.
//
// It implements the [DB] interface for type PrefixDB.
func (*prefixDB) Close() error {
	return nil
}

// Print prints all the key/value pairs in the database for debugging purposes.
//
// It implements the [DB] interface for type PrefixDB.
func (pDB *prefixDB) Print() error {
	fmt.Printf("prefix: %X\n", pDB.prefix)

	itr, err := pDB.Iterator(nil, nil)
	if err != nil {
		const format = "creating a prefixed DB namespace iterator for debug printing: %w"
		return fmt.Errorf(format, err)
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}

	return nil
}

// Stats implements the [DB] interface for type PebbleDB.
func (pDB *prefixDB) Stats() map[string]string {
	const (
		prefixStrKey = "prefixdb.prefix.string"
		prefixHexKey = "prefixdb.prefix.hex"
	)
	var (
		source = pDB.db.Stats()
		stats  = make(map[string]string, len(source)+2)
	)
	stats[prefixStrKey] = string(pDB.prefix)
	stats[prefixHexKey] = hex.EncodeToString(pDB.prefix)

	const prefixSrcKey = "prefixdb.source."
	for key, value := range source {
		stats[prefixSrcKey+key] = value
	}

	return stats
}

// prefixDBBatch is a sequence of database operations that are applied atomically.
// A batch is not safe for concurrent use; callers should use a batch per goroutine
// or provide their own synchronization methods.
//
// It wraps a [Batch] interface, which represents a batch to be applied to an
// underlying database, and adds operations to the wrapped batch after prepending
// the prefix to the key. That is, prefixDBBatch ensures that all keys written to or
// read from the underlying database are scoped to the provided prefix.
//
// It implements the [Batch] interface.
type prefixDBBatch struct {
	prefix []byte
	source Batch
}

var _ Batch = (*prefixDBBatch)(nil)

// newPrefixDBBatch returns a new prefixDBBatch wrapping the given [Batch] and
// scoping it to the given prefix.
// Use a batch for atomic database updates.
func newPrefixDBBatch(prefix []byte, source Batch) *prefixDBBatch {
	return &prefixDBBatch{
		prefix: prefix,
		source: source,
	}
}

// Set adds a set update to the batch that sets the key to map to the value.
// It is safe to modify the contents of the arguments after Set returns.
//
// It implements the [Batch] interface for type prefixDBBatch.
func (b *prefixDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}
	if value == nil {
		return ErrValueNil
	}

	prefixedKey := prependPrefix(b.prefix, key)
	if err := b.source.Set(prefixedKey, value); err != nil {
		return fmt.Errorf("prefixed DB namespace batch set: %w", err)
	}

	return nil
}

// Delete adds a delete update to the batch that deletes database the entry for
// key. It is safe to modify the contents of the arguments after Delete returns.
//
// It implements the [Batch] interface for type prefixDBBatch.
func (b *prefixDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}

	prefixedKey := prependPrefix(b.prefix, key)
	if err := b.source.Delete(prefixedKey); err != nil {
		return fmt.Errorf("prefixed DB namespace batch delete: %w", err)
	}

	return nil
}

// Write applies the batch to the database. Write does not guarantees that the batch
// is persisted to disk before returning.
//
// It implements the [Batch] interface for type prefixDBBatch.
func (b *prefixDBBatch) Write() error {
	if err := b.source.Write(); err != nil {
		return fmt.Errorf("prefixed DB namespace batch write: %w", err)
	}
	return nil
}

// WriteSync applies the batch to the database. WriteSync guarantees that the batch
// is persisted to disk before returning.
//
// It implements the [Batch] interface for type prefixDBBatch.
func (b *prefixDBBatch) WriteSync() error {
	if err := b.source.WriteSync(); err != nil {
		return fmt.Errorf("prefixed DB namespace batch write: %w", err)
	}
	return nil
}

// Close closes the batch without committing it. Close is idempotent, but calling
// other methods on the batch after closing it will return an error.
//
// It implements the [Batch] interface for type prefixDBBatch.
func (b *prefixDBBatch) Close() error {
	if err := b.source.Close(); err != nil {
		return fmt.Errorf("prefixed DB namespace batch close: %w", err)
	}
	return nil
}

// prefixDBIterator wraps an [Iterator] interface representing an iterator iterating
// over a range of key/value pairs in an underlying database. It will range over
// keys that begin with the given prefix. That is, prefixDBIterator ensures that all
// keys that it iterates over are scoped to the provided prefix.
//
// It is the caller's responsibility to call [Close] on it when done.
// prefixDBIterator is not safe for concurrent use, but it is safe to use multiple
// iterators concurrently. Callers cannot write to the underlying database
// while there exists an iterator over it.
// If there is an error during any operation, it is stored in the Iterator and can
// be retrieved via the [Error] method.
//
// It implements the [Iterator] interface.
type prefixDBIterator struct {
	source Iterator
	prefix []byte
	start  []byte
	end    []byte
	valid  bool
	err    error
}

// compile-time check: does *prefixDBIterator satisfy the Iterator interface?
var _ Iterator = (*prefixDBIterator)(nil)

// newPrefixDBIterator returns a new prefixDBIterator to iterate over a range of
// database key/value pairs that begin with the given prefix.
func newPrefixDBIterator(
	prefix, start, end []byte,
	source Iterator,
) (*prefixDBIterator, error) { //nolint:unparam
	// NOTE: Do not delete itInvalid and return an error below.
	// The current design allows the creation of an Iterator with the key range
	// [nil:nil], which we interpret as "from the first to the last key in the
	// database." However, there is a corner case: creating an Iterator with the
	// range [nil:some_key] when there is only one key in the database (or only one
	// key with the given prefix, i.e., prefix+some_key). Because the iterator's
	// upper bound is exclusive, this results in an empty key range. There might be
	// other corner cases that I haven't found, though.
	//
	// In the case of Pebble (and possibly other databases), such a scenario leads
	// to an invalid iterator state. According to Pebble's documentation, this
	// situation produces an iterator that "has a non-exhausted internalIterator,
	// but has reached a limit without any key for the caller." In other words, a
	// call to Valid() will fail, and therefore the creation of the Iterator itself
	// will fail.
	//
	// I think cometbft’s code depends on newPrefixDBIterator returning an invalid
	// iterator rather than an error. This design is not ideal, but changing it
	// would likely require revisiting a large portion of cometbft’s codebase.
	itInvalid := &prefixDBIterator{
		prefix: prefix,
		start:  start,
		end:    end,
		source: source,
		valid:  false,
	}

	// Empty keys are not allowed, so if a key exists in the database that exactly
	// matches the prefix, we need to skip it.
	if source.Valid() && bytes.Equal(source.Key(), prefix) {
		// The key is going to be lexicograpically smaller than the first
		// "correct" key of the form prefix+key, because
		// prefix < prefix+key, e.g., "foo" < "fooa".
		// Therefore, we only need to skip it to position the iterator at the first
		// "correct" key.
		source.Next()
	}

	if !source.Valid() || !bytes.HasPrefix(source.Key(), prefix) {
		return itInvalid, nil
	}

	it := &prefixDBIterator{
		prefix: prefix,
		start:  start,
		end:    end,
		source: source,
		valid:  true,
	}

	return it, nil
}

// Close closes the iterator, releasing any allocated resources.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Close() error {
	if err := it.source.Close(); err != nil {
		return fmt.Errorf("closing prefixed DB namespace iterator: %w", err)
	}

	return nil
}

// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
// Callers must not modify the returned slices. Instead, they should make a copy if
// they need to modify them.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Domain() (start []byte, end []byte) {
	return it.start, it.end
}

// Error returns the last error encountered by the iterator, if any.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Error() error {
	if err := it.source.Error(); err != nil {
		return err
	}
	return it.err
}

// Key returns the key, stripped of the prefix, at the current position or panics if
// the iterator is invalid.
// The caller should not modify the contents of the returned slice and its contents
// may change on the next call to [Next]. Therefore, callers should make a copy of
// the returned slice if they need to modify it.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Key() []byte {
	it.assertIsValid()
	key := it.source.Key()

	// we have checked that the key is valid in the call to assertIsValid()
	return key[len(it.prefix):]
}

// Next moves the iterator to the next key in the database, as defined by order
// of iteration, or panics if the iterator is invalid.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Next() {
	it.assertIsValid()
	it.source.Next()

	srcItInvalid := !it.source.Valid()
	if srcItInvalid || !bytes.HasPrefix(it.source.Key(), it.prefix) {
		it.valid = false
		return
	}

	if bytes.Equal(it.source.Key(), it.prefix) {
		// Empty keys are not allowed, so if a key exists in the database that
		// exactly matches the prefix we need to skip it.
		// The key is going to be lexicograpically smaller than the first "correct"
		// key of the form prefix+key, because
		// prefix < prefix+key, e.g., "foo" < "fooa".
		// Therefore, we only need to skip it to position the iterator at the first
		// "correct" key.
		it.Next()
	}
}

// Valid returns whether the current iterator is valid. Once invalid, the Iterator
// remains invalid forever.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Valid() bool {
	if !it.valid || it.err != nil || !it.source.Valid() {
		return false
	}

	var (
		key         = it.source.Key()
		prefixLen   = len(it.prefix)
		keyTooShort = len(key) < prefixLen
	)
	if keyTooShort || !bytes.Equal(key[:prefixLen], it.prefix) {
		const format = "received invalid key from backend: %X (expected prefix %X)"
		it.err = fmt.Errorf(format, key, it.prefix)

		return false
	}

	return true
}

// Value returns the value of the current key/value pair or panics if the iterator
// is invalid.
// The caller should not modify the contents of the returned slice, and its contents
// may change on the next call to [Next]. Therefore, callers should make a copy of
// the returned slice if they need to modify it.
//
// It implements the [Iterator] interface for type prefixDBIterator.
func (it *prefixDBIterator) Value() []byte {
	it.assertIsValid()
	return it.source.Value()
}

// assertIsValid panics if the iterator is invalid.
func (it *prefixDBIterator) assertIsValid() {
	if !it.Valid() {
		panic("prefixed DB namespace iterator is invalid")
	}
}

// prependPrefix concatenates a prefix with a key.
func prependPrefix(prefix, key []byte) []byte {
	prefixed := make([]byte, len(prefix)+len(key))
	copy(prefixed, prefix)
	copy(prefixed[len(prefix):], key)

	return prefixed
}

// incrementBigEndian treats the input slice as a big-endian unsigned integer.
// It creates a new slice of the same length, increments the value by one,
// and returns the result.
// If the input slice represents the maximum value for its length (all bytes are
// 0xFF), incrementBigEndian returns nil to indicate overflow.
// The input slice s remains unmodified.
func incrementBigEndian(s []byte) []byte {
	if len(s) == 0 {
		panic("incrementBigEndian called with empty slice")
	}

	result := make([]byte, len(s))
	copy(result, s)

	for i := len(result) - 1; i >= 0; i-- {
		if result[i] < 0xFF {
			result[i]++
			return result
		}
		result[i] = 0x00 // Carry over to the next byte
	}

	// Overflow if the loop finishes without returning
	return nil
}

// prefixedIteratorBounds takes the lower and upper bounds for an iterator and
// prepends the prefix. The bounds are ready for use in a prefixDBIterator.
func prefixedIteratorBounds(prefix, start, end []byte) ([]byte, []byte, error) {
	if start != nil && len(start) == 0 {
		return nil, nil, fmt.Errorf("lower bound: %w", ErrKeyEmpty)
	}
	if end != nil && len(end) == 0 {
		return nil, nil, fmt.Errorf("upper bound: %w", ErrKeyEmpty)
	}

	var (
		itStart = prependPrefix(prefix, start)
		itEnd   []byte
	)
	if end == nil {
		itEnd = incrementBigEndian(prefix)
	} else {
		itEnd = prependPrefix(prefix, end)
	}

	return itStart, itEnd, nil
}
