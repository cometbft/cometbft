package privval

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/libp2p/go-libp2p/core/peer"
)

// IsConnTimeout returns a boolean indicating whether the error is known to
// report that a connection timeout occurred. This detects both fundamental
// network timeouts, as well as ErrConnTimeout errors.
func IsConnTimeout(err error) bool {
	_, ok := errors.Unwrap(err).(timeoutError)
	switch {
	case errors.As(err, &EndpointTimeoutError{}):
		return true
	case ok:
		return true
	default:
		return false
	}
}

// NewSignerListener creates a new SignerListenerEndpoint using the corresponding listen address
func NewSignerListener(listenAddr string, logger log.Logger) (*SignerListenerEndpoint, error) {
	var listener net.Listener

	protocol, address := cmtnet.ProtocolAndAddress(listenAddr)
	ln, err := net.Listen(protocol, address)
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "unix":
		listener = NewUnixListener(ln)
	case "tcp":
		// TODO: persist this key so external signer can actually authenticate us
		listener = NewTCPListener(ln, ed25519.GenPrivKey())
	default:
		return nil, fmt.Errorf(
			"wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	pve := NewSignerListenerEndpoint(logger.With("module", "privval"), listener)

	return pve, nil
}

// NewSignerListenerFromAddr builds a SignerListenerEndpoint from a listen
// address, dispatching on scheme: tcp:// and unix:// behave exactly as
// NewSignerListener; noise://<signer-peer-id>@host:port secures the listener
// with libp2p Noise using nodeKey as identity and allowlists the given signer
// peer. nodeKey is only used by the noise scheme.
func NewSignerListenerFromAddr(listenAddr string, nodeKey crypto.PrivKey, logger log.Logger) (*SignerListenerEndpoint, error) {
	if strings.HasPrefix(listenAddr, noiseScheme) {
		signerPeer, hostport, err := ParseNoiseAddr(listenAddr)
		if err != nil {
			return nil, err
		}
		ln, err := net.Listen("tcp", hostport)
		if err != nil {
			return nil, err
		}
		nl, err := NewNoiseListener(ln, nodeKey, []peer.ID{signerPeer}, NoiseListenerLogger(logger.With("module", "privval")))
		if err != nil {
			_ = ln.Close()
			return nil, err
		}
		return NewSignerListenerEndpoint(logger.With("module", "privval"), nl), nil
	}
	return NewSignerListener(listenAddr, logger)
}

// GetFreeLocalhostAddrPort returns a free localhost:port address
func GetFreeLocalhostAddrPort() string {
	port, err := cmtnet.GetFreePort()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}
