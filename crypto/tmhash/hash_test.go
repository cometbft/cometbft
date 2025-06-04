package tmhash_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/v2/crypto/tmhash"
)

func TestHash(t *testing.T) {
	testVector := []byte("abc")
	hasher := tmhash.New()
	_, err := hasher.Write(testVector)
	require.NoError(t, err)
	bz := hasher.Sum(nil)

	bz2 := tmhash.Sum(testVector)

	hasher = sha256.New()
	_, err = hasher.Write(testVector)
	require.NoError(t, err)
	bz3 := hasher.Sum(nil)

	assert.Equal(t, bz, bz2)
	assert.Equal(t, bz, bz3)
}

func TestHashTruncated(t *testing.T) {
	testVector := []byte("abc")
	hasher := tmhash.NewTruncated()
	_, err := hasher.Write(testVector)
	require.NoError(t, err)
	bz := hasher.Sum(nil)

	bz2 := tmhash.SumTruncated(testVector)

	hasher = sha256.New()
	_, err = hasher.Write(testVector)
	require.NoError(t, err)
	bz3 := hasher.Sum(nil)
	bz3 = bz3[:tmhash.TruncatedSize]

	assert.Equal(t, bz, bz2)
	assert.Equal(t, bz, bz3)
}

func TestValidSHA256String(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr string
	}{
		{
			"ValidLowercase",
			"9e107d9d372bb6826bd81d3542a419d6e9c2e15d35b3d5d6b889def626eb8f23",
			"",
		},
		{
			"ValidUppercase",
			"9E107D9D372BB6826BD81D3542A419D6E9C2E15D35B3D5D6B889DEF626EB8F23",
			"",
		},
		{
			"TooShort",
			"9e107d9d372bb6826bd81d3542a419d6e9c2e15d35b3d5d6b889def626eb8f2",
			"expected 64 characters, but have 63",
		},
		{
			"TooLong",
			"9e107d9d372bb6826bd81d3542a419d6e9c2e15d35b3d5d6b889def626eb8f23a",
			"expected 64 characters, but have 65",
		},
		{
			"InvalidChar",
			"9e107d9d372bb6826bd81d3542a419d6e9c2e15d35b3d5d6b889def626eb8f2g",
			"contains non-hexadecimal characters",
		},
	}

	// success cases
	for i, tt := range tests[:2] {
		t.Run(tt.name, func(t *testing.T) {
			err := tmhash.ValidateSHA256(tt.hash)
			assert.NoError(t, err, "test %d", i)
		})
	}

	// failure cases
	for i, tt := range tests[2:] {
		t.Run(tt.name, func(t *testing.T) {
			err := tmhash.ValidateSHA256(tt.hash)
			assert.EqualError(t, err, tt.wantErr, "test %d", i)
		})
	}
}
