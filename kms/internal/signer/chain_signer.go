package signer

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"

	"github.com/cometbft/cometbft/kms/internal/backend"
)

// ChainSigner signs consensus messages for one chain, enforcing double-sign
// protection by delegating to a CometBFT *privval.FilePV (identical guard +
// crash-recovery logic). It is safe for concurrent use across multiple validator
// connections for the same chain.
type ChainSigner struct {
	chainID string
	mu      sync.Mutex
	fpv     *privval.FilePV
}

var _ types.PrivValidator = (*ChainSigner)(nil)

// NewChainSigner builds the signer. The backend is wrapped as a crypto.PrivKey
// and handed to privval.NewFilePV; any pre-existing sign-state at stateFile is
// reloaded so double-sign protection survives restarts. The directory containing
// stateFile must already exist (config validation guarantees this).
func NewChainSigner(chainID string, be backend.Signer, stateFile string) (*ChainSigner, error) {
	adapter, err := newBackendPrivKey(context.Background(), be)
	if err != nil {
		return nil, fmt.Errorf("chain %q: load pubkey: %w", chainID, err)
	}

	fpv := privval.NewFilePV(adapter, "", stateFile)

	if err := reloadState(fpv, stateFile); err != nil {
		return nil, fmt.Errorf("chain %q: reload sign-state: %w", chainID, err)
	}

	return &ChainSigner{chainID: chainID, fpv: fpv}, nil
}

// reloadState loads persisted FilePVLastSignState JSON into fpv.LastSignState,
// preserving the private filePath set by NewFilePV (JSON has no such field).
func reloadState(fpv *privval.FilePV, stateFile string) error {
	raw, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		return nil // fresh start; FilePV begins at height 0
	}
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return cmtjson.Unmarshal(raw, &fpv.LastSignState)
}

// GetPubKey implements types.PrivValidator.
func (c *ChainSigner) GetPubKey() (crypto.PubKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fpv.GetPubKey()
}

// SignVote implements types.PrivValidator.
func (c *ChainSigner) SignVote(chainID string, vote *cmtproto.Vote) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("chain %q: sign vote failed (state persistence): %v", c.chainID, r)
		}
	}()
	return c.fpv.SignVote(chainID, vote)
}

// SignProposal implements types.PrivValidator.
func (c *ChainSigner) SignProposal(chainID string, proposal *cmtproto.Proposal) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("chain %q: sign proposal failed (state persistence): %v", c.chainID, r)
		}
	}()
	return c.fpv.SignProposal(chainID, proposal)
}
