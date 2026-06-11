package server

import (
	"net"
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
