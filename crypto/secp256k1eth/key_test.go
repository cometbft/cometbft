package secp256k1eth_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	cmtjson "github.com/cometbft/cometbft/libs/json"

	"github.com/cometbft/cometbft/crypto/secp256k1eth"
)

func TestGenAndDeterministicGen(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	require.Len(t, priv.Bytes(), secp256k1eth.PrivKeySize)
	require.Equal(t, secp256k1eth.KeyType, priv.Type())

	pub := priv.PubKey()
	require.Len(t, pub.Bytes(), secp256k1eth.PubKeySize)
	require.Equal(t, secp256k1eth.KeyType, pub.Type())

	a := secp256k1eth.GenPrivKeySecp256k1Eth([]byte("seed-bytes"))
	b := secp256k1eth.GenPrivKeySecp256k1Eth([]byte("seed-bytes"))
	require.Len(t, a.Bytes(), secp256k1eth.PrivKeySize)
	require.True(t, a.Equals(b))
	require.False(t, a.Equals(secp256k1eth.GenPrivKeySecp256k1Eth([]byte("other"))))
}

func TestSignProducesEthFormat(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	sig, err := priv.Sign([]byte("hello cometbft"))
	require.NoError(t, err)
	require.Len(t, sig, secp256k1eth.SignatureSize) // 65 bytes: [R || S || V]
	require.LessOrEqual(t, sig[64], byte(1))        // V in {0,1}, go-ethereum form
}

func TestAddressKnownAnswer(t *testing.T) {
	// Widely published secp256k1 -> Ethereum address vector (Hardhat/Anvil
	// account #0). Proves legacy-Keccak address derivation matches the wider
	// Ethereum ecosystem, i.e. go-ethereum compatibility.
	const (
		privHex     = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
		wantAddrHex = "f39fd6e51aad88f6f4ce6ab8827279cfffb92266"
	)
	privBz, err := hex.DecodeString(privHex)
	require.NoError(t, err)

	addr := secp256k1eth.PrivKey(privBz).PubKey().Address()
	require.Equal(t, wantAddrHex, hex.EncodeToString(addr))
}

func TestVerifySignature(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	pub := priv.PubKey()
	msg := []byte("the quick brown fox")

	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	require.True(t, pub.VerifySignature(msg, sig))      // 65-byte form
	require.True(t, pub.VerifySignature(msg, sig[:64])) // 64-byte form
	require.False(t, pub.VerifySignature([]byte("tampered"), sig))
	require.False(t, pub.VerifySignature(msg, sig[:63])) // wrong length
}

func TestNewPubKeyFromBytes(t *testing.T) {
	pub := secp256k1eth.GenPrivKey().PubKey()
	got, err := secp256k1eth.NewPubKeyFromBytes(pub.Bytes())
	require.NoError(t, err)
	require.True(t, pub.Equals(got))

	_, err = secp256k1eth.NewPubKeyFromBytes([]byte{0x01, 0x02})
	require.Error(t, err)
}

func TestSignatureRecoversSigner(t *testing.T) {
	// Proves the [R || S || V] byte order is go-ethereum-correct: reconstruct
	// decred's compact form and recover the signer's pubkey from V.
	priv := secp256k1eth.GenPrivKey()
	want := priv.PubKey().Address()
	msg := []byte("recover me")

	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	// eth [R || S || V] -> decred compact [V+27 || R || S].
	compact := make([]byte, 65)
	compact[0] = sig[64] + 27
	copy(compact[1:], sig[:64])

	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write(msg)
	recovered, _, err := ecdsa.RecoverCompact(compact, h.Sum(nil))
	require.NoError(t, err)

	got := secp256k1eth.PubKey(recovered.SerializeCompressed()).Address()
	require.Equal(t, want, got)
}

func TestGenPrivKeySecp256k1EthIsValidFieldElement(t *testing.T) {
	// Every deterministically derived key must be a valid scalar: 0 < d < N.
	n := secp256k1.Params().N
	for _, seed := range [][]byte{[]byte("a"), []byte("b"), []byte("seed-bytes"), {}} {
		priv := secp256k1eth.GenPrivKeySecp256k1Eth(seed)
		d := new(big.Int).SetBytes(priv.Bytes())
		require.Equal(t, 1, d.Sign(), "scalar must be > 0")
		require.True(t, d.Cmp(n) < 0, "scalar must be < curve order N")
	}
}

func TestVerifyRejectsMalleableHighS(t *testing.T) {
	// A signature with S in the upper half (S' = N - S) is malleable and must
	// be rejected, even though it is otherwise a valid ECDSA signature.
	priv := secp256k1eth.GenPrivKey()
	pub := priv.PubKey()
	msg := []byte("malleable check")

	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	require.True(t, pub.VerifySignature(msg, sig), "canonical signature should verify")

	// Build the high-S variant: S' = N - S, keeping R unchanged.
	n := secp256k1.Params().N
	s := new(big.Int).SetBytes(sig[32:64])
	highS := new(big.Int).Sub(n, s)

	malleable := make([]byte, 64)
	copy(malleable[:32], sig[:32]) // R unchanged
	highSBytes := highS.Bytes()
	copy(malleable[64-len(highSBytes):64], highSBytes)

	require.False(t, pub.VerifySignature(msg, malleable),
		"high-S (malleable) signature must be rejected")
}

func TestJSONRoundTrip(t *testing.T) {
	// init() registers the amino/json routes; ensure both keys round-trip.
	priv := secp256k1eth.GenPrivKey()
	pub := priv.PubKey().(secp256k1eth.PubKey)

	privBz, err := cmtjson.Marshal(priv)
	require.NoError(t, err)
	var priv2 secp256k1eth.PrivKey
	require.NoError(t, cmtjson.Unmarshal(privBz, &priv2))
	require.True(t, priv.Equals(priv2))

	pubBz, err := cmtjson.Marshal(pub)
	require.NoError(t, err)
	var pub2 secp256k1eth.PubKey
	require.NoError(t, cmtjson.Unmarshal(pubBz, &pub2))
	require.True(t, pub.Equals(pub2))
}
