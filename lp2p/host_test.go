package lp2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	corepeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// Given sample protocol ids
	var (
		channelFoo = byte(0xaa)
		channelBar = byte(0xbb)
		protoFoo   = ProtocolID(channelFoo)
		protoBar   = ProtocolID(channelBar)
	)

	// Given 2 available ports
	ports := utils.GetFreePorts(t, 2)

	// Given two hosts that are connected to each other
	host1 := makeTestHost(t, ports[0], AddressBookConfig{})
	host2 := makeTestHost(t, ports[1], AddressBookConfig{
		Peers: []PeerConfig{
			{
				Host: fmt.Sprintf("127.0.0.1:%d", ports[0]),
				ID:   host1.ID().String(),
			},
		},
	})

	ConnectPeers(ctx, host2, host2.ConfigPeers())

	t.Logf("host1: %+v", host1.AddrInfo())
	t.Logf("host2: %+v", host2.AddrInfo())

	t.Cleanup(func() {
		host2.Close()
		host1.Close()
	})

	// Given sample envelope
	type envelope struct {
		protocol protocol.ID
		sender   corepeer.ID
		receiver corepeer.ID
		message  string
	}

	envelopes := []envelope{}
	mu := sync.Mutex{}

	// Given sample handler for both hosts
	handler := func(stream network.Stream) {
		var (
			conn     = stream.Conn()
			receiver = conn.LocalPeer()
			sender   = conn.RemotePeer()
		)

		if conn.ConnState().Transport != TransportQUIC {
			t.Fatalf("unexpected transport: %s", conn.ConnState().Transport)
			return
		}

		payload, err := StreamReadClose(stream)
		if err != nil {
			t.Fatalf("failed to read from stream originating from %s: %v", sender, err)
			return
		}

		msg := &types.Request{}
		require.NoError(t, msg.Unmarshal(payload))
		require.NotNil(t, msg.GetEcho())

		e := envelope{
			protocol: stream.Protocol(),
			sender:   sender,
			receiver: receiver,
			message:  msg.GetEcho().GetMessage(),
		}

		logMessage := e.message
		if len(logMessage) > 64 {
			logMessage = logMessage[:64] + "..."
		}

		t.Logf(
			"Received envelope: %s -> %s (proto %s): %s",
			e.sender.String(),
			e.receiver.String(),
			e.protocol,
			logMessage,
		)

		mu.Lock()
		defer mu.Unlock()

		envelopes = append(envelopes, e)
	}

	host1.SetStreamHandler(protoFoo, handler)
	host1.SetStreamHandler(protoBar, handler)

	host2.SetStreamHandler(protoFoo, handler)
	host2.SetStreamHandler(protoBar, handler)

	// Given counter peers
	host1Peer2, err := NewPeer(host1, host2.AddrInfo(), p2p.NopMetrics())
	require.NoError(t, err, "failed to create peer 1->2")
	require.NoError(t, host1Peer2.Start(), "failed to start peer 1->2")

	host2Peer1, err := NewPeer(host2, host1.AddrInfo(), p2p.NopMetrics())
	require.NoError(t, err, "failed to create peer 2->1")
	require.NoError(t, host2Peer1.Start(), "failed to start peer 2->1")

	t.Logf("host1Peer2: %+v", host1Peer2.ID())
	t.Logf("host2Peer1: %+v", host2Peer1.ID())

	// Given a long string
	// 300kb
	longStr := strings.Repeat("a", 300*1024)

	// ACT
	send1 := host1Peer2.Send(p2p.Envelope{
		ChannelID: channelFoo,
		Message:   types.ToRequestEcho("one two"),
	})

	send2 := host2Peer1.Send(p2p.Envelope{
		ChannelID: channelBar,
		Message:   types.ToRequestEcho("three four"),
	})

	send3 := host1Peer2.TrySend(p2p.Envelope{
		ChannelID: channelBar,
		Message:   types.ToRequestEcho(longStr),
	})

	// ASSERT
	// Ensure we've written to both streams
	require.True(t, send1, "failed to send message 1->2")
	require.True(t, send2, "failed to send message 2->1")
	require.True(t, send3, "failed to send message 1->2")

	// Ensure two envelopes have been received
	wait := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(envelopes) == 3
	}

	require.Eventually(t, wait, 500*time.Millisecond, 50*time.Millisecond)

	// Ensure both envelopes match the expected ones
	expectedEnvelopes := []envelope{
		{
			protocol: protoFoo,
			sender:   host1.ID(),
			receiver: host2.ID(),
			message:  "one two",
		},
		{
			protocol: protoBar,
			sender:   host2.ID(),
			receiver: host1.ID(),
			message:  "three four",
		},
		{
			protocol: protoBar,
			sender:   host1.ID(),
			receiver: host2.ID(),
			message:  longStr,
		},
	}

	require.ElementsMatch(t, expectedEnvelopes, envelopes)
}

func makeTestHost(t *testing.T, port int, addressBook AddressBookConfig) *Host {
	// config
	config := config.DefaultP2PConfig()
	config.RootDir = t.TempDir()
	config.ListenAddress = fmt.Sprintf("127.0.0.1:%d", port)
	config.ExternalAddress = fmt.Sprintf("127.0.0.1:%d", port)

	config.LibP2PConfig.Enabled = true

	// private key
	pk := ed25519.GenPrivKey()

	// lib p2p host
	host, err := NewHost(config, pk, addressBook, log.TestingLogger())
	require.NoError(t, err)

	return host
}
