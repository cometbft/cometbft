package mock

import (
	"net"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

type Peer struct {
	*service.BaseService
	ip                   net.IP
	id                   nodekey.ID
	addr                 *na.NetAddr
	kv                   map[string]any
	Outbound, Persistent bool
	server, client       net.Conn
}

// NewPeer creates and starts a new mock peer. If the ip
// is nil, random routable address is used.
func NewPeer(ip net.IP) *Peer {
	var netAddr *na.NetAddr
	if ip == nil {
		_, netAddr = na.CreateRoutableAddr()
	} else {
		netAddr = na.NewFromIPPort(ip, 26656)
	}
	nodeKey := nodekey.NodeKey{PrivKey: ed25519.GenPrivKey()}
	netAddr.ID = nodeKey.ID()
	server, client := net.Pipe()
	mp := &Peer{
		ip:     ip,
		id:     nodeKey.ID(),
		addr:   netAddr,
		kv:     make(map[string]any),
		server: server,
		client: client,
	}
	mp.BaseService = service.NewBaseService(nil, "MockPeer", mp)
	if err := mp.Start(); err != nil {
		panic(err)
	}
	return mp
}

func (mp *Peer) FlushStop() { mp.Stop() } //nolint:errcheck //ignore error
func (mp *Peer) OnStop() {
	mp.server.Close()
	mp.client.Close()
}
func (*Peer) HasChannel(_ byte) bool                           { return true }
func (*Peer) TrySend(_ p2p.Envelope) bool                      { return true }
func (*Peer) Send(_ p2p.Envelope) bool                         { return true }
func (mp *Peer) TrySendMarshalled(p2p.MarshalledEnvelope) bool { return true }
func (mp *Peer) NodeInfo() ni.NodeInfo {
	return ni.Default{
		DefaultNodeID: mp.addr.ID,
		ListenAddr:    mp.addr.DialString(),
	}
}
func (*Peer) Status() conn.ConnectionStatus { return conn.ConnectionStatus{} }
func (mp *Peer) ID() nodekey.ID             { return mp.id }
func (mp *Peer) IsOutbound() bool           { return mp.Outbound }
func (mp *Peer) IsPersistent() bool         { return mp.Persistent }
func (mp *Peer) Get(key string) any {
	if value, ok := mp.kv[key]; ok {
		return value
	}
	return nil
}

func (mp *Peer) Set(key string, value any) {
	mp.kv[key] = value
}
func (mp *Peer) RemoteIP() net.IP        { return mp.ip }
func (mp *Peer) SocketAddr() *na.NetAddr { return mp.addr }
func (mp *Peer) RemoteAddr() net.Addr    { return &net.TCPAddr{IP: mp.ip, Port: 8800} }
func (mp *Peer) Conn() net.Conn          { return mp.server }
func (*Peer) SetRemovalFailed()          {}
func (*Peer) GetRemovalFailed() bool     { return false }
