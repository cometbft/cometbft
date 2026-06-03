package manager

import (
	"net"
	"sync"
	"time"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/privval"
	privvalproto "github.com/cometbft/cometbft/proto/tendermint/privval"
	"github.com/cometbft/cometbft/types"
)

const (
	defaultDialTimeout    = 5 * time.Second
	defaultReadTimeout    = 10 * time.Second // must exceed the validator's ping interval (~3.3s at default)
	defaultWriteTimeout   = 5 * time.Second
	defaultBackoffInitial = 200 * time.Millisecond
	defaultBackoffMax     = 10 * time.Second

	// maxRemoteSignerMsgSize mirrors privval's framing cap (10 KiB).
	maxRemoteSignerMsgSize = 1024 * 10
)

// ValidatorConn describes one outbound signer connection.
type ValidatorConn struct {
	ChainID     string
	Addr        string
	IdentityKey crypto.PrivKey
	Signer      types.PrivValidator
	// Reconnect controls whether the manager re-dials after an established
	// connection drops. The initial connect always uses backoff regardless.
	Reconnect bool
}

// Manager supervises one validator connection per ValidatorConn, dialing out and
// re-dialing with backoff across outages (including graceful validator restarts).
type Manager struct {
	logger   log.Logger
	conns    []ValidatorConn
	stop     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// New builds a Manager.
func New(logger log.Logger, conns []ValidatorConn) *Manager {
	return &Manager{logger: logger, conns: conns, stop: make(chan struct{})}
}

// Start launches one supervised goroutine per connection. It never returns an
// error (kept in the signature for API stability and future validation).
func (m *Manager) Start() error {
	for _, c := range m.conns {
		m.wg.Add(1)
		go m.run(c)
	}
	return nil
}

// Stop signals all connections to close and waits for their goroutines to exit.
// Safe to call multiple times.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() { close(m.stop) })
	m.wg.Wait()
}

// run maintains one validator connection: dial (blocking, with backoff), serve
// until the connection breaks (EOF or timeout), then redial — until Stop.
func (m *Manager) run(c ValidatorConn) {
	defer m.wg.Done()

	logger := m.logger.With("chain", c.ChainID, "addr", c.Addr)
	base := privval.DialTCPFn(c.Addr, defaultDialTimeout, c.IdentityKey)
	dialer := backoffDialer(base, m.stop, logger, defaultBackoffInitial, defaultBackoffMax)

	for {
		select {
		case <-m.stop:
			return
		default:
		}

		conn, err := dialer()
		if err != nil {
			return // errDialerStopped: shutting down
		}
		logger.Info("cometkms: connected")

		// Close the conn promptly on shutdown so a blocked read unblocks.
		closed := make(chan struct{})
		go func() {
			select {
			case <-m.stop:
				_ = conn.Close()
			case <-closed:
			}
		}()

		m.serveConn(conn, c.ChainID, c.Signer, logger)

		close(closed)
		_ = conn.Close()

		if !c.Reconnect {
			logger.Info("cometkms: reconnect disabled; connection closed")
			return
		}
		logger.Info("cometkms: connection closed; will redial")
	}
}

// serveConn reads validation requests off conn and answers them with the reused
// privval.DefaultValidationRequestHandler, until any read/write error (EOF on a
// graceful restart, or a read-deadline timeout on a silently dead peer) or until
// stop. Returning signals the caller to redial.
func (m *Manager) serveConn(conn net.Conn, chainID string, signer types.PrivValidator, logger log.Logger) {
	reader := protoio.NewDelimitedReader(conn, maxRemoteSignerMsgSize)
	writer := protoio.NewDelimitedWriter(conn)

	for {
		select {
		case <-m.stop:
			return
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(defaultReadTimeout)); err != nil {
			logger.Error("cometkms: set read deadline", "err", err)
			return
		}
		var req privvalproto.Message
		if _, err := reader.ReadMsg(&req); err != nil {
			logger.Info("cometkms: read failed; dropping connection", "err", err)
			return
		}

		resp, err := privval.DefaultValidationRequestHandler(signer, req, chainID)
		kind, kv, ping := describeRequest(req)
		switch {
		case err != nil:
			// resp already carries an embedded RemoteSignerError; log loudly and still send it.
			logger.Error("cometkms: "+kind+" request rejected", append(kv, "err", err)...)
		case ping:
			logger.Debug("cometkms: ping")
		case kind == "pubkey":
			logger.Info("cometkms: served pubkey request")
		default:
			logger.Info("cometkms: signed "+kind, kv...)
		}

		if err := conn.SetWriteDeadline(time.Now().Add(defaultWriteTimeout)); err != nil {
			logger.Error("cometkms: set write deadline", "err", err)
			return
		}
		if _, err := writer.WriteMsg(&resp); err != nil {
			logger.Error("cometkms: write failed; dropping connection", "err", err)
			return
		}
	}
}

// describeRequest returns a short kind, structured log keyvals, and whether the
// request is a ping (which is logged at debug to avoid flooding the log every
// few seconds). It is used to emit one log line per served request.
func describeRequest(req privvalproto.Message) (kind string, kv []any, ping bool) {
	switch r := req.Sum.(type) {
	case *privvalproto.Message_SignVoteRequest:
		if v := r.SignVoteRequest.GetVote(); v != nil {
			return "vote", []any{"height", v.Height, "round", v.Round, "type", v.Type}, false
		}
		return "vote", nil, false
	case *privvalproto.Message_SignProposalRequest:
		if p := r.SignProposalRequest.GetProposal(); p != nil {
			return "proposal", []any{"height", p.Height, "round", p.Round}, false
		}
		return "proposal", nil, false
	case *privvalproto.Message_PubKeyRequest:
		return "pubkey", nil, false
	case *privvalproto.Message_PingRequest:
		return "ping", nil, true
	default:
		return "unknown", nil, false
	}
}
