package cat

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
)

func TestRequestSchedulerRerequest(t *testing.T) {
	var (
		requests        = newRequestScheduler(10*time.Millisecond, 1*time.Minute)
		tx              = types.Tx("tx")
		key             = tx.Key()
		peerA    uint16 = 1 // should be non-zero
		peerB    uint16 = 2
	)
	t.Cleanup(requests.Close)

	// check zero state
	require.Zero(t, requests.ForTx(key))
	require.False(t, requests.Has(peerA, key))
	// marking a tx that was never requested should return false
	require.False(t, requests.MarkReceived(peerA, key))

	// create a request
	closeCh := make(chan struct{})
	require.True(t, requests.Add(key, peerA, func(key types.TxKey) {
		require.Equal(t, key, key)
		// the first peer times out to respond so we ask the second peer
		require.True(t, requests.Add(key, peerB, func(key types.TxKey) {
			t.Fatal("did not expect to timeout")
		}))
		close(closeCh)
	}))

	// check that the request was added
	require.Equal(t, peerA, requests.ForTx(key))
	require.True(t, requests.Has(peerA, key))

	// should not be able to add the same request again
	require.False(t, requests.Add(key, peerA, nil))

	// wait for the scheduler to invoke the timeout
	<-closeCh

	// check that the request stil exists
	require.True(t, requests.Has(peerA, key))
	// check that peerB was requested
	require.True(t, requests.Has(peerB, key))

	// There should still be a request for the Tx
	require.Equal(t, peerB, requests.ForTx(key))

	// record a response from peerB
	require.True(t, requests.MarkReceived(peerB, key))

	// peerA comes in later with a response but it's still
	// considered a response from an earlier request
	require.True(t, requests.MarkReceived(peerA, key))
}

func TestRequestSchedulerNonResponsivePeer(t *testing.T) {
	var (
		requests        = newRequestScheduler(10*time.Millisecond, time.Millisecond)
		tx              = types.Tx("tx")
		key             = tx.Key()
		peerA    uint16 = 1 // should be non-zero
	)

	require.True(t, requests.Add(key, peerA, nil))
	require.Eventually(t, func() bool {
		return requests.ForTx(key) == 0
	}, 100*time.Millisecond, 5*time.Millisecond)
}

func TestRequestSchedulerConcurrencyAddsAndReads(t *testing.T) {
	leaktest.CheckTimeout(t, time.Second)()
	requests := newRequestScheduler(10*time.Millisecond, time.Millisecond)
	defer requests.Close()

	N := 5
	keys := make([]types.TxKey, N)
	for i := 0; i < N; i++ {
		tx := types.Tx(fmt.Sprintf("tx%d", i))
		keys[i] = tx.Key()
	}

	addWg := sync.WaitGroup{}
	receiveWg := sync.WaitGroup{}
	doneCh := make(chan struct{})
	for i := 1; i < N*N; i++ {
		addWg.Add(1)
		go func(peer uint16) {
			defer addWg.Done()
			requests.Add(keys[int(peer)%N], peer, nil)
		}(uint16(i))
	}
	for i := 1; i < N*N; i++ {
		receiveWg.Add(1)
		go func(peer uint16) {
			defer receiveWg.Done()
			markReceived := func() {
				for _, key := range keys {
					if requests.Has(peer, key) {
						requests.MarkReceived(peer, key)
					}
				}
			}
			for {
				select {
				case <-doneCh:
					// need to ensure this is run
					// at least once after all adds
					// are done
					markReceived()
					return
				default:
					markReceived()
				}
			}
		}(uint16(i))
	}
	addWg.Wait()
	close(doneCh)

	receiveWg.Wait()

	for _, key := range keys {
		require.Zero(t, requests.ForTx(key))
	}
}
