package cat

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/mocks"
)

func TestPeerLifecycle(t *testing.T) {
	ids := newMempoolIDs()
	peer1 := &mocks.Peer{}
	peerID := p2p.ID("peer1")
	peer1.On("ID").Return(peerID)

	require.Nil(t, ids.GetPeer(1))
	require.Zero(t, ids.GetIDForPeer(peerID))
	require.Len(t, ids.GetAll(), 0)
	ids.ReserveForPeer(peer1)

	id := ids.GetIDForPeer(peerID)
	require.Equal(t, uint16(1), id)
	require.Equal(t, peer1, ids.GetPeer(id))
	require.Len(t, ids.GetAll(), 1)

	// duplicate peer should panic
	require.Panics(t, func() {
		ids.ReserveForPeer(peer1)
	})

	require.Equal(t, ids.Reclaim(peerID), id)
	require.Nil(t, ids.GetPeer(id))
	require.Zero(t, ids.GetIDForPeer(peerID))
	require.Len(t, ids.GetAll(), 0)
}
