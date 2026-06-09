package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/config"
)

const good = `
[[chain]]
id = "cosmoshub-4"
state_file = "STATE"

[[validator]]
chain_id = "cosmoshub-4"
addr = "tcp://127.0.0.1:26659"
identity_key = "IDENT"

[[providers.softsign]]
chain_ids = ["cosmoshub-4"]
key_file = "KEY"
`

func writeCfg(t *testing.T, body string) (cfgPath, home string) {
	t.Helper()
	home = t.TempDir()
	cfgPath = filepath.Join(home, "cometkms.toml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(body), 0o600))
	return cfgPath, home
}

func TestLoadAndValidateGood(t *testing.T) {
	cfgPath, home := writeCfg(t, good)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.NoError(t, c.Validate(home))

	require.Len(t, c.Chains, 1)
	require.NotEmpty(t, c.Chains[0].StateFile)
	_, statErr := os.Stat(filepath.Dir(c.Chains[0].StateFile))
	require.NoError(t, statErr)
}

func TestStateFileDefaultsWhenOmitted(t *testing.T) {
	body := `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "i"

[[providers.softsign]]
chain_ids = ["c1"]
key_file = "k"
`
	cfgPath, home := writeCfg(t, body)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.NoError(t, c.Validate(home))
	require.Equal(t, filepath.Join(home, "state", "c1.json"), c.Chains[0].StateFile)
}

func TestRelativeKeyPathsResolvedAgainstHome(t *testing.T) {
	body := `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "identity.json"

[[providers.softsign]]
chain_ids = ["c1"]
key_file = "key.json"
`
	cfgPath, home := writeCfg(t, body)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.NoError(t, c.Validate(home))

	require.Equal(t, filepath.Join(home, "identity.json"), c.Validators[0].IdentityKey)
	require.Equal(t, filepath.Join(home, "key.json"), c.Providers.Softsign[0].KeyFile)
}

func TestAbsoluteKeyPathsLeftUnchanged(t *testing.T) {
	absIdent := filepath.Join(t.TempDir(), "abs-identity.json")
	absKey := filepath.Join(t.TempDir(), "abs-key.json")
	body := `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "` + absIdent + `"

[[providers.softsign]]
chain_ids = ["c1"]
key_file = "` + absKey + `"
`
	cfgPath, home := writeCfg(t, body)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.NoError(t, c.Validate(home))

	require.Equal(t, absIdent, c.Validators[0].IdentityKey)
	require.Equal(t, absKey, c.Providers.Softsign[0].KeyFile)
}

func TestValidatorReferencesUnknownChain(t *testing.T) {
	body := `
[[chain]]
id = "c1"
[[validator]]
chain_id = "nope"
addr = "tcp://127.0.0.1:1"
identity_key = "i"
[[providers.softsign]]
chain_ids = ["c1"]
key_file = "k"
`
	cfgPath, home := writeCfg(t, body)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.ErrorContains(t, c.Validate(home), "unknown chain")
}

func TestChainWithoutBackendRejected(t *testing.T) {
	body := `
[[chain]]
id = "c1"
[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "i"
`
	cfgPath, home := writeCfg(t, body)
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	require.ErrorContains(t, c.Validate(home), "no backend")
}
