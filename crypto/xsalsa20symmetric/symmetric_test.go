package xsalsa20symmetric

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSimple(t *testing.T) {
	plaintext := []byte("sometext")
	secret := []byte("somesecretoflengththirtytwo===32")
	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.NoError(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}

func TestSimpleWithKDF(t *testing.T) {
	plaintext := []byte("sometext")
	secretPass := []byte("somesecret")
	secret, err := bcrypt.GenerateFromPassword(secretPass, 12)
	if err != nil {
		t.Error(err)
	}
	sum := sha256.Sum256(secret)
	copy(secret, sum[:])
	secret = secret[:32]

	ciphertext := EncryptSymmetric(plaintext, secret)
	plaintext2, err := DecryptSymmetric(ciphertext, secret)

	require.NoError(t, err, "%+v", err)
	assert.Equal(t, plaintext, plaintext2)
}
