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
	assert.NotEqual(t, sha256.Sum256([]byte("hello")), sum, "SHA256 and SHA512 should not have the same checksum")

	Set(crypto.MD5)
	h2 := New()
	h2.Write([]byte("hello"))
	sum2 := h2.Sum(nil)
	assert.NotNil(t, sum2)
	assert.NotEqual(t, sum, sum2, "MD5 and SHA512 should not have the same checksum")
}

func TestSetFmtHash(t *testing.T) {
	Set(crypto.SHA256)

	bz := []byte("hello")
	assert.Equal(t, "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824", Sum(bz).String())

	SetFmtHash(func(bz []byte) string {
		return hex.EncodeToString(bz)
	})
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", Sum(bz).String())
}
