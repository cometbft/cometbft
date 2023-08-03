package pruningservice

import (
	context "context"
	"fmt"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cometbft/cometbft/libs/log"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/pruning/v1"
	sm "github.com/cometbft/cometbft/state"
)

type pruningServiceServer struct {
	pruner *sm.Pruner
	logger log.Logger
}

// New creates a new CometBFT pruning service server.
func New(pruner *sm.Pruner, logger log.Logger) v1.PruningServiceServer {
	return &pruningServiceServer{
		pruner: pruner,
		logger: logger.With("service", "PruningService"),
	}
}

// SetBlockRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) SetBlockRetainHeight(_ context.Context, req *v1.SetBlockRetainHeightRequest) (*v1.SetBlockRetainHeightResponse, error) {
	height := req.Height
	// Because we can't agree on a single type to represent block height.
	if height > uint64(math.MaxInt64) {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid height %d", height))
	}
	logger := s.logger.With("endpoint", "SetBlockRetainHeight")
	err := s.pruner.SetCompanionRetainHeight(int64(height))
	if err != nil {
		logger.Error("Cannot set block retain height", "err", err)
		return nil, status.Errorf(codes.Internal, "Failed to set block retain height")
	}
	return &v1.SetBlockRetainHeightResponse{}, nil
}

// GetBlockRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) GetBlockRetainHeight(_ context.Context, req *v1.GetBlockRetainHeightRequest) (*v1.GetBlockRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetBlockRetainHeight")
	svcHeight, err := s.pruner.GetCompanionBlockRetainHeight()
	if err != nil {
		logger.Error("Cannot get block retain height stored by companion", "err", err)
		return nil, status.Errorf(codes.Internal, "Failed to get companion block retain height")
	}
	appHeight, err := s.pruner.GetApplicationRetainHeight()
	if err != nil {
		logger.Error("Cannot get block retain height specified by application", "err", err)
		return nil, status.Errorf(codes.Internal, "Failed to get app block retain height")
	}
	return &v1.GetBlockRetainHeightResponse{
		PruningServiceRetainHeight: uint64(svcHeight),
		AppRetainHeight:            uint64(appHeight),
	}, nil
}

// SetBlockResultsRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) SetBlockResultsRetainHeight(_ context.Context, req *v1.SetBlockResultsRetainHeightRequest) (*v1.SetBlockResultsRetainHeightResponse, error) {
	height := req.Height
	// Because we can't agree on a single type to represent block height.
	if height > uint64(math.MaxInt64) {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid height %d", height))
	}
	logger := s.logger.With("endpoint", "SetBlockResultsRetainHeight")
	err := s.pruner.SetABCIResRetainHeight(int64(height))
	if err != nil {
		logger.Error("Cannot set block results retain height", "err", err)
		return nil, status.Errorf(codes.Internal, "Failed to set block results retain height")
	}
	return &v1.SetBlockResultsRetainHeightResponse{}, nil
}

// GetBlockResultsRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) GetBlockResultsRetainHeight(_ context.Context, req *v1.GetBlockResultsRetainHeightRequest) (*v1.GetBlockResultsRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetBlockResultsRetainHeight")
	height, err := s.pruner.GetABCIResRetainHeight()
	if err != nil {
		logger.Error("Cannot get block results retain height", "err", err)
		return nil, status.Errorf(codes.Internal, "Failed to get block results retain height")
	}
	return &v1.GetBlockResultsRetainHeightResponse{PruningServiceRetainHeight: uint64(height)}, nil
}
