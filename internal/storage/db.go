package storage

import "errors"

var (
	// errBatchClosed is returned when a closed or written batch is used.
	errBatchClosed = errors.New("batch has been written or closed")

	// errKeyEmpty is returned when attempting to use an empty or nil key.
	errKeyEmpty = errors.New("key cannot be empty")

	// errValueNil is returned when attempting to set a nil value.
	errValueNil = errors.New("value cannot be nil")
)

// DB is the main interface for all database backends. DBs are concurrency-safe.
// Callers must call Close on the database when done.
type DB interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// It is safe to modify the contents of key and of the returned slice after Get
	// returns.
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists.
	// It is safe to modify the contents of key after Has returns.
	Has(key []byte) (bool, error)

	// Set sets the value for the given key, ovewriting it if it already exists.
	// It is safe to modify the contents of the arguments after Set returns.
	//
	// Set does not synchronize the data to disk immediately. Instead, the write may
	// be cached in memory and written to disk later during a background flush or
	// compaction. Use [SetSync] to flush the write to disk immediately.
	Set(key []byte, value []byte) error

	// SetSync sets the value for the given key, ovewriting it if it already exists.
	// It is safe to modify the contents of the arguments after Set returns.
	//
	// SetSync flushes the data to disk immediately and the write operation is
	// completed only after the data has been successfully written to persistent
	// storage.
	SetSync(key []byte, value []byte) error

	// Delete deletes the value for the given key. Deletes will succeed even if the
	// key does not exist in the database.
	// It is safe to modify the contents of key after Delete returns.
	//
	// Delete does not synchronize the delete to disk immediately. Instead, it may be
	// cached in memory and synced to disk later during a background flush or
	// compaction. Use [DeleteSync] to flush the delete to disk immediately.
	// Delete is faster than [DeleteSync] because it does not incur the latency of
	// disk I/O.
	Delete(key []byte) error

	// DeleteSync deletes the value for the given key. Deletes will succeed even if
	// the key does not exist in the database.
	// It is safe to modify the contents of key after Delete returns.
	//
	// DeleteSync flushes the delete to disk immediately and the delete operation is
	// completed only after it synced with persistent storage.
	// Because it incurs the latency of disk I/O, it is slower than [Delete].
	DeleteSync(key []byte) error

	// Iterator returns an iterator over a domain of keys, in ascending order.
	// The caller must call [Close] when done. End is exclusive, and start must be
	// less than end. A nil start iterates from the first key, and a nil end
	// iterates to the last key (inclusive). Empty keys are not valid.
	// No writes may happen within a domain of keys while an iterator exists over it.
	//
	// It is unsafe to modify the contents of the arguments while the returned
	// iterator is in use.
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending
	// order. The caller must call Close when done. End is exclusive, and start must
	// be less than end. A nil end iterates from the last key (inclusive), and a nil
	// start iterates to the first key (inclusive). Empty keys are not valid.
	// No writes may happen within a domain of keys while an iterator exists over it.
	//
	// It is unsafe to modify the contents of the arguments while the returned
	// iterator is in use.
	ReverseIterator(start, end []byte) (Iterator, error)

	// Close closes the database connection.
	// It is not safe to close a DB until all outstanding iterators are closed
	// or to call Close concurrently with any other DB method. It is not valid
	// to call any of a DB's methods after the DB has been closed.
	Close() error

	// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
	NewBatch() Batch

	// Print prints all the key/value pairs in the database for debugging purposes.
	Print() error

	// Stats returns a map of property values for all keys and the size of the cache.
	Stats() map[string]string

	// Compact compacts the specified range of keys in the database.
	Compact(start, end []byte) error
}

// Batch represents a group of writes. Callers must call Close on the batch when
// done. A batch is not safe for concurrent use.
type Batch interface {
	// Set adds a set update to the batch that sets the key to map to the value.
	// It is safe to modify the contents of the arguments after Set returns.
	Set(key, value []byte) error

	// Delete adds a delete update to the batch that deletes database the entry for
	// key.
	// It is safe to modify the contents of the arguments after Delete returns.
	Delete(key []byte) error

	// Write writes the batch, possibly without flushing to disk.  Only Close() can
	// be called after, other methods will error.
	Write() error

	// WriteSync writes the batch and flushes it to disk. Only Close() can be called
	// after, other methods will error.
	WriteSync() error

	// Close closes the batch. It is idempotent, but calls to other methods
	// afterwards will error.
	Close() error
}

// Iterator represents an iterator over a domain of keys. Callers must call Close
// when done. No writes can happen to a domain of keys while there exists an iterator over it, some backends may use database locks to ensure this will not
// happen.
//
// Callers must make sure the iterator is valid before calling any methods on it,
// otherwise these methods will panic. This is in part caused by most backend
// databases using this convention.
//
// The Iterator's methods return values that are not safe to modify while the
// iterator is in use. Callers should make a copy of the values if they need to
// modify them.
//
// Typical usage:
//
// var itr Iterator = ...
// defer itr.Close()
//
//	for ; itr.Valid(); itr.Next() {
//	  k, v := itr.Key(); itr.Value()
//	  ...
//	}
//
//	if err := itr.Error(); err != nil {
//	  ...
//	}
type Iterator interface {
	// Domain returns the start (inclusive) and end (exclusive) limits of the
	// iterator.
	Domain() ([]byte, []byte)

	// Valid returns whether the current iterator is valid. Once invalid, the
	// Iterator remains invalid forever.
	Valid() bool

	// Next moves the iterator to the next key in the database, as defined by order
	// of iteration. It panics if the iterator is invalid.
	Next()

	// Key returns the key at the current position. It panics if the iterator is
	// invalid.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to Next.
	// Therefore, callers should make a copy of the returned slice if they need to
	// modify it.
	Key() []byte

	// Value returns the value of the current key/value pair. It panics if the
	// iterator is invalid.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to Next.
	// Therefore, callers should make a copy of the returned slice if they need to
	// modify it.
	Value() []byte

	// Error returns the last error encountered by the iterator, if any.
	Error() error

	// Close closes the iterator, relasing any allocated resources.
	Close() error
}
