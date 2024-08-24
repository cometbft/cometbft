package custom_test

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/cometbft/cometbft/crypto/custom" // This ensures that RegisterCustomCrypto works.

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
)

// Register custom crypto library.
func TestMain(m *testing.M) {
	var myPrivKey MySecpPrivKey
	var myPubKey MySecpPubKey
	var myBatchVerifier MySecpBatchVerifier
	ed25519.RegisterCustomCrypto(&myPrivKey, myPubKey, myBatchVerifier, options)
	os.Exit(m.Run())
}

// The below code is the same as TestSignAndValidateEd25519 in the ed25519_test.go file.
func TestSignAndValidateCustom(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifySignature(msg, sig))
}

// The below code is the same as TestBatchSafe in the ed25519_test.go file.
func TestBatchSafeCustom(t *testing.T) {
	v := ed25519.NewBatchVerifier()

	for i := 0; i <= 38; i++ {
		priv := ed25519.GenPrivKey()
		pub := priv.PubKey()

		var msg []byte
		if i%2 == 0 {
			msg = []byte("easter")
		} else {
			msg = []byte("egg")
		}

		sig, err := priv.Sign(msg)
		require.NoError(t, err)

		err = v.Add(pub, msg, sig)
		require.NoError(t, err)
	}

	ok, _ := v.Verify()
	require.True(t, ok)
}

// This test shows that the custom encryption implemented is compatible with the openssl implementation for secp256k1.
func TestProveCustomSecpCorrect(t *testing.T) {
	// Keys struct for ASN.1 decoding of openssl secp256k1 keys.
	type Keys struct {
		Version    int
		PrivateKey []byte
		Parameters asn1.RawValue  `asn1:"optional,explicit,tag:0"` // OID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey  asn1.BitString `asn1:"optional,explicit,tag:1"`
	}

	// Create a secp256k1 private and public key-pair using OpenSSL:
	// ```
	// openssl ecparam -genkey -name secp256k1 -noout | openssl ec 2> /dev/null | tee keys.pem | head -4 | tail -3
	// ```
	asnkeys, _ := base64.StdEncoding.DecodeString("MHQCAQEEIBu/bBBjTr7UBd1faMdjTsQ7AMdeLOMKsM0V6Ev2dhbeoAcGBSuBBAAK\noUQDQgAEtWVeKKkjqLlhCkSn/Dp3BC+xKzGRdZlIrbxG7RkyvhFjv23yT+PC/P7T\n1Whn7actJKwtq+ZPnUZxZrugSKYicg==")
	var keys Keys
	_, err := asn1.Unmarshal(asnkeys, &keys)
	require.NoError(t, err)

	// Create private key and public key using custom crypto implementation.
	privKey := ed25519.PrivKey(keys.PrivateKey) // This is equal to `privKey := MySecpPrivKey(keys.PrivateKey)` because of the custom crypto library registration.
	pubKey := privKey.PubKey()
	assert.Equal(t, keys.PublicKey.Bytes, pubKey.Bytes())

	// Create message hash to sign using OpenSSL:
	// ```
	// echo -n "Hello world" | openssl dgst -sha256 -binary | tee msg | base64
	// ```
	expectedHash, err := base64.StdEncoding.DecodeString("ZOyIygCyaOW6GjVnihtTFtIS9PNmskdyMlNKiuyjfzw=")
	require.NoError(t, err)

	// Create message to sign for custom crypto implementation. (It will be hashed internally.)
	msg := []byte("Hello world")

	// Compare internal SHA256 implementation with OpenSSL implementation, just in case.
	msgHash := sha256.Sum256(msg)
	assert.Equal(t, expectedHash, msgHash[:])

	// Create signature using OpenSSL:
	// ```
	// openssl pkeyutl -sign -in msg -inkey keys.pem | tee signature | base64
	// ```
	asnsignature, _ := base64.StdEncoding.DecodeString("MEUCIHTTt/ARtRcVSNwZsNMSMy6IGvQovxnPsEjnMsU/aKF+AiEA8eXQJMiqckQXH43bUzPgFc3qgvr47gY6GPHA5ttD9Ko=")

	// Create signature using custom crypto implementation:
	myasnsignature, err := privKey.Sign(msg)
	require.NoError(t, err)

	// We know that a deterministic signature of a hash of "Hello world" is this number.
	assert.Equal(t, hex.EncodeToString(myasnsignature), "3045022100bb63db53cb56989640dedd91bab680f8a57658b837e45381687a42e637fb81d4022075ca72b968d2ca04499423a0746d51e76e2b7a72b8a36acb3780b1ae71fa18b1")

	// Test signature using OpenSSL:
	// ```
	// openssl pkeyutl -verify -sigfile signature -in msg -inkey keys.pem -pubin
	// ```
	assert.True(t, pubKey.VerifySignature(msg, asnsignature))

	// Test signature using custom crypto implementation:
	assert.True(t, pubKey.VerifySignature(msg, myasnsignature))

	// Mutate the signature, just one bit.
	myasnsignature[7] ^= byte(0x01)
	asnsignature[7] ^= byte(0x01)
	assert.False(t, pubKey.VerifySignature(msg, myasnsignature))
	assert.False(t, pubKey.VerifySignature(msg, asnsignature))
}
