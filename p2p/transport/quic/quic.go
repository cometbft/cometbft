package quic

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	quic "github.com/quic-go/quic-go"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
)

type QUIC struct {
	*quic.Transport
	lintener *quic.Listener
}

var _ transport.Transport = (*QUIC)(nil)

// Listen starts listening for incoming QUIC connections.
//
// see net.ResolveUDPAddr.
func Listen(address string, tlsConfig *tls.Config) (*QUIC, error) {
	netUDPAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", netUDPAddr)
	if err != nil {
		return nil, err
	}

	tr := &quic.Transport{
		Conn: conn,
	}

	quicConfig := &quic.Config{
		KeepAlivePeriod: 5 * time.Second,
	}
	ln, err := tr.Listen(tlsConfig, quicConfig)

	return &QUIC{
		Transport: tr,
		lintener:  ln,
	}, nil
}

func (q *QUIC) NetAddr() na.NetAddr {
	panic("implement me")
}

func (q *QUIC) Accept() (transport.Conn, na.NetAddr, error) {
	conn, err := q.lintener.Accept(context.Background())
	if err != nil {
		return nil, na.NetAddr{}, err
	}

	return &Conn{Connection: conn}, na.NetAddr{}, nil
}

func (q *QUIC) Dial(addr na.NetAddr) (transport.Conn, error) {
	panic("implement me")
}
