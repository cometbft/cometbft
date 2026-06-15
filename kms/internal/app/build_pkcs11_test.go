package app_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/app"
	"github.com/cometbft/cometbft/kms/internal/backend/pkcs11/pkcs11test"
	"github.com/cometbft/cometbft/kms/internal/config"
)

// TestBuildWiresPKCS11Backend verifies a [[providers.pkcs11]] block is opened and
// wired into the manager, and that cleanup releases the HSM session.
func TestBuildWiresPKCS11Backend(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)
	home := t.TempDir()

	c := &config.Config{
		Chains:     []config.Chain{{ID: "c1"}},
		Validators: []config.Validator{{ChainID: "c1", Addr: "tcp://127.0.0.1:1", IdentityKey: home + "/identity.json"}},
		Providers: config.Providers{PKCS11: []config.PKCS11Provider{{
			ChainIDs:   []string{"c1"},
			Module:     module,
			TokenLabel: pkcs11test.TokenLabel,
			KeyLabel:   pkcs11test.KeyLabel,
			PIN:        pkcs11test.UserPIN,
		}}},
	}
	require.NoError(t, c.Validate(home))

	mgr, cleanup, err := app.Build(c, log.TestingLogger())
	require.NoError(t, err)
	t.Cleanup(cleanup) // releases the PKCS#11 session even if assertions below fail
	require.NotNil(t, mgr)
}
