package lp2p

import (
	"testing"

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
