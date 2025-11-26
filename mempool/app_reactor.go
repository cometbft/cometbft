package mempool

import (
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

	switchedOn           atomic.Bool
	waitForSwitchingOnCh chan struct{}
}

func NewAppReactor(
	config *config.MempoolConfig,
	mempool *AppMempool,
	waitSync bool,
) *AppReactor {
	r := &AppReactor{
		config:               config,
		mempool:              mempool,
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
	}

	return nil
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

	txHash := fmt.Sprintf("%X", tx.Hash())
	switch {
	case errors.Is(err, ErrSeenTx):
		r.Logger.Debug("Tx already seen", "tx", txHash, "peer", peerID)
	case errors.As(err, &ErrTxTooLarge{}):
		r.Logger.Debug("Tx too large", "err", err, "tx", txHash, "peer", peerID)
	default:
		r.Logger.Info("Failed to insert tx", "err", err, "tx", txHash, "peer", peerID)
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
