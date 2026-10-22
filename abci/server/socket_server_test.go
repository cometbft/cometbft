package server

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/types"
)

type panicReader struct{}

func (panicReader) Read(_ []byte) (int, error) {
	panic("boom")
}

func TestNewSocketServerWithListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := NewSocketServerWithListener(ln, types.NewBaseApplication())
	require.NoError(t, srv.Start())
	t.Cleanup(func() { require.NoError(t, srv.Stop()) })

	conn, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	conn.Close()
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

// TestHandleRequestsClosesResponsesOnExit ensures handleResponses is told to
// stop once the request side is gone.
func TestHandleRequestsClosesResponsesOnExit(t *testing.T) {
	s := &SocketServer{}
	closeConn := make(chan error, 2)
	responses := make(chan *types.Response, 1)

	go s.handleRequests(closeConn, strings.NewReader(""), responses)

	select {
	case err := <-closeConn:
		require.ErrorIs(t, err, io.EOF)
	case <-time.After(time.Second):
		t.Fatal("closeConn not signaled")
	}

	select {
	case _, ok := <-responses:
		require.False(t, ok, "responses channel should be closed")
	case <-time.After(time.Second):
		t.Fatal("responses channel not closed after handleRequests exited")
	}
}

func TestHandleResponsesReturnsWhenResponsesClosed(t *testing.T) {
	s := &SocketServer{}
	closeConn := make(chan error, 2)
	responses := make(chan *types.Response, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		s.handleResponses(closeConn, io.Discard, responses)
	}()

	close(responses)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleResponses did not return after the responses channel was closed")
	}
}
