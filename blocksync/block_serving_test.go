package blocksync

import (
	"bytes"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	p2pmock "github.com/cometbft/cometbft/p2p/mock"
	bcproto "github.com/cometbft/cometbft/proto/tendermint/blocksync"
	"github.com/cometbft/cometbft/store"
)

type blockServingPeer struct {
	p2p.Peer
	channels []conn.ChannelStatus
}

func (p *blockServingPeer) Status() conn.ConnectionStatus {
	return conn.ConnectionStatus{Channels: p.channels}
}

func TestBlockServingQueueStatus(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		channel  byte
		capacity int
		size     int
		full     bool
	}{
		{"empty", BlocksyncChannel, 1000, 0, false},
		{"one message", BlocksyncChannel, 1000, 1, false},
		{"below capacity", BlocksyncChannel, 1000, 999, false},
		{"at capacity", BlocksyncChannel, 1000, 1000, true},
		{"full plus in flight", BlocksyncChannel, 1000, 1001, true},
		{"unknown capacity", BlocksyncChannel, 0, 1, false},
		{"negative capacity", BlocksyncChannel, -1, 1, false},
		{"different channel", 0x20, 1000, 1001, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			peer := &blockServingPeer{channels: []conn.ChannelStatus{{
				ID: tc.channel, SendQueueCapacity: tc.capacity, SendQueueSize: tc.size,
			}}}
			require.Equal(t, tc.full, blockQueueFull(peer))
		})
	}
	require.False(t, blockQueueFull(nil))
	require.False(t, blockQueueFull(&blockServingPeer{}))
}

func TestBlockServingQueueDrain(t *testing.T) {
	t.Parallel()
	r := &Reactor{}
	r.BaseReactor = *p2p.NewBaseReactor("Reactor", r)
	var output bytes.Buffer
	r.BaseReactor.SetLogger(log.NewTMLogger(log.NewSyncWriter(&output)))
	peer := newBlockServingPeer(t, 1000)
	request := &bcproto.BlockRequest{Height: 17}
	r.Receive(p2p.Envelope{Src: peer, Message: request})
	require.False(t, r.respondToPeer(request, peer))
	require.Contains(t, output.String(), "send queue full")
	require.Contains(t, output.String(), string(peer.ID()))
	require.Contains(t, output.String(), "height=17")

	db := dbm.NewMemDB()
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	r.store = store.NewBlockStore(db)
	require.True(t, r.respondToPeer(request, newBlockServingPeer(t, 999)))
	require.False(t, r.respondToPeer(request, peer))
	peer.channels[0].SendQueueSize = 999
	require.True(t, r.respondToPeer(request, peer))
	peer.channels = nil
	require.True(t, r.respondToPeer(request, peer))
}

func newBlockServingPeer(t *testing.T, size int) *blockServingPeer {
	t.Helper()
	peer := &blockServingPeer{Peer: p2pmock.NewPeer(nil), channels: []conn.ChannelStatus{{
		ID: BlocksyncChannel, SendQueueCapacity: 1000, SendQueueSize: size,
	}}}
	t.Cleanup(func() { require.NoError(t, peer.Stop()) })
	return peer
}
