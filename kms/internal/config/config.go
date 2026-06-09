// Package config defines the cometkms TOML configuration and its validation.
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
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
}

// SoftsignProvider binds a softsign key file to one or more chains.
type SoftsignProvider struct {
	ChainIDs []string `toml:"chain_ids"`
	KeyFile  string   `toml:"key_file"`
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
