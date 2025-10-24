package types

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtrand "github.com/cometbft/cometbft/libs/rand"
)

const (
	testECPartSize = 65536 // 64KB ...  4096 // 4KB
)

func TestBasicECPartSet(t *testing.T) {
	// Construct random data of size partSize * 100
	nParts := 100
	expectedParts := 150 // 50% parity
	data := cmtrand.Bytes(testECPartSize * nParts)
	partSet, err := NewECPartSetFromData(data, testECPartSize)
	assert.NoError(t, err)

	assert.NotEmpty(t, partSet.Hash())
	assert.EqualValues(t, expectedParts, partSet.Total())
	assert.Equal(t, expectedParts, partSet.BitArray().Size())
	assert.True(t, partSet.HashesTo(partSet.Hash()))
	assert.True(t, partSet.IsComplete())
	assert.EqualValues(t, expectedParts, partSet.Count())
	assert.EqualValues(t, testECPartSize*expectedParts, partSet.ByteSize())

	// Test adding parts to a new partSet.
	partSet2 := NewECPartSetFromHeader(partSet.Header())

	assert.True(t, partSet2.HasHeader(partSet.Header()))
	for i := 0; i < int(partSet.Total()); i++ {
		part := partSet.GetPart(i)
		// t.Logf("\n%v", part)
		added, err := partSet2.AddPart(part)
		if !added || err != nil {
			t.Errorf("failed to add part %v, error: %v", i, err)
		}
	}
	// adding part with invalid index
	added, err := partSet2.AddPart(&Part{Index: 10000})
	assert.False(t, added)
	assert.Error(t, err)
	// adding existing part
	added, err = partSet2.AddPart(partSet2.GetPart(0))
	assert.False(t, added)
	assert.Nil(t, err)

	assert.Equal(t, partSet.Hash(), partSet2.Hash())
	assert.EqualValues(t, expectedParts, partSet2.Total())
	assert.EqualValues(t, expectedParts*testECPartSize, partSet.ByteSize())
	assert.True(t, partSet2.IsComplete())

	// Reconstruct data, assert that they are equal.
	data2Reader := partSet2.GetReader()
	data2, err := io.ReadAll(data2Reader)
	require.NoError(t, err)

	assert.Equal(t, data, data2)
}

func TestECWrongProof(t *testing.T) {
	// Construct random data of size partSize * 100
	data := cmtrand.Bytes(testECPartSize * 100)
	partSet, err := NewECPartSetFromData(data, testECPartSize)
	assert.NoError(t, err)

	// Test adding a part with wrong data.
	partSet2 := NewECPartSetFromHeader(partSet.Header())

	// Test adding a part with wrong trail.
	part := partSet.GetPart(0)
	part.Proof.Aunts[0][0] += byte(0x01)
	added, err := partSet2.AddPart(part)
	if added || err == nil {
		t.Errorf("expected to fail adding a part with bad trail.")
	}

	// Test adding a part with wrong bytes.
	part = partSet.GetPart(1)
	part.Bytes[0] += byte(0x01)
	added, err = partSet2.AddPart(part)
	if added || err == nil {
		t.Errorf("expected to fail adding a part with bad bytes.")
	}

	// Test adding a part with wrong proof index.
	part = partSet.GetPart(2)
	part.Proof.Index = 1
	added, err = partSet2.AddPart(part)
	if added || err == nil {
		t.Errorf("expected to fail adding a part with bad proof index.")
	}

	// Test adding a part with wrong proof total.
	part = partSet.GetPart(3)
	part.Proof.Total = int64(partSet.Total() - 1)
	added, err = partSet2.AddPart(part)
	if added || err == nil {
		t.Errorf("expected to fail adding a part with bad proof total.")
	}
}

func TestECPartSetHeaderValidateBasic(t *testing.T) {
	testCases := []struct {
		testName              string
		malleatePartSetHeader func(*ECPartSetHeader)
		expectErr             bool
	}{
		{"Good PartSet", func(psHeader *ECPartSetHeader) {}, false},
		{"Invalid Hash", func(psHeader *ECPartSetHeader) { psHeader.Hash = make([]byte, 1) }, true},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			data := cmtrand.Bytes(testECPartSize * 100)
			ps, err := NewECPartSetFromData(data, testECPartSize)
			assert.NoError(t, err)
			psHeader := ps.Header()
			tc.malleatePartSetHeader(&psHeader)
			assert.Equal(t, tc.expectErr, psHeader.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestECParSetHeaderProtoBuf(t *testing.T) {
	testCases := []struct {
		msg     string
		ps1     *ECPartSetHeader
		expPass bool
	}{
		{"success empty", &ECPartSetHeader{}, true},
		{
			"success",
			&ECPartSetHeader{Total: 1, Parity: 2, Hash: []byte("hash")}, true,
		},
	}

	for _, tc := range testCases {
		protoBlockID := tc.ps1.ToProto()

		psh, err := ECPartSetHeaderFromProto(&protoBlockID)
		if tc.expPass {
			require.Equal(t, tc.ps1, psh, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func BenchmarkMakeECPartSet(b *testing.B) {
	for nParts := 1; nParts <= 5; nParts++ {
		b.Run(fmt.Sprintf("nParts=%d", nParts), func(b *testing.B) {
			data := cmtrand.Bytes(testECPartSize * nParts)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = NewECPartSetFromData(data, testECPartSize)
			}
		})
	}
}
