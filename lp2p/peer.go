package lp2p

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/libs/cmap"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Peer struct {
	service.BaseService

	host *Host

	addrInfo peer.AddrInfo
	netAddr  *p2p.NetAddress

	data *cmap.CMap
}

var _ p2p.Peer = (*Peer)(nil)

func NewPeer(host *Host, addrInfo peer.AddrInfo) (*Peer, error) {
	netAddr, err := netAddressFromPeer(addrInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse net address: %w", err)
	}

	return &Peer{
		host:     host,
		addrInfo: addrInfo,
		netAddr:  netAddr,
		data:     cmap.NewCMap(),
	}, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{%s}", p.ID())
}

func (p *Peer) ID() p2p.ID {
	return p2p.ID(p.addrInfo.ID.String())
}

func (p *Peer) SocketAddr() *p2p.NetAddress {
	return p.netAddr
}

func (p *Peer) Get(key string) any {
	return p.data.Get(key)
}

func (p *Peer) Set(key string, value any) {
	p.data.Set(key, value)
}

// Send implements p2p.Peer.
func (p *Peer) Send(p2p.Envelope) bool {
	// todo implement
	// logic:
	// - skip if not running (todo how to check that peer is running?)
	// - skip if not having the channel (todo how to check that peer has the channel? do we need to check it at all?)
	// - marshal message
	// - SEND(channel_id, message_bytes) !!!! {just send to the peer via lib-p2p} [might return FALSE for timeout]
	// - collect metrics
	// - if okay, return TRUE
	return false
}

func (p *Peer) TrySend(p2p.Envelope) bool {
	// todo same as SEND, but if current queue is full (its cap=1), immediately return FALSE
	return false
}

func (p *Peer) CloseConn() error {
	return p.host.Network().ClosePeer(p.addrInfo.ID)
}

// These methods are not implemented as they're not used by reactors
// (only by PEX/p2p-transport which is not used with go-libp2p)

func (*Peer) Status() conn.ConnectionStatus { return conn.ConnectionStatus{} }
func (*Peer) NodeInfo() p2p.NodeInfo        { return nil }
func (*Peer) RemoteIP() net.IP              { return nil }
func (*Peer) RemoteAddr() net.Addr          { return nil }
func (*Peer) IsOutbound() bool              { return false }
func (*Peer) IsPersistent() bool            { return false }
func (*Peer) FlushStop()                    {}
func (*Peer) SetRemovalFailed()             {}
func (*Peer) GetRemovalFailed() bool        { return false }
