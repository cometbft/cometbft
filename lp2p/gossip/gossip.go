package gossip

import (
	"context"

	"github.com/cometbft/cometbft/libs/log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
)

// Service wraps lib-p2p gossip-sub.
// Supposed to be provisioned once on startup.
// NOTE: Not goroutine safe.
type Service struct {
	ctx    context.Context
	self   peer.ID
	ps     *pubsub.PubSub
	items  map[protocol.ID]item
	logger log.Logger
}

// Handler is a function that handles a gossip message.
type Handler func(protocolID protocol.ID, msg *pubsub.Message) error

type item struct {
	protocolID protocol.ID
	topic      *pubsub.Topic
	sub        *pubsub.Subscription
}

// New Service constructor.
func New(ctx context.Context, host host.Host, logger log.Logger) (*Service, error) {
	// todo configure options
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create gossip sub")
	}

	return &Service{
		ctx:    ctx,
		self:   host.ID(),
		ps:     pubSub,
		items:  make(map[protocol.ID]item),
		logger: logger,
	}, nil
}

// Join joins a gossip topic and registers a handler for incoming messages.
// We use protocolIDs for topics.
func (s *Service) Join(protocolID protocol.ID, handler Handler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}

	if _, ok := s.items[protocolID]; ok {
		return errors.New("already joined")
	}

	// todo: do we want to sign messages?
	// it's required to ensure that it's not possible to DDOS the network
	// by "relaying" malicious messages that pretend to be sent by other authentic peers
	topic, err := s.ps.Join(string(protocolID))
	if err != nil {
		return errors.Wrap(err, "unable to join")
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return errors.Wrap(err, "unable to subscribe")
	}

	i := item{
		protocolID: protocolID,
		topic:      topic,
		sub:        sub,
	}

	s.items[protocolID] = i

	s.runReceiveLoop(&i, handler)

	return nil
}

func (s *Service) Close() {
	for _, item := range s.items {
		item.sub.Cancel()

		if err := item.topic.Close(); err != nil {
			s.logger.Error("Error closing gossip topic", "protocol", item.protocolID, "err", err)
		}
	}
}

// Broadcast publishes a message to a gossip topic.
func (s *Service) Broadcast(protocolID protocol.ID, data []byte) error {
	if _, ok := s.items[protocolID]; !ok {
		return errors.Errorf("protocol %s not found", protocolID)
	}

	// todo explore publish options
	err := s.items[protocolID].topic.Publish(s.ctx, data)
	if err != nil {
		return errors.Wrap(err, "unable to publish message")
	}

	s.logger.Debug("Gossiped message", "protocol", protocolID, "data_len", len(data))

	return nil
}

func (s *Service) runReceiveLoop(item *item, handler Handler) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				s.logger.Error(
					"Panic in (*Registry).runReceiveLoop",
					"panic", p,
					"protocol", item.protocolID,
				)
			}
		}()

		err := s.runReceiveLoopBlocking(item, handler)
		switch {
		case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
			s.logger.Info("Context canceled or deadline exceeded for gossip", "protocol", item.protocolID)
		case err != nil:
			s.logger.Error("Error in gossip receive loop", "protocol", item.protocolID, "err", err)
		}
	}()
}

func (s *Service) runReceiveLoopBlocking(item *item, handler Handler) error {
	for {
		msg, err := item.sub.Next(s.ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get next message")
		}

		if msg.Local || msg.ReceivedFrom == s.self {
			s.logger.Debug(
				"Skipping local gossip message",
				"protocol", item.protocolID,
				"message_id", msg.ID,
			)
			continue
		}

		// ensure messages are processed concurrently
		go s.handleMessage(item, msg, handler)
	}
}

func (s *Service) handleMessage(item *item, msg *pubsub.Message, handler Handler) {
	defer func() {
		if p := recover(); p != nil {
			s.logger.Error(
				"Panic in (*Registry).handleMessage",
				"panic", p,
				"protocol", item.protocolID,
				"message_id", msg.ID,
			)
		}
	}()

	if err := handler(item.protocolID, msg); err != nil {
		s.logger.Error(
			"Error in gossip message handler",
			"protocol", item.protocolID,
			"message_id", msg.ID,
			"err", err,
		)
	}
}
