package mempool

import (
	"net"
	"testing"

	cfg "github.com/cometbft/cometbft/config"
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

func TestReactorAddPeerReservesIDReclaimedByEarlierConnection(t *testing.T) {
	config := cfg.TestMempoolConfig()
	config.Broadcast = false
	memR := &Reactor{config: config, ids: newMempoolIDs()}
	peer := mock.NewPeer(net.IP{127, 0, 0, 1})

	// Two connections with the same peer ID are initialized and share one reservation. The
	// first joins and disconnects before the second is added, which reclaims the ID.
	memR.InitPeer(peer)
	memR.InitPeer(peer)
	memR.RemovePeer(peer, nil)
	assert.Zero(t, memR.ids.GetForPeer(peer))

	memR.AddPeer(peer)
	assert.NotZero(t, memR.ids.GetForPeer(peer))
	assert.Len(t, memR.ids.activeIDs, 2)

	memR.RemovePeer(peer, nil)
	assert.Len(t, memR.ids.activeIDs, 1)
}
