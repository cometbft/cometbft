package mempool

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
	"github.com/cometbft/cometbft/types"
	"github.com/pkg/errors"
)

// AppReactor for interacting with AppMempool
type AppReactor struct {
	p2p.BaseReactor
	config  *config.MempoolConfig
	mempool *AppMempool

	ctx       context.Context
	cancelCtx context.CancelFunc

	switchedOn           atomic.Bool
	waitForSwitchingOnCh chan struct{}
}

func NewAppReactor(
	config *config.MempoolConfig,
	mempool *AppMempool,
	waitSync bool,
) *AppReactor {
	ctx, cancelCtx := context.WithCancel(context.Background())

	r := &AppReactor{
		config:               config,
		mempool:              mempool,
		ctx:                  ctx,
		cancelCtx:            cancelCtx,
		switchedOn:           atomic.Bool{},
		waitForSwitchingOnCh: nil,
	}

	r.BaseReactor = *p2p.NewBaseReactor("Mempool", r)

	if waitSync {
		r.switchedOn.Store(false)
		r.waitForSwitchingOnCh = make(chan struct{})
	} else {
		r.switchedOn.Store(true)
	}

	return r
}

// OnStart implements p2p.BaseReactor.
func (r *AppReactor) OnStart() error {
	if !r.switchedOn.Load() {
		r.Logger.Info("Waiting for mempool reactor to be switched on")
	}

	if !r.config.Broadcast {
		r.Logger.Info("Tx broadcasting is disabled")
		return nil
	}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				r.Logger.Error("Panic in broadcast routine", "panic", p)
			}
		}()

		// fallback to max tx bytes if max batch bytes is not set
		// most chains use 1MB which will definitely fit many small txs
		maxBatchSizeBytes := r.config.MaxTxBytes
		if r.config.MaxBatchBytes > 0 {
			maxBatchSizeBytes = r.config.MaxBatchBytes
		}

		r.broadcastTransactionsBatch(r.ctx, maxBatchSizeBytes)

		r.Logger.Info("Broadcast routine stopped")
	}()

	return nil
}

func (r *AppReactor) OnStop() {
	if !r.enabled() {
		return
	}

	// will close the context and cancel broadcast
	r.cancelCtx()
}

// GetChannels implements p2p.BaseReactor.
func (r *AppReactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, r.config.MaxTxBytes)
	batchMsg := protomem.Message{
		Sum: &protomem.Message_Txs{
			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
		},
	}

	return []*p2p.ChannelDescriptor{
		{
			ID:                  MempoolChannel,
			Priority:            5,
			RecvMessageCapacity: batchMsg.Size(),
			MessageType:         &protomem.Message{},
		},
	}
}

// WaitSync used for backward compatibility with external callers
func (r *AppReactor) WaitSync() bool {
	return !r.enabled()
}

// EnableInOutTxs enables inbound and outbound transactions
func (r *AppReactor) EnableInOutTxs() {
	if !r.switchedOn.CompareAndSwap(false, true) {
		// noop
		return
	}

	r.Logger.Info("Enabled inbound and outbound transactions")
	close(r.waitForSwitchingOnCh)
}

func (r *AppReactor) Receive(e p2p.Envelope) {
	if !r.enabled() {
		r.Logger.Debug("Ignored mempool message received while syncing")
		return
	}

	peerID := e.Src.ID()

	txs, err := txsFromEnvelope(e)
	if err != nil {
		r.Logger.Error("Failed to parse txs from envelope", "err", err, "peer", peerID)
		// todo disconnect peer for misbehaving?
		return
	}

	for _, tx := range txs {
		r.insertTx(peerID, tx)
	}
}

func (r *AppReactor) insertTx(peerID p2p.ID, tx types.Tx) {
	err := r.mempool.InsertTx(tx)
	if err == nil {
		// all good
		return
	}

	txHash := txHash(tx)
	switch {
	case errors.Is(err, ErrSeenTx):
		r.Logger.Debug("Tx already seen", "tx", txHash, "peer", peerID)
	case errors.As(err, &ErrTxTooLarge{}):
		r.Logger.Debug("Tx too large", "err", err, "tx", txHash, "peer", peerID)
	default:
		r.Logger.Info("Failed to insert tx", "err", err, "tx", txHash, "peer", peerID)
	}
}

// broadcastTransactionsBatch subscribes to new txs from app-mempool,
// accumulates them in batches and broadcasts them to all connected peers.
// Previously batching was disabled.
// @see https://github.com/tendermint/tendermint/issues/5796
func (r *AppReactor) broadcastTransactionsBatch(ctx context.Context, maxBatchSizeBytes int) {
	// will be closed when r.ctx is canceled, 8 is chan size,
	// which is ok since we send async
	stream := r.mempool.TxStream(ctx)

	for txs := range stream {
		batches := chunkTxs(txs, maxBatchSizeBytes)

		for _, batch := range batches {
			r.Switch.BroadcastAsync(p2p.Envelope{
				Message:   &protomem.Txs{Txs: batch.ToSliceOfBytes()},
				ChannelID: MempoolChannel,
			})
		}
	}
}

func (r *AppReactor) enabled() bool {
	return r.switchedOn.Load()
}

func txsFromEnvelope(e p2p.Envelope) ([]types.Tx, error) {
	msg, ok := e.Message.(*protomem.Txs)
	if !ok {
		return nil, fmt.Errorf("not a mempool.Txs message type: %T", e.Message)
	}

	txsRaw := msg.GetTxs()
	switch len(txsRaw) {
	case 0:
		return nil, fmt.Errorf("received empty txs list")
	case 1:
		// skip loop
		return []types.Tx{types.Tx(txsRaw[0])}, nil
	default:
		txs := make([]types.Tx, len(txsRaw))
		for i, tx := range txsRaw {
			txs[i] = types.Tx(tx)
		}
		return txs, nil
	}
}

// chunkTxs chunks transactions into batches of maxBatchSizeBytes
// example: [tx1, tx2, tx3, tx4, tx5, ...] -> [[tx1, tx2], [tx3, tx4], [tx5], ...]
//
// note: we can optimize []types.Txs to [][][]byte + have less allocs,
// but it's not worth it for now.
func chunkTxs(txs types.Txs, maxBatchSizeBytes int) []types.Txs {
	// should not happen
	if len(txs) == 0 {
		return nil
	}

	chunks := []types.Txs{}

	lastChunkSizeBytes := 0
	lastChunk := types.Txs{}

	for _, tx := range txs {
		txSizeBytes := len(tx)

		// tx won't fit into chunk, add current chunk to chunks and start a new one
		if (lastChunkSizeBytes + txSizeBytes) > maxBatchSizeBytes {
			// this check required to avoid adding empty chunk to chunks
			// when a single tx is bigger than maxBatchSizeBytes
			if len(lastChunk) > 0 {
				chunks = append(chunks, lastChunk)
			}

			// reset chunk size
			lastChunk = types.Txs{}
			lastChunkSizeBytes = 0
		}

		lastChunk = append(lastChunk, tx)
		lastChunkSizeBytes += txSizeBytes
	}

	// last chunk
	if len(lastChunk) > 0 {
		chunks = append(chunks, lastChunk)
	}

	return chunks
}

func txHash(tx types.Tx) string {
	return fmt.Sprintf("%X", tx.Hash())
}
