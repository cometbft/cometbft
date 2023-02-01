package core

import (
	"errors"
	"fmt"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/types"
)

// BroadcastEvidence broadcasts evidence of the misbehavior.
<<<<<<< HEAD
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/broadcast_evidence
func BroadcastEvidence(ctx *rpctypes.Context, ev types.Evidence) (*ctypes.ResultBroadcastEvidence, error) {
=======
// More: https://docs.cometbft.com/main/rpc/#/Evidence/broadcast_evidence
func (env *Environment) BroadcastEvidence(
	ctx *rpctypes.Context,
	ev types.Evidence) (*ctypes.ResultBroadcastEvidence, error) {

>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224))
	if ev == nil {
		return nil, errors.New("no evidence was provided")
	}

	if err := ev.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("evidence.ValidateBasic failed: %w", err)
	}

	if err := env.EvidencePool.AddEvidence(ev); err != nil {
		return nil, fmt.Errorf("failed to add evidence: %w", err)
	}
	return &ctypes.ResultBroadcastEvidence{Hash: ev.Hash()}, nil
}
