package lp2p

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// reactorSet manages multiple reactors as a single entrypoint for Switch.
type reactorSet struct {
	switchRef *Switch

	reactors []reactorItem

	// [reactor_name => reactor_idx] mapping
	reactorNames map[string]int

	// [protocol_id => reactorProtocol] mapping
	protocols map[protocol.ID]reactorProtocol
}

// reactorItem p2p.Reactor wrapper
type reactorItem struct {
	p2p.Reactor
	name         string
	receiveQueue chan pendingEnvelope
}

// reactorProtocol represents mapping between [reactor, protocol, comet's channel descriptor]
type reactorProtocol struct {
	reactorID  int
	descriptor *p2p.ChannelDescriptor
}

// pendingEnvelope is a wrapper around p2p.Envelope
type pendingEnvelope struct {
	p2p.Envelope
	messageType string
	addedAt     time.Time
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

		reactors:     []reactorItem{},
		reactorNames: make(map[string]int),
		protocols:    make(map[protocol.ID]reactorProtocol),
	}
}

// Add adds a new reactor to the set
// NOTE: not goroutine safe. Uses only for initialization.
func (rs *reactorSet) Add(reactor p2p.Reactor, name string) error {
	nextID := len(rs.reactors)

	if _, ok := rs.reactorNames[name]; ok {
		return fmt.Errorf("reactor %q is already registered", name)
	}

	// register channel descriptor to reactor & protocolID mapping
	for i := range reactor.GetChannels() {
		var (
			channelDescriptor = reactor.GetChannels()[i]
			protocolID        = ProtocolID(channelDescriptor.ID)
		)

		if _, ok := rs.protocols[protocolID]; ok {
			return fmt.Errorf("protocol %q is already registered", protocolID)
		}

		rs.protocols[protocolID] = reactorProtocol{
			reactorID:  nextID,
			descriptor: channelDescriptor,
		}
	}

	rs.reactors = append(rs.reactors, reactorItem{
		Reactor:      reactor,
		name:         name,
		receiveQueue: make(chan pendingEnvelope, reactorReceiveChanCapacity),
	})

	// add name to mapping
	rs.reactorNames[name] = nextID

	rs.switchRef.Logger.Info("Added reactor", "reactor", name)

	return nil
}

// Start starts all reactors with their receive queues
func (rs *reactorSet) Start(perProtocolCallback func(protocol.ID)) error {
	for idx, reactor := range rs.reactors {
		rs.switchRef.Logger.Info("Starting reactor", "reactor", reactor.name)
		reactor.SetSwitch(rs.switchRef)

		if err := reactor.Start(); err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", reactor.name, err)
		}

		rs.runReceiveQueue(idx, reactorReceiveConsumers)
	}

	for protocolID := range rs.protocols {
		perProtocolCallback(protocolID)
	}

	return nil
}

func (rs *reactorSet) Stop() {
	for _, reactor := range rs.reactors {
		// note: technically we should wait for the receive queue
		// to be empty before closing the channel
		close(reactor.receiveQueue)

		rs.switchRef.Logger.Info("Stopping reactor", "reactor", reactor.name)
		if err := reactor.Stop(); err != nil {
			rs.switchRef.Logger.Error("Failed to stop reactor", "name", reactor.name, "err", err)
		}
	}
}

func (rs *reactorSet) InitPeer(peer *Peer) {
	for _, reactor := range rs.reactors {
		reactor.InitPeer(peer)
	}
}

func (rs *reactorSet) AddPeer(peer *Peer) {
	for _, reactor := range rs.reactors {
		reactor.AddPeer(peer)
	}
}

func (rs *reactorSet) RemovePeer(peer *Peer, reason any) {
	for _, reactor := range rs.reactors {
		reactor.RemovePeer(peer, reason)
	}
}

func (rs *reactorSet) GetByName(name string) (p2p.Reactor, bool) {
	idx, ok := rs.reactorNames[name]
	if !ok {
		return nil, false
	}

	return rs.reactors[idx].Reactor, true
}

func (rs *reactorSet) getReactorWithProtocol(id protocol.ID) (reactorProtocol, reactorItem, error) {
	protocol, ok := rs.protocols[id]
	if !ok {
		return reactorProtocol{}, reactorItem{}, fmt.Errorf("protocol not found")
	}

	return protocol, rs.reactors[protocol.reactorID], nil
}

// SubmitReceive schedules receive operation for a reactor
func (rs *reactorSet) SubmitReceive(reactorName, messageType string, envelope p2p.Envelope) {
	idx, ok := rs.reactorNames[reactorName]
	if !ok {
		rs.switchRef.Logger.Error("SubmitReceive: reactor not found", "reactor", reactorName)
		return
	}

	reactor := rs.reactors[idx]

	labels := []string{
		"reactor", reactorName,
		"message_type", messageType,
	}

	// lp2p metrics
	rs.switchRef.metrics.MessagesReceived.With(labels...).Add(1)
	rs.switchRef.metrics.MessagesReactorInFlight.With(labels...).Add(1)
	now := time.Now()

	reactor.receiveQueue <- pendingEnvelope{
		Envelope:    envelope,
		messageType: messageType,
		addedAt:     now,
	}
}

func (rs *reactorSet) receive(reactorID int, e pendingEnvelope) {
	reactor := rs.reactors[reactorID]

	labels := []string{
		"reactor", reactor.name,
		"message_type", e.messageType,
	}

	// log envelopes that are older than 1 second with a dummy sampling of 10%
	if time.Since(e.addedAt) > time.Second && e.addedAt.UnixMilli()%10 == 0 {
		rs.switchRef.Logger.Info(
			"Envelope is pending for too long",
			"reactor", reactor.name,
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

func (rs *reactorSet) runReceiveQueue(reactorID, consumers int) {
	reactor := rs.reactors[reactorID]

	for i := 0; i < consumers; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					rs.switchRef.Logger.Error("Panic in receive queue", "reactor", reactor.name, "panic", r)
				}
			}()

			for pendingEnvelope := range reactor.receiveQueue {
				rs.receive(reactorID, pendingEnvelope)
			}

			rs.switchRef.Logger.Info("Receive queue closed", "reactor", reactor.name)
		}()
	}
}
