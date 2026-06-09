package manager

import (
	"errors"
	"net"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/privval"
)

// errDialerStopped is returned by a backoff dialer after its stop channel closes.
var errDialerStopped = errors.New("cometkms: dialer stopped")

// backoffDialer wraps a one-shot SocketDialer so it blocks, retrying with capped
// exponential backoff, until it obtains a connection or stop is closed. Because
// it blocks until success, privval's serviceLoop never exits on transient
// outages — giving cometkms full control over reconnect cadence.
func backoffDialer(base privval.SocketDialer, stop <-chan struct{}, logger log.Logger, initial, max time.Duration) privval.SocketDialer {
	return func() (net.Conn, error) {
		wait := initial
		for {
			select {
			case <-stop:
				return nil, errDialerStopped
			default:
			}

			conn, err := base()
			if err == nil {
				return conn, nil
			}
			logger.Error("cometkms: dial failed; backing off", "wait", wait, "err", err)

			select {
			case <-stop:
				return nil, errDialerStopped
			case <-time.After(wait):
			}
			if wait *= 2; wait > max {
				wait = max
			}
		}
	}
}
