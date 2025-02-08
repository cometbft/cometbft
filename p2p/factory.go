package p2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

// TransportFactory creates either the default TCP transport or libp2p transport
// based on configuration.
type TransportFactory struct {
	logger    log.Logger
	config    *config.P2PConfig
	nodeKey   nodekey.NodeKey
	useLibp2p bool // Feature flag to enable libp2p
}

// NewTransportFactory creates a new transport factory.
func NewTransportFactory(logger log.Logger, config *config.Config) *TransportFactory {
	// Obtain your nodeKey from file or generate it as needed.
	nk, err := LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(fmt.Sprintf("failed to load node key: %v", err))
	}
	return &TransportFactory{
		logger:    logger,
		config:    config.P2P,
		nodeKey:   *nk,
		useLibp2p: false, // TODO: optionally read from config or environment.
	}
}

// CreateTransport returns either TCP or libp2p transport based on configuration.
func (f *TransportFactory) CreateTransport() (transport.Transport, error) {
	if f.useLibp2p {
		// TODO: Implement libp2p transport here.
		return nil, fmt.Errorf("libp2p transport not yet implemented")
	}

	// Return the existing TCP transport.
	return tcp.NewMultiplexTransport(
		f.nodeKey,
		conn.DefaultMConnConfig(),
	), nil
}

// CreateSwitch creates a new switch with the appropriate transport.
func (f *TransportFactory) CreateSwitch() (*Switch, error) {
	tr, err := f.CreateTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	sw := NewSwitch(
		f.config,
		tr,
	)
	sw.SetLogger(f.logger)
	return sw, nil
}

// SetLibp2p enables or disables libp2p transport.
func (f *TransportFactory) SetLibp2p(enabled bool) {
	f.useLibp2p = enabled
}
