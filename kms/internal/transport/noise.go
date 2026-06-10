// Package transport provides privval connection transports for cometkms.
package transport

import (
	"context"
	"fmt"
	"net"
	"time"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/lp2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pnoise "github.com/libp2p/go-libp2p/p2p/security/noise"
)

// NoiseDialer returns a privval.SocketDialer that TCP-dials addr (host:port) and
// secures the connection with libp2p Noise, using identity as the KMS's libp2p
// key and pinning the validator's peer ID. The returned net.Conn is a
// sec.SecureConn and feeds the existing serve loop unchanged.
func NoiseDialer(addr string, identity cmtcrypto.PrivKey, validator peer.ID, timeout time.Duration) (privval.SocketDialer, error) {
	lpk, err := lp2p.PrivateKeyFromCosmosKey(identity)
	if err != nil {
		return nil, fmt.Errorf("cometkms: noise identity: %w", err)
	}
	tr, err := libp2pnoise.New(libp2pnoise.ID, lpk, nil)
	if err != nil {
		return nil, fmt.Errorf("cometkms: noise transport: %w", err)
	}
	return func() (net.Conn, error) {
		raw, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		secured, err := tr.SecureOutbound(ctx, raw, validator)
		if err != nil {
			_ = raw.Close()
			return nil, fmt.Errorf("cometkms: noise handshake with %s: %w", validator, err)
		}
		return secured, nil
	}, nil
}
