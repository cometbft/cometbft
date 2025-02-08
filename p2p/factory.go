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

type TransportFactory struct {
	logger    log.Logger
	config    *config.P2PConfig
	nodeKey   nodekey.NodeKey
	useLibp2p bool // Feature flag to enable libp2p
}

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

func (f *TransportFactory) CreateTransport() (transport.Transport, error) {
	if f.useLibp2p {
		return nil, fmt.Errorf("libp2p transport not yet implemented")
	}

	// Return TCP transport by default
	return tcp.NewMultiplexTransport(
		f.nodeKey,
		conn.DefaultMConnConfig(),
	), nil
}
