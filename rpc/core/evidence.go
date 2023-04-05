package core

import (
	"errors"
	"fmt"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/tendermint/tendermint/types"
)

// BroadcastEvidence broadcasts evidence of the misbehavior.
<<<<<<< HEAD
// More: https://docs.cometbft.com/v0.34/rpc/#/Info/broadcast_evidence
func BroadcastEvidence(ctx *rpctypes.Context, ev types.Evidence) (*ctypes.ResultBroadcastEvidence, error) {
=======
// More: https://docs.cometbft.com/main/rpc/#/Evidence/broadcast_evidence
func (env *Environment) BroadcastEvidence(
	_ *rpctypes.Context,
	ev types.Evidence,
) (*ctypes.ResultBroadcastEvidence, error) {
>>>>>>> 111d252d7 (Fix lints (#625))
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
