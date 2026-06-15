// Package awskms implements a backend.Signer backed by an AWS KMS asymmetric
// key. The private key never leaves KMS: signing is performed by the KMS Sign
// API. Ed25519 (ECC_NIST_EDWARDS25519 + ED25519_SHA_512, PureEdDSA over the
// canonical sign-bytes) is the only key algorithm today; see algo.go for the
// per-algorithm seam.
package awskms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/kms/internal/backend"
)

var _ backend.Signer = (*Signer)(nil)

// Config describes how to reach a signing key in AWS KMS. Credentials are
// resolved by the standard AWS default chain (environment, shared config, SSO,
// container/instance role); no secret material is read from this struct.
type Config struct {
	KeyID     string // KMS key id, ARN, or alias/<name>
	Region    string // optional; falls back to the AWS default chain
	Profile   string // optional shared-config profile
	Endpoint  string // optional endpoint override (LocalStack / testing)
	Algorithm string // key algorithm; defaults to "ed25519"
}

// kmsAPI is the subset of the AWS KMS client cometkms uses. *kms.Client
// satisfies it; tests inject a fake.
type kmsAPI interface {
	GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error)
}

// The concrete AWS KMS client must satisfy the interface we depend on.
var _ kmsAPI = (*kms.Client)(nil)

// Signer is a backend.Signer that signs via the AWS KMS Sign API. It is
// stateless beyond the cached public key and is safe for concurrent use (the
// AWS SDK client is concurrency-safe).
type Signer struct {
	client kmsAPI
	keyID  string
	pub    crypto.PubKey
	algo   keyAlgo
}

// Open resolves AWS configuration, builds a KMS client, fetches and caches the
// key's public key, and validates its spec against the configured algorithm. Any
// failure is returned (fatal at startup for the chain). It performs one KMS
// GetPublicKey call.
func Open(ctx context.Context, cfg Config) (*Signer, error) {
	algoName := cfg.Algorithm
	if algoName == "" {
		algoName = algoEd25519
	}
	algo, ok := algos[algoName]
	if !ok {
		return nil, fmt.Errorf("awskms: unknown algorithm %q", algoName)
	}

	var loadOpts []func(*awsconfig.LoadOptions) error
	if cfg.Region != "" {
		loadOpts = append(loadOpts, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.Profile != "" {
		loadOpts = append(loadOpts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("awskms: load AWS config: %w", err)
	}
	client := kms.NewFromConfig(awsCfg, func(o *kms.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})
	return open(ctx, client, cfg.KeyID, algo)
}

// open is the client-injectable core of Open: it fetches the public key,
// verifies the key spec, and decodes the public key. Tests call it with a fake
// kmsAPI.
func open(ctx context.Context, client kmsAPI, keyID string, algo keyAlgo) (*Signer, error) {
	out, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: aws.String(keyID)})
	if err != nil {
		return nil, fmt.Errorf("awskms: get public key for %q: %w", keyID, err)
	}
	if out.KeySpec != algo.keySpec {
		return nil, fmt.Errorf("awskms: key %q has spec %q, expected %q for algorithm %q",
			keyID, out.KeySpec, algo.keySpec, algo.name)
	}
	pub, err := algo.decodePub(out.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("awskms: decode public key for %q: %w", keyID, err)
	}
	return &Signer{client: client, keyID: keyID, pub: pub, algo: algo}, nil
}

// PubKey returns the validator public key cached at Open.
func (s *Signer) PubKey(context.Context) (crypto.PubKey, error) { return s.pub, nil }

// Sign signs the canonical consensus sign-bytes via the KMS Sign API.
func (s *Signer) Sign(ctx context.Context, signBytes []byte) ([]byte, error) {
	out, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:            aws.String(s.keyID),
		Message:          signBytes,
		MessageType:      types.MessageTypeRaw,
		SigningAlgorithm: s.algo.signAlgo,
	})
	if err != nil {
		return nil, fmt.Errorf("awskms: sign with %q: %w", s.keyID, err)
	}
	return s.algo.fixSig(out.Signature)
}
