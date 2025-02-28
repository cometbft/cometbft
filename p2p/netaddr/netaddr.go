package netaddr

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	nk "github.com/cometbft/cometbft/p2p/internal/nodekey"
)

// IDAddrString returns id@hostPort. It strips the leading
// protocol from protocolHostPort if it exists.
func IDAddrString(id nodekey.ID, protocolHostPort string) string {
	hostPort := removeProtocolIfDefined(protocolHostPort)
	return fmt.Sprintf("%s@%s", id, hostPort)
}

// NetAddr defines information about a peer on the network including its ID and
// ma.Multiaddr.
type NetAddr struct {
	ID        nodekey.ID   `json:"id"`
	Multiaddr ma.Multiaddr `json:"multiaddr"`
}

// New returns a new NetAddr using the provided net.Addr.
//
// Panics if ID is invalid or the convertion from net.Addr to ma.Multiaddr fails.
func New(id nodekey.ID, addr net.Addr) NetAddr {
	multiaddr, err := manet.FromNetAddr(addr)
	if err != nil {
		panic(fmt.Sprintf("net.Addr -> multiaddr: %v", err))
	}

	if err := ValidateID(id); err != nil {
		panic(fmt.Sprintf("Invalid ID %v: %v (addr: %v)", id, err, addr))
	}

	return NetAddr{
		ID:        id,
		Multiaddr: multiaddr,
	}
}

// NewFromString returns a new address using the provided address in the form
// of either:
//
// - "/ip4/127.0.0.1/tcp/65537/p2p/QmR8cFFb5GVDRujmof9anGetQQjSZFKEivQuYBZXNd89X4" (new format)
// - "QmR8cFFb5GVDRujmof9anGetQQjSZFKEivQuYBZXNd89X4@127.0.0.1:65537" (old format)
//
// Also resolves the host if host is not an IP.
//
// Errors are of type ErrXxx where Xxx is in (NoID, Invalid, Lookup).
func NewFromString(addr string) (NetAddr, error) {
	// If it's a ma.Multiaddr, return early.
	multiaddr, err := ma.NewMultiaddr(addr)
	if err == nil {
		multiaddr2, p2pComponent := ma.SplitLast(multiaddr)
		return NetAddr{
			ID:        nodekey.ID(p2pComponent.Value()),
			Multiaddr: multiaddr2,
		}, nil
	}

	// Fall back to the old format.
	{
		// Remove protocol (e.g., "http://").
		addrWithoutProtocol := removeProtocolIfDefined(addr)
		spl := strings.Split(addrWithoutProtocol, "@")
		if len(spl) != 2 {
			return NetAddr{}, ErrInvalid{Err: ErrNoID{addr}}
		}

		// Validate ID.
		if err := ValidateID(nodekey.ID(spl[0])); err != nil {
			return NetAddr{}, ErrInvalid{Err: err}
		}

		var id nodekey.ID
		id, addrWithoutProtocol = nodekey.ID(spl[0]), spl[1]
		// get host and port
		host, portStr, err := net.SplitHostPort(addrWithoutProtocol)
		if err != nil {
			return NetAddr{}, ErrInvalid{Err: err}
		}
		if len(host) == 0 {
			return NetAddr{}, ErrInvalid{
				Err: ErrEmptyHost,
			}
		}

		ip := net.ParseIP(host)
		if ip == nil {
			ips, err := net.LookupIP(host)
			if err != nil {
				return NetAddr{}, ErrLookup{host, err}
			}
			ip = ips[0]
		}

		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return NetAddr{}, ErrInvalid{err}
		}

		na := TCPFromIPPort(ip, uint16(port))
		na.ID = id
		return na, nil
	}
}

// NewFromStrings returns an array of Addr'es build using
// the provided strings.
func NewFromStrings(addrs []string) ([]NetAddr, []error) {
	netAddrs := make([]NetAddr, 0)
	errs := make([]error, 0)
	for _, addr := range addrs {
		netAddr, err := NewFromString(addr)
		if err != nil {
			errs = append(errs, err)
		} else {
			netAddrs = append(netAddrs, netAddr)
		}
	}
	return netAddrs, errs
}

// TCPFromIPPort returns a new Addr using the provided IP and port number.
// Panics if it fails to convert either IP or port to ma.Multiaddr.
func TCPFromIPPort(ip net.IP, port uint16) NetAddr {
	multiaddr, err := manet.FromIP(ip)
	if err != nil {
		panic(fmt.Sprintf("ip -> multiaddr: %v", err))
	}
	multiaddr2, err := ma.NewMultiaddr(fmt.Sprintf("/tcp/%d", port))
	if err != nil {
		panic(fmt.Sprintf("port -> multiaddr: %v", err))
	}
	return NetAddr{
		Multiaddr: multiaddr.Encapsulate(multiaddr2),
	}
}

// NewFromProto converts a Protobuf NetAddress into a native struct.
func NewFromProto(pb tmp2p.NetAddress) (NetAddr, error) {
	id := nodekey.ID(pb.ID)
	if err := ValidateID(id); err != nil {
		return NetAddr{}, ErrInvalid{Err: err}
	}

	if len(pb.Multiaddr) > 0 { // Multiaddr detected.
		multiaddr, err := ma.NewMultiaddrBytes(pb.Multiaddr)
		if err != nil {
			return NetAddr{}, ErrInvalid{Err: err}
		}
		return NetAddr{
			ID:        id,
			Multiaddr: multiaddr,
		}, nil
	} else { // Fallback to the old format.
		ip := net.ParseIP(pb.IP)
		if ip == nil {
			return NetAddr{}, ErrInvalid{Err: ErrInvalidIP}
		}
		if pb.Port >= 1<<16 {
			return NetAddr{}, ErrInvalid{Err: ErrInvalidPort{pb.Port}}
		}
		// CONTRACT: ip and port must be valid.
		netAddr := TCPFromIPPort(ip, uint16(pb.Port))
		netAddr.ID = id
		return netAddr, nil
	}
}

