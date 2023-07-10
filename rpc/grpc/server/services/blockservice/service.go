package blockservice

import (
	context "context"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	proto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
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

	partSetHeader := proto.PartSetHeader{
		Total: resp.BlockID.PartSetHeader.Total,
		Hash:  resp.BlockID.PartSetHeader.Hash,
	}

	blockID := proto.BlockID{
		Hash:          resp.BlockID.Hash,
		PartSetHeader: partSetHeader,
	}

	lastBlockIDPartSetHeader := proto.PartSetHeader{
		Total: resp.Block.LastBlockID.PartSetHeader.Total,
		Hash:  resp.Block.LastBlockID.PartSetHeader.Hash,
	}

	lastBlockID := proto.BlockID{
		Hash:          resp.Block.LastBlockID.Hash,
		PartSetHeader: lastBlockIDPartSetHeader,
	}

	version := cmtversion.Consensus{
		Block: resp.Block.Version.Block,
		App:   resp.Block.Version.Block,
	}

	header := proto.Header{
		Version:            version,
		ChainID:            resp.Block.ChainID,
		Height:             resp.Block.Height,
		Time:               resp.Block.Time,
		LastBlockId:        lastBlockID,
		LastCommitHash:     resp.Block.LastCommitHash,
		DataHash:           resp.Block.DataHash,
		ValidatorsHash:     resp.Block.ValidatorsHash,
		NextValidatorsHash: resp.Block.NextValidatorsHash,
		ConsensusHash:      resp.Block.ConsensusHash,
		AppHash:            resp.Block.AppHash,
		LastResultsHash:    resp.Block.LastResultsHash,
		EvidenceHash:       resp.Block.EvidenceHash,
		ProposerAddress:    resp.Block.ProposerAddress,
	}

	data := proto.Data{
		Txs: resp.Block.Data.Txs.ToSliceOfBytes(),
	}

	var commitSigs []proto.CommitSig

	for _, signature := range resp.Block.LastCommit.Signatures {
		commit := proto.CommitSig{
			BlockIdFlag:      proto.BlockIDFlag(signature.BlockIDFlag),
			ValidatorAddress: signature.ValidatorAddress,
			Timestamp:        signature.Timestamp,
			Signature:        signature.Signature,
		}
		commitSigs = append(commitSigs, commit)
	}

	lastCommitBlockIDPartSetHeader := proto.PartSetHeader{
		Total: resp.Block.LastCommit.BlockID.PartSetHeader.Total,
		Hash:  resp.Block.LastCommit.BlockID.PartSetHeader.Hash,
	}

	lastCommitBlockID := proto.BlockID{
		Hash:          resp.Block.LastCommit.BlockID.Hash,
		PartSetHeader: lastCommitBlockIDPartSetHeader,
	}

	var lastCommit = proto.Commit{
		Height:     resp.Block.LastCommit.Height,
		Round:      resp.Block.LastCommit.Round,
		BlockID:    lastCommitBlockID,
		Signatures: commitSigs,
	}
	block := proto.Block{
		Header:     header,
		Data:       data,
		Evidence:   proto.EvidenceList{}, //TODO: Convert evidence to proto
		LastCommit: &lastCommit,
	}

	return &v1.GetBlockResponse{
		BlockId: blockID,
		Block:   block,
	}, nil
}
