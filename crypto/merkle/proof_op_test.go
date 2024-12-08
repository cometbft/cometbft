package merkle

import (
	"errors"
	"testing"

	cmtcrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/stretchr/testify/require"
)

type MockProofOperator struct {
	Key  []byte
	Root []byte
}

func (m *MockProofOperator) Run(leaves [][]byte) ([][]byte, error) {
	if len(leaves) == 0 {
		return nil, errors.New("leaves cannot be empty")
	}
	hash := append(leaves[0], m.Key...)
	return [][]byte{hash}, nil
}

func (m *MockProofOperator) GetKey() []byte {
	return m.Key
}

func (m *MockProofOperator) ProofOp() cmtcrypto.ProofOp {
	return cmtcrypto.ProofOp{
		Type: "mock",
		Data: m.Key,
	}
}

func TestVerifyValue(t *testing.T) {
	prt := merkle.NewProofRuntime()
	prt.RegisterOpDecoder("mock", func(op cmtcrypto.ProofOp) (merkle.ProofOperator, error) {
		return &MockProofOperator{
			Key: op.Data,
		}, nil
	})

	key := []byte("test-key")
	value := []byte("test-value")
	expectedHash := append(value, key...)

	root := expectedHash
	keypath := "/test-key"
	mockProofOp := &MockProofOperator{Key: key}

	proofOps := &cmtcrypto.ProofOps{
		Ops: []cmtcrypto.ProofOp{
			mockProofOp.ProofOp(),
		},
	}

	err := prt.VerifyValue(proofOps, root, keypath, value)
	require.NoError(t, err, "Verify Failed")
}

func TestProofOperators_Verify(t *testing.T) {
	op1 := &MockProofOperator{Key: []byte("key1")}
	op2 := &MockProofOperator{Key: []byte("key2")}

	proofOperators := merkle.ProofOperators{op1, op2}

	initialValue := []byte("initial-value")
	intermediateHash := append(initialValue, []byte("key1")...)
	expectedRoot := append(intermediateHash, []byte("key2")...)

	root := expectedRoot
	keypath := "/key2/key1"
	args := [][]byte{initialValue}

	err := proofOperators.Verify(root, keypath, args)

	require.NoError(t, err, "Verify Failed")
}
