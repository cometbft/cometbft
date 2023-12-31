package types

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto/merkle"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// RowProof is a Merkle proof that a set of rows exist in a Merkle tree with a
// given data root.
type RowProof struct {
	// RowRoots are the roots of the rows being proven.
	RowRoots []tmbytes.HexBytes `json:"row_roots"`
	// Proofs is a list of Merkle proofs where each proof proves that a row
	// exists in a Merkle tree with a given data root.
	Proofs   []*merkle.Proof `json:"proofs"`
	StartRow uint32          `json:"start_row"`
	EndRow   uint32          `json:"end_row"`
}

// Validate performs checks on the fields of this RowProof. Returns an error if
// the proof fails validation. If the proof passes validation, this function
// attempts to verify the proof. It returns nil if the proof is valid.
func (rp RowProof) Validate(root []byte) error {
	// HACKHACK performing subtraction with unsigned integers is unsafe.
	if int(rp.EndRow-rp.StartRow+1) != len(rp.RowRoots) {
		return fmt.Errorf("the number of rows %d must equal the number of row roots %d", int(rp.EndRow-rp.StartRow+1), len(rp.RowRoots))
	}
	if len(rp.Proofs) != len(rp.RowRoots) {
		return fmt.Errorf("the number of proofs %d must equal the number of row roots %d", len(rp.Proofs), len(rp.RowRoots))
	}
	if !rp.VerifyProof(root) {
		return errors.New("row proof failed to verify")
	}

	return nil
}

// VerifyProof verifies that all the row roots in this RowProof exist in a
// Merkle tree with the given root. Returns true if all proofs are valid.
func (rp RowProof) VerifyProof(root []byte) bool {
	for i, proof := range rp.Proofs {
		err := proof.Verify(root, rp.RowRoots[i])
		if err != nil {
			return false
		}
	}
	return true
}

func RowProofFromProto(p *tmproto.RowProof) RowProof {
	if p == nil {
		return RowProof{}
	}
	rowRoots := make([]tmbytes.HexBytes, len(p.RowRoots))
	rowProofs := make([]*merkle.Proof, len(p.Proofs))
	for i := range p.Proofs {
		rowRoots[i] = p.RowRoots[i]
		rowProofs[i] = &merkle.Proof{
			Total:    p.Proofs[i].Total,
			Index:    p.Proofs[i].Index,
			LeafHash: p.Proofs[i].LeafHash,
			Aunts:    p.Proofs[i].Aunts,
		}
	}

	return RowProof{
		RowRoots: rowRoots,
		Proofs:   rowProofs,
		StartRow: p.StartRow,
		EndRow:   p.EndRow,
	}
}