// AddrsFromProtos converts a slice of Protobuf NetAddresses into a native slice.
func AddrsFromProtos(pbs []tmp2p.NetAddress) ([]NetAddr, error) {
	nas := make([]NetAddr, len(pbs))
	for _, pb := range pbs {
		na, err := NewFromProto(pb)
		if err != nil {
			return nil, err
		}
		nas = append(nas, na)
	}
	return nas, nil
}

// AddrsToProtos converts a slice of addresses into a Protobuf slice.
func AddrsToProtos(nas []NetAddr) []tmp2p.NetAddress {
	pbs := make([]tmp2p.NetAddress, len(nas))
	for _, na := range nas {
		pbs = append(pbs, na.ToProto())
	}
	return pbs
}

// IsEmpty returns true if the address is empty.
func (na NetAddr) IsEmpty() bool {
	return na.ID == "" && na.Multiaddr == nil
}

// ToProto converts an Addr to Protobuf.
func (na NetAddr) ToProto() tmp2p.NetAddress {
	return tmp2p.NetAddress{
		ID:        string(na.ID),
		Multiaddr: na.Multiaddr.Bytes(),
	}
}

// Equals reports whether na and other are the same addresses,
// including their ID, IP, and Port.
func (na NetAddr) Equals(other any) bool {
	if o, ok := other.(NetAddr); ok {
		return na.String() == o.String()
	}
	return false
}

// Same returns true is na has the same non-empty ID or DialString as other.
func (na NetAddr) Same(other any) bool {
	if o, ok := other.(NetAddr); ok {
		if na.DialString() == o.DialString() {
			return true
		}
		if na.ID != "" && na.ID == o.ID {
			return true
		}
	}
	return false
}

// String returns the string representation of the address.
// Example:
//
//	/ip4/192.168.1.0/tcp/26656/p2p/deadbeefdeadbeefdeadbeefdeadbeefdeadbeef
func (na NetAddr) String() string {
	p2pAddr, err := ma.NewMultiaddr("/p2p/" + string(na.ID))
	if err != nil {
		panic(err)
	}
	return na.Multiaddr.Encapsulate(p2pAddr).String()
}

// DialString returns a net.Addr String.
// Example:
//
//	192.168.1.0:26656
func (na NetAddr) DialString() string {
	netAddr, err := manet.ToNetAddr(na.Multiaddr)
	if err != nil {
		return ""
	}
	return netAddr.String()
}

// Dial calls net.Dial on the address.
func (na NetAddr) Dial() (net.Conn, error) {
	return manet.Dial(na.Multiaddr)
}

// DialTimeout calls net.DialTimeout on the address.
func (na NetAddr) DialTimeout(timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	d := &manet.Dialer{Dialer: net.Dialer{}}
	return d.DialContext(ctx, na.Multiaddr)
}

// Routable returns true if the address is routable.
//
// See manet.IsPublicAddr.
func (na NetAddr) Routable() bool {
	if err := na.Valid(); err != nil {
		return false
	}
	return manet.IsPublicAddr(na.Multiaddr)
}

// Valid checks ID and multiaddr.
//
// XXX: check if it's accurate.
// For IPv4 these are either a 0 or all bits set address. For IPv6 a zero
// address or one that matches the RFC3849 documentation address format.
//
// See IP.IsGlobalUnicast.
func (na NetAddr) Valid() error {
	if err := ValidateID(na.ID); err != nil {
		return ErrInvalidPeerID{na.ID, err}
	}

	if ip, err := manet.ToIP(na.Multiaddr); err == nil {
		if ip.IsGlobalUnicast() {
			return ErrInvalid{ErrInvalidIP}
		}
	}

	return nil
}

// Local returns true if it is a local address.
func (na NetAddr) Local() bool {
	return manet.IsIPLoopback(na.Multiaddr) || manet.IsIPUnspecified(na.Multiaddr)
}

// ToIP returns the IP address of the address when possible.
//
// See manet.ToIP.
func (na NetAddr) ToIP() (net.IP, error) {
	return manet.ToIP(na.Multiaddr)
}

// ToStdlibAddr converts a Multiaddr to a net.Addr Must be ThinWaist. acceptable protocol stacks are: /ip{4,6}/{tcp, udp}
//
// See manet.ToNetAddr.
func (na NetAddr) ToStdlibAddr() (net.Addr, error) {
	return manet.ToNetAddr(na.Multiaddr)
}

// ValidateID checks if the ID is valid.
// TODO: move to nodekey package
func ValidateID(id nodekey.ID) error {
	if len(id) == 0 {
		// invalid error
		return ErrNoID{""}
	}

	_, err := nk.DecodeID(string(id))
	return err
}

// Used for testing.
func CreateRoutableAddr() (addr string, netAddr NetAddr) {
	nodeKey := nodekey.NodeKey{
		PrivKey: ed25519.GenPrivKey(),
	}
	id := nodeKey.ID()
	for {
		var err error

		addr = fmt.Sprintf("%s@%v.%v.%v.%v:26656",
			id,
			cmtrand.Int()%256,
			cmtrand.Int()%256,
			cmtrand.Int()%256,
			cmtrand.Int()%256)
		netAddr, err = NewFromString(addr)
		if err != nil {
			panic(err)
		}
		if netAddr.Routable() {
			break
		}
	}
	return addr, netAddr
}

func removeProtocolIfDefined(addr string) string {
	if strings.Contains(addr, "://") {
		return strings.Split(addr, "://")[1]
	}
	return addr
}
