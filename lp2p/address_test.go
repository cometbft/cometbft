package lp2p

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddressToMultiAddr(t *testing.T) {
	for _, tt := range []struct {
		name        string
		addr        string
		transport   string
		want        string
		errContains string
	}{
		{
			name:      "tcp to quic",
			addr:      "tcp://1.1.1.1:5678",
			transport: TransportQUIC,
			want:      "/ip4/1.1.1.1/udp/5678/quic-v1",
		},
		{
			name:      "just ip and port to quic",
			addr:      "1.1.1.1:5678",
			transport: TransportQUIC,
			want:      "/ip4/1.1.1.1/udp/5678/quic-v1",
		},
		{
			name:        "no port provided",
			addr:        "1.1.1.1",
			transport:   TransportQUIC,
			errContains: "port is empty",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddressToMultiAddr(tt.addr, tt.transport)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got.String())
		})
	}
}
