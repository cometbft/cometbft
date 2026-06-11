// Package app wires a validated config into a runnable manager.
package app

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cometbft/cometbft/kms/internal/backend"
	"github.com/cometbft/cometbft/kms/internal/backend/softsign"
	"github.com/cometbft/cometbft/kms/internal/config"
	"github.com/cometbft/cometbft/kms/internal/identity"
	"github.com/cometbft/cometbft/kms/internal/manager"
	"github.com/cometbft/cometbft/kms/internal/signer"
	"github.com/cometbft/cometbft/kms/internal/transport"
)

// Build constructs a Manager from a validated config.
func Build(c *config.Config, logger log.Logger) (*manager.Manager, error) {
	// chainID -> backend (one backend per chain).
	backends := map[string]backend.Signer{}
	for _, p := range c.Providers.Softsign {
		s, err := softsign.Load(p.KeyFile)
		if err != nil {
			return nil, err
		}
		for _, id := range p.ChainIDs {
			if _, dup := backends[id]; dup {
				return nil, fmt.Errorf("app: multiple backends bound to chain %q", id)
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
		cs, err := signer.NewChainSigner(id, be, stateFiles[id])
		if err != nil {
			return nil, err
		}
		signers[id] = cs
	}

	// One ValidatorConn per [[validator]]; validators of a chain share its signer.
	var conns []manager.ValidatorConn
	for _, v := range c.Validators {
		cs, ok := signers[v.ChainID]
		if !ok {
			return nil, fmt.Errorf("app: chain %q has no backend", v.ChainID)
		}
		idKey, err := identity.LoadOrGen(v.IdentityKey)
		if err != nil {
			return nil, err
		}
		tr, addr, validatorPeer, err := v.ParsedTransport()
		if err != nil {
			return nil, err
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
				return nil, derr
			}
			vc.Dialer = d
		}
		conns = append(conns, vc)
	}

	return manager.New(logger, conns), nil
}
