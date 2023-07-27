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
	log := logger.With("service", "BlockService")
	return &blockServiceServer{
		store,
		eventBus,
		log,
	}
}

// GetByHeight implements v1.BlockServiceServer GetByHeight method
func (s *blockServiceServer) GetByHeight(_ context.Context, req *blocksvc.GetByHeightRequest) (*blocksvc.GetByHeightResponse, error) {
	logger := s.logger.With("endpoint", "GetByHeight")

	var height int64

	// retrieve the last height in the store
	latestHeight := s.store.Height()

	// validate height parameter, if height is 0 or
	// the request is nil, then use the latest height
	if req.Height == 0 {
		height = latestHeight
	} else if req.Height < 0 {
		logger.Error("negative height request", "height", req.Height)
		return nil, status.Error(codes.InvalidArgument, "Negative height request")
	} else {
		height = req.Height
	}

	// check if the height requested is not higher
	// than the latest height in the store
	if height > latestHeight {
		logger.Error("height requested higher than latest height", "requested", height, "latest", latestHeight)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("height requested %d higher than latest height %d", height, latestHeight))
	}

	block := s.store.LoadBlock(height)
	blockProto, err := block.ToProto()
	if err != nil {
		logger.Error("block not found", "height", height)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Block not found for height %d", height))
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
	logger := s.logger.With("endpoint", "GetLatestHeight")

	// Generate a unique subscriber ID using a UUID
	// The subscriber needs to be unique across all clients
	id, err := uuid.NewUUID()
	if err != nil {
		// cannot generate unique id
		logger.Error("error generating subscriber id", "err", err)
		return status.Error(codes.Internal, "Error generating subscriber id")
	}
	subscriber := id.String()

	sub, err := s.eventBus.Subscribe(context.Background(), subscriber, types.QueryForEvent(types.EventNewBlock), 1)
	if err != nil {
		logger.Error("cannot subscribe to new block events", "err", err)
		return status.Error(codes.Internal, "Cannot subscribe to new block events")
	}

	for {
		select {
		case msg := <-sub.Out():
			switch eventType := msg.Data().(type) {
			case types.EventDataNewBlock:
				if err := stream.Send(&blocksvc.GetLatestHeightResponse{Height: eventType.Block.Height}); err != nil {
					logger.Error("failed to stream new block", "err", err, "height", eventType.Block.Height, "subscriber", subscriber)
					return status.Error(codes.Unavailable, "Cannot send stream response")
				}
				logger.Debug("streamed new block", "height", eventType.Block.Height)
			}
		case <-sub.Canceled():
			switch sub.Err() {
			case cmtpubsub.ErrUnsubscribed:
				logger.Error("subscription terminated", "subscriber", subscriber)
				return status.Error(codes.Canceled, "Subscription terminated")
			case nil:
				logger.Info("subscription canceled without errors", "subscriber", subscriber)
				return status.Error(codes.Canceled, "Subscription canceled without errors")
			default:
				logger.Info("subscription canceled with errors", "err", sub.Err(), "subscriber", subscriber)
				return status.Error(codes.Canceled, "Subscription canceled with errors")
			}
		default:
			continue
		}
		if sub.Err() != nil {
			logger.Error("new block subscription error", "err", sub.Err(), "subscriber", subscriber)
			return status.Error(codes.Internal, "New block subscription error")
		}
	}
}
