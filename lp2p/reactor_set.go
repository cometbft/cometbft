package lp2p

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/lp2p/autopool"
	"github.com/cometbft/cometbft/lp2p/queue"
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
	priorityQueue *queue.PriorityQueue
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

	priorityQueue, consumerPool := rs.newReactorPriorityQueue(nextID, name)

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

// Receive schedules receive operation for a reactor
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

	pq := pendingEnvelope{
		Envelope:    envelope,
		messageType: messageType,
		addedAt:     now,
	}

	err := reactor.priorityQueue.Push(pq, priority)
	if err != nil {
		rs.switchRef.metrics.MessagesReactorInFlight.With(labels...).Add(-1)
		rs.switchRef.Logger.Error("Failed to push envelope to priority queue", "reactor", reactorName, "err", err)
	}
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

// newReactorPriorityQueue creates a consumer pool for reactor.Receive()
// It allows to dynamically adjust consumption concurrency based on the load,
// while maintaining the priority, order, and latency of messages.
func (rs *reactorSet) newReactorPriorityQueue(
	reactorID int,
	reactorName string,
) (*queue.PriorityQueue, *autopool.Pool[pendingEnvelope]) {
	// 1. create a priority queue for inbound messages (priority linked-list)
	// all new messages will be published here first ordered by priority and then by arrival time (FIFO)
	// cometbft has up to 10 priorities
	const priorities = 10
	priorityQueue := queue.NewPriorityQueue(priorities)

	// 2. create a queue for message processing (chan)
	// messages from the first queue will be published here for concurrent processing
	const concurrentPoolCapacity = 512

	poolQueue := make(chan pendingEnvelope, concurrentPoolCapacity)

	// 3. create a pipe from priority queue to the pool queue
	pipeStopCh := pipeQueues(priorityQueue, poolQueue)

	stopChannels := func() {
		// will be called only once
		close(pipeStopCh)
		close(poolQueue)
	}

	// 4. create a scaler for the pool
	scaler := rs.newReactorScaler(reactorName)

	concurrencyCounter := rs.switchRef.metrics.MessageReactorQueueConcurrency.With("reactor", reactorName)

	// 5. create a pool for message processing
	pool := autopool.New(
		scaler,
		poolQueue,
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
		autopool.WithOnStop[pendingEnvelope](stopChannels),
	)

	return priorityQueue, pool
}

func (rs *reactorSet) newReactorScaler(reactorName string) *autopool.ThroughputLatencyScaler {
	const (
		minWorkers        = 4
		defaultMaxWorkers = 32
		latencyThreshold  = 100 * time.Millisecond
		latencyPercentile = 90.0 // P90
		autoScaleInterval = 250 * time.Millisecond
	)

	maxWorkers := defaultMaxWorkers

	// bump max workers for mempool
	if reactorName == "MEMPOOL" {
		maxWorkers = 128
	}

	return autopool.NewThroughputLatencyScaler(
		minWorkers,
		maxWorkers,
		latencyPercentile,
		latencyThreshold,
		autoScaleInterval,
		rs.switchRef.Logger,
	)
}

func pipeQueues(producer *queue.PriorityQueue, consumer chan pendingEnvelope) chan struct{} {
	stop := make(chan struct{})

	go func() {
		for {
			value, ok := producer.Pop()
			if !ok {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			pe, ok := value.(pendingEnvelope)
			if !ok {
				// should never happen
				panic("unexpected type in priority queue")
			}

			select {
			case <-stop:
				// stop chan called before consumer close
				return
			default:
			}

			consumer <- pe
		}
	}()

	return stop
}
