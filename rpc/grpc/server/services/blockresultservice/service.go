package blockresultservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block_results/v1"
)

type blockResultsService struct {
	stateStore sm.Store
	blockStore *store.BlockStore
}

// New creates a new CometBFT block results service server.
func New(bs *store.BlockStore, ss sm.Store) v1.BlockResultsServiceServer {
	return &blockResultsService{stateStore: ss, blockStore: bs}
}

// GetBlockResults returns the block results of the requested height.
// If no height is given, the block results for the latest height are returned.
func (s *blockResultsService) GetBlockResults(_ context.Context, req *v1.GetBlockResultsRequest) (*v1.GetBlockResultsResponse, error) {
	height := req.Height
	latestHeight := s.blockStore.Height()
	if req.Height > latestHeight || req.Height < 0 {
		return &v1.GetBlockResultsResponse{}, status.Error(codes.InvalidArgument, "Height is invalid.")
	} else if req.Height == 0 {
		height = latestHeight
	}

	res, err := s.stateStore.LoadFinalizeBlockResponse(height)
	if err != nil {
		return &v1.GetBlockResultsResponse{}, err
	}

	return &v1.GetBlockResultsResponse{
		Height:              height,
		TxsResults:          res.TxResults,
		FinalizeBlockEvents: res.Events,
		ValidatorUpdates:    res.ValidatorUpdates,
		AppHash:             res.AppHash,
	}, nil
}
