package lp2p

import (
	"testing"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestNewPreMarshalledMessage(t *testing.T) {
	// ARRANGE
	// Given msg
	echo := &types.RequestEcho{Message: "foo"}
	msg := &types.Request{
		Value: &types.Request_Echo{Echo: echo},
	}

	// Given cached message
	cached := newPreMarshaledMessage(msg)

	// ACT
	bzOriginal, err := marshalProto(msg)
	require.NoError(t, err)

	bzCached, err := marshalProto(cached)
	require.NoError(t, err)

	// ASSERT
	// cached raw payload should match the direct proto.Marshal output
	require.Equal(t, bzOriginal, bzCached)

	// ACT 2
	// alter original message
	echo.Message = "bar"

	bzOriginal, err = marshalProto(msg)
	require.NoError(t, err)

	// fully drop the message so it would panic on marshal
	msg.Value = nil
	require.Nil(t, cached.Message.(*types.Request).Value)

	bzCached, err = marshalProto(cached)
	require.NoError(t, err)

	// ASSERT 2
	// messages should be different because bzCached.payload is persisted
	require.NotEqual(t, bzOriginal, bzCached)
}

func TestProtoTypeName(t *testing.T) {
	var (
		echoReq       = &types.RequestEcho{Message: "foo"}
		echoReqCached = newPreMarshaledMessage(echoReq)
	)

	// ensure that pre-marshaled message returns the same name as the original message
	for _, tt := range []struct {
		msg  proto.Message
		want string
	}{
		{
			msg:  echoReq,
			want: "RequestEcho",
		},
		{
			msg:  echoReqCached,
			want: "RequestEcho",
		},
	} {
		got := protoTypeName(tt.msg)
		require.Equal(t, tt.want, got)
	}
}

func TestKeyConversion(t *testing.T) {
	secret := []byte("lp2p-test-secret")

	t.Run("ecdsa", func(t *testing.T) {
		// ARRANGE
		cosmosPK := secp256k1.GenPrivKeySecp256k1(secret)

		// ACT
		libp2pPK, err := PrivateKeyFromCosmosKey(cosmosPK)
		require.NoError(t, err)
		peerID, err := IDFromPrivateKey(cosmosPK)

		// ASSERT
		require.NoError(t, err)
		require.NotNil(t, libp2pPK)
		require.Equal(t, "16Uiu2HAmNt4G6MjXTVwSz5HoajXdpbApHvwBNCc5WX6peH3rosN9", peerID.String())
	})

	t.Run("eddsa", func(t *testing.T) {
		// ARRANGE
		cosmosPK := ed25519.GenPrivKeyFromSecret(secret)

		// ACT
		libp2pPK, err := PrivateKeyFromCosmosKey(cosmosPK)
		require.NoError(t, err)
		peerID, err := IDFromPrivateKey(cosmosPK)

		// ASSERT
		require.NoError(t, err)
		require.NotNil(t, libp2pPK)
		require.Equal(t, "12D3KooWE3h4xxbYockU6Y5EKhffWqo4AtMGwNS3UzmGkEFKGPZK", peerID.String())
	})
}
