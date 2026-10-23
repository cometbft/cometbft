package core

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpcserver "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

type txSearchBlockStore struct {
	sm.BlockStore
	loads   atomic.Int64
	missing int64
}

func (s *txSearchBlockStore) LoadBlock(height int64) *types.Block {
	s.loads.Add(1)
	if height == s.missing {
		return nil
	}
	return s.BlockStore.LoadBlock(height)
}

// Use the real block store and KV indexer, including protobuf decoding on reads.
// Only matching transactions are indexed; proofs still cover the entire block.
func txSearchFixture(tb testing.TB, blocks, txsPerBlock, matchesPerBlock int) (*Environment, *txSearchBlockStore) {
	tb.Helper()
	blockDB, indexDB := dbm.NewMemDB(), dbm.NewMemDB()
	tb.Cleanup(func() {
		require.NoError(tb, blockDB.Close())
		require.NoError(tb, indexDB.Close())
	})
	bs := store.NewBlockStore(blockDB)
	indexer := kv.NewTxIndex(indexDB)
	indexer.SetLogger(log.NewNopLogger())
	for h := 1; h <= blocks; h++ {
		txs := make(types.Txs, txsPerBlock)
		for i := range txs {
			txs[i] = make(types.Tx, 256)
			binary.BigEndian.PutUint64(txs[i], uint64(h))
			binary.BigEndian.PutUint64(txs[i][8:], uint64(i))
		}
		block := types.MakeBlock(int64(h), txs, new(types.Commit), nil)
		block.ProposerAddress = make([]byte, 20)
		block.ValidatorsHash = make([]byte, 32)
		parts, err := block.MakePartSet(types.BlockPartSizeBytes)
		require.NoError(tb, err)
		bs.SaveBlock(block, parts, &types.Commit{Height: int64(h)})
		for i := 0; i < matchesPerBlock; i++ {
			// Non-consecutive indices also exercise proof selection after sorting.
			index := i * txsPerBlock / matchesPerBlock
			require.NoError(tb, indexer.Index(&abci.TxResult{
				Height: int64(h), Index: uint32(index), Tx: txs[index],
				Result: abci.ExecTxResult{Code: uint32(i % 2), Data: []byte{byte(i)}},
			}))
		}
	}
	counted := &txSearchBlockStore{BlockStore: bs}
	return &Environment{BlockStore: counted, TxIndexer: indexer}, counted
}

func TestTxSearchProofs(t *testing.T) {
	env, bs := txSearchFixture(t, 4, 12, 4)
	for _, order := range []struct {
		value      string
		descending bool
	}{{"asc", false}, {"desc", true}, {"", false}} {
		for _, prove := range []bool{false, true} {
			for _, page := range []int{1, 2, 3, 4} {
				t.Run(fmt.Sprintf("order=%s/prove=%t/page=%d", order.value, prove, page), func(t *testing.T) {
					perPage := 5
					result, err := env.TxSearch(&rpctypes.Context{}, "tx.height >= 1", prove, &page, &perPage, order.value)
					require.NoError(t, err)
					require.Equal(t, 16, result.TotalCount)
					require.Len(t, result.Txs, min(perPage, 16-(page-1)*perPage))
					for i, tx := range result.Txs {
						position := (page-1)*perPage + i
						if order.descending {
							position = 15 - position
						}
						height, index := int64(position/4+1), position%4*3
						block := bs.BlockStore.LoadBlock(height)
						require.Equal(t, height, tx.Height)
						require.EqualValues(t, index, tx.Index)
						require.Equal(t, block.Txs[index], tx.Tx)
						require.EqualValues(t, block.Txs[index].Hash(), tx.Hash)
						require.EqualValues(t, position%4%2, tx.TxResult.Code)
						if prove {
							require.Equal(t, block.Txs.Proof(index), tx.Proof)
							require.NoError(t, tx.Proof.Validate(block.DataHash))
						} else {
							require.Equal(t, types.TxProof{}, tx.Proof)
						}
					}
				})
			}
		}
	}
}

