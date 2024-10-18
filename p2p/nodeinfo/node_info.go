package nodeinfo

import (
	"bytes"
	"fmt"
	"reflect"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	cmtstrings "github.com/cometbft/cometbft/internal/strings"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
)

const (
	maxNodeInfoSize = 10240 // 10KB
	maxNumChannels  = 16    // plenty of room for upgrades, for now
)

// Max size of the NodeInfo struct.
func MaxNodeInfoSize() int {
	return maxNodeInfoSize
}

// -------------------------------------------------------------

// NodeInfo exposes basic info of a node
// and determines if we're compatible.
type NodeInfo interface {
	ID() nodekey.ID
	NetAddr() (*na.Addr, error)
	Validate() error
	CompatibleWith(other NodeInfo) error
}

// -------------------------------------------------------------

// ProtocolVersion contains the protocol versions for the software.
type ProtocolVersion struct {
	P2P   uint64 `json:"p2p"`
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

// NewProtocolVersion returns a fully populated ProtocolVersion.
func NewProtocolVersion(p2p, block, app uint64) ProtocolVersion {
	return ProtocolVersion{
		P2P:   p2p,
		Block: block,
		App:   app,
	}
}

// -------------------------------------------------------------

// Assert DefaultNodeInfo satisfies NodeInfo.
var _ NodeInfo = DefaultNodeInfo{}

// DefaultNodeInfo is the basic node information exchanged
// between two peers during the CometBFT P2P handshake.
type DefaultNodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`

	// Authenticate
	// TODO: replace with na.Addr
	DefaultNodeID nodekey.ID `json:"id"`          // authenticated identifier
	ListenAddr    string     `json:"listen_addr"` // accepting incoming

	// Check compatibility.
	// Channels are HexBytes so easier to read as JSON
	Network  string            `json:"network"`  // network/chain ID
	Version  string            `json:"version"`  // major.minor.revision
	Channels cmtbytes.HexBytes `json:"channels"` // channels this node knows about

	// ASCIIText fields
	Moniker string               `json:"moniker"` // arbitrary moniker
	Other   DefaultNodeInfoOther `json:"other"`   // other application specific data
}

// DefaultNodeInfoOther is the misc. application specific data.
type DefaultNodeInfoOther struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}

// ID returns the node's peer ID.
func (info DefaultNodeInfo) ID() nodekey.ID {
	return info.DefaultNodeID
}

// Validate checks the self-reported DefaultNodeInfo is safe.
// It returns an error if there
// are too many Channels, if there are any duplicate Channels,
// if the ListenAddr is malformed, or if the ListenAddr is a host name
// that can not be resolved to some IP.
// TODO: constraints for Moniker/Other? Or is that for the UI ?
// JAE: It needs to be done on the client, but to prevent ambiguous
// unicode characters, maybe it's worth sanitizing it here.
// In the future we might want to validate these, once we have a
// name-resolution system up.
// International clients could then use punycode (or we could use
// url-encoding), and we just need to be careful with how we handle that in our
// clients. (e.g. off by default).
func (info DefaultNodeInfo) Validate() error {
	// ID is already validated.

	// Validate ListenAddr.
	_, err := na.NewFromString(na.IDAddrString(info.ID(), info.ListenAddr))
	if err != nil {
		return err
	}

	// Network is validated in CompatibleWith.

	// Validate Version
	if len(info.Version) > 0 &&
		(!cmtstrings.IsASCIIText(info.Version) || cmtstrings.ASCIITrim(info.Version) == "") {
		return ErrInvalidNodeVersion{Version: info.Version}
	}

	// Validate Channels - ensure max and check for duplicates.
	if len(info.Channels) > maxNumChannels {
		return ErrChannelsTooLong{Length: len(info.Channels), Max: maxNumChannels}
	}

	channels := make(map[byte]struct{})
	for _, ch := range info.Channels {
		_, ok := channels[ch]
		if ok {
			return ErrDuplicateChannelID{ID: ch}
		}
		channels[ch] = struct{}{}
	}

	// Validate Moniker.
	if !cmtstrings.IsASCIIText(info.Moniker) || cmtstrings.ASCIITrim(info.Moniker) == "" {
		return ErrInvalidMoniker{Moniker: info.Moniker}
	}

	// Validate Other.
	other := info.Other
	txIndex := other.TxIndex
	switch txIndex {
	case "", "on", "off":
	default:
		return ErrInvalidTxIndex{TxIndex: txIndex}
	}
	// XXX: Should we be more strict about address formats?
	rpcAddr := other.RPCAddress
	if len(rpcAddr) > 0 && (!cmtstrings.IsASCIIText(rpcAddr) || cmtstrings.ASCIITrim(rpcAddr) == "") {
		return ErrInvalidRPCAddress{RPCAddress: rpcAddr}
	}

	return nil
}

// CompatibleWith checks if two DefaultNodeInfo are compatible with each other.
// CONTRACT: two nodes are compatible if the Block version and network match
// and they have at least one channel in common.
func (info DefaultNodeInfo) CompatibleWith(otherInfo NodeInfo) error {
	other, ok := otherInfo.(DefaultNodeInfo)
	if !ok {
		return ErrInvalidNodeInfoType{
			Type:     reflect.TypeOf(otherInfo).String(),
			Expected: fmt.Sprintf("%T", DefaultNodeInfo{}),
		}
	}

	if info.ProtocolVersion.Block != other.ProtocolVersion.Block {
		return ErrDifferentBlockVersion{
			Other: other.ProtocolVersion.Block,
			Our:   info.ProtocolVersion.Block,
		}
	}

	// nodes must be on the same network
	if info.Network != other.Network {
		return ErrDifferentNetwork{
			Other: other.Network,
			Our:   info.Network,
		}
	}

	// if we have no channels, we're just testing
	if len(info.Channels) == 0 {
		return nil
	}

	// for each of our channels, check if they have it
	found := false
OUTER_LOOP:
	for _, ch1 := range info.Channels {
		for _, ch2 := range other.Channels {
			if ch1 == ch2 {
				found = true
				break OUTER_LOOP // only need one
			}
		}
	}
	if !found {
		return ErrNoCommonChannels{
			OtherChannels: other.Channels,
			OurChannels:   info.Channels,
		}
	}
	return nil
}

// NetAddr returns a NetAddr derived from the DefaultNodeInfo -
// it includes the authenticated peer ID and the self-reported
// ListenAddr. Note that the ListenAddr is not authenticated and
// may not match that address actually dialed if its an outbound peer.
func (info DefaultNodeInfo) NetAddr() (*na.Addr, error) {
	idAddr := na.IDAddrString(info.ID(), info.ListenAddr)
	return na.NewFromString(idAddr)
}

func (info DefaultNodeInfo) HasChannel(chID byte) bool {
	return bytes.Contains(info.Channels, []byte{chID})
}

func (info DefaultNodeInfo) ToProto() *tmp2p.DefaultNodeInfo {
	dni := new(tmp2p.DefaultNodeInfo)
	dni.ProtocolVersion = tmp2p.ProtocolVersion{
		P2P:   info.ProtocolVersion.P2P,
		Block: info.ProtocolVersion.Block,
		App:   info.ProtocolVersion.App,
	}

	dni.DefaultNodeID = string(info.DefaultNodeID)
	dni.ListenAddr = info.ListenAddr
	dni.Network = info.Network
	dni.Version = info.Version
	dni.Channels = info.Channels
	dni.Moniker = info.Moniker
	dni.Other = tmp2p.DefaultNodeInfoOther{
		TxIndex:    info.Other.TxIndex,
		RPCAddress: info.Other.RPCAddress,
	}

	return dni
}

func DefaultNodeInfoFromToProto(pb *tmp2p.DefaultNodeInfo) (DefaultNodeInfo, error) {
	if pb == nil {
		return DefaultNodeInfo{}, ErrNoNodeInfo
	}

	dni := DefaultNodeInfo{
		ProtocolVersion: ProtocolVersion{
			P2P:   pb.ProtocolVersion.P2P,
			Block: pb.ProtocolVersion.Block,
			App:   pb.ProtocolVersion.App,
		},
		DefaultNodeID: nodekey.ID(pb.DefaultNodeID),
		ListenAddr:    pb.ListenAddr,
		Network:       pb.Network,
		Version:       pb.Version,
		Channels:      pb.Channels,
		Moniker:       pb.Moniker,
		Other: DefaultNodeInfoOther{
			TxIndex:    pb.Other.TxIndex,
			RPCAddress: pb.Other.RPCAddress,
		},
	}

	return dni, nil
}
