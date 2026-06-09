package signer

import (
	"context"

	"github.com/cometbft/cometbft/crypto"

	"github.com/cometbft/cometbft/kms/internal/backend"
)

// backendPrivKey adapts a backend.Signer to crypto.PrivKey so it can be handed
// to privval.NewFilePV. Only Sign, PubKey, and Type are exercised by the FilePV
// signing path; Bytes and Equals are intentionally unsupported for remote keys.
type backendPrivKey struct {
	ctx context.Context
	be  backend.Signer
	pub crypto.PubKey
}

var _ crypto.PrivKey = (*backendPrivKey)(nil)

// newBackendPrivKey caches the public key (so PubKey is cheap and FilePV's
// address computation works) and returns the adapter.
func newBackendPrivKey(ctx context.Context, be backend.Signer) (crypto.PrivKey, error) {
	pub, err := be.PubKey(ctx)
	if err != nil {
		return nil, err
	}
	return &backendPrivKey{ctx: ctx, be: be, pub: pub}, nil
}

func (k *backendPrivKey) Sign(msg []byte) ([]byte, error) { return k.be.Sign(k.ctx, msg) }
func (k *backendPrivKey) PubKey() crypto.PubKey           { return k.pub }
func (k *backendPrivKey) Type() string                    { return k.pub.Type() }

// Bytes is unsupported: remote/HSM keys never expose private material. It returns
// nil rather than panicking because crypto.PrivKey requires the method; it is not
// called on the FilePV signing path.
func (k *backendPrivKey) Bytes() []byte { return nil }

// Equals compares by public key (private material is unavailable).
func (k *backendPrivKey) Equals(other crypto.PrivKey) bool {
	return k.pub.Equals(other.PubKey())
}
