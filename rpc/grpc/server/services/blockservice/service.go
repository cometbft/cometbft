package blockservice

import (
	context "context"
	"fmt"

	"github.com/cometbft/cometbft/internal/rpctrace"
	"github.com/cometbft/cometbft/libs/log"
	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
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
	return &blockServiceServer{
		store:    store,
		eventBus: eventBus,
		logger:   logger.With("service", "BlockService"),
	}
}

// GetByHeight implements v1.BlockServiceServer GetByHeight method
func (s *blockServiceServer) GetByHeight(_ context.Context, req *blocksvc.GetByHeightRequest) (*blocksvc.GetByHeightResponse, error) {
	height, err := validateOrUpdateBlockHeight(
		req.Height,
		s.store.Base(),
		s.store.Height(),
	)
	if err != nil {
		return nil, err
	}

	block := s.store.LoadBlock(height)
	blockProto, err := block.ToProto()
	if err != nil {
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

	traceID, err := rpctrace.New()
	if err != nil {
		logger.Error("Error generating RPC trace ID", "err", err)
		return status.Error(codes.Internal, "Internal server error")
	}

	// The trace ID is reused as a unique subscriber ID
	sub, err := s.eventBus.Subscribe(context.Background(), traceID, types.QueryForEvent(types.EventNewBlock), 1)
	if err != nil {
		logger.Error("Cannot subscribe to new block events", "err", err, "traceID", traceID)
		return status.Errorf(codes.Internal, "Cannot subscribe to new block events (see logs for trace ID: %s)", traceID)
	}

	for {
		select {
		case msg := <-sub.Out():
			height, err := getHeightFromMsg(msg)
			if err != nil {
				logger.Error("Failed to extract height from subscription message", "err", err, "traceID", traceID)
				return status.Errorf(codes.Internal, "Internal server error (see logs for trace ID: %s)", traceID)
			}
			if err := stream.Send(&blocksvc.GetLatestHeightResponse{Height: height}); err != nil {
				logger.Error("Failed to stream new block", "err", err, "height", height, "traceID", traceID)
				return status.Errorf(codes.Unavailable, "Cannot send stream response (see logs for trace ID: %s)", traceID)
			}
		case <-sub.Canceled():
			switch sub.Err() {
			case cmtpubsub.ErrUnsubscribed:
				return status.Error(codes.Canceled, "Subscription terminated")
			case nil:
				return status.Error(codes.Canceled, "Subscription canceled without errors")
			default:
				logger.Info("Subscription canceled with errors", "err", sub.Err(), "traceID", traceID)
				return status.Errorf(codes.Canceled, "Subscription canceled with errors (see logs for trace ID: %s)", traceID)
			}
		default:
			continue
		}
		if sub.Err() != nil {
			logger.Error("New block subscription error", "err", sub.Err(), "traceID", traceID)
			return status.Errorf(codes.Internal, "New block subscription error (see logs for trace ID: %s)", traceID)
		}
	}
}

func validateOrUpdateBlockHeight(height, baseHeight, latestHeight int64) (int64, error) {
	switch {
	case height == 0:
		return latestHeight, nil
	case height < 0:
		return -1, status.Error(codes.InvalidArgument, "Height cannot be negative")
	case height < baseHeight:
		return -1, status.Errorf(codes.InvalidArgument, "Requested height %d is below base height %d", height, baseHeight)
	case height > latestHeight:
		return -1, status.Errorf(codes.InvalidArgument, "Requested height %d is higher than latest height %d", height, latestHeight)
	}
	return height, nil
}

func getHeightFromMsg(msg cmtpubsub.Message) (int64, error) {
	switch eventType := msg.Data().(type) {
	case types.EventDataNewBlock:
		return eventType.Block.Height, nil
	default:
		return -1, fmt.Errorf("unexpected event type: %v", eventType)
	}
}
