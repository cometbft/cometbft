package lp2p

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/lp2p/autopool"
	pq "github.com/cometbft/cometbft/lp2p/priorityqueue"
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
	name          string
	priorityQueue *pq.Queue
	consumerPool  *autopool.Pool[pendingEnvelope]
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

	consumerPool, priorityQueue := rs.newReactorQueue(nextID, name)

	rs.reactors = append(rs.reactors, reactorItem{
		Reactor:       reactor,
		name:          name,
		priorityQueue: priorityQueue,
		consumerPool:  consumerPool,
	})

	// add name to mapping
	rs.reactorNames[name] = nextID

	rs.switchRef.Logger.Info("Added reactor", "reactor", name)

	return nil
}

// Start starts all reactors with their receive queues
func (rs *reactorSet) Start(perProtocolCallback func(protocol.ID)) error {
	for _, reactor := range rs.reactors {
		rs.switchRef.Logger.Info("Starting reactor", "reactor", reactor.name)
		reactor.SetSwitch(rs.switchRef)

		if err := reactor.Start(); err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", reactor.name, err)
		}

		reactor.consumerPool.Start()
	}

	for protocolID := range rs.protocols {
		perProtocolCallback(protocolID)
	}

	return nil
}

func (rs *reactorSet) Stop() {
	for _, reactor := range rs.reactors {
		reactor.consumerPool.Stop()

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

// Receive schedules receive operation for a reactor. How it works:
// 1) pendingEnvelope is added to a priority queue that sorts msgs (heap)
// 2) priorityQueue exposes a chan of sorted items available to consumption
// 3) autopool picks this channel, receives the message and calls reactorSet.receiveQueued(pendingEnvelope)
//
// This setup allows to handle lots of incoming message in a timely manner. System ensures that
// - We can process as many concurrent messages as possible
// - All messages are sorted by priority, most important are processed first
// - In case of latency degradation, the system is downscale to preserve processing speed.
func (rs *reactorSet) Receive(reactorName, messageType string, envelope p2p.Envelope, priority int) {
	idx, ok := rs.reactorNames[reactorName]
	if !ok {
		rs.switchRef.Logger.Error("Receive: reactor not found", "reactor", reactorName)
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

	pe := pendingEnvelope{
		Envelope:    envelope,
		messageType: messageType,
		addedAt:     now,
	}

	reactor.priorityQueue.Push(pe, uint64(priority))
}

func (rs *reactorSet) receiveQueued(reactorID int, e pendingEnvelope) {
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

/*
// not used
func (rs *reactorSet) receiveSync(reactorID int, e pendingEnvelope) {
	reactor := rs.reactors[reactorID]
	m := rs.switchRef.metrics
	labels := []string{
		"reactor", reactor.name,
		"message_type", e.messageType,
	}

	// for sync receiver, queue concurrency equals to in-flight messages
	m.MessageReactorQueueConcurrency.With("reactor", reactor.name).Add(1)

	now := time.Now()
	reactor.Receive(e.Envelope)
	timeTaken := time.Since(now)

	m.MessagesReactorInFlight.With(labels...).Add(-1)
	m.MessageReactorReceiveDuration.With(labels...).Observe(timeTaken.Seconds())
	m.MessageReactorQueueConcurrency.With("reactor", reactor.name).Add(-1)
}
*/

// newConsumerPool creates a pool of envelope consumers for a reactor
// the idea to dynamically adjust concurrency based on the load.
func (rs *reactorSet) newReactorQueue(
	reactorID int,
	reactorName string,
) (*autopool.Pool[pendingEnvelope], *pq.Queue) {
	// Constants fro pool
	const (
		// how many message we can accept to this before blocking (per reactor)
		reactorReceiveChanCapacity = 1024

		minWorkers              = 4
		defaultMaxWorkers       = 32
		defaultLatencyThreshold = 100 * time.Millisecond
		latencyPercentile       = 90.0 // P90
		autoScaleInternal       = 250 * time.Millisecond
	)

	var (
		latencyThreshold = defaultLatencyThreshold
		maxWorkers       = defaultMaxWorkers
	)

	// bump max workers for mempool
	if reactorName == "MEMPOOL" {
		maxWorkers = 128
	}

	// 2. Create channel that receives messages from priority queue
	// new p2p messages are written to this chan
	// then, pool consumes these messages concurrently and calls rs.receiveQueued()
	pendingMessagesChan := make(chan pendingEnvelope, reactorReceiveChanCapacity)

	// 3. Create consumer pool with scaler
	concurrencyCounter := rs.switchRef.metrics.MessageReactorQueueConcurrency.With("reactor", reactorName)

	scaler := autopool.NewThroughputLatencyScaler(
		minWorkers,
		maxWorkers,
		latencyPercentile,
		latencyThreshold,
		autoScaleInternal,
		rs.switchRef.Logger,
	)

	pool := autopool.New(
		scaler,
		pendingMessagesChan,
		func(e pendingEnvelope) {
			rs.receiveQueued(reactorID, e)
		},
		rs.switchRef.Logger,
		autopool.WithOnScale[pendingEnvelope](func() {
			concurrencyCounter.Add(1)
		}),
		autopool.WithOnShrink[pendingEnvelope](func() {
			concurrencyCounter.Add(-1)
		}),
	)

	// 4. Create priority queue. It stores all messages in heap based on message priority,
	// And then publishes them in ordered way to consumer
	queue := pq.New(reactorReceiveChanCapacity)

	// 5. Create pipe that writes ordered messages to consumer
	go func() {
		ch, cancel := queue.Consumer()
		defer cancel()

		for v := range ch {
			pe, ok := v.(pendingEnvelope)
			if !ok {
				// should not happen
				panic("invalid type")
			}

			pendingMessagesChan <- pe
		}
	}()

	return pool, queue
}
