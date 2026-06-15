package awskms

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/stretchr/testify/require"
)

// fakeKMS is an in-process stand-in for AWS KMS backed by a real Ed25519 key. It
// lets the public-key-parse and sign->verify path run offline, exercising
// exactly the conversion logic in the backend.
type fakeKMS struct {
	priv      ed25519.PrivateKey
	keySpec   types.KeySpec
	getErr    error
	signErr   error
	badPubDER []byte // when set, GetPublicKey returns this instead of a valid SPKI
}

func newFakeKMS(t *testing.T) *fakeKMS {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return &fakeKMS{priv: priv, keySpec: types.KeySpecEccNistEdwards25519}
}

func (f *fakeKMS) GetPublicKey(_ context.Context, _ *kms.GetPublicKeyInput, _ ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	der := f.badPubDER
	if der == nil {
		var err error
		der, err = x509.MarshalPKIXPublicKey(f.priv.Public())
		if err != nil {
			return nil, err
		}
	}
	return &kms.GetPublicKeyOutput{PublicKey: der, KeySpec: f.keySpec}, nil
}

func (f *fakeKMS) Sign(_ context.Context, in *kms.SignInput, _ ...func(*kms.Options)) (*kms.SignOutput, error) {
	if f.signErr != nil {
		return nil, f.signErr
	}
	if in.MessageType != types.MessageTypeRaw {
		return nil, errors.New("fakeKMS: unexpected message type")
	}
	if in.SigningAlgorithm != types.SigningAlgorithmSpecEd25519Sha512 {
		return nil, errors.New("fakeKMS: unexpected signing algorithm")
	}
	// Mirror KMS ED25519_SHA_512 + MessageType=RAW: PureEd25519 over the message.
	return &kms.SignOutput{Signature: ed25519.Sign(f.priv, in.Message)}, nil
}

func TestOpenAndSignRoundtrip(t *testing.T) {
	f := newFakeKMS(t)
	s, err := open(context.Background(), f, "alias/validator", algos[algoEd25519])
	require.NoError(t, err)

	pub, err := s.PubKey(context.Background())
	require.NoError(t, err)
	require.Equal(t, algoEd25519, pub.Type())
	require.Equal(t, []byte(f.priv.Public().(ed25519.PublicKey)), pub.Bytes())

	msg := []byte("canonical consensus sign-bytes")
	sig, err := s.Sign(context.Background(), msg)
	require.NoError(t, err)
	require.Len(t, sig, ed25519.SignatureSize)
	require.True(t, pub.VerifySignature(msg, sig), "consensus pubkey must verify the KMS signature")
}

func TestOpenRejectsWrongKeySpec(t *testing.T) {
	f := newFakeKMS(t)
	f.keySpec = types.KeySpecEccSecgP256k1
	_, err := open(context.Background(), f, "k", algos[algoEd25519])
	require.ErrorContains(t, err, "spec")
}

func TestOpenPropagatesGetPublicKeyError(t *testing.T) {
	f := newFakeKMS(t)
	f.getErr = errors.New("access denied")
	_, err := open(context.Background(), f, "k", algos[algoEd25519])
	require.ErrorContains(t, err, "access denied")
}

func TestOpenRejectsUndecodablePublicKey(t *testing.T) {
	f := newFakeKMS(t)
	f.badPubDER = []byte("not-a-valid-spki")
	_, err := open(context.Background(), f, "k", algos[algoEd25519])
	require.ErrorContains(t, err, "decode public key")
}

func TestSignPropagatesError(t *testing.T) {
	f := newFakeKMS(t)
	s, err := open(context.Background(), f, "k", algos[algoEd25519])
	require.NoError(t, err)
	f.signErr = errors.New("throttled")
	_, err = s.Sign(context.Background(), []byte("m"))
	require.ErrorContains(t, err, "throttled")
}

func TestOpenUnknownAlgorithm(t *testing.T) {
	_, err := Open(context.Background(), Config{KeyID: "k", Algorithm: "rsa-9000"})
	require.ErrorContains(t, err, "unknown algorithm")
}
