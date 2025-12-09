package autopool

import (
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// Pool primitive auto-scaling pool for concurrent message processing.
// It accepts a function to process messages and scales the number of workers
// dynamically based on the message processing time. It also supports priority-based message processing.
type Pool[T any] struct {
	// processing channel for messages
	inbound chan T

	// a queue on top of Pool for priority-based message processing.
	// acts as a buffer + priority-based FIFO queue. Routes messages to Pool.inbound
	priorityQueue *PriorityQueue

	// consumer function that is used to process messages
	receive func(T)

	// latest sequence number of the worker (worker id)
	seqNum int

	scaler *ThroughputLatencyScaler

	workers   map[int]*worker[T]
	workersWg sync.WaitGroup

	// callbacks to be called when the pool is scaled or shrunk
	onScale  func()
	onShrink func()
	onStay   func()

	mu        sync.RWMutex
	stoppedCh chan struct{}

	logger log.Logger
}

type worker[T any] struct {
	seqNum  int
	pool    *Pool[T]
	closeCh chan struct{}
}

type Option[T any] func(*Pool[T])

func WithLogger[T any](logger log.Logger) Option[T] {
	return func(p *Pool[T]) { p.logger = logger }
}

func WithPriorityQueue[T any](pq *PriorityQueue) Option[T] {
	return func(p *Pool[T]) { p.priorityQueue = pq }
}

func WithOnScale[T any](onScale func()) Option[T] {
	return func(p *Pool[T]) { p.onScale = onScale }
}

func WithOnShrink[T any](onShrink func()) Option[T] {
	return func(p *Pool[T]) { p.onShrink = onShrink }
}

func WithOnStay[T any](onStay func()) Option[T] {
	return func(p *Pool[T]) { p.onStay = onStay }
}

// interval to sleep when the priority queue is idle
// note that this should be fine because most of the time the queue will be busy.
// I also explored var cond for a more efficient solution, but it's not worth the complexity for now.
const priorityQueueIdleInterval = 50 * time.Millisecond

// New Pool constructor.
func New[T any](
	scaler *ThroughputLatencyScaler,
	receiveFN func(T),
	capacity int,
	opts ...Option[T],
) *Pool[T] {
	const defaultPriorities = 10

	pool := &Pool[T]{
		inbound:       make(chan T, capacity),
		priorityQueue: NewPriorityQueue(defaultPriorities),

		receive: receiveFN,
		scaler:  scaler,

		workers:   make(map[int]*worker[T]),
		workersWg: sync.WaitGroup{},

		onScale:  nil,
		onShrink: nil,
		onStay:   nil,

		stoppedCh: make(chan struct{}),

		mu:     sync.RWMutex{},
		logger: log.NewNopLogger(),
	}

	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

func (p *Pool[T]) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// already started or stopped
	if p.stopped() || len(p.workers) > 0 {
		return
	}

	for i := 0; i < p.scaler.Min(); i++ {
		p.scale()
	}

	go p.monitor()
	go p.pipePriorityQueue()
}

// Stop stops the pool and all workers
// safe to call multiple times
func (p *Pool[T]) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped() || len(p.workers) == 0 {
		return
	}

	// collect all ids to avoid map loop-and-delete
	workerIDs := make([]int, 0, len(p.workers))
	for id := range p.workers {
		workerIDs = append(workerIDs, id)
	}

	for _, id := range workerIDs {
		p.removeWorker(id)
	}

	p.logger.Info("Waiting for workers to finish")
	p.workersWg.Wait()

	close(p.inbound)
	close(p.stoppedCh)
}

// Push adds a message directly to the pool FIFO queue.
// Blocks if the queue is full.
func (p *Pool[T]) Push(msg T) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stopped() {
		p.logger.Error("Cannot push a message to a stopped pool (Push)")
		return
	}

	p.inbound <- msg
}

// PushPriority adds a message to the priority queue first.
// Non-blocking as priority queue is a linked-list
func (p *Pool[T]) PushPriority(msg T, priority int) error {
	return p.priorityQueue.Push(msg, priority)
}

func (p *Pool[T]) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stopped() {
		return 0
	}

	return len(p.inbound)
}

func (p *Pool[T]) Cap() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stopped() {
		return 0
	}

	return cap(p.inbound)
}

func (w *worker[T]) run() {
	w.pool.workersWg.Add(1)

	defer func() {
		w.pool.workersWg.Done()
		if r := recover(); r != nil {
			w.pool.logger.Error("Panic in pool worker", "panic", r)
		}
	}()

	for {
		select {
		case <-w.closeCh:
			// worker received a close signal
			return
		case msg, ok := <-w.pool.inbound:
			// channel is closed for all workers, stop the whole pool
			if !ok {
				w.pool.Stop()
				return
			}

			w.pool.handleMessage(msg)
		}
	}
}

func (p *Pool[T]) handleMessage(msg T) {
	now := time.Now()
	p.receive(msg)
	timeTaken := time.Since(now)

	// record metrics
	p.scaler.Track(timeTaken)
}

// monitor the pool and autoscale it based on the load
func (p *Pool[T]) monitor() {
	ticker := time.NewTicker(p.scaler.EpochDuration())
	defer ticker.Stop()

	for range ticker.C {
		if exit := p.autoscale(); exit {
			return
		}
	}
}

func (p *Pool[T]) autoscale() (exit bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped() {
		return true
	}

	decision := p.scaler.Decide(len(p.workers), len(p.inbound), cap(p.inbound))

	switch decision {
	case ShouldScale:
		p.scale()
	case ShouldShrink:
		p.shrink()
	case ShouldStay:
		if p.onStay != nil {
			p.onStay()
		}
	}

	return false
}

// lock should be hold by the caller
func (p *Pool[T]) scale() {
	if p.stopped() || len(p.workers) >= p.scaler.Max() {
		return
	}

	p.seqNum++

	// register new worker
	w := &worker[T]{
		seqNum:  p.seqNum,
		pool:    p,
		closeCh: make(chan struct{}),
	}

	p.workers[p.seqNum] = w

	go w.run()

	if p.onScale != nil {
		p.onScale()
	}
}

// lock should be hold by the caller
func (p *Pool[T]) shrink() {
	if p.stopped() || len(p.workers) == 0 {
		return
	}

	// stop any worker (non deterministic)
	// it's okay to do so because worker maps a relatively small
	for id := range p.workers {
		p.removeWorker(id)

		if p.onShrink != nil {
			p.onShrink()
		}

		return
	}
}

// lock should be hold by the caller
func (p *Pool[T]) removeWorker(id int) {
	w, ok := p.workers[id]
	if !ok {
		// should not happen
		p.logger.Error("Worker not found", "id", id)
		return
	}

	// send close signal to worker
	close(w.closeCh)
	delete(p.workers, id)
}

// pipePriorityQueue pipes messages from the priority queue to the inbound channel
func (p *Pool[T]) pipePriorityQueue() {
	for {
		if p.stopped() {
			return
		}

		value, ok := p.priorityQueue.Pop()
		if !ok {
			time.Sleep(priorityQueueIdleInterval)
			continue
		}

		tt, ok := value.(T)
		if !ok {
			// should never happen
			panic("unexpected type in priority queue")
		}

		// an idiomatic way of "push or exit" pattern
		select {
		case <-p.stoppedCh:
			return
		default:
			select {
			case <-p.stoppedCh:
				return
			case p.inbound <- tt:
				// message sent successfully
			}
		}
	}
}

func (p *Pool[T]) stopped() bool {
	select {
	case <-p.stoppedCh:
		return true
	default:
		return false
	}
}
