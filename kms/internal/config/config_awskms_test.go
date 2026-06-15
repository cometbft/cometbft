package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/config"
)

func awskmsCfg(extra string) string {
	return `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "i"

[[providers.awskms]]
chain_ids = ["c1"]
` + extra
}

func validateAWSKMS(t *testing.T, extra string) error {
	t.Helper()
	home := t.TempDir()
	cfgPath := filepath.Join(home, "cometkms.toml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(awskmsCfg(extra)), 0o600))
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	return c.Validate(home)
}

func TestAWSKMSGood(t *testing.T) {
	require.NoError(t, validateAWSKMS(t, `key_id = "alias/validator"
region = "us-east-1"
`))
}

func TestAWSKMSSatisfiesBackendRequirement(t *testing.T) {
	// A chain whose only provider is awskms must not trip the "no backend" check.
	require.NoError(t, validateAWSKMS(t, `key_id = "arn:aws:kms:us-east-1:1:key/abcd"
`))
}

func TestAWSKMSRequiresKeyID(t *testing.T) {
	err := validateAWSKMS(t, `region = "us-east-1"
`)
	require.ErrorContains(t, err, "key_id")
}

func TestAWSKMSUnknownAlgorithm(t *testing.T) {
	err := validateAWSKMS(t, `key_id = "k"
algorithm = "rsa-9000"
`)
	require.ErrorContains(t, err, "algorithm")
}

func TestAWSKMSUnknownChain(t *testing.T) {
	home := t.TempDir()
	body := `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "i"

[[providers.awskms]]
chain_ids = ["does-not-exist"]
key_id = "k"
`
	cfgPath := filepath.Join(home, "cometkms.toml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(body), 0o600))
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	// Must be the awskms unknown-chain error, not the generic "no backend" one.
	require.ErrorContains(t, c.Validate(home), "unknown chain")
}

func TestAWSKMSNoChainIDs(t *testing.T) {
	home := t.TempDir()
	body := `
[[chain]]
id = "c1"

[[validator]]
chain_id = "c1"
addr = "tcp://127.0.0.1:1"
identity_key = "i"

[[providers.awskms]]
key_id = "k"
chain_ids = []
`
	cfgPath := filepath.Join(home, "cometkms.toml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(body), 0o600))
	c, err := config.Load(cfgPath)
	require.NoError(t, err)
	// An awskms provider with empty chain_ids is rejected.
	require.ErrorContains(t, c.Validate(home), "chain_ids")
}
