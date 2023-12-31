package core

import (
	"errors"
	"fmt"
	"sort"

	// <celestia-core>
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/pkg/consts"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/state"
	// </celestia-core>

	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtquery "github.com/cometbft/cometbft/libs/pubsub/query"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/state/txindex/null"
	"github.com/cometbft/cometbft/types"
)

// Tx allows you to query the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/tx
func (env *Environment) Tx(ctx *rpctypes.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, fmt.Errorf("transaction indexing is disabled")
	}

	r, err := env.TxIndexer.Get(hash)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	// var proof types.TxProof
	// if prove {
	// 	block := env.BlockStore.LoadBlock(r.Height)
	// 	proof = block.Data.Txs.Proof(int(r.Index))
	// }

	// return &ctypes.ResultTx{
	// 	Hash:     hash,
	// 	Height:   r.Height,
	// 	Index:    r.Index,
	// 	TxResult: r.Result,
	// 	Tx:       r.Tx,
	// 	Proof:    proof,
	// }, nil

	// <celestia-core>
	height := r.Height
	index := r.Index

	var shareProof types.ShareProof
	if prove {
		shareProof, err = env.proveTx(ctx, height, index)
		if err != nil {
			return nil, err
		}
	}

	return &ctypes.ResultTx{
		Hash:     hash,
		Height:   height,
		Index:    index,
		TxResult: r.Result,
		Tx:       r.Tx,
		Proof:    shareProof,
	}, nil
	// </celestia-core>
}

// TxSearch allows you to query for multiple transactions results. It returns a
// list of transactions (maximum ?per_page entries) and the total count.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/tx_search
func (env *Environment) TxSearch(
	ctx *rpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*ctypes.ResultTxSearch, error) {
	// if index is disabled, return error
	if _, ok := env.TxIndexer.(*null.TxIndex); ok {
		return nil, errors.New("transaction indexing is disabled")
	} else if len(query) > maxQueryLength {
		return nil, errors.New("maximum query length exceeded")
	}

	q, err := cmtquery.New(query)
	if err != nil {
		return nil, err
	}

	results, err := env.TxIndexer.Search(ctx.Context(), q)
	if err != nil {
		return nil, err
	}

	// sort results (must be done before pagination)
	switch orderBy {
	case "desc":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index > results[j].Index
			}
			return results[i].Height > results[j].Height
		})
	case "asc", "":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index < results[j].Index
			}
			return results[i].Height < results[j].Height
		})
	default:
		return nil, errors.New("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := env.validatePerPage(perPagePtr)

	page, err := validatePage(pagePtr, perPage, totalCount)
	if err != nil {
		return nil, err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := cmtmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*ctypes.ResultTx, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		r := results[i]

		// var proof types.TxProof
		// if prove {
		// 	block := env.BlockStore.LoadBlock(r.Height)
		// 	proof = block.Data.Txs.Proof(int(r.Index))
		// }

		// apiResults = append(apiResults, &ctypes.ResultTx{
		// 	Hash:     types.Tx(r.Tx).Hash(),
		// 	Height:   r.Height,
		// 	Index:    r.Index,
		// 	TxResult: r.Result,
		// 	Tx:       r.Tx,
		// 	Proof:    proof,
		// })

		// <celestia-core>
		var shareProof types.ShareProof
		if prove {
			shareProof, err = env.proveTx(ctx, r.Height, r.Index)
			if err != nil {
				return nil, err
			}
		}

		apiResults = append(apiResults, &ctypes.ResultTx{
			Hash:     types.Tx(r.Tx).Hash(),
			Height:   r.Height,
			Index:    r.Index,
			TxResult: r.Result,
			Tx:       r.Tx,
			Proof:    shareProof,
		})
		// </celestia-core>
	}

	return &ctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount}, nil
}

// <celestia-core>

func (env *Environment) proveTx(ctx *rpctypes.Context, height int64, index uint32) (types.ShareProof, error) {
	var (
		pShareProof cmtproto.ShareProof
		shareProof  types.ShareProof
	)
	rawBlock, err := loadRawBlock(env.BlockStore, height)
	if err != nil {
		return shareProof, err
	}
	res, err := env.ProxyAppQuery.Query(ctx.Context(), &abcitypes.RequestQuery{
		Data: rawBlock,
		Path: fmt.Sprintf(consts.TxInclusionProofQueryPath, index),
	})
	if err != nil {
		return shareProof, err
	}
	err = pShareProof.Unmarshal(res.Value)
	if err != nil {
		return shareProof, err
	}
	shareProof, err = types.ShareProofFromProto(pShareProof)
	if err != nil {
		return shareProof, err
	}
	return shareProof, nil
}

// ProveShares creates an NMT proof for a set of shares to a set of rows. It is
// end exclusive.
func (env *Environment) ProveShares(
	ctx *rpctypes.Context,
	height int64,
	startShare uint64,
	endShare uint64,
) (types.ShareProof, error) {
	var (
		pShareProof cmtproto.ShareProof
		shareProof  types.ShareProof
	)
	rawBlock, err := loadRawBlock(env.BlockStore, height)
	if err != nil {
		return shareProof, err
	}
	res, err := env.ProxyAppQuery.Query(ctx.Context(), &abcitypes.RequestQuery{
		Data: rawBlock,
		Path: fmt.Sprintf(consts.ShareInclusionProofQueryPath, startShare, endShare),
	})
	if err != nil {
		return shareProof, err
	}
	if res.Value == nil && res.Log != "" {
		// we can make the assumption that for custom queries, if the value is nil
		// and some logs have been emitted, then an error happened.
		return types.ShareProof{}, errors.New(res.Log)
	}
	err = pShareProof.Unmarshal(res.Value)
	if err != nil {
		return shareProof, err
	}
	shareProof, err = types.ShareProofFromProto(pShareProof)
	if err != nil {
		return shareProof, err
	}
	return shareProof, nil
}

func loadRawBlock(bs state.BlockStore, height int64) ([]byte, error) {
	var blockMeta = bs.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, fmt.Errorf("no block found for height %d", height)
	}

	buf := []byte{}
	for i := 0; i < int(blockMeta.BlockID.PartSetHeader.Total); i++ {
		part := bs.LoadBlockPart(height, i)
		// If the part is missing (e.g. since it has been deleted after we
		// loaded the block meta) we consider the whole block to be missing.
		if part == nil {
			return nil, fmt.Errorf("missing block part at height %d part %d", height, i)
		}
		buf = append(buf, part.Bytes...)
	}
	return buf, nil
}

// </celestia-core>
