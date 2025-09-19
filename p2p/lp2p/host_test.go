package lp2p

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	corepeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// Given sample protocol id
	const protocolID = "/cometbft/foobar"

	// Given 2 available ports
	ports := utils.GetFreePorts(t, 2)

	// Given two hosts ...
	host1 := makeTestHost(t, ports[0])
	host2 := makeTestHost(t, ports[1])

	t.Logf("host1: %+v", host1.AddrInfo())
	t.Logf("host2: %+v", host2.AddrInfo())

	t.Cleanup(func() {
		host2.Close()
		host1.Close()
	})

	// Given sample handler for both hosts
	type envelope struct {
		sender   corepeer.ID
		receiver corepeer.ID
		message  string
	}

	envelopes := []envelope{}
	mu := sync.Mutex{}

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

		msg, err := io.ReadAll(stream)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("failed to read from stream originating from %s: %v", sender, err)
			return
		}

		if err := stream.Close(); err != nil {
			t.Fatalf("failed to close stream originating from %s: %v", sender, err)
			return
		}

		e := envelope{
			sender:   sender,
			receiver: receiver,
			message:  string(msg),
		}

		mu.Lock()
		defer mu.Unlock()

		t.Logf(
			"Received envelope: %s -> %s: %s",
			sender.String(),
			receiver.String(),
			string(msg),
		)

		envelopes = append(envelopes, e)
	}

	host1.SetStreamHandler(protocolID, handler)
	host2.SetStreamHandler(protocolID, handler)

	// Given hosts are connected
	err := host1.Connect(ctx, host2.AddrInfo())
	require.NoError(t, err, "failed to connect hosts")

	// Given streams are created
	stream1to2, err := host1.NewStream(ctx, host2.ID(), protocolID)
	require.NoError(t, err, "failed to create stream 1->2")

	stream2to1, err := host2.NewStream(ctx, host1.ID(), protocolID)
	require.NoError(t, err, "failed to create stream 2->1")

	t.Cleanup(func() {
		stream1to2.Close()
		stream2to1.Close()
	})

	// ACT
	// Write host1 -> host2
	_, err1 := stream1to2.Write([]byte("one two"))
	require.NoError(t, stream1to2.CloseWrite(), "failed to close write stream 1->2")

	// Write host2 -> host1
	_, err2 := stream2to1.Write([]byte("three four"))
	require.NoError(t, stream2to1.CloseWrite(), "failed to close write stream 2->1")

	// ASSERT
	// Ensure we've written to both streams
	require.NoError(t, err1, "failed to write to stream 1->2")
	require.NoError(t, err2, "failed to write to stream 2->1")

	// Ensure two envelopes have been received
	wait := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(envelopes) == 2
	}

	require.Eventually(t, wait, 500*time.Millisecond, 50*time.Millisecond)

	// Ensure both envelopes match the expected ones
	expectedEnvelopes := []envelope{
		{sender: host1.ID(), receiver: host2.ID(), message: "one two"},
		{sender: host2.ID(), receiver: host1.ID(), message: "three four"},
	}

	require.ElementsMatch(t, expectedEnvelopes, envelopes)
}

func makeTestHost(t *testing.T, port int) *Host {
	// config
	config := config.DefaultP2PConfig()
	config.UseLibP2P = true
	config.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", port)
	config.ExternalAddress = fmt.Sprintf("tcp://127.0.0.1:%d", port)

	// private key
	pk := ed25519.GenPrivKey()

	// lib p2p host
	host, err := NewHost(config, pk)
	require.NoError(t, err)

	return host
}
