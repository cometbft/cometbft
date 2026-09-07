package http

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

func TestPublishEventDoesNotBlockOnUnbufferedSubscription(t *testing.T) {
	const (
		unbufferedQuery = "tm.event = 'Tx'"
		bufferedQuery   = "tm.event = 'NewBlock'"
	)

	buffered := make(chan ctypes.ResultEvent, 1)
	w := &WSEvents{
		subscriptions: map[string]chan ctypes.ResultEvent{
			unbufferedQuery: make(chan ctypes.ResultEvent),
			bufferedQuery:   buffered,
		},
	}
	w.Logger = log.NewNopLogger()

	done := make(chan struct{})
	go func() {
		w.publishEvent(&ctypes.ResultEvent{Query: unbufferedQuery})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publishing to an unread unbuffered subscription blocked")
	}

	w.publishEvent(&ctypes.ResultEvent{Query: bufferedQuery})

	select {
	case event := <-buffered:
		if event.Query != bufferedQuery {
			t.Fatalf("received event for query %q, expected %q", event.Query, bufferedQuery)
		}
	case <-time.After(time.Second):
		t.Fatal("event was not delivered to the buffered subscription")
	}
}
