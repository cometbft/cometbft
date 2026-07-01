package mempool

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/internal/protowire"
)

// Field numbers we care about. Both happen to be 1: the Message oneof's only
// member is Txs (field 1), and within Txs the repeated tx entries are field 1.
const (
	fieldMessageTxs = 1 // Message.sum: `Txs txs = 1`
	fieldTxsEntry   = 1 // Txs: `repeated bytes txs = 1`
)

var (
	errNoTransactions   = errors.New("mempool message contains no transactions")
	errEmptyTransaction = errors.New("mempool batch contains an empty transaction")
)

// batchTally accumulates what has been seen so far while scanning a batch.
type batchTally struct {
	count      int // number of transaction entries
	totalBytes int // sum of their sizes
}

// filterMempoolMsgBytes enforces the filtering rules above against
// the raw wire bytes of a tendermint.mempool.Message.
//
// A single tx can be as large as maxTxBytes even when maxBatchBytes is smaller,
// so the whole-message ceiling is the larger of the two. A non-positive value
// for either limit disables that check.
//
// Filtering rules:
//
//  1. Well-formed: every varint, tag and length prefix must stay within the
//     buffer. Truncated, overflowing or out-of-bounds encodings are rejected.
//  2. No empty entries: every transaction must have at least one byte. An empty
//     entry carries no payload yet still costs an allocation, so it is never
//     legitimate — this is the core of the amplification attack.
//  3. Per-tx bound: no single transaction may exceed maxTxBytes.
//  4. Per-batch bound: the sum of all transaction sizes may not exceed
//     max(maxTxBytes, maxBatchBytes).
//  5. Non-empty message: the message must carry at least one transaction.
func filterMempoolMsgBytes(msgBytes []byte, maxTxBytes, maxBatchBytes int) error {
	batchBudget := max(maxTxBytes, maxBatchBytes)
	var tally batchTally

	msg := protowire.NewWireCursor(msgBytes)
	for !msg.AtEnd() {
		fieldNum, wireType, err := msg.ReadTag()
		if err != nil {
			return err
		}

		// Anything that is not the Txs submessage is skipped, so unknown or
		// future fields do not break the scan.
		if fieldNum != fieldMessageTxs || wireType != protowire.WireBytes {
			if err := msg.SkipField(wireType); err != nil {
				return err
			}
			continue
		}

		txsBytes, err := msg.ReadLengthDelimited()
		if err != nil {
			return err
		}
		if err := scanTxsSubmessage(txsBytes, maxTxBytes, batchBudget, &tally); err != nil {
			return err
		}
	}

	if tally.count == 0 {
		return errNoTransactions
	}
	return nil
}

// scanTxsSubmessage walks a Txs submessage (`repeated bytes txs = 1`),
// validating each transaction entry against the limits and folding it into the
// tally.
func scanTxsSubmessage(txsBytes []byte, maxTxBytes, batchBudget int, tally *batchTally) error {
	txs := protowire.NewWireCursor(txsBytes)
	for !txs.AtEnd() {
		fieldNum, wireType, err := txs.ReadTag()
		if err != nil {
			return err
		}

		if fieldNum != fieldTxsEntry || wireType != protowire.WireBytes {
			if err := txs.SkipField(wireType); err != nil {
				return err
			}
			continue
		}

		tx, err := txs.ReadLengthDelimited()
		if err != nil {
			return err
		}
		if err := applyRules(tx, maxTxBytes, batchBudget, tally); err != nil {
			return err
		}
	}
	return nil
}

// applyRules applies the per-entry rules (2-4) to a single transaction.
func applyRules(tx []byte, maxTxBytes, batchBudget int, tally *batchTally) error {
	// No empty transaction
	if len(tx) == 0 {
		return errEmptyTransaction
	}
	// Transaction cannot exceed maxTxBytes
	if maxTxBytes > 0 && len(tx) > maxTxBytes {
		return fmt.Errorf("transaction size %d exceeds max_tx_bytes %d", len(tx), maxTxBytes)
	}

	tally.count++
	tally.totalBytes += len(tx)

	// The sum of all transaction sizes may not exceed the batch budget.
	if batchBudget > 0 && tally.totalBytes > batchBudget {
		return fmt.Errorf("mempool batch exceeds %d byte budget", batchBudget)
	}
	return nil
}
