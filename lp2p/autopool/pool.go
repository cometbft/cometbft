package autopool

import (
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// Pool primitive auto-scaling pool for concurrent message processing.
// It accepts a channel of messages and a function to process them and scales
// the number of workers dynamically based on the message processing time.
type Pool[T any] struct {
	// channel what is used to consume messages
	ch <-chan T

	// consumer function that is used to process messages
	receive func(T)

	// latest sequence number of the worker (worker id)
	seqNum int

	scaler *ThroughputLatencyScaler

	workers   map[int]*worker[T]
	workersWg sync.WaitGroup

	mu      sync.Mutex
	stopped bool

	logger log.Logger
}

type worker[T any] struct {
	seqNum  int
	pool    *Pool[T]
	closeCh chan struct{}
}

// New Pool constructor.
func New[T any](
	scaler *ThroughputLatencyScaler,
	producer <-chan T,
	receiveFN func(T),
	logger log.Logger,
) *Pool[T] {
	return &Pool[T]{
		ch:        producer,
		receive:   receiveFN,
		scaler:    scaler,
		workers:   make(map[int]*worker[T]),
		workersWg: sync.WaitGroup{},
		stopped:   false,
		mu:        sync.Mutex{},
		logger:    logger,
	}
}

func (p *Pool[T]) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.scaler.Min(); i++ {
		p.scale()
	}

	go p.monitor()
}

// Stop stops the pool and all workers
// safe to call multiple times
func (p *Pool[T]) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped || len(p.workers) == 0 {
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
	p.stopped = true
}

func (w *worker[T]) run() {
	defer func() {
		if r := recover(); r != nil {
			w.pool.logger.Error("Panic in pool worker", "panic", r)
		}
	}()

	for {
		select {
		case <-w.closeCh:
			// worker received a close signal
			return
		case msg, ok := <-w.pool.ch:
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
	p.workersWg.Add(1)
	defer p.workersWg.Done()

	now := time.Now()
	p.receive(msg)
	timeTaken := time.Since(now)

	// record metrics
	p.scaler.Track(timeTaken)
}

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

	if p.stopped {
		return true
	}

	decision := p.scaler.Decide(len(p.workers))

	switch decision {
	case ShouldScale:
		p.scale()
	case ShouldShrink:
		p.shrink()
	case ShouldStay:
		// do nothing
	}

	return false
}

// lock should be hold by the caller
func (p *Pool[T]) scale() {
	if p.stopped || len(p.workers) >= p.scaler.Max() {
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
}

// lock should be hold by the caller
func (p *Pool[T]) shrink() {
	if p.stopped || len(p.workers) == 0 {
		return
	}

	// stop any worker (non deterministic)
	// it's okay to do so because worker maps a relatively small
	for id := range p.workers {
		p.removeWorker(id)
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
