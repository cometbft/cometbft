package client

import (
	"context"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	protoType "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cosmos/gogoproto/grpc"
	"time"
)

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlock(ctx context.Context, request *v1.GetBlockRequest) (*v1.GetBlockResponse, error)
}

type blockServiceClient struct {
	client v1.BlockServiceClient
}

func newBlockServiceClient(conn grpc.ClientConn) BlockServiceClient {
	return &blockServiceClient{
		client: v1.NewBlockServiceClient(conn),
	}
}

// GetBlock implements BlockServiceClient
func (c *blockServiceClient) GetBlock(ctx context.Context, request *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	blockID := protoType.BlockID{
		Hash:          nil,
		PartSetHeader: protoType.PartSetHeader{},
	}

	block := protoType.Block{
		Header: protoType.Header{
			Version:            cmtversion.Consensus{},
			ChainID:            "",
			Height:             request.Height,
			Time:               time.Now(),
			LastCommitHash:     nil,
			DataHash:           nil,
			ValidatorsHash:     nil,
			NextValidatorsHash: nil,
			ConsensusHash:      nil,
			AppHash:            nil,
			LastResultsHash:    nil,
			EvidenceHash:       nil,
			ProposerAddress:    nil,
		},
		Data:       protoType.Data{},
		Evidence:   protoType.EvidenceList{},
		LastCommit: nil,
	}
	blockResp := &v1.GetBlockResponse{
		BlockId: &blockID,
		Block:   &block,
	}
	return blockResp, nil
}

type disabledBlockServiceClient struct{}

func newDisabledBlockServiceClient() BlockServiceClient {
	return &disabledBlockServiceClient{}
}

// GetBlock implements BlockServiceClient
func (*disabledBlockServiceClient) GetBlock(context.Context, *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	panic("block service client is disabled")
}
