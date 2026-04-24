package lp2p

import (
	"context"
	"net"
	"testing"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeer(t *testing.T) {
	t.Run("PeerInfo", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		hosts := makeTestHosts(t, 3)
		hostA := hosts[0]
		hostB := hosts[1]
		hostC := hosts[2]

		// hostA dials hostB (outbound from A's perspective)
		err := hostA.Connect(ctx, hostB.AddrInfo())
		require.NoError(t, err)

		// hostC dials hostA (inbound from A's perspective)
		err = hostC.Connect(ctx, hostA.AddrInfo())
		require.NoError(t, err)

		peerB, err := NewPeer(hostA, hostB.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		peerC, err := NewPeer(hostA, hostC.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		// ACT & ASSERT — outbound peer (we dialed hostB)
		// NodeInfo must type-assert to DefaultNodeInfo (this is what /net_info does)
		nodeInfo, ok := peerB.NodeInfo().(p2p.DefaultNodeInfo)
		require.True(t, ok, "NodeInfo must type-assert to p2p.DefaultNodeInfo")
		assert.Equal(t, peerB.ID(), nodeInfo.DefaultNodeID)
		assert.NotEmpty(t, nodeInfo.ListenAddr)

		assert.True(t, peerB.RemoteIP().Equal(net.IPv4(127, 0, 0, 1)))
		assert.True(t, peerB.IsOutbound())

		remoteAddr, ok := peerB.RemoteAddr().(*net.TCPAddr)
		require.True(t, ok)
		assert.NotZero(t, remoteAddr.Port)

		// ACT & ASSERT — inbound peer (hostC dialed us)
		_, ok = peerC.NodeInfo().(p2p.DefaultNodeInfo)
		require.True(t, ok)
		assert.True(t, peerC.IsOutbound(), "all lp2p peers are bi-directional")
	})

	t.Run("NetInfoIteration", func(t *testing.T) {
		// Simulates the /net_info ForEach loop from rpc/core/net.go
		// to verify the libp2p Peer works end-to-end with the RPC handler.
		ctx := context.Background()

		hosts := makeTestHosts(t, 4)
		hostA := hosts[0]

		ps := NewPeerSet(hostA, p2p.NopMetrics(), log.NewNopLogger())

		for i := 1; i < len(hosts); i++ {
			err := hostA.Connect(ctx, hosts[i].AddrInfo())
			require.NoError(t, err)

			_, err = ps.Add(hosts[i].AddrInfo(), PeerAddOptions{})
			require.NoError(t, err)
		}

		// ACT — exact pattern from rpc/core/net.go NetInfo()
		count := 0

		ps.ForEach(func(peer p2p.Peer) {
			// These are the fields /net_info accesses
			_, ok := peer.NodeInfo().(p2p.DefaultNodeInfo)
			require.True(t, ok, "NodeInfo must type-assert to p2p.DefaultNodeInfo")
			_ = peer.IsOutbound()
			_ = peer.RemoteIP().String()
			_ = peer.Status()

			count++
		})

		assert.Equal(t, 3, count)
	})

	t.Run("GetSet", func(t *testing.T) {
		// ARRANGE
		hosts := makeTestHosts(t, 2)
		hostA := hosts[0]
		hostB := hosts[1]

		peerB, err := NewPeer(hostA, hostB.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		const key = "test-key"
		expected := map[string]int{"value": 42}

		// ACT
		got1 := peerB.Get(key)
		peerB.Set(key, expected)
		actual := peerB.Get(key)

		// ASSERT
		require.Nil(t, got1)
		require.NotNil(t, actual)
		require.Equal(t, expected, actual)
	})

	t.Run("Send", func(t *testing.T) {
		// ARRANGE
		var (
			ctx   = context.Background()
			hosts = makeTestHosts(t, 2)
			hostA = hosts[0]
			hostB = hosts[1]
		)

		err := hostA.Connect(ctx, hostB.AddrInfo())
		require.NoError(t, err)

		// given peerB in hostA
		peerB, err := NewPeer(hostA, hostB.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		const channelID = byte(0xaa)

		payloadCh := make(chan []byte, 10)
		defer close(payloadCh)

		hostB.SetStreamHandler(ProtocolID(channelID), func(stream network.Stream) {
			payload, err := StreamReadClose(stream)
			require.NoError(t, err)

			payloadCh <- payload
		})

		// ACT
		// send two messages
		send1 := peerB.Send(p2p.Envelope{
			ChannelID: channelID,
			Message:   types.ToRequestEcho("hello-send"),
		})
		send2 := peerB.TrySend(p2p.Envelope{
			ChannelID: channelID,
			Message:   types.ToRequestEcho("hello-try-send"),
		})

		// ASSERT
		require.True(t, send1)
		require.True(t, send2)

		// check msg1
		msg1 := &types.Request{}
		require.NoError(t, msg1.Unmarshal(<-payloadCh))

		// check msg2
		msg2 := &types.Request{}
		require.NoError(t, msg2.Unmarshal(<-payloadCh))

		require.ElementsMatch(
			t,
			[]string{"hello-send", "hello-try-send"},
			[]string{msg1.GetEcho().GetMessage(), msg2.GetEcho().GetMessage()},
		)

		t.Run("Failure", func(t *testing.T) {
			// ACT
			send := peerB.Send(p2p.Envelope{
				ChannelID: channelID,
				Message:   nil, // invalid proto message
			})

			// ASSERT
			require.False(t, send)
		})
	})
}
