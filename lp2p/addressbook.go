package lp2p

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type AddressBookConfig struct {
	Peers []PeerConfig `mapstructure:"peers" toml:"peers"`
}

type PeerConfig struct {
	// ip:port example: "192.0.2.0:65432"
	Host string `toml:"host"`
	// id example: "12D3KooWJx9i35Vx1h6T6nVqQz4YW1r2J1Y2P2nY3N4N5N6N7N8N9N0"
	ID string `toml:"id"`

	// TODO: port peer flavors
	// Private       bool   `mapstructure:"private"`
	// Persistent    bool   `mapstructure:"persistent"`
	// Unconditional bool   `mapstructure:"unconditional"`
}

func AddressBookFromFilePath(filepath string) (*AddressBookConfig, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read address book file %q: %w", filepath, err)
	}

	ab := &AddressBookConfig{}
	if err := toml.Unmarshal(raw, ab); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address book file %q: %w", filepath, err)
	}

	return ab, nil
}

func (ab *AddressBookConfig) Validate() error {
	_, err := ab.DecodePeers()
	if err != nil {
		return fmt.Errorf("failed to decode peers: %w", err)
	}

	return nil
}

func (ab *AddressBookConfig) DecodePeers() ([]peer.AddrInfo, error) {
	var (
		out   = make([]peer.AddrInfo, 0, len(ab.Peers))
		cache = make(map[string]struct{})
	)

	for _, pc := range ab.Peers {
		addr, err := pc.AddrInfo()
		if err != nil {
			return nil, err
		}

		// dedup by peer id
		if _, ok := cache[addr.ID.String()]; ok {
			continue
		}

		out = append(out, addr)
		cache[addr.ID.String()] = struct{}{}
	}

	return out, nil
}

func (pc *PeerConfig) AddrInfo() (peer.AddrInfo, error) {
	addr, err := AddressToMultiAddr(pc.Host, TransportQUIC)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to convert host to multiaddr: %w", err)
	}

	id, err := peer.Decode(pc.ID)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to decode id: %w", err)
	}

	return peer.AddrInfo{ID: id, Addrs: []ma.Multiaddr{addr}}, nil
}

func (ab *AddressBookConfig) Save(filepath string) error {
	if err := ab.Validate(); err != nil {
		return fmt.Errorf("failed to validate address book: %w", err)
	}

	raw, err := toml.Marshal(ab)
	if err != nil {
		return fmt.Errorf("failed to marshal address book: %w", err)
	}

	// rw-r--r--
	if err := os.WriteFile(filepath, raw, 0o644); err != nil {
		return fmt.Errorf("failed to save address book to file: %w", err)
	}

	return nil
}
