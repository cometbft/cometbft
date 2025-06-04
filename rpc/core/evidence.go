package core

import (
	"reflect"

	ctypes "github.com/cometbft/cometbft/v2/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/v2/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/v2/types"
)

// BroadcastEvidence broadcasts evidence of the misbehavior.
// More: https://docs.cometbft.com/main/rpc/#/Evidence/broadcast_evidence
func (env *Environment) BroadcastEvidence(
	_ *rpctypes.Context,
	ev types.Evidence,
) (*ctypes.ResultBroadcastEvidence, error) {
	if ev == nil {
		return nil, ErrNoEvidence
	}

	if err := ev.ValidateBasic(); err != nil {
		return nil, ErrValidation{
			Source:  err,
			ValType: reflect.TypeOf(ev).String(),
		}
	}

	if err := env.EvidencePool.AddEvidence(ev); err != nil {
		return nil, ErrAddEvidence{err}
	}

	return &ctypes.ResultBroadcastEvidence{Hash: ev.Hash()}, nil
}