func TestTxSearchProofBlockLoads(t *testing.T) {
	env, bs := txSearchFixture(t, 3, 100, 100)
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			bs.loads.Store(0)
			result, err := env.TxSearch(&rpctypes.Context{}, "tx.height = 1", true, nil, &n, "asc")
			require.NoError(t, err)
			require.Len(t, result.Txs, n)
			require.EqualValues(t, 1, bs.loads.Load())
		})
	}
	bs.missing = 2 // The index can retain transactions after their block is pruned.
	perPage, page := 100, 2
	for _, prove := range []bool{true, false} {
		bs.loads.Store(0)
		result, err := env.TxSearch(&rpctypes.Context{}, "tx.height >= 1", prove, &page, &perPage, "asc")
		require.NoError(t, err)
		require.Len(t, result.Txs, perPage)
		for _, tx := range result.Txs {
			require.EqualValues(t, 2, tx.Height)
			require.Equal(t, types.TxProof{}, tx.Proof)
		}
		if prove {
			require.EqualValues(t, 1, bs.loads.Load())
		} else {
			require.Zero(t, bs.loads.Load())
		}
	}
}

func TestTxSearchProofsRPC(t *testing.T) {
	env, bs := txSearchFixture(t, 3, 12, 4)
	bs.missing = 2
	mux := http.NewServeMux()
	rpcserver.RegisterRPCFuncs(mux, env.GetRoutes(), log.NewNopLogger())
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"tx_search","params":{"query":"tx.height >= 1","prove":true,"per_page":"100","order_by":"desc"}}`))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var envelope rpctypes.RPCResponse
	require.NoError(t, cmtjson.Unmarshal(response.Body.Bytes(), &envelope))
	require.Nil(t, envelope.Error)
	var result ctypes.ResultTxSearch
	require.NoError(t, cmtjson.Unmarshal(envelope.Result, &result))
	require.Equal(t, 12, result.TotalCount)
	require.Len(t, result.Txs, 12)
	for _, tx := range result.Txs {
		block := bs.BlockStore.LoadBlock(tx.Height)
		if tx.Height == 2 {
			require.Equal(t, types.TxProof{}, tx.Proof)
		} else {
			require.Equal(t, block.Txs.Proof(int(tx.Index)), tx.Proof)
			require.NoError(t, tx.Proof.Validate(block.DataHash))
		}
	}
}

func TestTxSearchProofsErrors(t *testing.T) {
	env, bs := txSearchFixture(t, 1, 10, 10)
	_, err := env.TxSearch(&rpctypes.Context{}, "invalid query", true, nil, nil, "asc")
	require.Error(t, err)
	result, err := env.TxSearch(&rpctypes.Context{}, "tx.height = 100", true, nil, nil, "asc")
	require.NoError(t, err)
	require.Empty(t, result.Txs)
	require.Zero(t, bs.loads.Load())
	page := 2
	_, err = env.TxSearch(&rpctypes.Context{}, "tx.height = 1", true, &page, nil, "asc")
	require.Error(t, err)
	_, err = env.TxSearch(&rpctypes.Context{}, "tx.height = 1", true, nil, nil, "invalid")
	require.Error(t, err)
	require.Zero(t, bs.loads.Load())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	// The existing KV indexer returns an empty result when already canceled.
	result, err = env.TxSearch(&rpctypes.Context{HTTPReq: req}, "tx.height = 1", true, nil, nil, "asc")
	require.NoError(t, err)
	require.Empty(t, result.Txs)
	require.Zero(t, bs.loads.Load())
}

func BenchmarkTxSearchProofs(b *testing.B) {
	for _, blockSize := range []int{100, 1000} {
		for _, layout := range []string{"same-block", "distinct-blocks"} {
			blocks, matches := 1, 100
			if layout == "distinct-blocks" {
				blocks, matches = 100, 1
			}
			env, bs := txSearchFixture(b, blocks, blockSize, matches)
			for _, prove := range []bool{false, true} {
				for _, n := range []int{1, 10, 100} {
					b.Run(fmt.Sprintf("block=%d/%s/prove=%t/N=%d", blockSize, layout, prove, n), func(b *testing.B) {
						bs.loads.Store(0)
						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							result, err := env.TxSearch(&rpctypes.Context{}, "tx.height >= 1", prove, nil, &n, "asc")
							if err != nil || len(result.Txs) != n {
								b.Fatalf("unexpected result: %v, %v", result, err)
							}
						}
						b.ReportMetric(float64(bs.loads.Load())/float64(b.N), "block-loads/op")
					})
				}
			}
		}
	}
}
