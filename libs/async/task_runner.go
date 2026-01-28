package async

import (
	"runtime/debug"
	"sync"
)

// AsyncRunner executes a task asynchronously.
type AsyncRunner func(func())

// TaskRunner executes tasks sequentially in a dedicated goroutine.
// It provides FIFO ordering, panic recovery, and graceful shutdown.
type TaskRunner struct {
	taskCh     chan func()
	done       chan struct{}
	workerDone chan struct{}
	onPanic    func(r any, stack []byte)
	once       sync.Once
}

// NewTaskRunner creates a TaskRunner with the given buffer size.
// onPanic is called when a task panics; if nil, panics are silently recovered.
func NewTaskRunner(bufferSize int, onPanic func(r any, stack []byte)) *TaskRunner {
	if bufferSize < 0 {
		bufferSize = 0
	}
	tr := &TaskRunner{
		taskCh:     make(chan func(), bufferSize),
		done:       make(chan struct{}),
		workerDone: make(chan struct{}),
		onPanic:    onPanic,
	}
	go tr.loop()
	return tr
}

func (tr *TaskRunner) loop() {
	defer close(tr.workerDone)
	for {
		select {
		case f := <-tr.taskCh:
			tr.run(f)
		case <-tr.done:
			return
		}
	}
}

func (tr *TaskRunner) run(f func()) {
	defer func() {
		if r := recover(); r != nil && tr.onPanic != nil {
			tr.onPanic(r, debug.Stack())
		}
	}()
	f()
}

// Enqueue adds a task to be executed.
// Returns false if the runner is stopped; a true return means the task was
// accepted, but it may still be skipped if Stop() races with execution.
func (tr *TaskRunner) Enqueue(f func()) bool {
	// Check if already stopped (non-blocking).
	// This ensures Enqueue returns false after Stop() returns.
	select {
	case <-tr.done:
		return false
	default:
	}

	select {
	case tr.taskCh <- f:
		return true
	case <-tr.done:
		return false
	}
}

// Stop signals the runner to stop and waits for in-flight tasks to finish.
// Safe to call multiple times.
func (tr *TaskRunner) Stop() {
	tr.once.Do(func() {
		close(tr.done)
		<-tr.workerDone
	})
}
