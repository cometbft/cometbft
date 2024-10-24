package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/netutil"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p/internal/fuzz"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

const (
	defaultDialTimeout      = time.Second
	defaultFilterTimeout    = 5 * time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// IPResolver is a behavior subset of net.Resolver.
type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// accept is the container to carry the upgraded connection from an
// asynchronously running routine to the Accept method.
type accept struct {
	netAddr *na.NetAddr
	conn    net.Conn
	err     error
}

// transportLifecycle bundles the methods for callers to control start and stop
// behavior.
type transportLifecycle interface {
	Close() error
	Listen(addr na.NetAddr) error
}

// ConnFilterFunc to be implemented by filter hooks after a new connection has
// been established. The set of existing connections is passed along together
// with all resolved IPs for the new connection.
type ConnFilterFunc func(ConnSet, net.Conn, []net.IP) error

// ConnDuplicateIPFilter resolves and keeps all ips for an incoming connection
// and refuses new ones if they come from a known ip.
func ConnDuplicateIPFilter() ConnFilterFunc {
	return func(cs ConnSet, c net.Conn, ips []net.IP) error {
		for _, ip := range ips {
			if cs.HasIP(ip) {
				return ErrRejected{
					conn:        c,
					err:         fmt.Errorf("ip<%v> already connected", ip),
					isDuplicate: true,
				}
			}
		}

		return nil
	}
}

// MultiplexTransportOption sets an optional parameter on the
// MultiplexTransport.
type MultiplexTransportOption func(*MultiplexTransport)

// MultiplexTransportConnFilters sets the filters for rejection new connections.
func MultiplexTransportConnFilters(
	filters ...ConnFilterFunc,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.connFilters = filters }
}

// MultiplexTransportFilterTimeout sets the timeout waited for filter calls to
// return.
func MultiplexTransportFilterTimeout(
	timeout time.Duration,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.filterTimeout = timeout }
}

// MultiplexTransportResolver sets the Resolver used for ip lokkups, defaults to
// net.DefaultResolver.
func MultiplexTransportResolver(resolver IPResolver) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.resolver = resolver }
}

// MultiplexTransportMaxIncomingConnections sets the maximum number of
// simultaneous connections (incoming). Default: 0 (unlimited).
func MultiplexTransportMaxIncomingConnections(n int) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.maxIncomingConnections = n }
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	netAddr                na.NetAddr
	listener               net.Listener
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	conns       ConnSet
	connFilters []ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeKey          nodekey.NodeKey
	resolver         IPResolver

	// TODO(xla): This config is still needed as we parameterise peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig conn.MConnConfig
}

// Test multiplexTransport for interface completeness.
var (
	_ transportLifecycle = (*MultiplexTransport)(nil)
)

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(
	nodeKey nodekey.NodeKey,
	mConfig conn.MConnConfig,
) *MultiplexTransport {
	return &MultiplexTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      defaultDialTimeout,
		filterTimeout:    defaultFilterTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          mConfig,
		nodeKey:          nodeKey,
		conns:            NewConnSet(),
		resolver:         net.DefaultResolver,
	}
}

// NetAddr implements Transport.
func (mt *MultiplexTransport) NetAddr() na.NetAddr {
	return mt.netAddr
}

// Accept implements Transport.
func (mt *MultiplexTransport) Accept() (net.Conn, *na.NetAddr, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, nil, a.err
		}

		// cfg.outbound = false

		return a.conn, a.netAddr, nil
	case <-mt.closec:
		return nil, nil, ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *MultiplexTransport) Dial(
	addr na.NetAddr,
) (net.Conn, error) {
	c, err := addr.DialTimeout(mt.dialTimeout)
	if err != nil {
		return nil, err
	}

	if mt.mConfig.TestFuzz {
		// so we have time to do peer handshakes and get set up.
		c = fuzz.ConnAfterFromConfig(c, 10*time.Second, mt.mConfig.TestFuzzConfig)
	}

	// TODO(xla): Evaluate if we should apply filters if we explicitly dial.
	if err := mt.filterConn(c); err != nil {
		return nil, err
	}

	secretConn, err := mt.upgrade(c, &addr)
	if err != nil {
		return nil, err
	}

	return secretConn, nil
}

// Close implements transportLifecycle.
func (mt *MultiplexTransport) Close() error {
	close(mt.closec)

	if mt.listener != nil {
		return mt.listener.Close()
	}

	return nil
}

