package mempool

import (
	"crypto/sha256"
	"fmt"

	abcicli "github.com/cometbft/cometbft/v2/abci/client"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/p2p"
	"github.com/cometbft/cometbft/v2/types"
)

const (
	MempoolChannel        = byte(0x30)
	MempoolControlChannel = byte(0x31)

	// PeerCatchupSleepIntervalMS defines how much time to sleep if a peer is behind.
	PeerCatchupSleepIntervalMS = 100
)

//go:generate ../scripts/mockery_generate.sh Mempool

// Mempool defines the mempool interface.
//
// Updates to the mempool need to be synchronized with committing a block so
// applications can reset their transient state on Commit.
type Mempool interface {
	// CheckTx executes a new transaction against the application to determine
	// its validity and whether it should be added to the mempool.
	CheckTx(tx types.Tx, sender p2p.ID) (*abcicli.ReqRes, error)

	// RemoveTxByKey removes a transaction, identified by its key,
	// from the mempool.
	RemoveTxByKey(txKey types.TxKey) error

	// ReapMaxBytesMaxGas reaps transactions from the mempool up to maxBytes
	// bytes total with the condition that the total gasWanted must be less than
	// maxGas.
	//
	// If both maxes are negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs

	// ReapMaxTxs reaps up to max transactions from the mempool. If max is
	// negative, there is no cap on the size of all returned transactions
	// (~ all available transactions).
	ReapMaxTxs(max int) types.Txs

	// GetTxByHash returns the types.Tx with the given hash if found in the mempool,
	// otherwise returns nil.
	GetTxByHash(hash []byte) types.Tx

	// Lock locks the mempool. The consensus must be able to hold lock to safely
	// update.
	Lock()

	// Unlock unlocks the mempool.
	Unlock()

	// PreUpdate signals that a new update is coming, before acquiring the mempool lock.
	// If the mempool is still rechecking at this point, it should be considered full.
	PreUpdate()

	// Update informs the mempool that the given txs were committed and can be
	// discarded.
	//
	// NOTE:
	// 1. This should be called *after* block is committed by consensus.
	// 2. Lock/Unlock must be managed by the caller.
	Update(
		blockHeight int64,
		blockTxs types.Txs,
		deliverTxResponses []*abci.ExecTxResult,
		newPreFn PreCheckFunc,
		newPostFn PostCheckFunc,
	) error

	// FlushAppConn flushes the mempool connection to ensure async callback calls
	// are done, e.g. from CheckTx.
	//
	// NOTE:
	// 1. Lock/Unlock must be managed by caller.
	FlushAppConn() error

	// Flush removes all transactions from the mempool and caches.
	Flush()

	// Contains returns true iff the transaction, identified by its key, is in
	// the mempool.
	Contains(txKey types.TxKey) bool

	// TxsAvailable returns a channel which fires once for every height, and only
	// when transactions are available in the mempool.
	//
	// NOTE:
	// 1. The returned channel may be nil if EnableTxsAvailable was not called.
	TxsAvailable() <-chan struct{}

	// EnableTxsAvailable initializes the TxsAvailable channel, ensuring it will
	// trigger once every height when transactions are available.
	EnableTxsAvailable()

	// Size returns the number of transactions in the mempool.
	Size() int

	// SizeBytes returns the total size of all txs in the mempool.
	SizeBytes() int64

	// GetSenders returns the list of node IDs from which we receive the given transaction.
	GetSenders(txKey types.TxKey) ([]p2p.ID, error)
}

// PreCheckFunc is an optional filter executed before CheckTx and rejects
// transaction if false is returned. An example would be to ensure that a
// transaction doesn't exceeded the block size.
type PreCheckFunc func(types.Tx) error

// PostCheckFunc is an optional filter executed after CheckTx and rejects
// transaction if false is returned. An example would be to ensure a
// transaction doesn't require more gas than available for the block.
type PostCheckFunc func(types.Tx, *abci.CheckTxResponse) error

// PreCheckMaxBytes checks that the size of the transaction is smaller or equal
// to the expected maxBytes.
func PreCheckMaxBytes(maxBytes int64) PreCheckFunc {
	return func(tx types.Tx) error {
		txSize := types.ComputeProtoSizeForTxs([]types.Tx{tx})

		if txSize > maxBytes {
			return fmt.Errorf("tx size is too big: %d, max: %d", txSize, maxBytes)
		}

		return nil
	}
}

// PostCheckMaxGas checks that the wanted gas is smaller or equal to the passed
// maxGas. Returns nil if maxGas is -1.
func PostCheckMaxGas(maxGas int64) PostCheckFunc {
	return func(_ types.Tx, res *abci.CheckTxResponse) error {
		if maxGas == -1 {
			return nil
		}
		if res.GasWanted < 0 {
			return fmt.Errorf("gas wanted %d is negative",
				res.GasWanted)
		}
		if res.GasWanted > maxGas {
			return fmt.Errorf("gas wanted %d is greater than max gas %d",
				res.GasWanted, maxGas)
		}

		return nil
	}
}

// TxKey is the fixed length array key used as an index.
type TxKey [sha256.Size]byte

// An Entry represents a transaction stored in the mempool.
type Entry interface {
	// Tx returns the transaction stored in the entry.
	Tx() types.Tx

	// Height returns the height of the latest block at the moment the entry was created.
	Height() int64

	// GasWanted returns the amount of gas required by the transaction.
	GasWanted() int64

	// IsSender returns whether we received the transaction from the given peer ID.
	IsSender(peerID p2p.ID) bool

	// Senders returns the list of registered peers that sent us the transaction.
	Senders() []p2p.ID
}

// An Iterator is used to iterate through the mempool entries.
// It allows multiple iterators to run concurrently, enabling
// parallel processing of mempool entries.
type Iterator interface {
	// WaitNextCh returns a channel on which to wait for the next available entry.
	WaitNextCh() <-chan Entry
}
