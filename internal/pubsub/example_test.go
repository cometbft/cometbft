package pubsub_test

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/internal/pubsub"
	"github.com/cometbft/cometbft/internal/pubsub/query"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	s := pubsub.NewServer()
	s.SetLogger(log.TestingLogger())
	err := s.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Error(err)
		}
	})

	ctx := context.Background()
	subscription, err := s.Subscribe(ctx, "example-client", query.MustCompile("abci.account.name='John'"))
	require.NoError(t, err)
	err = s.PublishWithEvents(ctx, "Tombstone", map[string][]string{"abci.account.name": {"John"}})
	require.NoError(t, err)
	assertReceive(t, "Tombstone", subscription.Out())
}
