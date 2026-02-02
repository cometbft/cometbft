package lp2p

import (
	"sort"
	"sync"

	"github.com/cometbft/cometbft/libs/log"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
)

// PeerSet represents a single entrypoint for managing Peer's lifecycle
type PeerSet struct {
	host *Host

	peers map[peer.ID]*Peer
	mu    sync.RWMutex

	metrics *p2p.Metrics
	logger  log.Logger
}

var _ p2p.IPeerSet = (*PeerSet)(nil)

var (
	ErrPeerExists = errors.New("peer already exists")
	ErrSelfPeer   = errors.New("peer is self")
)

// NewPeerSet manager peers for a given switch
func NewPeerSet(host *Host, metrics *p2p.Metrics, logger log.Logger) *PeerSet {
	return &PeerSet{
		host:    host,
		peers:   make(map[peer.ID]*Peer),
		metrics: metrics,
		logger:  logger,
	}
}

func (ps *PeerSet) Has(key p2p.ID) bool {
	_, exists := ps.getByKey(key)

	return exists
}

func (ps *PeerSet) Get(key p2p.ID) p2p.Peer {
	peer, exists := ps.getByKey(key)
	if !exists {
		return nil
	}

	return peer
}

// PeerAddOptions options for adding a peer to the PeerSet.
// It includes behavioral flags and lifecycle callbacks for peer initialization.
type PeerAddOptions struct {
	Private       bool
	Persistent    bool
	Unconditional bool
	OnBeforeStart func(p *Peer)
	OnAfterStart  func(p *Peer)
	OnStartFailed func(p *Peer, reason any)
}

// Add adds a new peer to the peer set.
// Fails if peer is already present or self.
func (ps *PeerSet) Add(id peer.ID, opts PeerAddOptions) (*Peer, error) {
	if id == ps.host.ID() {
		return nil, ErrSelfPeer
	}

	ps.logger.Info("Adding peer", "peer_id", id.String())

	addrInfo := ps.host.Peerstore().PeerInfo(id)
	if len(addrInfo.Addrs) == 0 {
		return nil, errors.New("peer has no addresses in peerstore")
	}

	p, err := NewPeer(ps.host, addrInfo, ps.metrics, opts.Private, opts.Persistent, opts.Unconditional)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create peer")
	}

	set := ps.set(id, p)
	if !set {
		return nil, ErrPeerExists
	}

	if opts.OnBeforeStart != nil {
		opts.OnBeforeStart(p)
	}

	if err := p.Start(); err != nil {
		ps.unset(id)
		if opts.OnStartFailed != nil {
			opts.OnStartFailed(p, err)
		}
		return nil, errors.Wrap(err, "unable to start peer")
	}

	if opts.OnAfterStart != nil {
		opts.OnAfterStart(p)
	}

	ps.metrics.Peers.Add(1)

	return p, nil
}

// PeerRemovalOptions options for removing a peer from the PeerSet.
// If OnAfterStop is provided, it will be called after the peer is stopped.
// Note that Reason is any due to backwards compatibility.
type PeerRemovalOptions struct {
	Reason      any
	OnAfterStop func(p *Peer, reason any)
}

func (ps *PeerSet) Remove(key p2p.ID, opts PeerRemovalOptions) error {
	id := ps.keyToPeerID(key)
	if id == "" {
		return errors.New("invalid peer key")
	}

	if id == ps.host.ID() {
		return ErrSelfPeer
	}

	ps.logger.Info("Removing peer", "peer_id", id.String(), "reason", opts.Reason)

	p, ok := ps.unset(id)
	if !ok {
		return errors.New("peer not found")
	}

	if err := p.Stop(); err != nil {
		return errors.Wrap(err, "failed to stop peer")
	}

	if opts.OnAfterStop != nil {
		opts.OnAfterStop(p, opts.Reason)
	}

	if err := ps.host.Network().ClosePeer(id); err != nil {
		ps.logger.Error("Failed to close peer", "peer_id", id, "err", err)
		// tolerate this error.
	}

	ps.metrics.Peers.Add(-1)

	return nil
}

func (ps *PeerSet) RemoveAll(opts PeerRemovalOptions) {
	peers := ps.Copy()

	for _, peer := range peers {
		id := peer.ID()
		if err := ps.Remove(id, opts); err != nil {
			ps.logger.Error("Failed to remove peer", "peer_id", id, "err", err)
		}
	}
}

func (ps *PeerSet) ForEach(fn func(p2p.Peer)) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, p := range ps.peers {
		fn(p)
	}
}

// Random returns a random peer from the PeerSet (or nil if no peers are present).
// This method is not expected to be called frequently as it has O(n log n) complexity.
//
// Deprecated: use only for backwards compatibility.
func (ps *PeerSet) Random() p2p.Peer {
	peers := ps.Copy()
	if len(peers) == 0 {
		return nil
	}

	idx := cmtrand.Intn(len(peers))

	return peers[idx]
}

// Copy returns a copy of the peers list.
//
// Deprecated: use only for backwards compatibility.
func (ps *PeerSet) Copy() []p2p.Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	results := make([]p2p.Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		results = append(results, p)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID() < results[j].ID()
	})

	return results
}

// Size returns the number of peers in the peerSet.
func (ps *PeerSet) Size() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return len(ps.peers)
}

func (ps *PeerSet) getByKey(key p2p.ID) (*Peer, bool) {
	id := ps.keyToPeerID(key)
	if id == "" {
		return nil, false
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	p, ok := ps.peers[id]
	if !ok {
		return nil, false
	}

	return p, true
}

// set adds a peer to the peer set and returns true if it was added
func (ps *PeerSet) set(id peer.ID, p *Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, ok := ps.peers[id]; ok {
		return false
	}

	ps.peers[id] = p

	return true
}

// unset removes a peer from the peer set and returns it
// returns nil if peer is not found
func (ps *PeerSet) unset(id peer.ID) (*Peer, bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	p, ok := ps.peers[id]
	if !ok {
		return nil, false
	}

	delete(ps.peers, id)

	return p, true
}

func (ps *PeerSet) keyToPeerID(key p2p.ID) peer.ID {
	if key == "" {
		return ""
	}

	b, err := base58.Decode(string(key))
	if err != nil {
		ps.logger.Error("Failed to decode base58 key", "peer_key", key, "err", err)
		return ""
	}

	id, err := peer.IDFromBytes(b)
	if err != nil {
		ps.logger.Error("Failed to convert bytes to peer ID", "peer_key", key, "err", err)
		return ""
	}

	return id
}

// note that peerID.String() is base58 encoded and
// raw peerID is string([]byte)!
func peerIDToKey(id peer.ID) p2p.ID {
	return p2p.ID(id.String())
}
