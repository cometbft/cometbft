package lp2p

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
)

func TestProtocolID(t *testing.T) {
	for _, tt := range []struct {
		channel  byte
		expected string
	}{
		{channel: 0x00, expected: "/p2p/cometbft/1.0.0/channel/0x00"},
		{channel: 0x01, expected: "/p2p/cometbft/1.0.0/channel/0x01"},
		{channel: 0x10, expected: "/p2p/cometbft/1.0.0/channel/0x10"},
		{channel: 0xff, expected: "/p2p/cometbft/1.0.0/channel/0xff"},
	} {
		require.Equal(t, protocol.ID(tt.expected), ProtocolID(tt.channel))
	}
}

func TestStreamRead(t *testing.T) {
	t.Run("ReadTooLargePayload", func(t *testing.T) {
		// ARRANGE
		var (
			ctx     = context.Background()
			protoID = ProtocolID(0xAA)
			ports   = utils.GetFreePorts(t, 2)
			host1   = makeTestHost(t, ports[0], withLogging())
			host2   = makeTestHost(t, ports[1], withLogging())
		)

		t.Cleanup(func() {
			host2.Close()
			host1.Close()
		})

		// connect hosts
		require.NoError(t, host2.Connect(ctx, host1.AddrInfo()))

		readErr := make(chan error, 1)
		host1.SetStreamHandler(protoID, func(stream network.Stream) {
			defer stream.Close()

			_, err := StreamRead(stream)
			readErr <- err
		})

		// create stream
		stream, err := host2.NewStream(ctx, host1.ID(), protoID)
		require.NoError(t, err)
		t.Cleanup(func() { _ = stream.Close() })

		tooLargeHeader := uint64ToUvarint(MaxStreamSize + 1)

		// ACT
		_, err = stream.Write(tooLargeHeader)
		require.NoError(t, err)
		require.NoError(t, stream.Close())

		// ASSERT
		err = <-readErr

		require.Error(t, err)
		require.ErrorContains(t, err, "payload is too large")
	})
}
