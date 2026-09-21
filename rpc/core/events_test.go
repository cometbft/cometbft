package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/types"
)

// recordingWSConn records the context of every WriteRPCResponse call.
type recordingWSConn struct {
	mtx    sync.Mutex
	ctxs   []context.Context
	writes chan struct{}
}

func (*recordingWSConn) GetRemoteAddr() string { return "test" }

func (c *recordingWSConn) WriteRPCResponse(ctx context.Context, _ rpctypes.RPCResponse) error {
	c.mtx.Lock()
	c.ctxs = append(c.ctxs, ctx)
	c.mtx.Unlock()
	c.writes <- struct{}{}
	return nil
}

func (*recordingWSConn) TryWriteRPCResponse(rpctypes.RPCResponse) bool { return true }

func (*recordingWSConn) Context() context.Context { return context.Background() }

func (c *recordingWSConn) writeCtx(i int) context.Context {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.ctxs[i]
}

func TestSubscribeCancelsWriteContextAfterEachEvent(t *testing.T) {
	eventBus := types.NewEventBus()
	eventBus.SetLogger(log.TestingLogger())
	require.NoError(t, eventBus.Start())
	t.Cleanup(func() {
		if err := eventBus.Stop(); err != nil {
			t.Error(err)
		}
	})

	env := &Environment{
		EventBus: eventBus,
		Logger:   log.TestingLogger(),
		Config:   *cfg.DefaultRPCConfig(),
	}
	conn := &recordingWSConn{writes: make(chan struct{}, 2)}
	ctx := &rpctypes.Context{
		JSONReq: &rpctypes.RPCRequest{ID: rpctypes.JSONRPCIntID(1)},
		WSConn:  conn,
	}

	_, err := env.Subscribe(ctx, "tm.event='Tx'")
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		require.NoError(t, eventBus.PublishEventTx(types.EventDataTx{}))
		select {
		case <-conn.writes:
		case <-time.After(5 * time.Second):
			t.Fatalf("event %d was not forwarded", i)
		}
	}

	// The context used for the first write must be released once that
	// write is done, not when the subscription eventually ends.
	require.ErrorIs(t, conn.writeCtx(0).Err(), context.Canceled)
}
