package txhash

import (
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	h := New()
	h.Write([]byte("hello"))

	assert.NotNil(t, h.Sum(nil))
}

func TestSum(t *testing.T) {
	bz := []byte("hello")

	assert.NotNil(t, Sum(bz))
}

func TestSet(t *testing.T) {
	Set(crypto.SHA512)
	h := New()
	h.Write([]byte("hello"))
	sum := h.Sum(nil)

	assert.NotNil(t, sum)
	assert.NotEqual(t, sha256.Sum256([]byte("hello")), sum)
}

func TestSetFmtHash(t *testing.T) {
	bz := []byte("hello")
	assert.Equal(t, "9B71D224BD62F3785D96D46AD3EA3D73319BFBC2890CAADAE2DFF72519673CA72323C3D99BA5C11D7C7ACC6E14B8C5DA0C4663475C2E5C3ADEF46F73BCDEC043", Sum(bz).String())

	SetFmtHash(func(bz []byte) string {
		return hex.EncodeToString(bz)
	})
	assert.Equal(t, "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", Sum(bz).String())
}
