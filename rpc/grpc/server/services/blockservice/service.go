package blockservice

import (
	context "context"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/rpc/core"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

type blockServiceServer struct {
	nodeEnv *core.Environment
}

// New creates a new CometBFT version service server.
func New(env *core.Environment) v1.BlockServiceServer {
	return &blockServiceServer{nodeEnv: env}
}

// GetBlock implements v1.BlockServiceServer
func (s *blockServiceServer) GetBlock(ctx context.Context, req *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	resp, err := s.nodeEnv.Block(&rpctypes.Context{}, &req.Height)
	if err != nil {
		return nil, err
	}

	block, err := resp.Block.ToProto()
	if err != nil {
		return nil, err
	}
	return &v1.GetBlockResponse{
		BlockId: resp.BlockID.ToProto(),
		Block:   *block,
	}, nil
}
