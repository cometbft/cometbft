package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validate resolves defaults and enforces fail-fast invariants. home is the base
// directory used to resolve relative paths and default state files.
func (c *Config) Validate(home string) error {
	if len(c.Chains) == 0 {
		return fmt.Errorf("config: no [[chain]] declared")
	}

	chainIDs := make(map[string]int, len(c.Chains)) // id -> index
	for i, ch := range c.Chains {
		if ch.ID == "" {
			return fmt.Errorf("config: [[chain]] #%d has empty id", i)
		}
		if _, dup := chainIDs[ch.ID]; dup {
			return fmt.Errorf("config: duplicate chain id %q", ch.ID)
		}
		chainIDs[ch.ID] = i
	}

	// Resolve + ensure writable state files.
	for i := range c.Chains {
		sf := c.Chains[i].StateFile
		if sf == "" {
			sf = filepath.Join(home, "state", c.Chains[i].ID+".json")
		} else if !filepath.IsAbs(sf) {
			sf = filepath.Join(home, sf)
		}
		if err := os.MkdirAll(filepath.Dir(sf), 0o700); err != nil {
			return fmt.Errorf("config: state dir for chain %q: %w", c.Chains[i].ID, err)
		}
		if err := checkWritable(filepath.Dir(sf)); err != nil {
			return fmt.Errorf("config: state file for chain %q not writable: %w", c.Chains[i].ID, err)
		}
		c.Chains[i].StateFile = sf
	}

	// Every validator references a known chain.
	for i := range c.Validators {
		v := c.Validators[i]
		if _, ok := chainIDs[v.ChainID]; !ok {
			return fmt.Errorf("config: validator references unknown chain %q", v.ChainID)
		}
		if v.Addr == "" {
			return fmt.Errorf("config: validator for chain %q has empty addr", v.ChainID)
		}
		if v.IdentityKey == "" {
			return fmt.Errorf("config: validator for chain %q has empty identity_key", v.ChainID)
		}
		// Resolve relative identity_key against home so app.Build consumes the
		// resolved path (CWD-relative resolution would silently mint a new key).
		if !filepath.IsAbs(v.IdentityKey) {
			c.Validators[i].IdentityKey = filepath.Join(home, v.IdentityKey)
		}
	}

	// Every provider references a known chain; collect which chains have a backend.
	hasBackend := make(map[string]bool)
	for i := range c.Providers.Softsign {
		p := c.Providers.Softsign[i]
		if p.KeyFile == "" {
			return fmt.Errorf("config: softsign provider has empty key_file")
		}
		for _, id := range p.ChainIDs {
			if _, ok := chainIDs[id]; !ok {
				return fmt.Errorf("config: softsign provider references unknown chain %q", id)
			}
			hasBackend[id] = true
		}
		// Resolve relative key_file against home.
		if !filepath.IsAbs(p.KeyFile) {
			c.Providers.Softsign[i].KeyFile = filepath.Join(home, p.KeyFile)
		}
	}

	// Every chain must have at least one backend.
	for _, ch := range c.Chains {
		if !hasBackend[ch.ID] {
			return fmt.Errorf("config: chain %q has no backend", ch.ID)
		}
	}
	return nil
}

func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".writecheck-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}
