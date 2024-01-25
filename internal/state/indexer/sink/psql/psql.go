// Package psql implements an event sink backed by a PostgreSQL database.
package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cometbft/cometbft/internal/rand"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/pubsub/query"
	"github.com/cometbft/cometbft/types"
)

const (
	tableBlocks     = "blocks"
	tableTxResults  = "tx_results"
	tableEvents     = "events"
	tableAttributes = "attributes"
	driverName      = "postgres"
)

// EventSink is an indexer backend providing the tx/block index services.  This
// implementation stores records in a PostgreSQL database using the schema
// defined in state/indexer/sink/psql/schema.sql.
type EventSink struct {
	store   *sql.DB
	chainID string
}

// NewEventSink constructs an event sink associated with the PostgreSQL
// database specified by connStr. Events written to the sink are attributed to
// the specified chainID.
func NewEventSink(connStr, chainID string) (*EventSink, error) {
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, err
	}
	return &EventSink{
		store:   db,
		chainID: chainID,
	}, nil
}

// DB returns the underlying Postgres connection used by the sink.
// This is exported to support testing.
func (es *EventSink) DB() *sql.DB { return es.store }

// runInTransaction executes query in a fresh database transaction.
// If query reports an error, the transaction is rolled back and the
// error from query is reported to the caller.
// Otherwise, the result of committing the transaction is returned.
func runInTransaction(db *sql.DB, query func(*sql.Tx) error) error {
	dbtx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := query(dbtx); err != nil {
		_ = dbtx.Rollback() // report the initial error, not the rollback
		return err
	}
	return dbtx.Commit()
}

func runBulkInsert(db *sql.DB, tableName string, columns []string, inserts [][]any) error {
	return runInTransaction(db, func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(pq.CopyIn(tableName, columns...))
		if err != nil {
			return fmt.Errorf("preparing bulk insert statement: %w", err)
		}
		for _, insert := range inserts {
			if _, err := stmt.Exec(insert...); err != nil {
				return fmt.Errorf("executing insert statement: %w", err)
			}
		}
		if _, err := stmt.Exec(); err != nil {
			return fmt.Errorf("flushing bulk insert: %w", err)
		}
		if err := stmt.Close(); err != nil {
			return fmt.Errorf("closing bulk insert statement: %w", err)
		}
		return nil
	})
}

