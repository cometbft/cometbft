package lp2p

import (
	"fmt"

	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type reactorSet struct {
	reactors []ReactorItem

	// [reactor_name => reactor] mapping
	reactorsByName map[string]p2p.Reactor

	// [protocol_id => reactor] mapping
	reactorsByProtocolID map[protocol.ID]p2p.Reactor

	// [protocol_id => reactor_name] mapping
	reactorNamesByProtocolID map[protocol.ID]string

	// [protocol_id => channel_descriptor] mapping
	descriptorByProtocolID map[protocol.ID]*p2p.ChannelDescriptor
}

// ReactorItem is a pair of name and reactor.
// Preserves order when adding.
type ReactorItem struct {
	Name    string
	Reactor p2p.Reactor
}

type reactorWithDescriptor struct {
	p2p.Reactor
	Name       string
	Descriptor *p2p.ChannelDescriptor
}

func newReactorSet() *reactorSet {
	return &reactorSet{
		reactors:                 []ReactorItem{},
		reactorsByName:           make(map[string]p2p.Reactor),
		reactorsByProtocolID:     make(map[protocol.ID]p2p.Reactor),
		reactorNamesByProtocolID: make(map[protocol.ID]string),
		descriptorByProtocolID:   make(map[protocol.ID]*p2p.ChannelDescriptor),
	}
}

func (rs *reactorSet) Add(item ReactorItem, switcher p2p.Switcher) error {
	for i := range item.Reactor.GetChannels() {
		var (
			channelDescriptor = item.Reactor.GetChannels()[i]
			protocolID        = ProtocolID(channelDescriptor.ID)
		)

		if _, ok := rs.reactorsByProtocolID[protocolID]; ok {
			return fmt.Errorf("protocol %q is already registered", protocolID)
		}

		rs.reactorsByProtocolID[protocolID] = item.Reactor
		rs.descriptorByProtocolID[protocolID] = channelDescriptor
		rs.reactorNamesByProtocolID[protocolID] = item.Name
	}

	rs.reactors = append(rs.reactors, item)
	rs.reactorsByName[item.Name] = item.Reactor

	return nil
}

func (rs *reactorSet) Start(switcher p2p.Switcher, perProtocolCallback func(protocol.ID)) error {
	for _, el := range rs.reactors {
		name, reactor := el.Name, el.Reactor

		switcher.Log().Info("Starting reactor", "reactor", name)

		reactor.SetSwitch(switcher)

		if err := reactor.Start(); err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	for protocolID := range rs.reactorsByProtocolID {
		perProtocolCallback(protocolID)
	}

	return nil
}

func (rs *reactorSet) Stop(switcher p2p.Switcher) {
	for name, reactor := range rs.reactorsByName {
		if err := reactor.Stop(); err != nil {
			switcher.Log().Error("failed to stop reactor", "name", name, "err", err)
		}
	}
}

func (rs *reactorSet) InitPeer(peer *Peer) {
	for _, el := range rs.reactors {
		el.Reactor.InitPeer(peer)
	}
}

func (rs *reactorSet) AddPeer(peer *Peer) {
	for _, el := range rs.reactors {
		el.Reactor.AddPeer(peer)
	}
}

func (rs *reactorSet) RemovePeer(peer *Peer, reason any) {
	for _, reactor := range rs.reactorsByName {
		reactor.RemovePeer(peer, reason)
	}
}

func (rs *reactorSet) GetByName(name string) (p2p.Reactor, bool) {
	reactor, exists := rs.reactorsByName[name]
	return reactor, exists
}

func (rs *reactorSet) GetWithDescriptorByProtocolID(id protocol.ID) (reactorWithDescriptor, error) {
	reactor, ok := rs.reactorsByProtocolID[id]
	if !ok {
		return reactorWithDescriptor{}, fmt.Errorf("reactor not found")
	}

	descriptor, ok := rs.descriptorByProtocolID[id]
	if !ok {
		return reactorWithDescriptor{}, fmt.Errorf("descriptor not found")
	}

	return reactorWithDescriptor{
		Name:       rs.reactorNamesByProtocolID[id],
		Reactor:    reactor,
		Descriptor: descriptor,
	}, nil
}
