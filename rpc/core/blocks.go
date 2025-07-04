package core

import (
	"sort"

	"github.com/cometbft/cometbft/v2/libs/bytes"
	cmtmath "github.com/cometbft/cometbft/v2/libs/math"
	cmtquery "github.com/cometbft/cometbft/v2/libs/pubsub/query"
	ctypes "github.com/cometbft/cometbft/v2/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/v2/rpc/jsonrpc/types"
	blockidxnull "github.com/cometbft/cometbft/v2/state/indexer/block/null"
	"github.com/cometbft/cometbft/v2/types"
)

// BlockchainInfo gets block headers for minHeight <= height <= maxHeight.
// More: https://docs.cometbft.com/main/rpc/#/Info/blockchain
func (env *Environment) BlockchainInfo(
	_ *rpctypes.Context,
	minHeight, maxHeight int64,
) (*ctypes.ResultBlockchainInfo, error) {
	const limit int64 = 20
	var err error
	minHeight, maxHeight, err = filterMinMax(
		env.BlockStore.Base(),
		env.BlockStore.Height(),
		minHeight,
		maxHeight,
		limit)
	if err != nil {
		return nil, err
	}
	env.Logger.Debug("BlockchainInfoHandler", "maxHeight", maxHeight, "minHeight", minHeight)

	blockMetas := []*types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := env.BlockStore.LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ctypes.ResultBlockchainInfo{
		LastHeight: env.BlockStore.Height(),
		BlockMetas: blockMetas,
	}, nil
}

// error if either min or max are negative or min > max
// if 0, use blockstore base for min, latest block height for max
// enforce limit.
func filterMinMax(base, height, min, max, limit int64) (minHeight, maxHeight int64, err error) {
	// filter negatives
	if min < 0 || max < 0 {
		return min, max, ErrNegativeHeight
	}

	// adjust for default values
	if min == 0 {
		min = 1
	}
	if max == 0 {
		max = height
	}

	// limit max to the height
	max = cmtmath.MinInt64(height, max)

	// limit min to the base
	min = cmtmath.MaxInt64(base, min)

	// limit min to within `limit` of max
	// so the total number of blocks returned will be `limit`
	min = cmtmath.MaxInt64(min, max-limit+1)

	if min > max {
		return min, max, ErrHeightMinGTMax{Min: min, Max: max}
	}
	return min, max, nil
}

// Header gets block header at a given height.
// If no height is provided, it will fetch the latest header.
// More: https://docs.cometbft.com/main/rpc/#/Info/header
func (env *Environment) Header(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultHeader, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	blockMeta := env.BlockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return &ctypes.ResultHeader{}, nil
	}

	return &ctypes.ResultHeader{Header: &blockMeta.Header}, nil
}

// HeaderByHash gets header by hash.
// More: https://docs.cometbft.com/main/rpc/#/Info/header_by_hash
func (env *Environment) HeaderByHash(_ *rpctypes.Context, hash bytes.HexBytes) (*ctypes.ResultHeader, error) {
	// N.B. The hash parameter is HexBytes so that the reflective parameter
	// decoding logic in the HTTP service will correctly translate from JSON.
	// See https://github.com/tendermint/tendermint/issues/6802 for context.

	blockMeta := env.BlockStore.LoadBlockMetaByHash(hash)
	if blockMeta == nil {
		return &ctypes.ResultHeader{}, nil
	}

	return &ctypes.ResultHeader{Header: &blockMeta.Header}, nil
}

// Block gets block at a given height.
// If no height is provided, it will fetch the latest block.
// More: https://docs.cometbft.com/main/rpc/#/Info/block
func (env *Environment) Block(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlock, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	block, blockMeta := env.BlockStore.LoadBlock(height)
	if blockMeta == nil {
		return &ctypes.ResultBlock{BlockID: types.BlockID{}, Block: block}, nil
	}
	return &ctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// BlockByHash gets block by hash.
// More: https://docs.cometbft.com/main/rpc/#/Info/block_by_hash
func (env *Environment) BlockByHash(_ *rpctypes.Context, hash []byte) (*ctypes.ResultBlock, error) {
	block, blockMeta := env.BlockStore.LoadBlockByHash(hash)
	if blockMeta == nil {
		return &ctypes.ResultBlock{BlockID: types.BlockID{}, Block: nil}, nil
	}
	return &ctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// Commit gets block commit at a given height.
// If no height is provided, it will fetch the commit for the latest block.
// More: https://docs.cometbft.com/main/rpc/#/Info/commit
func (env *Environment) Commit(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultCommit, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	blockMeta := env.BlockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil, nil
	}
	header := blockMeta.Header

	// If the next block has not been committed yet,
	// use a non-canonical commit
	if height == env.BlockStore.Height() {
		commit := env.BlockStore.LoadSeenCommit(height)
		return ctypes.NewResultCommit(&header, commit, false), nil
	}

	// Return the canonical commit (comes from the block at height+1)
	commit := env.BlockStore.LoadBlockCommit(height)
	return ctypes.NewResultCommit(&header, commit, true), nil
}

// BlockResults gets ABCIResults at a given height.
// If no height is provided, it will fetch results for the latest block.
//
// Results are for the height of the block containing the txs.
// Thus response.results.deliver_tx[5] is the results of executing
// getBlock(h).Txs[5]
// More: https://docs.cometbft.com/main/rpc/#/Info/block_results
func (env *Environment) BlockResults(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlockResults, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	results, err := env.StateStore.LoadFinalizeBlockResponse(height)
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultBlockResults{
		Height:                height,
		TxResults:             results.TxResults,
		FinalizeBlockEvents:   results.Events,
		ValidatorUpdates:      results.ValidatorUpdates,
		ConsensusParamUpdates: results.ConsensusParamUpdates,
		AppHash:               results.AppHash,
	}, nil
}

// BlockSearch searches for a paginated set of blocks matching
// FinalizeBlock event search criteria.
func (env *Environment) BlockSearch(
	ctx *rpctypes.Context,
	query string,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*ctypes.ResultBlockSearch, error) {
	// skip if block indexing is disabled
	if _, ok := env.BlockIndexer.(*blockidxnull.BlockerIndexer); ok {
		return nil, ErrBlockIndexing
	}

	q, err := cmtquery.New(query)
	if err != nil {
		return nil, err
	}

	results, err := env.BlockIndexer.Search(ctx.Context(), q)
	if err != nil {
		return nil, err
	}

	// sort results (must be done before pagination)
	switch orderBy {
	case Descending, "":
		sort.Slice(results, func(i, j int) bool { return results[i] > results[j] })

	case Ascending:
		sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })

	default:
		return nil, ErrInvalidOrderBy{orderBy}
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

	apiResults := make([]*ctypes.ResultBlock, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		block, blockMeta := env.BlockStore.LoadBlock(results[i])
		if blockMeta != nil {
			apiResults = append(apiResults, &ctypes.ResultBlock{
				Block:   block,
				BlockID: blockMeta.BlockID,
			})
		}
	}

	return &ctypes.ResultBlockSearch{Blocks: apiResults, TotalCount: totalCount}, nil
}