// queryWithID executes the specified SQL query with the given arguments,
// expecting a single-row, single-column result containing an ID. If the query
// succeeds, the ID from the result is returned.
func queryWithID(tx *sql.Tx, query string, args ...interface{}) (uint32, error) {
	var id uint32
	if err := tx.QueryRow(query, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// insertEvents inserts a slice of events and any indexed attributes of those
// events into the database associated with dbtx.
//
// If txID > 0, the event is attributed to the transaction with that
// ID; otherwise it is recorded as a block event.
func insertEvents(dbtx *sql.Tx, blockID, txID uint32, evts []abci.Event) error {
	// Populate the transaction ID field iff one is defined (> 0).
	var txIDArg interface{}
	if txID > 0 {
		txIDArg = txID
	}

	const (
		insertEventQuery = `
			INSERT INTO ` + tableEvents + ` (block_id, tx_id, type)
			VALUES ($1, $2, $3)
			RETURNING rowid;
		`
		insertAttributeQuery = `
			INSERT INTO ` + tableAttributes + ` (event_id, key, composite_key, value)
			VALUES ($1, $2, $3, $4);
		`
	)

	// Add each event to the events table, and retrieve its row ID to use when
	// adding any attributes the event provides.
	for _, evt := range evts {
		// Skip events with an empty type.
		if evt.Type == "" {
			continue
		}

		eid, err := queryWithID(dbtx, insertEventQuery, blockID, txIDArg, evt.Type)
		if err != nil {
			return err
		}

		// Add any attributes flagged for indexing.
		for _, attr := range evt.Attributes {
			if !attr.Index {
				continue
			}
			compositeKey := evt.Type + "." + attr.Key
			if _, err := dbtx.Exec(insertAttributeQuery, eid, attr.Key, compositeKey, attr.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func bulkInsertEvents(blockID, txID uint32, events []v1.Event) (eventInserts, attrInserts [][]any) {
	for _, event := range events {
		// Skip events with an empty type.
		if event.Type == "" {
			continue
		}
		eventID := rand.Uint32() + 1
		eventInserts = append(eventInserts, []any{eventID, blockID, txID, event.Type})
		for _, attr := range event.Attributes {
			if !attr.Index {
				continue
			}
			compositeKey := event.Type + "." + attr.Key
			attrInserts = append(attrInserts, []any{eventID, attr.Key, compositeKey, attr.Value})
		}
	}
	return eventInserts, attrInserts
}

// makeIndexedEvent constructs an event from the specified composite key and
// value. If the key has the form "type.name", the event will have a single
// attribute with that name and the value; otherwise the event will have only
// a type and no attributes.
func makeIndexedEvent(compositeKey, value string) abci.Event {
	i := strings.Index(compositeKey, ".")
	if i < 0 {
		return abci.Event{Type: compositeKey}
	}
	return abci.Event{Type: compositeKey[:i], Attributes: []abci.EventAttribute{
		{Key: compositeKey[i+1:], Value: value, Index: true},
	}}
}

// IndexBlockEvents indexes the specified block header, part of the
// indexer.EventSink interface.
func (es *EventSink) IndexBlockEvents(h types.EventDataNewBlockEvents) error {
	ts := time.Now().UTC()

	return runInTransaction(es.store, func(dbtx *sql.Tx) error {
		// Add the block to the blocks table and report back its row ID for use
		// in indexing the events for the block.
		blockID, err := queryWithID(dbtx, `
INSERT INTO `+tableBlocks+` (height, chain_id, created_at)
  VALUES ($1, $2, $3)
  ON CONFLICT DO NOTHING
  RETURNING rowid;
`, h.Height, es.chainID, ts)
		if err == sql.ErrNoRows {
			return nil // we already saw this block; quietly succeed
		} else if err != nil {
			return fmt.Errorf("indexing block header: %w", err)
		}

		// Insert the special block meta-event for height.
		if err := insertEvents(dbtx, blockID, 0, []abci.Event{
			makeIndexedEvent(types.BlockHeightKey, strconv.FormatInt(h.Height, 10)),
		}); err != nil {
			return fmt.Errorf("block meta-events: %w", err)
		}
		// Insert all the block events. Order is important here,
		if err := insertEvents(dbtx, blockID, 0, h.Events); err != nil {
			return fmt.Errorf("finalizeblock events: %w", err)
		}
		return nil
	})
}

// getBlockIDs returns corresponding block ids for the provided heights
func (es *EventSink) getBlockIDs(heights []int64) ([]int64, error) {
	var blockIDs pq.Int64Array
	if err := es.store.QueryRow(`
SELECT array_agg((
	SELECT rowid FROM `+tableBlocks+` WHERE height = txr.height AND chain_id = $1
)) FROM unnest($2::bigint[]) AS txr(height);`,
		es.chainID, pq.Array(heights)).Scan(&blockIDs); err != nil {
		return nil, fmt.Errorf("getting block ids for txs from sql: %w", err)
	}
	return blockIDs, nil
}

func prefetchTxrExistence(db *sql.DB, blockIDs []int64, indexes []uint32) ([]bool, error) {
	var existence []bool
	if err := db.QueryRow(`
SELECT array_agg((
	SELECT EXISTS(SELECT 1 FROM `+tableTxResults+` WHERE block_id = txr.block_id AND index = txr.index)
)) FROM UNNEST($1::bigint[], $2::integer[]) as txr(block_id, index);`,
		pq.Array(blockIDs), pq.Array(indexes)).Scan((*pq.BoolArray)(&existence)); err != nil {
		return nil, fmt.Errorf("fetching already indexed txrs: %w", err)
	}
	return existence, nil
}

func (es *EventSink) IndexTxEvents(txrs []*abci.TxResult) error {
	ts := time.Now().UTC()
	heights := make([]int64, len(txrs))
	indexes := make([]uint32, len(txrs))
	for i, txr := range txrs {
		heights[i] = txr.Height
		indexes[i] = txr.Index
	}
	// prefetch blockIDs for all txrs. Every block header must have been indexed
	// prior to the transactions belonging to it.
	blockIDs, err := es.getBlockIDs(heights)
	if err != nil {
		return fmt.Errorf("getting block ids for txs: %w", err)
	}
	alreadyIndexed, err := prefetchTxrExistence(es.store, blockIDs, indexes)
	if err != nil {
		return fmt.Errorf("failed to prefetch which txrs were already indexed: %w", err)
	}
	txrInsertColumns := []string{"rowid", "block_id", "index", "created_at", "tx_hash", "tx_result"}
	eventInsertColumns := []string{"rowid", "block_id", "tx_id", "type"}
	attrInsertColumns := []string{"event_id", "key", "composite_key", "value"}
	var txrInserts, attrInserts, eventInserts [][]any
	for i, txr := range txrs {
		if alreadyIndexed[i] {
			continue
		}
		// Encode the result message in protobuf wire format for indexing.
		resultData, err := proto.Marshal(txr)
		if err != nil {
			return fmt.Errorf("marshaling tx_result: %w", err)
		}
		// Index the hash of the underlying transaction as a hex string.
		txHash := fmt.Sprintf("%X", types.Tx(txr.Tx).Hash())
		// Generate random ID for this tx_result and insert a record for it
		txID := rand.Uint32() + 1
		txrInserts = append(txrInserts, []any{txID, blockIDs[i], txr.Index, ts, txHash, resultData})
		// Insert the special transaction meta-events for hash and height.
		events := append(txr.Result.Events,
			makeIndexedEvent(types.TxHashKey, txHash),
			makeIndexedEvent(types.TxHeightKey, strconv.FormatInt(txr.Height, 10)),
		)
		newEventInserts, newAttrInserts := bulkInsertEvents(uint32(blockIDs[i]), txID, events)
		eventInserts = append(eventInserts, newEventInserts...)
		attrInserts = append(attrInserts, newAttrInserts...)
	}
	if err := runBulkInsert(es.store, tableTxResults, txrInsertColumns, txrInserts); err != nil {
		return fmt.Errorf("bulk inserting txrs: %w", err)
	}
	if err := runBulkInsert(es.store, tableEvents, eventInsertColumns, eventInserts); err != nil {
		return fmt.Errorf("bulk inserting events: %w", err)
	}
	if err := runBulkInsert(es.store, tableAttributes, attrInsertColumns, attrInserts); err != nil {
		return fmt.Errorf("bulk inserting attributes: %w", err)
	}
	return nil
}

// SearchBlockEvents is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) SearchBlockEvents(_ context.Context, _ *query.Query) ([]int64, error) {
	return nil, errors.New("block search is not supported via the postgres event sink")
}

// SearchTxEvents is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) SearchTxEvents(_ context.Context, _ *query.Query) ([]*abci.TxResult, error) {
	return nil, errors.New("tx search is not supported via the postgres event sink")
}

// GetTxByHash is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) GetTxByHash(_ []byte) (*abci.TxResult, error) {
	return nil, errors.New("getTxByHash is not supported via the postgres event sink")
}

// HasBlock is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) HasBlock(_ int64) (bool, error) {
	return false, errors.New("hasBlock is not supported via the postgres event sink")
}

// Stop closes the underlying PostgreSQL database.
func (es *EventSink) Stop() error { return es.store.Close() }
