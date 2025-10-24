package lp2p

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type reactorSet struct {
	switchRef *Switch

	reactors []ReactorItem

	// [reactor_name => reactor] mapping
	reactorsByName map[string]p2p.Reactor

	// [protocol_id => reactor] mapping
	reactorsByProtocolID map[protocol.ID]p2p.Reactor

	// [protocol_id => reactor_name] mapping
	reactorNamesByProtocolID map[protocol.ID]string

	// [protocol_id => channel_descriptor] mapping
	descriptorByProtocolID map[protocol.ID]*p2p.ChannelDescriptor

	pendingEnvelopesByReactorName map[string]receiveQueue
}

type receiveQueue struct {
	ch      chan pendingEnvelope
	closeCh chan struct{}
}

// ReactorItem is a pair of name and reactor.
// Preserves order when adding.
type ReactorItem struct {
	Name    string
	Reactor p2p.Reactor
}

type reactorWithDescriptor struct {
	p2p.Reactor
	Name       string
	Descriptor *p2p.ChannelDescriptor
}

const (
	// how many message we can accept to this before blocking
	reactorReceiveChanCapacity = 1024

	// how many messages we can process concurrently
	reactorReceiveConsumers = 4
)

func newReactorSet(switchRef *Switch) *reactorSet {
	return &reactorSet{
		switchRef: switchRef,

		reactors:                      []ReactorItem{},
		reactorsByName:                make(map[string]p2p.Reactor),
		reactorsByProtocolID:          make(map[protocol.ID]p2p.Reactor),
		reactorNamesByProtocolID:      make(map[protocol.ID]string),
		descriptorByProtocolID:        make(map[protocol.ID]*p2p.ChannelDescriptor),
		pendingEnvelopesByReactorName: make(map[string]receiveQueue),
	}
}

func (rs *reactorSet) Add(item ReactorItem, switcher p2p.Switcher) error {
	for i := range item.Reactor.GetChannels() {
		var (
			channelDescriptor = item.Reactor.GetChannels()[i]
			protocolID        = ProtocolID(channelDescriptor.ID)
		)

		if _, ok := rs.reactorsByProtocolID[protocolID]; ok {
			return fmt.Errorf("protocol %q is already registered", protocolID)
		}

		rs.reactorsByProtocolID[protocolID] = item.Reactor
		rs.descriptorByProtocolID[protocolID] = channelDescriptor
		rs.reactorNamesByProtocolID[protocolID] = item.Name
	}

	rs.reactors = append(rs.reactors, item)
	rs.reactorsByName[item.Name] = item.Reactor

	rs.provisionReceiveQueue(
		item.Name,
		item.Reactor,
		reactorReceiveChanCapacity,
		reactorReceiveConsumers,
	)

	return nil
}

func (rs *reactorSet) Start(perProtocolCallback func(protocol.ID)) error {
	for _, el := range rs.reactors {
		name, reactor := el.Name, el.Reactor

		rs.switchRef.Logger.Info("Starting reactor", "reactor", name)

		reactor.SetSwitch(rs.switchRef)

		if err := reactor.Start(); err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	for protocolID := range rs.reactorsByProtocolID {
		perProtocolCallback(protocolID)
	}

	return nil
}

func (rs *reactorSet) Stop(switcher p2p.Switcher) {
	for name, reactor := range rs.reactorsByName {
		if err := reactor.Stop(); err != nil {
			switcher.Log().Error("failed to stop reactor", "name", name, "err", err)
		}

		close(rs.pendingEnvelopesByReactorName[name].closeCh)
	}
}

func (rs *reactorSet) InitPeer(peer *Peer) {
	for _, el := range rs.reactors {
		el.Reactor.InitPeer(peer)
	}
}

func (rs *reactorSet) AddPeer(peer *Peer) {
	for _, el := range rs.reactors {
		el.Reactor.AddPeer(peer)
	}
}

func (rs *reactorSet) RemovePeer(peer *Peer, reason any) {
	for _, reactor := range rs.reactorsByName {
		reactor.RemovePeer(peer, reason)
	}
}

func (rs *reactorSet) GetByName(name string) (p2p.Reactor, bool) {
	reactor, exists := rs.reactorsByName[name]
	return reactor, exists
}

func (rs *reactorSet) GetWithDescriptorByProtocolID(id protocol.ID) (reactorWithDescriptor, error) {
	reactor, ok := rs.reactorsByProtocolID[id]
	if !ok {
		return reactorWithDescriptor{}, fmt.Errorf("reactor not found")
	}

	descriptor, ok := rs.descriptorByProtocolID[id]
	if !ok {
		return reactorWithDescriptor{}, fmt.Errorf("descriptor not found")
	}

	return reactorWithDescriptor{
		Name:       rs.reactorNamesByProtocolID[id],
		Reactor:    reactor,
		Descriptor: descriptor,
	}, nil
}

type pendingEnvelope struct {
	p2p.Envelope
	messageType string
	addedAt     time.Time
}

// SubmitReceive schedules receive operation for a reactor
func (rs *reactorSet) SubmitReceive(reactorName, messageType string, envelope p2p.Envelope) {
	labels := []string{
		"reactor", reactorName,
		"message_type", messageType,
	}

	// lp2p metrics
	rs.switchRef.metrics.MessagesReceived.With(labels...).Add(1)
	rs.switchRef.metrics.MessagesReactorInFlight.With(labels...).Add(1)
	now := time.Now()

	rs.pendingEnvelopesByReactorName[reactorName].ch <- pendingEnvelope{
		Envelope:    envelope,
		messageType: messageType,
		addedAt:     now,
	}
}

func (rs *reactorSet) receive(reactorName string, reactor p2p.Reactor, e pendingEnvelope) {
	labels := []string{
		"reactor", reactorName,
		"message_type", e.messageType,
	}

	// log envelopes that are older than 1 second with a dummy sampling of 10%
	if time.Since(e.addedAt) > time.Second && e.addedAt.UnixMilli()%10 == 0 {
		rs.switchRef.Logger.Info(
			"Envelope is pending for too long",
			"reactor", reactorName,
			"message_type", e.messageType,
			"pending_dur", time.Since(e.addedAt).String(),
		)
	}

	now := time.Now()

	reactor.Receive(e.Envelope)

	timeTaken := time.Since(now)

	rs.switchRef.metrics.MessagesReactorInFlight.With(labels...).Add(-1)
	rs.switchRef.metrics.MessageReactorReceiveDuration.With(labels...).Observe(timeTaken.Seconds())
}

func (rs *reactorSet) provisionReceiveQueue(reactorName string, reactor p2p.Reactor, capacity, consumers int) {
	rq := receiveQueue{
		ch:      make(chan pendingEnvelope, capacity),
		closeCh: make(chan struct{}),
	}

	for i := 0; i < consumers; i++ {
		go func(index int) {
			defer func() {
				if r := recover(); r != nil {
					rs.switchRef.Logger.Error("Panic in receive queue", "reactor", reactorName, "panic", r)
				}
			}()

			for {
				select {
				case pendingEnvelope := <-rq.ch:
					rs.receive(reactorName, reactor, pendingEnvelope)
				case <-rq.closeCh:
					rs.switchRef.Logger.Info("Receive queue closed", "reactor", reactorName, "index", index)
					return
				}
			}
		}(i + 1)
	}

	rs.pendingEnvelopesByReactorName[reactorName] = rq
}
