package blockservice

import (
	context "context"
	"fmt"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/store"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	block := s.store.LoadBlock(height)
	blockProto, err := block.ToProto()
	if err != nil {
		return nil, err
	}

	blockID := s.store.LoadBlockMeta(height)
	return &v1.GetBlockResponse{
		BlockId: blockID.BlockID.ToProto(),
		Block:   *blockProto,
	}, nil
}
