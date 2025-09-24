package lp2p

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/cosmos/gogoproto/proto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Peer struct {
	service.BaseService

	host *Host

	addrInfo peer.AddrInfo
	netAddr  *p2p.NetAddress
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
	v, err := p.host.Peerstore().Get(p.addrInfo.ID, key)
	if err != nil {
		return nil
	}

	return v
}

func (p *Peer) Set(key string, value any) {
	//nolint:errcheck // always returns err=nil
	p.host.Peerstore().Put(p.addrInfo.ID, key, value)
}

// Send implements p2p.Peer.
func (p *Peer) Send(e p2p.Envelope) bool {
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "err", err)
		return false
	}

	return true
}

func (p *Peer) TrySend(e p2p.Envelope) bool {
	// todo same as SEND, but if current queue is full (its cap=1), immediately return FALSE
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "err", err)
		return false
	}

	return true
}

func (p *Peer) CloseConn() error {
	return p.host.Network().ClosePeer(p.addrInfo.ID)
}

func (p *Peer) send(e p2p.Envelope) error {
	// todo implement
	// - skip if not running (todo how to check that peer is running?)
	// - skip if not having the channel (todo how to check that peer has the channel? do we need to check it at all?)
	// - collect metrics

	payload, err := proto.Marshal(e.Message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	protocolID := ProtocolID(e.ChannelID)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutStream)
	defer cancel()

	s, err := p.host.NewStream(ctx, p.addrInfo.ID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream %s: %w", protocolID, err)
	}

	defer s.Close()

	return StreamWriteClose(s, payload)
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

// PeerSet represents lazy-initialized peer set adapter for go-libp2p.
// TODO: cache calculations
type PeerSet struct {
	host *Host

	peers map[peer.ID]*Peer
	mu    sync.RWMutex

	logger log.Logger
}

var _ p2p.IPeerSet = (*PeerSet)(nil)

func NewPeerSet(host *Host, logger log.Logger) *PeerSet {
	const initialCapacity = 64

	logger = logger.With("module", "lp2p_peer_set")

	return &PeerSet{
		host:   host,
		peers:  make(map[peer.ID]*Peer, initialCapacity),
		mu:     sync.RWMutex{},
		logger: logger,
	}
}

func (p *PeerSet) Has(key p2p.ID) bool {
	id := peer.ID(key)

	return len(p.host.Peerstore().Addrs(id)) > 0
}

func (p *PeerSet) HasIP(ip net.IP) bool {
	peers := p.existingPeerIDs()

	for _, peer := range peers {
		addrInfo := p.host.Peerstore().PeerInfo(peer)

		netAddr, err := netAddressFromPeer(addrInfo)
		if err == nil && netAddr.IP.Equal(ip) {
			return true
		}
	}

	return false
}

func (p *PeerSet) Get(key p2p.ID) p2p.Peer {
	id := peer.ID(key)

	// use cache
	if peer, ok := p.cacheGet(id); ok {
		return peer
	}

	addrInfo := p.host.Peerstore().PeerInfo(id)
	if len(addrInfo.Addrs) == 0 {
		p.logger.Error("Peer not found", "peer", id)
		return nil
	}

	peer, err := NewPeer(p.host, addrInfo)
	if err != nil {
		p.logger.Error("Failed to create peer", "peer", id, "err", err)
		return nil
	}

	p.cacheSet(peer)

	return peer
}

func (p *PeerSet) Remove(key p2p.ID) {
	id := peer.ID(key)

	if _, ok := p.cacheGet(id); !ok {
		// noop
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// delete *Peer
	delete(p.peers, id)

	// drop kv if any
	p.host.Peerstore().RemovePeer(id)
}

func (p *PeerSet) Size() int {
	return len(p.existingPeerIDs())
}

func (p *PeerSet) Copy() []p2p.Peer {
	peers := p.existingPeerIDs()
	results := make([]p2p.Peer, 0, len(peers))

	for _, id := range peers {
		key := p2p.ID(id)

		if peer := p.Get(key); peer != nil {
			results = append(results, peer)
		}
	}

	return results
}

func (p *PeerSet) ForEach(lambda func(p2p.Peer)) {
	peers := p.existingPeerIDs()

	for _, id := range peers {
		key := p2p.ID(id)

		peer := p.Get(key)
		if peer == nil {
			continue
		}

		lambda(peer)
	}
}

func (p *PeerSet) Random() p2p.Peer { return nil }

func (p *PeerSet) existingPeerIDs() []peer.ID {
	return p.host.Peerstore().PeersWithAddrs()
}

func (p *PeerSet) cacheGet(id peer.ID) (*Peer, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	peer, ok := p.peers[id]

	return peer, ok
}

func (p *PeerSet) cacheSet(peer *Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers[peer.addrInfo.ID] = peer
}
