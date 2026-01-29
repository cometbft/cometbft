package server

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
)

type panicReader struct{}

func (panicReader) Read(_ []byte) (int, error) {
	panic("boom")
}

// TestHandleRequestsPanicBeforeLock ensures the panic recovery block does not
// attempt to unlock appMtx when the lock was never acquired.
func TestHandleRequestsPanicBeforeLock(t *testing.T) {
	s := &SocketServer{}
	closeConn := make(chan error, 1)
	responses := make(chan *types.Response, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		s.handleRequests(closeConn, panicReader{}, responses)
	}()

	select {
	case <-closeConn:
	case <-time.After(time.Second):
		t.Fatal("closeConn not signaled")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleRequests did not exit")
	}
}
