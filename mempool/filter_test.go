package mempool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/config"
	protomem "github.com/cometbft/cometbft/proto/tendermint/mempool"
)

// marshalTxsMsg builds the wire bytes for a wrapped tendermint.mempool.Message
// carrying the given raw transactions, exactly as it would appear on the wire.
func marshalTxsMsg(t *testing.T, txs [][]byte) []byte {
	t.Helper()
	msg := &protomem.Message{Sum: &protomem.Message_Txs{Txs: &protomem.Txs{Txs: txs}}}
	b, err := msg.Marshal()
	require.NoError(t, err)
	return b
}

func TestFilterMempoolMsgBytes(t *testing.T) {
	const maxTxBytes = 1024
	const maxBatchBytes = 4096

	// A batch packed with empty entries: cheap on the wire, but len(txs) is huge.
	emptyEntries := make([][]byte, 10000)
	for i := range emptyEntries {
		emptyEntries[i] = []byte{}
	}

	// A batch of one-byte entries whose declared sizes blow the byte budget.
	overBudget := make([][]byte, maxBatchBytes+1)
	for i := range overBudget {
		overBudget[i] = []byte{0x01}
	}

	testCases := []struct {
		name    string
		txs     [][]byte
		wantErr bool
	}{
		{
			name:    "valid single tx",
			txs:     [][]byte{[]byte("hello world")},
			wantErr: false,
		},
		{
			name:    "valid batch",
			txs:     [][]byte{[]byte("tx-one"), []byte("tx-two"), []byte("tx-three")},
			wantErr: false,
		},
		{
			name:    "valid tx at max size",
			txs:     [][]byte{make([]byte, maxTxBytes)},
			wantErr: false,
		},
		{
			name:    "empty txs list",
			txs:     [][]byte{},
			wantErr: true,
		},
		{
			name:    "single empty entry",
			txs:     [][]byte{{}},
			wantErr: true,
		},
		{
			name:    "empty-entry packing attack",
			txs:     emptyEntries,
			wantErr: true,
		},
		{
			name:    "tx exceeds max_tx_bytes",
			txs:     [][]byte{make([]byte, maxTxBytes+1)},
			wantErr: true,
		},
		{
			name:    "batch exceeds byte budget",
			txs:     overBudget,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msgBytes := marshalTxsMsg(t, tc.txs)
			err := filterMempoolMsgBytes(msgBytes, maxTxBytes, maxBatchBytes)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFilterMempoolMsgBytes_Unlimited(t *testing.T) {
	// With maxTxBytes == 0 and maxBatchBytes == 0 the size checks are disabled,
	// but empty entries are still rejected (they are never valid txs).
	big := marshalTxsMsg(t, [][]byte{make([]byte, 1<<20)})
	require.NoError(t, filterMempoolMsgBytes(big, 0, 0))

	require.Error(t, filterMempoolMsgBytes(marshalTxsMsg(t, [][]byte{{}}), 0, 0))
}

func TestFilterMempoolMsgBytes_Malformed(t *testing.T) {
	// Truncated varint for the outer Message field length.
	require.Error(t, filterMempoolMsgBytes([]byte{0x0a, 0xff}, 1024, 4096))

	// Length that runs past the end of the buffer.
	require.Error(t, filterMempoolMsgBytes([]byte{0x0a, 0x7f}, 1024, 4096))
}

func TestGossipBatchByteBudget(t *testing.T) {
	testCases := []struct {
		maxTxBytes    int
		maxBatchBytes int
		want          int
	}{
		{maxTxBytes: 1024, maxBatchBytes: 0, want: 1024},
		{maxTxBytes: 1024, maxBatchBytes: 4096, want: 4096},
		{maxTxBytes: 4096, maxBatchBytes: 1024, want: 4096}, // single large tx must still pass
		{maxTxBytes: 0, maxBatchBytes: 0, want: 0},
	}
	for _, tc := range testCases {
		cfg := &config.MempoolConfig{MaxTxBytes: tc.maxTxBytes, MaxBatchBytes: tc.maxBatchBytes}
		assert.Equal(t, tc.want, gossipBatchByteBudget(cfg))
	}
}

func TestReactorFilterMsgBytes_ChannelGuard(t *testing.T) {
	cfg := config.DefaultMempoolConfig()
	r := &AppReactor{config: cfg}

	bad := marshalTxsMsg(t, [][]byte{{}, {}, {}})

	// Non-mempool channel: not our concern, must pass through untouched.
	assert.NoError(t, r.FilterMsgBytes(byte(0x00), nil, bad))
	// Empty payload: nothing to validate.
	assert.NoError(t, r.FilterMsgBytes(MempoolChannel, nil, nil))
	// Mempool channel with abusive payload: rejected.
	assert.Error(t, r.FilterMsgBytes(MempoolChannel, nil, bad))
}
