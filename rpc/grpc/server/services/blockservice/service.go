package blockservice

import (
	context "context"
	"fmt"

	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/store"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type blockServiceServer struct {
	store *store.BlockStore
}

// New creates a new CometBFT version service server.
func New(store *store.BlockStore) blocksvc.BlockServiceServer {
	return &blockServiceServer{
		store,
	}
}

// GetByHeight implements v1.BlockServiceServer GetByHeight method
func (s *blockServiceServer) GetByHeight(_ context.Context, req *blocksvc.GetByHeightRequest) (*blocksvc.GetByHeightResponse, error) {
	var height int64

	// retrieve the last height in the store
	latestHeight := s.store.Height()

	// validate height parameter, if height is 0 or
	// the request is nil, then use the latest height
	if req.Height == 0 {
		height = latestHeight
	} else {
		height = req.Height
	}

	// check if the height requested is not higher
	// than the latest height in the store
	if height > latestHeight {
		st := status.New(codes.InvalidArgument, "invalid height")
		description := fmt.Sprintf("height requested (%d) is higher than the latest available height (%d)", height, latestHeight)
		v := &errdetails.BadRequest_FieldViolation{
			Field:       "height",
			Description: description,
		}
		br := &errdetails.BadRequest{}
		br.FieldViolations = append(br.FieldViolations, v)
		st, err := st.WithDetails(br)
		// if there is an issue adding details just return a simple
		// error message without details
		if err != nil {
			err := status.Error(codes.InvalidArgument, description)
			return nil, err
		}
		return nil, st.Err()
	}
	var blockProto *cmtproto.Block
	var blockIDProto cmtproto.BlockID

	block := s.store.LoadBlock(height)
	blockProto, err := block.ToProto()
	if err != nil {
		description := fmt.Sprintf("block at height %d not found", height)
		err := status.Error(codes.NotFound, description)
		return nil, err
	}

	blockMeta := s.store.LoadBlockMeta(height)

	blockIDProto = blockMeta.BlockID.ToProto()

	return &blocksvc.GetByHeightResponse{
		BlockId: &blockIDProto,
		Block:   blockProto,
	}, nil
}

// GetLastestHeight implements v1.BlockServiceServer GetLatestHeight method
func (s *blockServiceServer) GetLatestHeight(req *blocksvc.GetLatestHeightRequest, srv blocksvc.BlockService_GetLatestHeightServer) error {
	err := status.Error(codes.Unimplemented, "not implemented")
	return err
}
