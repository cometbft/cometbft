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
	log := logger.With("service", "BlockService")
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
		return nil, status.Error(codes.InvalidArgument, "Negative height request")
	} else {
		height = req.Height
	}

	// check if the height requested is not higher
	// than the latest height in the store
	if height > latestHeight {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("height requested %d higher than latest height %d", height, latestHeight))
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
			switch eventType := msg.Data().(type) {
			case types.EventDataNewBlock:
				if err := stream.Send(&blocksvc.GetLatestHeightResponse{Height: eventType.Block.Height}); err != nil {
					logger.Error("Failed to stream new block", "err", err, "height", eventType.Block.Height, "traceID", traceID)
					return status.Errorf(codes.Unavailable, "Cannot send stream response (see logs for trace ID: %s)", traceID)
				}
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
