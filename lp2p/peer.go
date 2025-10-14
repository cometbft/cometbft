package lp2p

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
)

type Peer struct {
	service.BaseService

	host *Host

	addrInfo peer.AddrInfo
	netAddr  *p2p.NetAddress
	metrics  *p2p.Metrics
}

var _ p2p.Peer = (*Peer)(nil)

func NewPeer(host *Host, addrInfo peer.AddrInfo, metrics *p2p.Metrics) (*Peer, error) {
	netAddr, err := netAddressFromPeer(addrInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse net address: %w", err)
	}

	p := &Peer{
		host:     host,
		addrInfo: addrInfo,
		netAddr:  netAddr,
		metrics:  metrics,
	}

	p.BaseService = *service.NewBaseService(nil, "Peer", p)

	return p, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{%s}", p.ID())
}

func (p *Peer) ID() p2p.ID {
	return peerIDToKey(p.addrInfo.ID)
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
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "Send", "err", err)
		return false
	}

	return true
}

func (p *Peer) TrySend(e p2p.Envelope) bool {
	// todo same as SEND, but if current queue is full (its cap=1), immediately return FALSE
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "TrySend", "err", err)
		return false
	}

	return true
}

func (p *Peer) CloseConn() error {
	return p.host.Network().ClosePeer(p.addrInfo.ID)
}

func (p *Peer) send(e p2p.Envelope) (err error) {
	if !p.IsRunning() {
		return errors.New("peer is not running")
	}

	// todo: skip if not having the channel

	peerID := p.addrInfo.ID
	protocolID := ProtocolID(e.ChannelID)

	payload, err := marshalProto(e.Message)
	if err != nil {
		return err
	}

	var (
		peerIDStr    = peerID.String()
		messageType  = protoTypeName(e.Message)
		payloadLen   = float64(len(payload))
		metricLabels = []string{
			"peer_id", peerIDStr,
			"chID", fmt.Sprintf("%#x", e.ChannelID),
		}

		// note metric's name is misleading, it's a counter, not sum(bytes_pending)
		pendingMessagesCounter = p.metrics.PeerPendingSendBytes.With("peer_id", peerIDStr)
	)

	pendingMessagesCounter.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutStream)
	defer cancel()

	start := time.Now()

	defer func() {
		pendingMessagesCounter.Add(-1)

		if err != nil {
			return
		}

		p.metrics.PeerSendBytesTotal.With(metricLabels...).Add(payloadLen)
		p.metrics.MessageSendBytesTotal.With("message_type", messageType).Add(payloadLen)

		p.Logger.Debug(
			"sent envelope",
			"protocol", protocolID,
			"peer_id", peerIDStr,
			"stream_opened_duration", time.Since(start).String(),
		)
	}()

	// if no streams are available, it will block or return an error
	s, err := p.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream %s: %w", protocolID, err)
	}

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
type PeerSet struct {
	host *Host

	peers map[peer.ID]*Peer
	mu    sync.RWMutex

	metrics *p2p.Metrics
	logger  log.Logger
}

var _ p2p.IPeerSet = (*PeerSet)(nil)

func NewPeerSet(host *Host, metrics *p2p.Metrics, logger log.Logger) *PeerSet {
	const initialCapacity = 64

	return &PeerSet{
		host:    host,
		peers:   make(map[peer.ID]*Peer, initialCapacity),
		mu:      sync.RWMutex{},
		metrics: metrics,
		logger:  logger,
	}
}

func (p *PeerSet) Has(key p2p.ID) bool {
	id := p.keyToPeerID(key)
	if id == "" {
		return false
	}

	return len(p.host.Peerstore().Addrs(id)) > 0
}

func (p *PeerSet) GetByID(id peer.ID) p2p.Peer {
	peer, err := p.getOrAdd(id)
	if err != nil {
		p.logger.Error("PeerSet.Get failed", "peer_id", id.String(), "err", err)
		return nil
	}

	return peer
}

func (p *PeerSet) Get(key p2p.ID) p2p.Peer {
	id := p.keyToPeerID(key)
	if id == "" {
		return nil
	}

	return p.GetByID(id)
}

func (p *PeerSet) getOrAdd(id peer.ID) (*Peer, error) {
	// use cache
	if peer, ok := p.cacheGet(id); ok {
		return peer, nil
	}

	// we don't want to return self
	if p.host.ID() == id {
		return nil, nil
	}

	addrInfo := p.host.Peerstore().PeerInfo(id)
	if len(addrInfo.Addrs) == 0 {
		return nil, errors.New("peer has no addresses in peerstore")
	}

	peer, err := NewPeer(p.host, addrInfo, p.metrics)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create peer")
	}

	peer.SetLogger(p.logger.With("peer_id", id.String()))

	return p.cacheSet(peer), nil
}

func (p *PeerSet) Remove(key p2p.ID) {
	if id := p.keyToPeerID(key); id != "" {
		p.remove(id)
	}
}

func (p *PeerSet) remove(id peer.ID) {
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
		key := peerIDToKey(id)

		if peer := p.Get(key); peer != nil {
			results = append(results, peer)
		}
	}

	return results
}

func (p *PeerSet) ForEach(lambda func(p2p.Peer)) {
	peers := p.existingPeerIDs()

	for _, id := range peers {
		key := peerIDToKey(id)
		peer := p.Get(key)

		if peer == nil {
			continue
		}

		lambda(peer)
	}
}

func (p *PeerSet) Random() p2p.Peer { return nil }

func (p *PeerSet) existingPeerIDs() []peer.ID {
	hostID := p.host.ID()
	peers := p.host.Peerstore().PeersWithAddrs()

	// exclude self
	for i := 0; i < len(peers); i++ {
		if peers[i] == hostID {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}

	p.logger.Debug("Existing peer IDs", "host_id", hostID, "peers", peers)

	return peers
}

func (p *PeerSet) cacheGet(id peer.ID) (*Peer, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	peer, ok := p.peers[id]

	return peer, ok
}

func (p *PeerSet) cacheSet(peer *Peer) *Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	// noop
	if peer, ok := p.peers[peer.addrInfo.ID]; ok {
		return peer
	}

	p.peers[peer.addrInfo.ID] = peer

	return peer
}

func (p *PeerSet) keyToPeerID(key p2p.ID) peer.ID {
	if key == "" {
		// todo drop debug
		// todo this might happen because some reactors treat self as "" peer id
		stack := string(debug.Stack())
		p.logger.Debug("Attempt to get an empty peer id", "stack", stack)
		return ""
	}

	id, err := keyToPeerID(key)
	if err != nil {
		p.logger.Error("Failed to convert key to peer ID!", "peer_key", key, "err", err)
		return ""
	}

	return id
}

func keyToPeerID(key p2p.ID) (peer.ID, error) {
	b, err := base58.Decode(string(key))
	if err != nil {
		return "", errors.Wrap(err, "failed to decode base58 key")
	}

	id, err := peer.IDFromBytes(b)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert bytes to peer ID")
	}

	return id, nil
}

// note that peerID.String() is base58 encoded and
// raw peerID is string([]byte)!
func peerIDToKey(id peer.ID) p2p.ID {
	return p2p.ID(id.String())
}
