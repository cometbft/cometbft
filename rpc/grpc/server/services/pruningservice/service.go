package pruningservice

import (
	context "context"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/pruning/v1"
	sm "github.com/cometbft/cometbft/state"
)

type pruningServiceServer struct {
	pruner *sm.Pruner
}

// New creates a new CometBFT pruning service server.
func New(pruner *sm.Pruner) v1.PruningServiceServer {
	return &pruningServiceServer{
		pruner: pruner,
	}
}

// SetBlockRetainHeight implements v1.PruningServiceServer.
func (s *pruningServiceServer) SetBlockRetainHeight(_ context.Context, req *v1.SetBlockRetainHeightRequest) (*v1.SetBlockRetainHeightResponse, error) {
	panic("unimplemented")
}

// GetBlockRetainHeight implements v1.PruningServiceServer.
func (*pruningServiceServer) GetBlockRetainHeight(context.Context, *v1.GetBlockRetainHeightRequest) (*v1.GetBlockRetainHeightResponse, error) {
	panic("unimplemented")
}

// SetBlockResultsRetainHeight implements v1.PruningServiceServer.
func (*pruningServiceServer) SetBlockResultsRetainHeight(context.Context, *v1.SetBlockResultsRetainHeightRequest) (*v1.SetBlockResultsRetainHeightResponse, error) {
	panic("unimplemented")
}

// GetBlockResultsRetainHeight implements v1.PruningServiceServer.
func (*pruningServiceServer) GetBlockResultsRetainHeight(context.Context, *v1.GetBlockResultsRetainHeightRequest) (*v1.GetBlockResultsRetainHeightResponse, error) {
	panic("unimplemented")
}
