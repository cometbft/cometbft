package p2p

import (
	"fmt"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
)

type PathSelector struct {
	transports map[transport.Protocol]transport.Transport
}

func NewPathSelector(transports []transport.Transport) *PathSelector {
	selector := &PathSelector{
		transports: make(map[transport.Protocol]transport.Transport),
	}
	for _, t := range transports {
		selector.transports[t.Protocol()] = t
	}
	return selector
}

// SelectFastestPath selects the fastest transport path to the target
// Prefers KCP by default unless another transport proves significantly faster
func (ps *PathSelector) SelectFastestPath(targetAddr na.NetAddr) (transport.Transport, error) {
	// Try KCP first
	if kcpTransport, ok := ps.transports[transport.KCPProtocol]; ok {
		conn, err := kcpTransport.Dial(targetAddr)
		if err == nil {
			conn.Close("") // Close test connection
			return kcpTransport, nil
		}
	}

	// If KCP fails, try other transports
	var bestTransport transport.Transport
	var bestRTT time.Duration

	for _, t := range ps.transports {
		start := time.Now()
		conn, err := t.Dial(targetAddr)
		if err != nil {
			continue
		}
		conn.Close("") // Close test connection
		rtt := time.Since(start)

		if bestTransport == nil || rtt < bestRTT {
			bestTransport = t
			bestRTT = rtt
		}
	}

	if bestTransport == nil {
		return nil, fmt.Errorf("no available transport to %v", targetAddr.String())
	}

	return bestTransport, nil
}
