package lp2p

import (
	"fmt"
	"reflect"

	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/gogoproto/proto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

func privateKeyFromCosmosKey(key cmcrypto.PrivKey) (crypto.PrivKey, error) {
	keyType := key.Type()

	switch keyType {
	case ed25519.KeyType:
		return crypto.UnmarshalEd25519PrivateKey(key.Bytes())
	case secp256k1.KeyType:
		return crypto.UnmarshalSecp256k1PrivateKey(key.Bytes())
	default:
		return nil, fmt.Errorf("unsupported private key type %q", keyType)
	}
}

func withAddressFactory(addr ma.Multiaddr) libp2p.Option {
	fn := func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return []ma.Multiaddr{addr}
	}

	return libp2p.AddrsFactory(fn)
}

func marshalProto(msg proto.Message) ([]byte, error) {
	// comet compatibility
	// @see p2p/peer.go (*peer).send()
	if w, ok := msg.(p2p.Wrapper); ok {
		msg = w.Wrap()
	}

	payload, err := proto.Marshal(msg)
	switch {
	case err != nil:
		return nil, errors.Wrapf(err, "failed to marshal proto")
	case len(payload) == 0:
		return nil, errors.New("payload is empty")
	}

	return payload, nil
}

func unmarshalProto(descriptor *p2p.ChannelDescriptor, payload []byte) (proto.Message, error) {
	var (
		msg = proto.Clone(descriptor.MessageType)
		err = proto.Unmarshal(payload, msg)
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal message")
	}

	// comet compatibility
	// @see p2p/peer.go createMConnection()
	if w, ok := msg.(p2p.Unwrapper); ok {
		msg, err = w.Unwrap()
		if err != nil {
			return nil, errors.Wrap(err, "failed to unwrap message")
		}
	}

	return msg, nil
}

func protoTypeName(msg proto.Message) string {
	return reflect.TypeOf(msg).Elem().Name()
}
