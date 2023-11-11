package cat

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

func TestSeenTxSet(t *testing.T) {
	var (
		tx1Key = types.Tx("tx1").Key()
		tx2Key = types.Tx("tx2").Key()
		tx3Key = types.Tx("tx3").Key()
		peer1  = p2p.ID("1")
		peer2  = p2p.ID("2")
	)

	seenSet := NewSeenTxSet()
	require.Zero(t, seenSet.Pop(tx1Key))

	seenSet.Add(tx1Key, peer1)
	seenSet.Add(tx1Key, peer1)
	require.Equal(t, 1, seenSet.Len())
	seenSet.Add(tx1Key, peer2)
	peers := seenSet.Get(tx1Key)
	require.NotNil(t, peers)
	require.Equal(t, map[p2p.ID]struct{}{peer1: {}, peer2: {}}, peers)
	seenSet.Add(tx2Key, peer1)
	seenSet.Add(tx3Key, peer1)
	require.Equal(t, 3, seenSet.Len())
	seenSet.RemoveKey(tx2Key)
	require.Equal(t, 2, seenSet.Len())
	require.Zero(t, seenSet.Pop(tx2Key))
	require.Equal(t, peer1, *seenSet.Pop(tx3Key))
}

func TestSeenTxSetConcurrency(_ *testing.T) {
	seenSet := NewSeenTxSet()

	const (
		concurrency = 10
		numTx       = 100
	)

	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i uint16) {
			defer wg.Done()
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.Add(tx.Key(), p2p.ID(fmt.Sprintf("%d", i)))
			}
		}(uint16(i % 2))
	}
	time.Sleep(time.Millisecond)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(peer uint16) {
			defer wg.Done()
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.Has(tx.Key(), p2p.ID(fmt.Sprintf("%d", i)))
			}
		}(uint16(i % 2))
	}
	time.Sleep(time.Millisecond)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(peer uint16) {
			defer wg.Done()
			for i := numTx - 1; i >= 0; i-- {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.RemoveKey(tx.Key())
			}
		}(uint16(i % 2))
	}
	wg.Wait()
}