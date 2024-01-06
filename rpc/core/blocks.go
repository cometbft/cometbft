package core

import (
	"errors"
	"fmt"
	"sort"

	// <celestia-core>
	"encoding/hex"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/pkg/consts"
	"strconv"
	// </celestia-core>

	"github.com/cometbft/cometbft/libs/bytes"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtquery "github.com/cometbft/cometbft/libs/pubsub/query"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	blockidxnull "github.com/cometbft/cometbft/state/indexer/block/null"
	"github.com/cometbft/cometbft/types"
)

// BlockchainInfo gets block headers for minHeight <= height <= maxHeight.
//
// If maxHeight does not yet exist, blocks up to the current height will be
// returned. If minHeight does not exist (due to pruning), earliest existing
// height will be used.
//
// At most 20 items will be returned. Block headers are returned in descending
// order (highest first).
//
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/blockchain
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
func filterMinMax(base, height, min, max, limit int64) (int64, int64, error) {
	// filter negatives
	if min < 0 || max < 0 {
		return min, max, fmt.Errorf("heights must be non-negative")
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
		return min, max, fmt.Errorf("min height %d can't be greater than max height %d", min, max)
	}
	return min, max, nil
}

// Header gets block header at a given height.
// If no height is provided, it will fetch the latest header.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/header
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
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/header_by_hash
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
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/block
func (env *Environment) Block(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlock, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	block := env.BlockStore.LoadBlock(height)
	blockMeta := env.BlockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return &ctypes.ResultBlock{BlockID: types.BlockID{}, Block: block}, nil
	}
	return &ctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// <celestia-core>

// SignedBlock fetches the set of transactions at a specified height and all the relevant
// data to verify the transactions (i.e. using light client verification).
func (env *Environment) SignedBlock(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultSignedBlock, error) {
	height, err := env.getHeight(env.BlockStore.Height(), heightPtr)
	if err != nil {
		return nil, err
	}

	block := env.BlockStore.LoadBlock(height)
	if block == nil {
		return nil, errors.New("block not found")
	}
	seenCommit := env.BlockStore.LoadSeenCommit(height)
	if seenCommit == nil {
		return nil, errors.New("seen commit not found")
	}
	validatorSet, err := env.StateStore.LoadValidators(height)
	if validatorSet == nil || err != nil {
		return nil, err
	}

	return &ctypes.ResultSignedBlock{
		Header:       block.Header,
		Commit:       *seenCommit,
		ValidatorSet: *validatorSet,
		Data:         block.Data,
	}, nil
}

// </celestia-core>

