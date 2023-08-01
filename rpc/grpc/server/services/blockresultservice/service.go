package blockresultservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cometbft/cometbft/libs/log"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"

	brs "github.com/cometbft/cometbft/proto/tendermint/services/block_results/v1"
)

type blockResultsService struct {
	stateStore sm.Store
	blockStore *store.BlockStore
	logger     log.Logger
}

// New creates a new CometBFT block results service server.
func New(bs *store.BlockStore, ss sm.Store, logger log.Logger) brs.BlockResultsServiceServer {
	return &blockResultsService{
		stateStore: ss,
		blockStore: bs,
		logger:     logger.With("service", "BlockResultsService"),
	}
}

// GetBlockResults returns the block results of the requested height.
// If no height is given, the block results for the latest height are returned.
func (s *blockResultsService) GetBlockResults(_ context.Context, req *brs.GetBlockResultsRequest) (*brs.GetBlockResultsResponse, error) {
	logger := s.logger.With("endpoint", "GetBlockResults")
	height := req.Height
	ss, err := s.stateStore.Load()
	if err != nil {
		logger.Error("Error loading store", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	if req.Height > ss.LastBlockHeight || req.Height < 0 {
		logger.Error("Error validating GetBlockResults request height")
		return nil, status.Errorf(codes.InvalidArgument, "Height must be between 0 and the last effective height (%d)", ss.LastBlockHeight)
	} else if req.Height == 0 {
		height = ss.LastBlockHeight
	}

	res, err := s.stateStore.LoadFinalizeBlockResponse(height)
	if err != nil {
		logger.Error("Error fetching BlockResults", "height", height, "err", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &brs.GetBlockResultsResponse{
		Height:              height,
		TxsResults:          res.TxResults,
		FinalizeBlockEvents: formatProtoToRef(res.Events),
		ValidatorUpdates:    formatProtoToRef(res.ValidatorUpdates),
		AppHash:             res.AppHash,
	}, nil
}

func formatProtoToRef[T any](collection []T) []*T {
	res := []*T{}
	for i := range collection {
		res = append(res, &collection[i])
	}
	return res
}
