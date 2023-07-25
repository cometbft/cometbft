package blockservice

import (
	context "context"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type blockServiceServer struct {
	store    *store.BlockStore
	eventBus *types.EventBus
	logger   log.Logger
}

// New creates a new CometBFT version service server.
func New(store *store.BlockStore, eventBus *types.EventBus, logger log.Logger) blocksvc.BlockServiceServer {
	log := logger.With("module", "grpc-block-service")
	return &blockServiceServer{
		store,
		eventBus,
		log,
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
		errMsg := fmt.Sprintf("got negative height (%d), please specify a height >= 0", req.Height)
		s.logger.Error("GetByHeight", "err", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	} else {
		height = req.Height
	}

	// check if the height requested is not higher
	// than the latest height in the store
	if height > latestHeight {
		errMsg := fmt.Sprintf("height requested (%d) is higher than the latest available height (%d)", height, latestHeight)
		s.logger.Error("GetByHeight", "err", errMsg)
		return nil, status.Errorf(codes.InvalidArgument, errMsg)
	}

	block := s.store.LoadBlock(height)
	blockProto, err := block.ToProto()
	if err != nil {
		errMsg := fmt.Sprintf("block at height %d not found", height)
		s.logger.Error("GetByHeight", "err", errMsg)
		return nil, status.Errorf(codes.NotFound, errMsg)
	}

	blockMeta := s.store.LoadBlockMeta(height)

	blockIDProto := blockMeta.BlockID.ToProto()

	return &blocksvc.GetByHeightResponse{
		BlockId: &blockIDProto,
		Block:   blockProto,
	}, nil
}

// GetLatestHeight implements v1.BlockServiceServer GetLatestHeight method
func (s *blockServiceServer) GetLatestHeight(_ *blocksvc.GetLatestHeightRequest, stream blocksvc.BlockService_GetLatestHeightServer) error {
	// Generate a unique subscriber ID using a UUID
	// The subscriber needs to be unique across all clients
	id, err := uuid.NewUUID()
	if err != nil {
		// cannot generate unique id
		errMsg := "error generating a subscriber id, cannot subscribe to new block events"
		s.logger.Error("GetLatestHeight", "err", errMsg)
		return status.Error(codes.Internal, errMsg)
	}
	subscriber := id.String()

	sub, err := s.eventBus.Subscribe(context.Background(), subscriber, types.QueryForEvent(types.EventNewBlock), 1)
	if err != nil {
		errMsg := "cannot subscribe to new block events"
		s.logger.Error("GetLatestHeight", "err", errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	for {
		select {
		case msg := <-sub.Out():
			switch eventType := msg.Data().(type) {
			case types.EventDataNewBlock:
				if err := stream.Send(&blocksvc.GetLatestHeightResponse{Height: eventType.Block.Height}); err != nil {
					s.logger.Error("Failed to stream new block height", "err", err, "height", eventType.Block.Height, "subscriber", subscriber)
					return status.Error(codes.Unavailable, "cannot send stream response")
				}
				s.logger.Debug("GetLatestHeight", "msg", fmt.Sprintf("streamed new block height %d", eventType.Block.Height))
			}
		case <-sub.Canceled():
			switch sub.Err() {
			case cmtpubsub.ErrUnsubscribed:
				s.logger.Error("GetLatestHeight", "err", fmt.Sprintf("subscriber %s unsubscribed", subscriber))
				return status.Error(codes.Canceled, "client unsubscribed")
			case nil:
				s.logger.Info("GetLatestHeight", "msg", fmt.Sprintf("subscription for %s canceled without errors", subscriber))
				return status.Error(codes.Canceled, "subscription canceled without errors")
			default:
				s.logger.Info("GetLatestHeight", "msg", fmt.Sprintf("subscription for %s canceled with errors %s", subscriber, sub.Err()))
				return status.Error(codes.Canceled, "subscription canceled with errors")
			}
		default:
			continue
		}
		if sub.Err() != nil {
			s.logger.Error("GetLatestHeight", "err", fmt.Sprintf("error in new block subscription for subscriber %s", subscriber))
			return status.Error(codes.Internal, "error in new block subscription")
		}
	}
}
