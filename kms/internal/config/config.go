// Package config defines the cometkms TOML configuration and its validation.
package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cometbft/cometbft/privval"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Config is the top-level cometkms configuration.
type Config struct {
	Chains     []Chain     `toml:"chain"`
	Validators []Validator `toml:"validator"`
	Providers  Providers   `toml:"providers"`
}

// Chain declares a chain and its double-sign state file.
type Chain struct {
	ID        string `toml:"id"`
	StateFile string `toml:"state_file"` // optional; defaulted by Validate
}

// Validator declares one outbound connection to a validator's listener.
type Validator struct {
	ChainID     string `toml:"chain_id"`
	Addr        string `toml:"addr"`         // tcp://host:port
	IdentityKey string `toml:"identity_key"` // ed25519 node-key file for SecretConnection
	Reconnect   *bool  `toml:"reconnect"`    // default true
}

// Providers groups the configured key backends.
type Providers struct {
	Softsign []SoftsignProvider `toml:"softsign"`
	PKCS11   []PKCS11Provider   `toml:"pkcs11"`
	AWSKMS   []AWSKMSProvider   `toml:"awskms"`
}

// SoftsignProvider binds a softsign key file to one or more chains.
type SoftsignProvider struct {
	ChainIDs []string `toml:"chain_ids"`
	KeyFile  string   `toml:"key_file"`
}

// PKCS11Provider binds a key on a PKCS#11 token/HSM to one or more chains.
// Exactly one of TokenLabel/Slot selects the token, at least one of
// KeyLabel/KeyID selects the key, and exactly one of PIN/PINEnv/PINFile supplies
// the user PIN. Algorithm defaults to "ed25519" when empty.
type PKCS11Provider struct {
	ChainIDs   []string `toml:"chain_ids"`
	Module     string   `toml:"module"`      // path to the PKCS#11 .so
	TokenLabel string   `toml:"token_label"` // CKA_LABEL of the token (XOR slot)
	Slot       *uint    `toml:"slot"`        // slot number (XOR token_label)
	KeyLabel   string   `toml:"key_label"`   // CKA_LABEL of the key
	KeyID      string   `toml:"key_id"`      // hex-encoded CKA_ID of the key
	PIN        string   `toml:"pin"`         // inline PIN
	PINEnv     string   `toml:"pin_env"`     // name of env var holding the PIN
	PINFile    string   `toml:"pin_file"`    // path to a file holding the PIN
	Algorithm  string   `toml:"algorithm"`   // key algorithm; defaults to "ed25519"
}

// AWSKMSProvider binds an Ed25519 key stored in AWS KMS to one or more chains.
// The private key never leaves KMS; signing is performed by the KMS Sign API.
// Credentials are resolved by the AWS default credential chain — no secret
// material is placed here. Algorithm defaults to "ed25519" when empty.
type AWSKMSProvider struct {
	ChainIDs  []string `toml:"chain_ids"`
	KeyID     string   `toml:"key_id"`    // KMS key id, ARN, or alias/<name>
	Region    string   `toml:"region"`    // optional; AWS default chain otherwise
	Profile   string   `toml:"profile"`   // optional shared-config profile
	Endpoint  string   `toml:"endpoint"`  // optional endpoint override (LocalStack/testing)
	Algorithm string   `toml:"algorithm"` // key algorithm; defaults to "ed25519"
}

// Load parses a TOML config file.
func Load(path string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}
	return &c, nil
}

// ReconnectEnabled reports the effective reconnect setting (default true).
func (v Validator) ReconnectEnabled() bool { return v.Reconnect == nil || *v.Reconnect }

// Transport identifies the privval connection transport selected by a validator
// address scheme.
type Transport int

const (
	// TransportTCP is tcp:// with cometbft SecretConnection (the default).
	TransportTCP Transport = iota
	// TransportNoise is noise://<peer-id>@host:port with libp2p Noise.
	TransportNoise
)

// ParsedTransport classifies v.Addr. For TCP it returns the full address
// unchanged (DialTCPFn consumes the tcp:// form) and an empty peer ID. For Noise
// it returns the host:port and the pinned validator peer ID.
func (v Validator) ParsedTransport() (tr Transport, addr string, validatorPeer peer.ID, err error) {
	if strings.HasPrefix(v.Addr, "noise://") {
		pid, hostport, perr := privval.ParseNoiseAddr(v.Addr)
		if perr != nil {
			return TransportNoise, "", "", perr
		}
		return TransportNoise, hostport, pid, nil
	}
	return TransportTCP, v.Addr, "", nil
}
