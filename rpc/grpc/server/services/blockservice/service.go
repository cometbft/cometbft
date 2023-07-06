package blockservice

import (
	context "context"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
)

type blockServiceServer struct{}

// New creates a new CometBFT version service server.
func New() v1.BlockServiceServer {
	return &blockServiceServer{}
}

// GetBlock implements v1.BlockServiceServer
func (s *blockServiceServer) GetBlock(context.Context, *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	return &v1.GetBlockResponse{}, nil
}
