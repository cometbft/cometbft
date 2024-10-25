package mock

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
)

type Reactor struct {
	p2p.BaseReactor

	Channels []p2p.StreamDescriptor
}

func NewReactor() *Reactor {
	r := &Reactor{}
	r.BaseReactor = *p2p.NewBaseReactor("Mock-PEX", r)
	r.SetLogger(log.TestingLogger())
	return r
}

func (r *Reactor) StreamDescriptors() []p2p.StreamDescriptor { return r.Channels }
func (*Reactor) AddPeer(_ p2p.Peer)                          {}
func (*Reactor) RemovePeer(_ p2p.Peer, _ any)                {}
func (*Reactor) Receive(_ p2p.Envelope)                      {}
