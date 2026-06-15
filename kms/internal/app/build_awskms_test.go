package app_test

import (
	"path/filepath"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/app"
	"github.com/cometbft/cometbft/kms/internal/config"
)

func TestBuildAWSKMSProviderUnreachableErrors(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_MAX_ATTEMPTS", "1")

	home := t.TempDir()
	c := &config.Config{
		Chains:     []config.Chain{{ID: "c1"}},
		Validators: []config.Validator{{ChainID: "c1", Addr: "tcp://127.0.0.1:1", IdentityKey: filepath.Join(home, "id.json")}},
		Providers: config.Providers{AWSKMS: []config.AWSKMSProvider{{
			ChainIDs: []string{"c1"},
			KeyID:    "alias/validator",
			Region:   "us-east-1",
			Endpoint: "http://127.0.0.1:1", // closed port -> connection refused
		}}},
	}
	require.NoError(t, c.Validate(home))

	_, cleanup, err := app.Build(c, log.TestingLogger())
	t.Cleanup(cleanup)
	// The error must come from awskms.Open's GetPublicKey call (connection
	// refused against the closed port), proving the provider was wired in. Before
	// the wiring exists, Build instead errors with "chain has no backend", which
	// does NOT contain this substring — so this assertion is a true red->green.
	require.ErrorContains(t, err, "get public key")
}
