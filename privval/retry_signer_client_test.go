package privval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// TestRetrySignerClientCloseAborts checks Close aborts an in-flight retry loop
// instead of draining the full retry budget (which would wedge node shutdown).
func TestRetrySignerClientCloseAborts(t *testing.T) {
	// Listener with no remote signer ever dialing in: every attempt blocks in
	// WaitConnection until timeoutAccept, then errors.
	endpoint := newSignerListenerEndpoint(log.TestingLogger(), "tcp://127.0.0.1:0", testTimeoutReadWrite)
	require.NoError(t, endpoint.Start())
	t.Cleanup(func() { _ = endpoint.Stop() })

	sc, err := NewSignerClient(endpoint, "chain-id")
	require.NoError(t, err)

	// Large budget: if Close didn't abort, this would take many seconds.
	rsc := NewRetrySignerClient(sc, 100, 50*time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- rsc.SignVote("chain-id", &cmtproto.Vote{ValidatorAddress: ed25519.GenPrivKey().PubKey().Address()})
	}()

	// Let one attempt get underway, then close.
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, rsc.Close())

	select {
	case err := <-done:
		require.Error(t, err) // aborted, not signed
	case <-time.After(testTimeoutAccept + 2*time.Second):
		t.Fatal("SignVote did not abort after Close")
	}

	require.NotPanics(t, func() { _ = rsc.Close() }) // idempotent
}
