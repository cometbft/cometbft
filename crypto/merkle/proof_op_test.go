package merkle

import (
	"testing"
	"errors"

	"github.com/stretchr/testify/require"
	cmtcrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto/merkle"
)

type MockProofOperator struct {
	Key  []byte
	Root []byte // 添加 Root 字段以控制返回的哈希值
}

// func (m *MockProofOperator) Run(leaves [][]byte) ([][]byte, error) {
// 	// 返回预期的根哈希值
// 	return [][]byte{m.Root}, nil
// }
func (m *MockProofOperator) Run(leaves [][]byte) ([][]byte, error) {
	if len(leaves) == 0 {
		return nil, errors.New("leaves cannot be empty")
	}
	// 拼接叶子节点和 Key，模拟生成中间哈希值
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
	// 创建 ProofRuntime
	prt := merkle.NewProofRuntime()
	prt.RegisterOpDecoder("mock", func(op cmtcrypto.ProofOp) (merkle.ProofOperator, error) {
		return &MockProofOperator{
			Key: op.Data,
		}, nil
	})

	// 构造测试数据
	key := []byte("test-key")
	value := []byte("test-value")
	expectedHash := append(value, key...) // 计算期望的根哈希值

	root := expectedHash         // 设置预期的根哈希
	keypath := "/test-key"       // 确保以斜杠开头
	mockProofOp := &MockProofOperator{Key: key}

	// 构造 ProofOps
	proofOps := &cmtcrypto.ProofOps{
		Ops: []cmtcrypto.ProofOp{
			mockProofOp.ProofOp(),
		},
	}

	// 执行验证
	err := prt.VerifyValue(proofOps, root, keypath, value)
	require.NoError(t, err, "验证失败")
}


func TestProofOperators_Verify(t *testing.T) {
	// 构造多个 MockProofOperator
	op1 := &MockProofOperator{Key: []byte("key1")}
	op2 := &MockProofOperator{Key: []byte("key2")}

	proofOperators := merkle.ProofOperators{op1, op2}

	// 构造测试数据
	initialValue := []byte("initial-value")
	intermediateHash := append(initialValue, []byte("key1")...) // op1 生成的中间值
	expectedRoot := append(intermediateHash, []byte("key2")...) // op2 生成的最终值

	root := expectedRoot
	keypath := "/key2/key1" // 确保以斜杠开头
	args := [][]byte{initialValue}

	// 调用 Verify
	err := proofOperators.Verify(root, keypath, args)

	// 检查结果
	require.NoError(t, err, "验证失败")
}
