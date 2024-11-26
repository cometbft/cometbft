package quic

import (
	"net"

	quic "github.com/quic-go/quic-go"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
)

type QUIC struct {
	*quic.Transport
	lintener *quic.EarlyListener
}

var _ transport.Transport = (*QUIC)(nil)

// Listen starts listening for incoming QUIC connections.
//
// see net.ResolveUDPAddr.
func Listen(address string, certFile, keyFile string) (*QUIC, error) {
	tlsConf, err := CreateTLSConfig(certFile, keyFile)
	if err != nil {
		return nil, err
	}

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

	quicConf := &quic.Config{Allow0RTT: true}
	ln, err := tr.ListenEarly(tlsConf, quicConf)

	return &QUIC{
		Transport: tr,
		lintener:  ln,
	}, nil
}

func (q *QUIC) NetAddr() na.NetAddr {
	panic("implement me")
}

func (q *QUIC) Accept() (transport.Conn, *na.NetAddr, error) {
	panic("implement me")
}

func (q *QUIC) Dial(addr na.NetAddr) (transport.Conn, error) {
	panic("implement me")
}
