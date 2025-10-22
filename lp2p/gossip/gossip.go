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

// Registry wraps lib-p2p gossip-sub. Supposed to be provisioned once on startup.
// NOTE: Not goroutine safe.
type Registry struct {
	ctx    context.Context
	self   peer.ID
	ps     *pubsub.PubSub
	items  map[protocol.ID]item
	logger log.Logger
}

type Handler func(protocolID protocol.ID, msg *pubsub.Message) error

type item struct {
	protocolID protocol.ID
	topic      *pubsub.Topic
	sub        *pubsub.Subscription
}

func New(
	ctx context.Context,
	host host.Host,
	logger log.Logger,
) (*Registry, error) {
	// todo configure options
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create gossip sub")
	}

	return &Registry{
		ctx:    ctx,
		self:   host.ID(),
		ps:     pubSub,
		items:  make(map[protocol.ID]item),
		logger: logger,
	}, nil
}

func (r *Registry) Join(protocolID protocol.ID, handler Handler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}

	if _, ok := r.items[protocolID]; ok {
		return errors.New("already joined")
	}

	// todo: do we want to sign messages?
	// it's required to ensure that it's not possible to DDOS the network
	// by "relaying" malicious messages that pretend to be sent by other authentic peers
	topic, err := r.ps.Join(string(protocolID))
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

	r.items[protocolID] = i

	r.runReceiveLoop(&i, handler)

	return nil
}

func (r *Registry) Close() {
	for _, item := range r.items {
		item.sub.Cancel()

		if err := item.topic.Close(); err != nil {
			r.logger.Error("Error closing gossip topic", "protocol", item.protocolID, "err", err)
		}
	}
}

func (r *Registry) Broadcast(protocolID protocol.ID, data []byte) error {
	if _, ok := r.items[protocolID]; !ok {
		return errors.Errorf("protocol %s not found", protocolID)
	}

	// todo explore publish options
	err := r.items[protocolID].topic.Publish(r.ctx, data)
	if err != nil {
		return errors.Wrap(err, "unable to publish message")
	}

	r.logger.Debug("Gossiped message", "protocol", protocolID, "data_len", len(data))

	return nil
}

func (r *Registry) runReceiveLoop(item *item, handler Handler) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				r.logger.Error(
					"Panic in (*Registry).runReceiveLoop",
					"panic", p,
					"protocol", item.protocolID,
				)
			}
		}()

		err := r.runReceiveLoopBlocking(item, handler)
		switch {
		case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
			r.logger.Info("Context canceled or deadline exceeded for gossip", "protocol", item.protocolID)
		case err != nil:
			r.logger.Error("Error in gossip receive loop", "protocol", item.protocolID, "err", err)
		}
	}()
}

func (r *Registry) runReceiveLoopBlocking(item *item, handler Handler) error {
	for {
		msg, err := item.sub.Next(r.ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get next message")
		}

		if msg.Local || msg.ReceivedFrom == r.self {
			r.logger.Debug(
				"Skipping local gossip message",
				"protocol", item.protocolID,
				"message_id", msg.ID,
			)
			continue
		}

		// ensure messages are processed concurrently
		go r.handleMessage(item, msg, handler)
	}
}

func (r *Registry) handleMessage(item *item, msg *pubsub.Message, handler Handler) {
	defer func() {
		if p := recover(); p != nil {
			r.logger.Error(
				"Panic in (*Registry).handleMessage",
				"panic", p,
				"protocol", item.protocolID,
				"message_id", msg.ID,
			)
		}
	}()

	if err := handler(item.protocolID, msg); err != nil {
		r.logger.Error(
			"Error in gossip message handler",
			"protocol", item.protocolID,
			"message_id", msg.ID,
			"err", err,
		)
	}
}
