package privval

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/lp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/sec"
	libp2pnoise "github.com/libp2p/go-libp2p/p2p/security/noise"
)

const noiseScheme = "noise://"

type NoiseListener struct {
	net.Listener

	transport        *libp2pnoise.Transport
	allowed          map[peer.ID]struct{}
	timeoutAccept    time.Duration
	timeoutReadWrite time.Duration
	logger           log.Logger
}

// sec.SecureConn (returned by Accept) is a net.Conn, so SignerListenerEndpoint
// consumes a NoiseListener exactly like a TCP/Unix listener.
var _ net.Conn = (sec.SecureConn)(nil)

type NoiseListenerOption func(*NoiseListener)

func NoiseListenerTimeoutAccept(d time.Duration) NoiseListenerOption {
	return func(nl *NoiseListener) { nl.timeoutAccept = d }
}

func NoiseListenerTimeoutReadWrite(d time.Duration) NoiseListenerOption {
	return func(nl *NoiseListener) { nl.timeoutReadWrite = d }
}

func NoiseListenerLogger(l log.Logger) NoiseListenerOption {
	return func(nl *NoiseListener) { nl.logger = l }
}

// NewNoiseListener wraps ln. key is the node's cometbft private key (its libp2p
// identity); allowed is the set of signer peer IDs permitted to connect.
//
// An empty allowed set means "accept any authenticated peer" and MUST NOT be
// used in production — it is intended only for tests. The production caller
// (NewSignerListenerFromAddr) always passes a non-empty allowlist.
func NewNoiseListener(ln net.Listener, key crypto.PrivKey, allowed []peer.ID, opts ...NoiseListenerOption) (*NoiseListener, error) {
	lpk, err := lp2p.PrivateKeyFromCosmosKey(key)
	if err != nil {
		return nil, fmt.Errorf("privval: noise identity: %w", err)
	}
	tr, err := libp2pnoise.New(libp2pnoise.ID, lpk, nil)
	if err != nil {
		return nil, fmt.Errorf("privval: noise transport: %w", err)
	}
	set := make(map[peer.ID]struct{}, len(allowed))
	for _, p := range allowed {
		set[p] = struct{}{}
	}
	nl := &NoiseListener{
		Listener:         ln,
		transport:        tr,
		allowed:          set,
		timeoutAccept:    defaultTimeoutAcceptSeconds * time.Second,
		timeoutReadWrite: defaultTimeoutReadWriteSeconds * time.Second,
		logger:           log.NewNopLogger(),
	}
	for _, o := range opts {
		o(nl)
	}
	return nl, nil
}

func (nl *NoiseListener) Accept() (net.Conn, error) {
	if tcpLn, ok := nl.Listener.(*net.TCPListener); ok {
		_ = tcpLn.SetDeadline(time.Now().Add(nl.timeoutAccept))
	}
	rawConn, err := nl.Listener.Accept()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), nl.timeoutReadWrite)
	defer cancel()

	secured, err := nl.transport.SecureInbound(ctx, rawConn, "")
	if err != nil {
		_ = rawConn.Close()
		return nil, fmt.Errorf("privval: noise handshake: %w", err)
	}

	if len(nl.allowed) > 0 {
		if _, ok := nl.allowed[secured.RemotePeer()]; !ok {
			nl.logger.Error("privval: rejecting non-allowlisted signer peer", "peer", secured.RemotePeer())
			_ = secured.Close()
			return nil, fmt.Errorf("privval: signer peer %s not allowlisted", secured.RemotePeer())
		}
	}
	return secured, nil
}

// ParseNoiseAddr parses "noise://<peer-id>@host:port".
func ParseNoiseAddr(addr string) (peer.ID, string, error) {
	if !strings.HasPrefix(addr, noiseScheme) {
		return "", "", fmt.Errorf("privval: not a noise address: %q", addr)
	}
	rest := strings.TrimPrefix(addr, noiseScheme)
	at := strings.LastIndex(rest, "@")
	if at < 0 {
		return "", "", fmt.Errorf("privval: noise address missing peer id: %q", addr)
	}
	pidStr, hostport := rest[:at], rest[at+1:]
	if hostport == "" {
		return "", "", fmt.Errorf("privval: noise address missing host:port: %q", addr)
	}
	pid, err := peer.Decode(pidStr)
	if err != nil {
		return "", "", fmt.Errorf("privval: invalid peer id %q: %w", pidStr, err)
	}
	return pid, hostport, nil
}
