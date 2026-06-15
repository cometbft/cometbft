// Package app wires a validated config into a runnable manager.
package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cometbft/cometbft/kms/internal/backend"
	"github.com/cometbft/cometbft/kms/internal/backend/awskms"
	"github.com/cometbft/cometbft/kms/internal/backend/pkcs11"
	"github.com/cometbft/cometbft/kms/internal/backend/softsign"
	"github.com/cometbft/cometbft/kms/internal/config"
	"github.com/cometbft/cometbft/kms/internal/identity"
	"github.com/cometbft/cometbft/kms/internal/manager"
	"github.com/cometbft/cometbft/kms/internal/signer"
	"github.com/cometbft/cometbft/kms/internal/transport"
)

// Build constructs a Manager from a validated config. The returned cleanup
// function releases backend resources (e.g. PKCS#11 sessions) and must be called
// on shutdown. cleanup is non-nil even when an error is returned, so callers can
// always defer it.
func Build(c *config.Config, logger log.Logger) (mgr *manager.Manager, cleanup func(), err error) {
	// Backends that hold OS/HSM resources are closed by cleanup.
	var closers []io.Closer
	cleanup = func() {
		for _, cl := range closers {
			_ = cl.Close()
		}
	}
	// On error, release anything already opened before returning.
	defer func() {
		if err != nil {
			cleanup()
		}
	}()

	// chainID -> backend (one backend per chain).
	backends := map[string]backend.Signer{}
	for _, p := range c.Providers.Softsign {
		s, lerr := softsign.Load(p.KeyFile)
		if lerr != nil {
			return nil, cleanup, lerr
		}
		for _, id := range p.ChainIDs {
			if _, dup := backends[id]; dup {
				return nil, cleanup, fmt.Errorf("app: multiple backends bound to chain %q", id)
			}
			backends[id] = s
		}
	}

	for _, p := range c.Providers.PKCS11 {
		var keyID []byte
		if p.KeyID != "" {
			keyID, err = hex.DecodeString(p.KeyID)
			if err != nil {
				return nil, cleanup, fmt.Errorf("app: pkcs11 provider key_id %q: %w", p.KeyID, err)
			}
		}
		s, oerr := pkcs11.Open(pkcs11.Config{
			Module:     p.Module,
			TokenLabel: p.TokenLabel,
			Slot:       p.Slot,
			KeyLabel:   p.KeyLabel,
			KeyID:      keyID,
			PIN:        p.PIN,
			PINEnv:     p.PINEnv,
			PINFile:    p.PINFile,
			Algorithm:  p.Algorithm,
		})
		if oerr != nil {
			return nil, cleanup, oerr
		}
		closers = append(closers, s)
		for _, id := range p.ChainIDs {
			if _, dup := backends[id]; dup {
				return nil, cleanup, fmt.Errorf("app: multiple backends bound to chain %q", id)
			}
			backends[id] = s
		}
	}

	for _, p := range c.Providers.AWSKMS {
		s, oerr := awskms.Open(context.Background(), awskms.Config{
			KeyID:     p.KeyID,
			Region:    p.Region,
			Profile:   p.Profile,
			Endpoint:  p.Endpoint,
			Algorithm: p.Algorithm,
		})
		if oerr != nil {
			return nil, cleanup, oerr
		}
		// awskms holds no closable resource (unlike pkcs11), so it is not added
		// to closers.
		for _, id := range p.ChainIDs {
			if _, dup := backends[id]; dup {
				return nil, cleanup, fmt.Errorf("app: multiple backends bound to chain %q", id)
			}
			backends[id] = s
		}
	}

	// chainID -> state file.
	stateFiles := map[string]string{}
	for _, ch := range c.Chains {
		stateFiles[ch.ID] = ch.StateFile
	}

	// chainID -> *ChainSigner.
	signers := map[string]*signer.ChainSigner{}
	for id, be := range backends {
		cs, cerr := signer.NewChainSigner(id, be, stateFiles[id])
		if cerr != nil {
			return nil, cleanup, cerr
		}
		signers[id] = cs
	}

	// One ValidatorConn per [[validator]]; validators of a chain share its signer.
	var conns []manager.ValidatorConn
	for _, v := range c.Validators {
		cs, ok := signers[v.ChainID]
		if !ok {
			return nil, cleanup, fmt.Errorf("app: chain %q has no backend", v.ChainID)
		}
		idKey, lerr := identity.LoadOrGen(v.IdentityKey)
		if lerr != nil {
			return nil, cleanup, lerr
		}
		tr, addr, validatorPeer, perr := v.ParsedTransport()
		if perr != nil {
			return nil, cleanup, perr
		}
		vc := manager.ValidatorConn{
			ChainID:     v.ChainID,
			Addr:        v.Addr,
			IdentityKey: idKey,
			Signer:      cs,
			Reconnect:   v.ReconnectEnabled(),
		}
		if tr == config.TransportNoise {
			d, derr := transport.NoiseDialer(addr, idKey, validatorPeer, manager.DefaultDialTimeout)
			if derr != nil {
				return nil, cleanup, derr
			}
			vc.Dialer = d
		}
		conns = append(conns, vc)
	}

	return manager.New(logger, conns), cleanup, nil
}
