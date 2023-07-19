package blockservice

import (
	context "context"
	"fmt"

	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type blockServiceServer struct {
	store    *store.BlockStore
	eventBus *types.EventBus
}

// New creates a new CometBFT version service server.
func New(store *store.BlockStore, eventBus *types.EventBus) blocksvc.BlockServiceServer {
	return &blockServiceServer{
		store,
		eventBus,
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
	} else if req.Height < 0 {
		description := fmt.Sprintf("got negative height (%d), please specify a height >= 0", req.Height)
		err := status.Error(codes.InvalidArgument, description)
		return nil, err
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

// GetLatestHeight implements v1.BlockServiceServer GetLatestHeight method
func (s *blockServiceServer) GetLatestHeight(_ *blocksvc.GetLatestHeightRequest, stream blocksvc.BlockService_GetLatestHeightServer) error {

	// Generate a unique subscriber ID based on client address, if not possible, generate a random ID using UUID
	// The subscriber needs to be unique across all clients
	var subscriber string
	ctx := stream.Context()
	p, ok := peer.FromContext(ctx)
	if ok {
		subscriber = fmt.Sprintf("subscriber_%s", p.Addr.String())
	} else {
		id, err := uuid.NewUUID()
		if err != nil {
			// cannot generate unique id
			err := status.Error(codes.Internal, "error generating a subscriber id, cannot subscribe to new block events")
			return err
		}
		subscriber = fmt.Sprintf("subscriber_%s", id.String())
	}

	var sub types.Subscription
	sub, err := s.eventBus.Subscribe(context.Background(), subscriber, types.QueryForEvent(types.EventNewBlock), 1)
	if err != nil {
		err := status.Error(codes.Internal, "cannot subscribe to new block events")
		return err
	}

	for {
		select {
		case msg := <-sub.Out():
			switch eventType := msg.Data().(type) {
			case types.EventDataNewBlock:
				if err := stream.Send(&blocksvc.GetLatestHeightResponse{Height: eventType.Block.Height}); err != nil {
					err := status.Error(codes.Unimplemented, "cannot send stream response")
					return err
				}
			}
		case <-sub.Canceled():
			if sub.Err() == cmtpubsub.ErrUnsubscribed {
				err := status.Error(codes.Canceled, "client unsubscribed")
				return err
			}
		default:
			continue
		}
		err := sub.Err()
		if err != nil {
			err := status.Error(codes.Internal, "error in new block subscription")
			return err
		}
	}
}
