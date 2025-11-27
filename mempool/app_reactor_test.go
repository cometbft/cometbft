package mempool

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

func TestChunkTxs(t *testing.T) {
	makeTx := func(size int) types.Tx {
		return types.Tx(rand.Bytes(size))
	}

	toTxs := func(sizes []int) types.Txs {
		txs := make([]types.Tx, 0, len(sizes))
		for _, size := range sizes {
			txs = append(txs, makeTx(size))
		}
		return txs
	}

	for _, tt := range []struct {
		name   string
		input  []int
		size   int
		output [][]int
	}{
		{
			name:   "single tx smaller than size",
			input:  []int{100},
			size:   200,
			output: [][]int{{100}},
		},
		{
			name:   "single tx bigger than size",
			input:  []int{100},
			size:   50,
			output: [][]int{{100}},
		},
		{
			name:   "basic",
			input:  []int{100, 100, 100},
			size:   200,
			output: [][]int{{100, 100}, {100}},
		},
		{
			name:   "txs equal size",
			input:  []int{100, 100, 100},
			size:   100,
			output: [][]int{{100}, {100}, {100}},
		},
		{
			name:   "edge-case",
			input:  []int{101, 20, 30, 50, 2, 102, 3},
			size:   100,
			output: [][]int{{101}, {20, 30, 50}, {2}, {102}, {3}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			input := toTxs(tt.input)

			expected := make([]types.Txs, 0, len(tt.output))
			for _, chunk := range tt.output {
				expected = append(expected, toTxs(chunk))
			}

			// ACT
			actual := chunkTxs(input, tt.size)

			// ASSERT
			require.Equal(t, len(expected), len(actual), "output length mismatch")

			for i, chunk := range actual {
				require.Equal(t, len(expected[i]), len(chunk), "chunk length mismatch (#%d)", i)
			}
		})
	}
}
