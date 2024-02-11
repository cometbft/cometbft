package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	cmtsync "github.com/cometbft/cometbft/internal/sync"
	"github.com/cometbft/cometbft/libs/log"
	types "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

var wsCallTimeout = 5 * time.Second

type myHandler struct {
	closeConnAfterRead bool
	mtx                cmtsync.RWMutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for {
		messageType, in, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req types.RPCRequest
		err = json.Unmarshal(in, &req)
		if err != nil {
			panic(err)
		}

		h.mtx.RLock()
		if h.closeConnAfterRead {
			if err := conn.Close(); err != nil {
				panic(err)
			}
		}
		h.mtx.RUnlock()

		res := json.RawMessage(`{}`)
		emptyRespBytes, err := json.Marshal(types.RPCResponse{Result: res, ID: req.ID})
		if err != nil {
			panic(err)
		}
		if err := conn.WriteMessage(messageType, emptyRespBytes); err != nil {
			return
		}
	}
}

func TestWSClientReconnectsAfterReadFailure(t *testing.T) {
	var wg sync.WaitGroup

	// start server
	h := &myHandler{}
	s := httptest.NewServer(h)
	defer s.Close()

	c := startClient(t, "//"+s.Listener.Addr().String())
	defer c.Stop()               //nolint:errcheck // ignore for tests
	errCh := make(chan error, 1) // Create the error channel

	wg.Add(1)

	callWgDoneOnResult(t, c, &wg, errCh) // Pass the error channel as an argument

	h.mtx.Lock()
	h.closeConnAfterRead = true
	h.mtx.Unlock()

	// results in WS read error, no send retry because write succeeded
	err := call(t, "a", c)
	require.NoError(t, err) // Handle the error using require.NoError

	// expect to reconnect almost immediately
	time.Sleep(10 * time.Millisecond)
	h.mtx.Lock()
	h.closeConnAfterRead = false
	h.mtx.Unlock()

	// should succeed
	err = call(t, "b", c)
	require.NoError(t, err) // Handle the error using require.NoError

	wg.Wait()
}

func TestWSClientReconnectsAfterWriteFailure(t *testing.T) {
	var wg sync.WaitGroup

	// start server
	h := &myHandler{}
	s := httptest.NewServer(h)

	c := startClient(t, "//"+s.Listener.Addr().String())
	defer c.Stop() //nolint:errcheck // ignore for tests

	wg.Add(2)
	errCh := make(chan error, 2) // Channel to collect errors from goroutines

	go func() {
		defer wg.Done()
		// hacky way to abort the connection before write
		if err := c.conn.Close(); err != nil {
			errCh <- err // Send error to the channel instead of calling t.Error directly
		}

		// results in WS write error, the client should resend on reconnect
		err := call(t, "a", c) // Use the updated call function that returns an error
		if err != nil {
			errCh <- err // Send error to the channel
		}
	}()

	go func() {
		defer wg.Done()
		// expect to reconnect almost immediately
		time.Sleep(10 * time.Millisecond)

		// should succeed
		err := call(t, "b", c) // Again, use the updated call function for the second call
		if err != nil {
			errCh <- err // Send error to the channel
		}
	}()

	wg.Wait()
	close(errCh) // Close the channel after all goroutines are done

	// Check for errors outside of the goroutines
	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestWSClientReconnectFailure(t *testing.T) {
	// start server
	h := &myHandler{}
	s := httptest.NewServer(h)

	c := startClient(t, "//"+s.Listener.Addr().String())
	defer c.Stop() //nolint:errcheck // ignore for tests

	go func() {
		for {
			select {
			case <-c.ResponsesCh:
			case <-c.Quit():
				return
			}
		}
	}()

	// hacky way to abort the connection before write
	if err := c.conn.Close(); err != nil {
		t.Error(err)
	}
	s.Close()

	// results in WS write error
	// provide timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), wsCallTimeout)
	defer cancel()
	if err := c.Call(ctx, "a", make(map[string]interface{})); err != nil {
		t.Error(err)
	}

	// expect to reconnect almost immediately
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	errCh := make(chan error, 1) // Create an error channel to communicate errors from the goroutine

	go func() {
		// client should block on this
		err := call(t, "b", c)
		if err != nil {
			errCh <- err // Send the error to the main goroutine
			return
		}
		close(done) // Close done to signal successful call
	}()

	// test that client blocks on the second send
	select {
	case <-done:
		t.Fatal("client should block on calling 'b' during reconnect")
	case err := <-errCh: // Receive any errors from the goroutine
		require.NoError(t, err) // Assert no error occurred
	case <-time.After(5 * time.Second):
		t.Log("All good")
	}
}

func TestNotBlockingOnStop(t *testing.T) {
	timeout := 2 * time.Second
	s := httptest.NewServer(&myHandler{})
	c := startClient(t, "//"+s.Listener.Addr().String())
	c.Call(context.Background(), "a", make(map[string]interface{})) //nolint:errcheck // ignore for tests
	// Let the readRoutine get around to blocking
	time.Sleep(time.Second)
	passCh := make(chan struct{})
	errCh := make(chan error, 1) // Create an error channel to communicate errors from the goroutine

	go func() {
		// Unless we have a non-blocking write to ResponsesCh from readRoutine
		// this blocks forever on the waitgroup
		err := c.Stop()
		if err != nil {
			errCh <- err // Send the error to the main goroutine
			return
		}
		close(passCh) // Close passCh to signal successful stop
	}()

	select {
	case <-passCh:
		// Pass
	case err := <-errCh: // Receive any errors from the goroutine
		require.NoError(t, err) // Assert no error occurred
	case <-time.After(timeout):
		t.Fatalf("WSClient failed to stop within %v seconds - is one of the read/write routines blocking?",
			timeout.Seconds())
	}
}

func startClient(t *testing.T, addr string) *WSClient {
	t.Helper()
	c, err := NewWS(addr, "/websocket")
	require.NoError(t, err)
	err = c.Start()
	require.NoError(t, err)
	c.SetLogger(log.TestingLogger())
	return c
}

func call(t *testing.T, method string, c *WSClient) error {
	t.Helper()
	// Call the method on the WSClient and return any error encountered.
	return c.Call(context.Background(), method, make(map[string]interface{}))
}

func callWgDoneOnResult(t *testing.T, c *WSClient, wg *sync.WaitGroup, errCh chan<- error) {
	t.Helper()
	go func() {
		defer close(errCh)
		for {
			select {
			case resp := <-c.ResponsesCh:
				if resp.Error != nil {
					errCh <- fmt.Errorf("unexpected error: %v", resp.Error) // Send error to the main goroutine.
					return
				}
				if resp.Result != nil {
					wg.Done()
				}
			case <-c.Quit():
				return
			}
		}
	}()
}
