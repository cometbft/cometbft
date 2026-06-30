package mempool

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/config"
)

// Protobuf wire types (the low 3 bits of a field tag).
const (
	wireVarint  = 0
	wireFixed64 = 1
	wireBytes   = 2
	wireFixed32 = 5
)

// Field numbers we care about. Both happen to be 1: the Message oneof's only
// member is Txs (field 1), and within Txs the repeated tx entries are field 1.
const (
	fieldMessageTxs = 1 // Message.sum: `Txs txs = 1`
	fieldTxsEntry   = 1 // Txs: `repeated bytes txs = 1`
)

var (
	errVarintOverflow   = errors.New("malformed mempool message: varint overflow")
	errTruncatedVarint  = errors.New("malformed mempool message: truncated varint")
	errOutOfBounds      = errors.New("malformed mempool message: length out of bounds")
	errNoTransactions   = errors.New("mempool message contains no transactions")
	errEmptyTransaction = errors.New("mempool batch contains an empty transaction")
)

// gossipBatchByteBudget is the max raw tx bytes a peer may pack into one gossip
// message. A single tx can be MaxTxBytes even when MaxBatchBytes is smaller, so
// the budget is the larger of the two. A non-positive result means unbounded.
func gossipBatchByteBudget(cfg *config.MempoolConfig) int {
	return max(cfg.MaxTxBytes, cfg.MaxBatchBytes)
}

// txLimits holds the size ceilings applied to a gossip batch. A non-positive
// value disables that limit.
type txLimits struct {
	maxTxBytes    int // per-transaction ceiling
	maxBatchBytes int // whole-message ceiling
}

// batchTally accumulates what has been seen so far while scanning a batch.
type batchTally struct {
	count      int // number of transaction entries
	totalBytes int // sum of their sizes
}

// filterMempoolMsgBytes enforces the filtering rules above against
// the raw wire bytes of a tendermint.mempool.Message.
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
//     maxBatchBytes.
//  5. Non-empty message: the message must carry at least one transaction.
func filterMempoolMsgBytes(msgBytes []byte, maxTxBytes, maxBatchBytes int) error {
	limits := txLimits{maxTxBytes: maxTxBytes, maxBatchBytes: maxBatchBytes}
	var tally batchTally

	msg := wireCursor{buf: msgBytes}
	for !msg.atEnd() {
		fieldNum, wireType, err := msg.readTag()
		if err != nil {
			return err
		}

		// Anything that is not the Txs submessage is skipped, so unknown or
		// future fields do not break the scan.
		if fieldNum != fieldMessageTxs || wireType != wireBytes {
			if err := msg.skipField(wireType); err != nil {
				return err
			}
			continue
		}

		txsBytes, err := msg.readLengthDelimited()
		if err != nil {
			return err
		}
		if err := scanTxsSubmessage(txsBytes, limits, &tally); err != nil {
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
func scanTxsSubmessage(txsBytes []byte, limits txLimits, tally *batchTally) error {
	txs := wireCursor{buf: txsBytes}
	for !txs.atEnd() {
		fieldNum, wireType, err := txs.readTag()
		if err != nil {
			return err
		}

		if fieldNum != fieldTxsEntry || wireType != wireBytes {
			if err := txs.skipField(wireType); err != nil {
				return err
			}
			continue
		}

		tx, err := txs.readLengthDelimited()
		if err != nil {
			return err
		}
		if err := applyRules(tx, limits, tally); err != nil {
			return err
		}
	}
	return nil
}

// applyRules applies the per-entry rules (2-4) to a single transaction.
func applyRules(tx []byte, limits txLimits, tally *batchTally) error {
	// No empty transaction
	if len(tx) == 0 {
		return errEmptyTransaction
	}
	// Transaction cannot exceed maxTxBytes
	if limits.maxTxBytes > 0 && len(tx) > limits.maxTxBytes {
		return fmt.Errorf("transaction size %d exceeds max_tx_bytes %d", len(tx), limits.maxTxBytes)
	}

	tally.count++
	tally.totalBytes += len(tx)

	// The sum of all transaction sizes may not exceed maxBatchBytes
	if limits.maxBatchBytes > 0 && tally.totalBytes > limits.maxBatchBytes {
		return fmt.Errorf("mempool batch exceeds %d byte budget", limits.maxBatchBytes)
	}
	return nil
}

// wireCursor walks a protobuf buffer left to right. Every read is bounds-checked
// and advances the cursor; a read past the end returns an error rather than
// panicking, which is what enforces rule 1 (well-formedness).
type wireCursor struct {
	buf []byte
	pos int
}

func (c *wireCursor) atEnd() bool { return c.pos >= len(c.buf) }

// readVarint decodes a base-128 varint at the cursor and advances past it.
func (c *wireCursor) readVarint() (uint64, error) {
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return 0, errVarintOverflow
		}
		if c.pos >= len(c.buf) {
			return 0, errTruncatedVarint
		}
		b := c.buf[c.pos]
		c.pos++
		v |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return v, nil
		}
	}
}

// readTag decodes a field tag, splitting it into field number and wire type.
func (c *wireCursor) readTag() (fieldNum, wireType int, err error) {
	v, err := c.readVarint()
	if err != nil {
		return 0, 0, err
	}
	fieldNum = int(v >> 3)
	wireType = int(v & 0x7)
	if fieldNum <= 0 {
		return 0, 0, fmt.Errorf("malformed mempool message: illegal field number %d", fieldNum)
	}
	return fieldNum, wireType, nil
}

// readLengthDelimited reads a length-prefixed run of bytes (wire type 2) and
// returns it as a sub-slice of the underlying buffer, advancing past it.
func (c *wireCursor) readLengthDelimited() ([]byte, error) {
	length, err := c.readVarint()
	if err != nil {
		return nil, err
	}
	start := c.pos
	end := start + int(length)
	if end < start || end > len(c.buf) {
		return nil, errOutOfBounds
	}
	c.pos = end
	return c.buf[start:end], nil
}

// skipField advances the cursor past the body of a field with the given wire
// type, used to ignore fields the filter does not care about.
func (c *wireCursor) skipField(wireType int) error {
	switch wireType {
	case wireVarint:
		_, err := c.readVarint()
		return err
	case wireFixed64:
		return c.advance(8)
	case wireBytes:
		_, err := c.readLengthDelimited()
		return err
	case wireFixed32:
		return c.advance(4)
	default:
		return fmt.Errorf("malformed mempool message: unsupported wire type %d", wireType)
	}
}

// advance moves the cursor forward by n bytes, bounds-checked.
func (c *wireCursor) advance(n int) error {
	end := c.pos + n
	if end < c.pos || end > len(c.buf) {
		return errOutOfBounds
	}
	c.pos = end
	return nil
}
