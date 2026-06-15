package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// supportedPKCS11Algorithms mirrors the algo registry in
// internal/backend/pkcs11. It is duplicated here so config validation does not
// have to import the cgo-backed pkcs11 package. Keep the two in sync when adding
// a key type.
var supportedPKCS11Algorithms = map[string]bool{"ed25519": true}

// supportedAWSKMSAlgorithms mirrors the algo registry in
// internal/backend/awskms. It is duplicated here so config validation does not
// have to import the awskms package. Keep the two in sync when adding a key type.
var supportedAWSKMSAlgorithms = map[string]bool{"ed25519": true}

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
		if _, _, _, perr := v.ParsedTransport(); perr != nil {
			return fmt.Errorf("config: validator for chain %q has invalid addr: %w", v.ChainID, perr)
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
		if len(p.ChainIDs) == 0 {
			return fmt.Errorf("config: softsign provider with key_file %q has no chain_ids", p.KeyFile)
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

	for i := range c.Providers.PKCS11 {
		if err := c.validatePKCS11Provider(i, home, chainIDs, hasBackend); err != nil {
			return err
		}
	}

	for i := range c.Providers.AWSKMS {
		if err := c.validateAWSKMSProvider(i, chainIDs, hasBackend); err != nil {
			return err
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

// validatePKCS11Provider checks one [[providers.pkcs11]] block, resolves its
// relative paths against home, and records which chains it backs.
func (c *Config) validatePKCS11Provider(i int, home string, chainIDs map[string]int, hasBackend map[string]bool) error {
	p := c.Providers.PKCS11[i]

	if p.Module == "" {
		return fmt.Errorf("config: pkcs11 provider has empty module")
	}
	if len(p.ChainIDs) == 0 {
		return fmt.Errorf("config: pkcs11 provider with module %q has no chain_ids", p.Module)
	}

	// Token selector: exactly one of token_label / slot.
	switch {
	case p.TokenLabel != "" && p.Slot != nil:
		return fmt.Errorf("config: pkcs11 provider sets both token_label and slot (use exactly one)")
	case p.TokenLabel == "" && p.Slot == nil:
		return fmt.Errorf("config: pkcs11 provider must set token_label or slot")
	}

	// Key selector: at least one of key_label / key_id.
	if p.KeyLabel == "" && p.KeyID == "" {
		return fmt.Errorf("config: pkcs11 provider must set key_label or key_id")
	}
	if p.KeyID != "" {
		if _, err := hex.DecodeString(p.KeyID); err != nil {
			return fmt.Errorf("config: pkcs11 provider key_id %q is not valid hex: %w", p.KeyID, err)
		}
	}

	// PIN source: exactly one of pin / pin_env / pin_file.
	n := 0
	for _, set := range []bool{p.PIN != "", p.PINEnv != "", p.PINFile != ""} {
		if set {
			n++
		}
	}
	if n != 1 {
		return fmt.Errorf("config: pkcs11 provider must set exactly one of pin, pin_env, pin_file (got %d)", n)
	}

	// Algorithm: empty defaults to ed25519; otherwise must be registered.
	if p.Algorithm != "" && !supportedPKCS11Algorithms[p.Algorithm] {
		return fmt.Errorf("config: pkcs11 provider has unknown algorithm %q", p.Algorithm)
	}

	// Resolve relative paths against home before checking the module is readable.
	if !filepath.IsAbs(p.Module) {
		p.Module = filepath.Join(home, p.Module)
	}
	if p.PINFile != "" && !filepath.IsAbs(p.PINFile) {
		p.PINFile = filepath.Join(home, p.PINFile)
	}
	if _, err := os.Stat(p.Module); err != nil {
		return fmt.Errorf("config: pkcs11 provider module %q not readable: %w", p.Module, err)
	}
	c.Providers.PKCS11[i] = p

	for _, id := range p.ChainIDs {
		if _, ok := chainIDs[id]; !ok {
			return fmt.Errorf("config: pkcs11 provider references unknown chain %q", id)
		}
		hasBackend[id] = true
	}
	return nil
}

// validateAWSKMSProvider checks one [[providers.awskms]] block and records which
// chains it backs. There is no path resolution or local readability check:
// credentials and reachability are an AWS concern surfaced at Open (startup).
func (c *Config) validateAWSKMSProvider(i int, chainIDs map[string]int, hasBackend map[string]bool) error {
	p := c.Providers.AWSKMS[i]

	if p.KeyID == "" {
		return fmt.Errorf("config: awskms provider has empty key_id")
	}
	if len(p.ChainIDs) == 0 {
		return fmt.Errorf("config: awskms provider with key_id %q has no chain_ids", p.KeyID)
	}
	if p.Algorithm != "" && !supportedAWSKMSAlgorithms[p.Algorithm] {
		return fmt.Errorf("config: awskms provider has unknown algorithm %q", p.Algorithm)
	}

	for _, id := range p.ChainIDs {
		if _, ok := chainIDs[id]; !ok {
			return fmt.Errorf("config: awskms provider references unknown chain %q", id)
		}
		hasBackend[id] = true
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
