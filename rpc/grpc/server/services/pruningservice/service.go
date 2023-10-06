package pruningservice

import (
	context "context"
	"fmt"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/cometbft/cometbft/api/cometbft/services/pruning/v1"
	"github.com/cometbft/cometbft/internal/rpctrace"
	"github.com/cometbft/cometbft/libs/log"
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

func (s *pruningServiceServer) SetBlockIndexerRetainHeight(_ context.Context, request *v1.SetBlockIndexerRetainHeightRequest) (*v1.SetBlockIndexerRetainHeightResponse, error) {
	height := request.Height
	// Because we can't agree on a single type to represent tx indexer height.
	if height > uint64(math.MaxInt64) {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid height %d", height))
	}
	logger := s.logger.With("endpoint", "SetBlockIndexerRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	if err := s.pruner.SetBlockIndexerRetainHeight(int64(height)); err != nil {
		logger.Error("Cannot set block indexer retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to set block indexer retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.SetBlockIndexerRetainHeightResponse{}, nil
}

func (s *pruningServiceServer) GetBlockIndexerRetainHeight(_ context.Context, _ *v1.GetBlockIndexerRetainHeightRequest) (*v1.GetBlockIndexerRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetBLockIndexerRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	height, err := s.pruner.GetBlockIndexerRetainHeight()
	if err != nil {
		logger.Error("Cannot get block indexer retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to get block indexer retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.GetBlockIndexerRetainHeightResponse{Height: uint64(height)}, nil
}

func (s *pruningServiceServer) SetTxIndexerRetainHeight(_ context.Context, request *v1.SetTxIndexerRetainHeightRequest) (*v1.SetTxIndexerRetainHeightResponse, error) {
	height := request.Height
	// Because we can't agree on a single type to represent tx indexer height.
	if height > uint64(math.MaxInt64) {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid height %d", height))
	}
	logger := s.logger.With("endpoint", "SetTxIndexerRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	if err := s.pruner.SetTxIndexerRetainHeight(int64(height)); err != nil {
		logger.Error("Cannot set tx indexer retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to set tx indexer retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.SetTxIndexerRetainHeightResponse{}, nil
}

func (s *pruningServiceServer) GetTxIndexerRetainHeight(_ context.Context, _ *v1.GetTxIndexerRetainHeightRequest) (*v1.GetTxIndexerRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetTxIndexerRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	height, err := s.pruner.GetTxIndexerRetainHeight()
	if err != nil {
		logger.Error("Cannot get tx indexer retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to get tx indexer retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.GetTxIndexerRetainHeightResponse{Height: uint64(height)}, nil
}

// SetBlockRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) SetBlockRetainHeight(_ context.Context, req *v1.SetBlockRetainHeightRequest) (*v1.SetBlockRetainHeightResponse, error) {
	height := req.Height
	// Because we can't agree on a single type to represent block height.
	if height > uint64(math.MaxInt64) {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid height %d", height))
	}
	logger := s.logger.With("endpoint", "SetBlockRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	if err := s.pruner.SetCompanionBlockRetainHeight(int64(height)); err != nil {
		logger.Error("Cannot set block retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to set block retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.SetBlockRetainHeightResponse{}, nil
}

// GetBlockRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) GetBlockRetainHeight(_ context.Context, _ *v1.GetBlockRetainHeightRequest) (*v1.GetBlockRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetBlockRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	svcHeight, err := s.pruner.GetCompanionBlockRetainHeight()
	if err != nil {
		logger.Error("Cannot get block retain height stored by companion", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to get companion block retain height (see logs for trace ID: %s)", traceID)
	}
	appHeight, err := s.pruner.GetApplicationRetainHeight()
	if err != nil {
		logger.Error("Cannot get block retain height specified by application", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to get app block retain height (see logs for trace ID: %s)", traceID)
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
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	if err := s.pruner.SetABCIResRetainHeight(int64(height)); err != nil {
		logger.Error("Cannot set block results retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to set block results retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.SetBlockResultsRetainHeightResponse{}, nil
}

// GetBlockResultsRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) GetBlockResultsRetainHeight(_ context.Context, _ *v1.GetBlockResultsRetainHeightRequest) (*v1.GetBlockResultsRetainHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetBlockResultsRetainHeight")
	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return nil, status.Error(codes.Internal, "Internal server error - see logs for details")
	}
	height, err := s.pruner.GetABCIResRetainHeight()
	if err != nil {
		logger.Error("Cannot get block results retain height", "err", err, "traceID", traceID)
		return nil, status.Errorf(codes.Internal, "Failed to get block results retain height (see logs for trace ID: %s)", traceID)
	}
	return &v1.GetBlockResultsRetainHeightResponse{PruningServiceRetainHeight: uint64(height)}, nil
}
