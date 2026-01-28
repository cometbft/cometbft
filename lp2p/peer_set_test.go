package lp2p

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeerSet(t *testing.T) {
	t.Run("CRUD", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		hosts := makeTestHosts(t, 2)
		hostA := hosts[0]
		hostB := hosts[1]

		// Connect hosts so they know each other's addresses
		err := hostA.Connect(ctx, hostB.AddrInfo())
		require.NoError(t, err)

		ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())
		peerBKey := peerIDToKey(hostB.ID())

		// ACT & ASSERT #1: has(false) -> get(nil)
		hasPeer := ps.Has(peerBKey)
		assert.False(t, hasPeer)

		gotPeer := ps.Get(peerBKey)
		assert.Nil(t, gotPeer)

		// ACT & ASSERT #2: add -> get(ok) -> has(ok)
		peerB, err := ps.Add(hostB.ID(), PeerAddOptions{
			Private:       true,
			Persistent:    true,
			Unconditional: true,
		})
		require.NoError(t, err)
		require.NotNil(t, peerB)

		// check flags
		require.True(t, peerB.IsPrivate())
		require.True(t, peerB.IsPersistent())
		require.True(t, peerB.IsUnconditional())

		hasPeer = ps.Has(peerBKey)
		assert.True(t, hasPeer)

		gotPeer = ps.Get(peerBKey)
		require.NotNil(t, gotPeer)
		assert.Equal(t, peerBKey, gotPeer.ID())

		// ACT #3: remove -> has(false) -> get(nil)
		err = ps.Remove(peerBKey, PeerRemovalOptions{Reason: "test removal"})

		// ASSERT
		require.NoError(t, err)

		hasPeer = ps.Has(peerBKey)
		assert.False(t, hasPeer)

		gotPeer = ps.Get(peerBKey)
		assert.Nil(t, gotPeer)
	})

	t.Run("Safeguards", func(t *testing.T) {
		t.Run("Not exists in addr book", func(t *testing.T) {
			// ARRANGE
			hosts := makeTestHosts(t, 2)
			hostA := hosts[0]
			hostB := hosts[1]

			ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

			// ACT
			_, err := ps.Add(hostB.ID(), PeerAddOptions{})

			// ASSERT
			require.Error(t, err)
		})

		t.Run("Add existing", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()

			hosts := makeTestHosts(t, 2)
			hostA := hosts[0]
			hostB := hosts[1]

			err := hostA.Connect(ctx, hostB.AddrInfo())
			require.NoError(t, err)

			ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

			// ACT
			_, err1 := ps.Add(hostB.ID(), PeerAddOptions{})
			_, err2 := ps.Add(hostB.ID(), PeerAddOptions{})

			// ASSERT
			require.NoError(t, err1)
			require.ErrorIs(t, err2, ErrPeerExists)
		})

		t.Run("Remove non-existing", func(t *testing.T) {
			// ARRANGE
			hosts := makeTestHosts(t, 2)
			hostA := hosts[0]
			hostB := hosts[1]

			ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())
			nonExistentKey := peerIDToKey(hostB.ID())

			// ACT
			err := ps.Remove(nonExistentKey, PeerRemovalOptions{Reason: "test"})

			// ASSERT
			require.ErrorContains(t, err, "peer not found")
		})
	})

	t.Run("ForEach", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		const peers = 6

		hosts := makeTestHosts(t, peers)
		hostA := hosts[0]

		// Given peer set for hostA
		ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

		// Given connected hosts
		for i := 1; i < peers; i++ {
			err := hostA.Connect(ctx, hosts[i].AddrInfo())
			require.NoError(t, err)

			_, err = ps.Add(hosts[i].ID(), PeerAddOptions{})
			require.NoError(t, err)
		}

		// ACT
		// collect ids
		collectedIDs := make(map[p2p.ID]struct{})
		ps.ForEach(func(p p2p.Peer) {
			collectedIDs[p.ID()] = struct{}{}
		})

		// ASSERT
		assert.Equal(t, peers-1, ps.Size())
		assert.Equal(t, peers-1, len(collectedIDs))
	})

	t.Run("Backwards compatibility", func(t *testing.T) {
		t.Run("Random", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()

			hosts := makeTestHosts(t, 4)
			hostA := hosts[0]
			ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

			// ACT & ASSERT #1: Random returns nil when empty
			randomPeer := ps.Random()
			assert.Nil(t, randomPeer)

			// ARRANGE #2: Add some peers
			for i := 1; i < 4; i++ {
				err := hostA.Connect(ctx, hosts[i].AddrInfo())
				require.NoError(t, err)

				_, err = ps.Add(hosts[i].ID(), PeerAddOptions{})
				require.NoError(t, err)
			}

			// ACT & ASSERT #2: Random returns a valid peer
			randomPeer = ps.Random()
			require.NotNil(t, randomPeer)
			require.Contains(
				t,
				[]peer.ID{hosts[1].ID(), hosts[2].ID(), hosts[3].ID()},
				ps.keyToPeerID(randomPeer.ID()),
			)
		})

		t.Run("Copy", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()

			hosts := makeTestHosts(t, 4)
			hostA := hosts[0]
			ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

			expectedIDs := make([]p2p.ID, 0, 3)

			for i := 1; i < 4; i++ {
				err := hostA.Connect(ctx, hosts[i].AddrInfo())
				require.NoError(t, err)

				_, err = ps.Add(hosts[i].ID(), PeerAddOptions{})
				require.NoError(t, err)

				expectedIDs = append(expectedIDs, peerIDToKey(hosts[i].ID()))
			}

			// ACT
			copied := ps.Copy()

			// ASSERT
			require.Len(t, copied, len(expectedIDs))

			copiedIDs := make([]p2p.ID, len(copied))
			for i, p := range copied {
				copiedIDs[i] = p.ID()
			}

			// ensure copy is sorted
			for i := 1; i < len(copied); i++ {
				assert.True(t, copied[i-1].ID() < copied[i].ID())
			}

			assert.ElementsMatch(t, expectedIDs, copiedIDs)
		})
	})
}
