package blockservice

import (
	context "context"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/store"
)

type blockServiceServer struct {
	store *store.BlockStore
}

// New creates a new CometBFT version service server.
func New(store *store.BlockStore) v1.BlockServiceServer {
	return &blockServiceServer{
		store,
	}
}

// GetBlock implements v1.BlockServiceServer
func (s *blockServiceServer) GetBlock(ctx context.Context, req *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	block := s.store.LoadBlock(req.Height)
	blockProto, err := block.ToProto()
	if err != nil {
		return nil, err
	}

	blockID := s.store.LoadBlockMeta(req.Height)
	return &v1.GetBlockResponse{
		BlockId: blockID.BlockID.ToProto(),
		Block:   *blockProto,
	}, nil
}