// BlockByHash gets block by hash.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/block_by_hash
func (env *Environment) BlockByHash(_ *rpctypes.Context, hash []byte) (*ctypes.ResultBlock, error) {
	block := env.BlockStore.LoadBlockByHash(hash)
	if block == nil {
		return &ctypes.ResultBlock{BlockID: types.BlockID{}, Block: nil}, nil
	}
	// If block is not nil, then blockMeta can't be nil.
	blockMeta := env.BlockStore.LoadBlockMeta(block.Height)
	return &ctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// Commit gets block commit at a given height.
// If no height is provided, it will fetch the commit for the latest block.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/commit
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

// <celestia>

// DataCommitment collects the data roots over a provided ordered range of blocks,
// and then creates a new Merkle root of those data roots. The range is end exclusive.
func (env *Environment) DataCommitment(ctx *rpctypes.Context, start, end uint64) (*ctypes.ResultDataCommitment, error) {
	err := env.validateDataCommitmentRange(start, end)
	if err != nil {
		return nil, err
	}
	tuples, err := env.fetchDataRootTuples(start, end)
	if err != nil {
		return nil, err
	}
	root, err := hashDataRootTuples(tuples)
	if err != nil {
		return nil, err
	}
	// Create data commitment
	return &ctypes.ResultDataCommitment{DataCommitment: root}, nil
}

// DataRootInclusionProof creates an inclusion proof for the data root of block
// height `height` in the set of blocks defined by `start` and `end`. The range
// is end exclusive.
func (env *Environment) DataRootInclusionProof(
	ctx *rpctypes.Context,
	height int64,
	start,
	end uint64,
) (*ctypes.ResultDataRootInclusionProof, error) {
	err := env.validateDataRootInclusionProofRequest(uint64(height), start, end)
	if err != nil {
		return nil, err
	}
	tuples, err := env.fetchDataRootTuples(start, end)
	if err != nil {
		return nil, err
	}
	proof, err := proveDataRootTuples(tuples, height)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultDataRootInclusionProof{Proof: *proof}, nil
}

// padBytes Pad bytes to given length
func padBytes(byt []byte, length int) ([]byte, error) {
	l := len(byt)
	if l > length {
		return nil, fmt.Errorf(
			"cannot pad bytes because length of bytes array: %d is greater than given length: %d",
			l,
			length,
		)
	}
	if l == length {
		return byt, nil
	}
	tmp := make([]byte, length)
	copy(tmp[length-l:], byt)
	return tmp, nil
}

// To32PaddedHexBytes takes a number and returns its hex representation padded to 32 bytes.
// Used to mimic the result of `abi.encode(number)` in Ethereum.
func To32PaddedHexBytes(number uint64) ([]byte, error) {
	hexRepresentation := strconv.FormatUint(number, 16)
	// Make sure hex representation has even length.
	// The `strconv.FormatUint` can return odd length hex encodings.
	// For example, `strconv.FormatUint(10, 16)` returns `a`.
	// Thus, we need to pad it.
	if len(hexRepresentation)%2 == 1 {
		hexRepresentation = "0" + hexRepresentation
	}
	hexBytes, hexErr := hex.DecodeString(hexRepresentation)
	if hexErr != nil {
		return nil, hexErr
	}
	paddedBytes, padErr := padBytes(hexBytes, 32)
	if padErr != nil {
		return nil, padErr
	}
	return paddedBytes, nil
}

// DataRootTuple contains the data that will be used to create the QGB commitments.
// The commitments will be signed by orchestrators and submitted to an EVM chain via a relayer.
// For more information: https://github.com/celestiaorg/quantum-gravity-bridge/blob/master/src/DataRootTuple.sol
type DataRootTuple struct {
	height   uint64
	dataRoot [32]byte
}

// EncodeDataRootTuple takes a height and a data root, and returns the equivalent of
// `abi.encode(...)` in Ethereum.
// The encoded type is a DataRootTuple, which has the following ABI:
//
//	{
//	  "components":[
//	     {
//	        "internalType":"uint256",
//	        "name":"height",
//	        "type":"uint256"
//	     },
//	     {
//	        "internalType":"bytes32",
//	        "name":"dataRoot",
//	        "type":"bytes32"
//	     },
//	     {
//	        "internalType":"structDataRootTuple",
//	        "name":"_tuple",
//	        "type":"tuple"
//	     }
//	  ]
//	}
//
// padding the hex representation of the height padded to 32 bytes concatenated to the data root.
// For more information, refer to:
// https://github.com/celestiaorg/quantum-gravity-bridge/blob/master/src/DataRootTuple.sol
func EncodeDataRootTuple(height uint64, dataRoot [32]byte) ([]byte, error) {
	paddedHeight, err := To32PaddedHexBytes(height)
	if err != nil {
		return nil, err
	}
	return append(paddedHeight, dataRoot[:]...), nil
}

// validateDataCommitmentRange runs basic checks on the asc sorted list of
// heights that will be used subsequently in generating data commitments over
// the defined set of heights.
func (env *Environment) validateDataCommitmentRange(start uint64, end uint64) error {
	if start == 0 {
		return fmt.Errorf("the first block is 0")
	}
	heightsRange := end - start
	if heightsRange > uint64(consts.DataCommitmentBlocksLimit) {
		return fmt.Errorf("the query exceeds the limit of allowed blocks %d", consts.DataCommitmentBlocksLimit)
	}
	if heightsRange == 0 {
		return fmt.Errorf("cannot create the data commitments for an empty set of blocks")
	}
	if start >= end {
		return fmt.Errorf("last block is smaller than first block")
	}
	// the data commitment range is end exclusive
	if end > uint64(env.BlockStore.Height())+1 {
		return fmt.Errorf(
			"end block %d is higher than current chain height %d",
			end,
			env.BlockStore.Height(),
		)
	}
	return nil
}

// hashDataRootTuples hashes a list of blocks data root tuples, i.e. height, data root and square size,
// then returns their merkle root.
func hashDataRootTuples(tuples []DataRootTuple) ([]byte, error) {
	dataRootEncodedTuples := make([][]byte, 0, len(tuples))
	for _, tuple := range tuples {
		encodedTuple, err := EncodeDataRootTuple(
			tuple.height,
			tuple.dataRoot,
		)
		if err != nil {
			return nil, err
		}
		dataRootEncodedTuples = append(dataRootEncodedTuples, encodedTuple)
	}
	root := merkle.HashFromByteSlices(dataRootEncodedTuples)
	return root, nil
}

// validateDataRootInclusionProofRequest validates the request to generate a data root
// inclusion proof.
func (env *Environment) validateDataRootInclusionProofRequest(height uint64, start uint64, end uint64) error {
	err := env.validateDataCommitmentRange(start, end)
	if err != nil {
		return err
	}
	if height < start || height >= end {
		return fmt.Errorf(
			"height %d should be in the end exclusive interval first_block %d last_block %d",
			height,
			start,
			end,
		)
	}
	return nil
}

// proveDataRootTuples returns the merkle inclusion proof for a height.
func proveDataRootTuples(tuples []DataRootTuple, height int64) (*merkle.Proof, error) {
	dataRootEncodedTuples := make([][]byte, 0, len(tuples))
	for _, tuple := range tuples {
		encodedTuple, err := EncodeDataRootTuple(
			tuple.height,
			tuple.dataRoot,
		)
		if err != nil {
			return nil, err
		}
		dataRootEncodedTuples = append(dataRootEncodedTuples, encodedTuple)
	}
	_, proofs := merkle.ProofsFromByteSlices(dataRootEncodedTuples)
	return proofs[height-int64(tuples[0].height)], nil
}

// </celestia-core>

// BlockResults gets ABCIResults at a given height.
// If no height is provided, it will fetch results for the latest block.
//
// Results are for the height of the block containing the txs.
// Thus response.results.deliver_tx[5] is the results of executing
// getBlock(h).Txs[5]
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/block_results
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
		TxsResults:            results.TxResults,
		FinalizeBlockEvents:   results.Events,
		ValidatorUpdates:      results.ValidatorUpdates,
		ConsensusParamUpdates: results.ConsensusParamUpdates,
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
		return nil, errors.New("block indexing is disabled")
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
	case "desc", "":
		sort.Slice(results, func(i, j int) bool { return results[i] > results[j] })

	case "asc":
		sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })

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

	apiResults := make([]*ctypes.ResultBlock, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		block := env.BlockStore.LoadBlock(results[i])
		if block != nil {
			blockMeta := env.BlockStore.LoadBlockMeta(block.Height)
			if blockMeta != nil {
				apiResults = append(apiResults, &ctypes.ResultBlock{
					Block:   block,
					BlockID: blockMeta.BlockID,
				})
			}
		}
	}

	return &ctypes.ResultBlockSearch{Blocks: apiResults, TotalCount: totalCount}, nil
}

// <celestia-core>

// fetchDataRootTuples takes an end exclusive range of heights and fetches its
// corresponding data root tuples.
func (env *Environment) fetchDataRootTuples(start, end uint64) ([]DataRootTuple, error) {
	tuples := make([]DataRootTuple, 0, end-start)
	for height := start; height < end; height++ {
		block := env.BlockStore.LoadBlock(int64(height))
		if block == nil {
			return nil, fmt.Errorf("couldn't load block %d", height)
		}
		tuples = append(tuples, DataRootTuple{
			height:   uint64(block.Height),
			dataRoot: *(*[32]byte)(block.DataHash),
		})
	}
	return tuples, nil
}

// </celestia-core>