// Listen implements transportLifecycle.
func (mt *MultiplexTransport) Listen(addr na.NetAddr) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return err
	}

	if mt.maxIncomingConnections > 0 {
		ln = netutil.LimitListener(ln, mt.maxIncomingConnections)
	}

	mt.netAddr = addr
	mt.listener = ln

	go mt.acceptPeers()

	return nil
}

func (mt *MultiplexTransport) acceptPeers() {
	for {
		c, err := mt.listener.Accept()
		if err != nil {
			// If Close() has been called, silently exit.
			select {
			case _, ok := <-mt.closec:
				if !ok {
					return
				}
			default:
				// Transport is not closed
			}

			mt.acceptc <- accept{err: err}
			return
		}

		// Connection upgrade and filtering should be asynchronous to avoid
		// Head-of-line blocking[0].
		// Reference:  https://github.com/tendermint/tendermint/issues/2047
		//
		// [0] https://en.wikipedia.org/wiki/Head-of-line_blocking
		go func(c net.Conn) {
			defer func() {
				if r := recover(); r != nil {
					err := ErrRejected{
						conn:          c,
						err:           fmt.Errorf("recovered from panic: %v", r),
						isAuthFailure: true,
					}
					select {
					case mt.acceptc <- accept{err: err}:
					case <-mt.closec:
						// Give up if the transport was closed.
						_ = c.Close()
						return
					}
				}
			}()

			var (
				secretConn *conn.SecretConnection
				netAddr    *na.NetAddr
			)

			err := mt.filterConn(c)
			if err == nil {
				secretConn, err = mt.upgrade(c, nil)
				if err == nil {
					addr := c.RemoteAddr()
					id := nodekey.PubKeyToID(secretConn.RemotePubKey())
					netAddr = na.New(id, addr)
				}
			}

			select {
			case mt.acceptc <- accept{netAddr, secretConn, err}:
				// Make the upgraded peer available.
			case <-mt.closec:
				// Give up if the transport was closed.
				_ = c.Close()
				return
			}
		}(c)
	}
}

// Cleanup removes the given address from the connections set and
// closes the connection.
func (mt *MultiplexTransport) Cleanup(c net.Conn) error {
	mt.conns.Remove(c)
	return c.Close()
}

func (mt *MultiplexTransport) filterConn(c net.Conn) (err error) {
	defer func() {
		if err != nil {
			_ = c.Close()
		}
	}()

	// Reject if connection is already present.
	if mt.conns.Has(c) {
		return ErrRejected{conn: c, isDuplicate: true}
	}

	// Resolve ips for incoming conn.
	ips, err := resolveIPs(mt.resolver, c)
	if err != nil {
		return err
	}

	errc := make(chan error, len(mt.connFilters))

	for _, f := range mt.connFilters {
		go func(f ConnFilterFunc, c net.Conn, ips []net.IP, errc chan<- error) {
			errc <- f(mt.conns, c, ips)
		}(f, c, ips, errc)
	}

	for i := 0; i < cap(errc); i++ {
		select {
		case err := <-errc:
			if err != nil {
				return ErrRejected{conn: c, err: err, isFiltered: true}
			}
		case <-time.After(mt.filterTimeout):
			return ErrFilterTimeout{}
		}
	}

	mt.conns.Set(c, ips)

	return nil
}

func (mt *MultiplexTransport) upgrade(
	c net.Conn,
	dialedAddr *na.NetAddr,
) (secretConn *conn.SecretConnection, err error) {
	defer func() {
		if err != nil {
			_ = mt.Cleanup(c)
		}
	}()

	secretConn, err = upgradeSecretConn(c, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	if err != nil {
		return nil, ErrRejected{
			conn:          c,
			err:           fmt.Errorf("secret conn failed: %w", err),
			isAuthFailure: true,
		}
	}

	// For outgoing conns, ensure connection key matches dialed key.
	connID := nodekey.PubKeyToID(secretConn.RemotePubKey())
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID != dialedID {
			return nil, ErrRejected{
				conn: c,
				id:   connID,
				err: fmt.Errorf(
					"conn.ID (%v) dialed ID (%v) mismatch",
					connID,
					dialedID,
				),
				isAuthFailure: true,
			}
		}
	}

	return secretConn, nil
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func resolveIPs(resolver IPResolver, c net.Conn) ([]net.IP, error) {
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	addrs, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}

	ips := []net.IP{}

	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}

	return ips, nil
}
