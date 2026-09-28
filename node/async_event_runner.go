package node

import "sync"

const defaultAsyncEventQueueSize = 128

type asyncEventRunner struct {
	tasks chan func()
	done  chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	mu        sync.RWMutex
	stopped   bool
}

func newAsyncEventRunner(bufferSize int) *asyncEventRunner {
	return &asyncEventRunner{
		tasks: make(chan func(), bufferSize),
		done:  make(chan struct{}),
	}
}

func (r *asyncEventRunner) Start() {
	r.startOnce.Do(func() {
		go func() {
			defer close(r.done)
			for task := range r.tasks {
				task()
			}
		}()
	})
}

func (r *asyncEventRunner) Submit(task func()) {
	r.Start()

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.stopped {
		return
	}
	r.tasks <- task
}

func (r *asyncEventRunner) Stop() {
	r.stopOnce.Do(func() {
		r.Start()

		r.mu.Lock()
		r.stopped = true
		close(r.tasks)
		r.mu.Unlock()

		<-r.done
	})
}
