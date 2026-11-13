package mempool

import (
	"net"
	"testing"

	"github.com/cometbft/cometbft/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestMempoolIDsBasic(t *testing.T) {
	ids := newMempoolIDs()

	peer := mock.NewPeer(net.IP{127, 0, 0, 1})

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 1, ids.GetForPeer(peer))
	ids.Reclaim(peer)

	ids.ReserveForPeer(peer)
	assert.EqualValues(t, 2, ids.GetForPeer(peer))
	ids.Reclaim(peer)
}

func TestMempoolIDsReserveIsIdempotent(t *testing.T) {
	ids := newMempoolIDs()

	peer := mock.NewPeer(net.IP{127, 0, 0, 1})

	ids.ReserveForPeer(peer)
	first := ids.GetForPeer(peer)
	ids.ReserveForPeer(peer)

	assert.Equal(t, first, ids.GetForPeer(peer))
	assert.Len(t, ids.activeIDs, 2) // UnknownPeerID + the peer's slot

	ids.Reclaim(peer)
	assert.Len(t, ids.activeIDs, 1)
	assert.Empty(t, ids.peerMap)
}

func TestMempoolIDsPanicsIfNodeRequestsOvermaxActiveIDs(t *testing.T) {
	if testing.Short() {
		return
	}

	// 0 is already reserved for UnknownPeerID
	ids := newMempoolIDs()

	for i := 0; i < MaxActiveIDs-1; i++ {
		peer := mock.NewPeer(net.IP{127, 0, 0, 1})
		ids.ReserveForPeer(peer)
	}

	assert.Panics(t, func() {
		peer := mock.NewPeer(net.IP{127, 0, 0, 1})
		ids.ReserveForPeer(peer)
	})
}
