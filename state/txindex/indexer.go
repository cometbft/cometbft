package txindex

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/pubsub/query"
)

// XXX/TODO: These types should be moved to the indexer package.

//go:generate ../../scripts/mockery_generate.sh TxIndexer

// TxIndexer defines an interface for indexing, storing, and retrieving transactions
// against an underlying database. Implementations of TxIndexer are expected to
// maintain an indexed view of transactions that can be queried efficiently.
//
// All methods that write to or read from storage ultimately interact with an
// underlying database. Therefore, the caller should ensure that the database is
// properly closed by calling Close() on the TxIndexer when finished. Failing to do
// so may result in resource leaks and inconsistent data states.
type TxIndexer interface {
	// AddBatch analyzes, indexes, and stores a batch of transactions in the
	// underlying database. The provided batch should contain all the transaction
	// data needed for indexing. An error is returned if any part of the batch
	// fails to be stored.
	AddBatch(b *Batch) error

	// Index analyzes, indexes, and stores a single transaction in the underlying
	// database. This method is typically used for incrementally updating the index
	// as new transactions arrive.
	Index(result *abci.TxResult) error

	// Get retrieves a transaction by its hash from the underlying database. If the
	// transaction is not found or has not been indexed, it returns nil.
	Get(hash []byte) (*abci.TxResult, error)

	// Search queries the underlying database for transactions matching the provided
	// query. It returns a slice of transaction results, the total number of
	// matching transactions, and an error if the search operation encounters any
	// issues. Because the function can do a lot of database I/O, callers should
	// provide a valid context to cancel long-running searches.
	Search(ctx context.Context, q *query.Query, pagSettings Pagination) ([]*abci.TxResult, int, error)

	// SetLogger configures a logger for this TxIndexer. This logger may be used
	// to report database I/O operations, indexing progress, or errors encountered
	// during storage and retrieval.
	SetLogger(l log.Logger)

	// Prune removes any data older than the specified retainHeight from the
	// underlying database. It returns the number of records pruned and the
	// resulting height after pruning, or an error if the prune operation fails.
	Prune(retainHeight int64) (int64, int64, error)

	// GetRetainHeight retrieves the current retain height from the underlying
	// database, indicating the earliest block height from which transactions are
	// retained.
	GetRetainHeight() (int64, error)

	// SetRetainHeight updates the retain height in the underlying database. This
	// value determines how old transactions can be before they are pruned. An error
	// is returned if the value cannot be persisted.
	SetRetainHeight(retainHeight int64) error

	// Close closes the underlying database used by the TxIndexer. This method
	// should be called by the entity that created the TxIndexer once it is done
	// with all indexing and queries. Failing to call Close may leave database
	// connections open, resulting in resource leaks and potential inconsistencies.
	Close() error
}

// Batch groups together multiple Index operations to be performed at the same time.
// NOTE: Batch is NOT thread-safe and must not be modified after starting its execution.
type Batch struct {
	Ops []*abci.TxResult
}

// Pagination provides pagination information for queries.
// This allows us to use the same TxSearch API for pruning to return all relevant data,
// while still limiting public queries to pagination.
type Pagination struct {
	OrderDesc   bool
	IsPaginated bool
	Page        int
	PerPage     int
}

// NewBatch creates a new Batch.
func NewBatch(n int64) *Batch {
	return &Batch{
		Ops: make([]*abci.TxResult, n),
	}
}

// Add or update an entry for the given result.Index.
func (b *Batch) Add(result *abci.TxResult) error {
	b.Ops[result.Index] = result
	return nil
}

// Size returns the total number of operations inside the batch.
func (b *Batch) Size() int {
	return len(b.Ops)
}

// ErrorEmptyHash indicates empty hash.
var ErrorEmptyHash = errors.New("transaction hash cannot be empty")
