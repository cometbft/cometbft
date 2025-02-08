package p2p

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	"github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/kcp"
	"github.com/cometbft/cometbft/p2p/transport/quic"
	"github.com/cometbft/cometbft/p2p/transport/tcp"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

type TransportFactory struct {
	logger    log.Logger
	config    *config.P2PConfig
	nodeKey   nodekey.NodeKey
	useLibp2p bool // Feature flag to enable libp2p
	useQUIC   bool // Feature flag to enable QUIC
	useKCP    bool // Feature flag to enable KCP
}

func NewTransportFactory(logger log.Logger, config *config.Config) *TransportFactory {
	nk, err := LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(fmt.Sprintf("failed to load node key: %v", err))
	}
	return &TransportFactory{
		logger:    logger,
		config:    config.P2P,
		nodeKey:   *nk,
		useLibp2p: false,
		useQUIC:   config.P2P.UseQUIC,
		useKCP:    config.P2P.UseKCP,
	}
}

// CreateTransports creates all supported transports
func (f *TransportFactory) CreateTransports() ([]transport.Transport, error) {
	var transports []transport.Transport

	// KCP Transport (default)
	kcpOpts := &kcp.Options{
		DataShards:     10,
		ParityShards:   3,
		MaxWindowSize:  32768,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		NoDelay:        1,
		FastResend:     1,
		CongestionCtrl: false,
		RTO:            200,
		MTU:            1400,
	}
	kcpTransport, err := kcp.NewTransport(kcpOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create KCP transport: %w", err)
	}

	kcpAddr, err := netaddr.NewFromString("0.0.0.0:28656")
	if err != nil {
		return nil, fmt.Errorf("failed to create KCP address: %w", err)
	}

	if err := kcpTransport.Listen(*kcpAddr); err != nil {
		return nil, fmt.Errorf("failed to listen on KCP: %w", err)
	}
	transports = append(transports, kcpTransport)

	// TCP Transport
	tcpTransport := tcp.NewMultiplexTransport(f.nodeKey, conn.DefaultMConnConfig())
	tcpAddr, err := netaddr.NewFromString("0.0.0.0:26656")
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP address: %w", err)
	}
	if err := tcpTransport.Listen(*tcpAddr); err != nil {
		return nil, fmt.Errorf("failed to listen on TCP: %w", err)
	}
	transports = append(transports, tcpTransport)

	// QUIC Transport
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // TODO: Configure proper TLS for production
	}
	quicOpts := &quic.Options{
		TLSConfig:          tlsConfig,
		MaxIncomingStreams: 100,
		KeepAlivePeriod:    30 * time.Second,
		IdleTimeout:        5 * time.Minute,
	}
	quicTransport, err := quic.NewTransport(quicOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create QUIC transport: %w", err)
	}

	quicAddr, err := netaddr.NewFromString("0.0.0.0:27656")
	if err != nil {
		return nil, fmt.Errorf("failed to create QUIC address: %w", err)
	}
	if err := quicTransport.Listen(*quicAddr); err != nil {
		return nil, fmt.Errorf("failed to listen on QUIC: %w", err)
	}
	transports = append(transports, quicTransport)

	return transports, nil
}
