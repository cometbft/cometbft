---
order: 2
---

# Custom Database

This guide will explain to you how to use a custom database with CometBFT. By
default, CometBFT uses the `pebbledb` key-value store. However, you can use a
custom database with CometBFT. This guide will show you how to use any
key-value store, as long as it implements the `DB` interface.

Let's get started!

Say you have been using `goleveldb` and don't want to switch to `pebbledb`.
First, you will need to implement [the `DB` interface][db] for `goleveldb`.

NOTE: You can find old implementations of `DB` interface in
[cometbft-db](https://github.com/cometbft/cometbft-db) repo, which is archived
now.

```go
package goleveldb

import (
	"fmt"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type GoLevelDB struct {
	db *leveldb.DB
}

func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
    dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, o)
	if err != nil {
		return nil, err
	}

	database := &GoLevelDB{
		db: db,
	}
	return database, nil
}

// Get implements DB.
func (db *GoLevelDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errKeyEmpty
	}
	res, err := db.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

// Has implements DB.
func (db *GoLevelDB) Has(key []byte) (bool, error) {
	bytes, err := db.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements DB.
func (db *GoLevelDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	err := db.db.Put(key, value, nil)
	if err != nil {
		return err
	}
	return nil
}

// SetSync implements DB.
func (db *GoLevelDB) SetSync(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}

	err := db.db.Put(key, value, &opt.WriteOptions{Sync: true})
	if err != nil {
		return err
	}
	return nil
}

// Delete implements DB.
func (db *GoLevelDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}

	err := db.db.Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSync implements DB.
func (db *GoLevelDB) DeleteSync(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	err := db.db.Delete(key, &opt.WriteOptions{Sync: true})
	if err != nil {
		return err
	}
	return nil
}

// Close implements DB.
func (db *GoLevelDB) Close() error {
	return db.db.Close()
}

// Print implements DB.
func (db *GoLevelDB) Print() error {
	str, err := db.db.GetProperty("leveldb.stats")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", str)

	itr := db.db.NewIterator(nil, nil)
	for itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
	return nil
}

// Stats implements DB.
func (db *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
		"leveldb.sstables",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.openedtables",
		"leveldb.alivesnaps",
		"leveldb.aliveiters",
	}

	stats := make(map[string]string)
	for _, key := range keys {
		str, err := db.db.GetProperty(key)
		if err == nil {
			stats[key] = str
		}
	}
	return stats
}

// NewBatch implements DB.
func (db *GoLevelDB) NewBatch() Batch {
	return newGoLevelDBBatch(db)
}


//  Iterator implements DB.
func (db *GoLevelDB) Iterator(start, end []byte) (Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	itr := db.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DB.
func (db *GoLevelDB) ReverseIterator(start, end []byte) (Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	itr := db.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, true), nil
}

// Compact implements DB.
func (db *GoLevelDB) Compact(start, end []byte) error {
	return db.db.CompactRange(util.Range{Start: start, Limit: end})
}
```

Most of the functions in the `DB` interface are self-explanatory. We set some
keys, delete others, save and close the database.

`Stats` should output some useful information about the database. The `Print`
function is used to print the contents of the database (for debugging
purposes).

`Compact` is used to compact the database. If the database does not support
compaction, it can be a no-op.

The database must support batch writes. The `NewBatch` function should return a
new batch object.

With that, let's move on to [the `Batch`][batch] interface.

```go
package goleveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type goLevelDBBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}

func newGoLevelDBBatch(db *GoLevelDB) *goLevelDBBatch {
	return &goLevelDBBatch{
		db:    db,
		batch: new(leveldb.Batch),
	}
}

// Set implements Batch.
func (b *goLevelDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	if b.batch == nil {
		return errBatchClosed
	}
	b.batch.Put(key, value)
	return nil
}

// Delete implements Batch.
func (b *goLevelDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if b.batch == nil {
		return errBatchClosed
	}
	b.batch.Delete(key)
	return nil
}

// Write implements Batch.
func (b *goLevelDBBatch) Write() error {
	return b.write(false)
}

// WriteSync implements Batch.
func (b *goLevelDBBatch) WriteSync() error {
	return b.write(true)
}

func (b *goLevelDBBatch) write(sync bool) error {
	if b.batch == nil {
		return errBatchClosed
	}

	err := b.db.db.Write(b.batch, &opt.WriteOptions{Sync: sync})
	if err != nil {
		return err
	}
	// Make sure batch cannot be used afterwards. Callers should still call Close(), for errors.
	return b.Close()
}

// Close implements Batch.
func (b *goLevelDBBatch) Close() error {
	if b.batch != nil {
		b.batch.Reset()
		b.batch = nil
	}
	return nil
}
```

CONTRACT: either `Write` or `WriteSync` must be called before `Close`. `Close`
is always called 1 time.

Also, pretty straightforward. Now, let's move on to [the `Iterator` interface][iterator].

```go
package goleveldb

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type goLevelDBIterator struct {
	source    iterator.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

func newGoLevelDBIterator(source iterator.Iterator, start, end []byte, isReverse bool) *goLevelDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		} else {
			valid := source.Seek(end)
			if valid {
				eoakey := source.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.Last()
			}
		}
	} else {
		if start == nil {
			source.First()
		} else {
			source.Seek(start)
		}
	}
	return &goLevelDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Domain implements Iterator.
func (itr *goLevelDBIterator) Domain() (start []byte, end []byte) {
	return itr.start, itr.end
}

// Valid implements Iterator.
func (itr *goLevelDBIterator) Valid() bool {
	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// If source errors, invalid.
	if err := itr.Error(); err != nil {
		itr.isInvalid = true
		return false
	}

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	start := itr.start
	end := itr.end
	key := itr.source.Key()

	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

// Key implements Iterator.
// The caller should not modify the contents of the returned slice.
// Instead, the caller should make a copy and work on the copy.
func (itr *goLevelDBIterator) Key() []byte {
	itr.assertIsValid()
	return itr.source.Key()
}

// Value implements Iterator.
// The caller should not modify the contents of the returned slice.
// Instead, the caller should make a copy and work on the copy.
func (itr *goLevelDBIterator) Value() []byte {
	itr.assertIsValid()
	return itr.source.Value()
}

// Next implements Iterator.
func (itr *goLevelDBIterator) Next() {
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Error implements Iterator.
func (itr *goLevelDBIterator) Error() error {
	return itr.source.Error()
}

// Close implements Iterator.
func (itr *goLevelDBIterator) Close() error {
	itr.source.Release()
	return nil
}

func (itr goLevelDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
```

If there's an error, the iterator is invalid. `Valid` must return false in such
case and `Error` must return the error itself.

Glad you made it this far! Now you can use `goleveldb` as a custom database
with CometBFT. For that, you need to provide a different
[`DBProvider`][db-provider] when constructing a [new node object][new-node].

```go
customDBProvider := func(ctx *cmtdb.DBContext) (cmtdb.DB, error) {
    return goleveldb.NewGoLevelDB(ctx.ID, ctx.Config.DBDir())
}

node, err := nm.NewNode(
    context.Background(),
    config,
    pv,
    nodeKey,
    proxy.NewLocalClientCreator(app),
    nm.DefaultGenesisDocProviderFunc(config),
    customDBProvider,
    nm.DefaultMetricsProvider(config.Instrumentation),
    logger)
```

That's it! CometBFT will now use `goleveldb` as the database.

[db]: https://pkg.go.dev/github.com/cometbft/cometbft-db#DB
[batch]: https://pkg.go.dev/github.com/cometbft/cometbft-db#Batch
[iterator]: https://pkg.go.dev/github.com/cometbft/cometbft-db#Iterator
[db-provider]: https://pkg.go.dev/github.com/cometbft/cometbft@v1.0.1/config#DBProvider
[new-node]: https://pkg.go.dev/github.com/cometbft/cometbft/node#NewNode
