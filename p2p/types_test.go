package p2p

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	p2pproto "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
)

func TestEnvelopeMarshalMessage(t *testing.T) {
	msg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "0",
			},
		},
	}
	expectedMsgBytes, err := proto.Marshal(msg.Wrap())
	require.NoError(t, err)

	envelope := Envelope{
		Message: msg,
	}
	msgBytes, err := envelope.marshalMessage()
	require.NoError(t, err)
	require.Equal(t, expectedMsgBytes, msgBytes)
	require.Equal(t, expectedMsgBytes, envelope.messageBytes)
}
