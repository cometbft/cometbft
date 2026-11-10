package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/test"
)

func ensureFiles(t *testing.T, rootDir string, files ...string) {
	for _, f := range files {
		p := filepath.Join(rootDir, f)
		_, err := os.Stat(p)
		assert.NoError(t, err, p)
	}
}

func TestEnsureRoot(t *testing.T) {
	require := require.New(t)

	// setup temp dir for test
	tmpDir, err := os.MkdirTemp("", "config-test")
	require.Nil(err)
	defer os.RemoveAll(tmpDir)

	// create root dir
	config.EnsureRoot(tmpDir)

	// make sure config is set properly
	data, err := os.ReadFile(filepath.Join(tmpDir, config.DefaultConfigDir, config.DefaultConfigFileName))
	require.Nil(err)

	assertValidConfig(t, string(data))

	ensureFiles(t, tmpDir, "data")
}

func TestEnsureTestRoot(t *testing.T) {
	require := require.New(t)

	// create root dir
	cfg := test.ResetTestRoot("ensureTestRoot")
	defer os.RemoveAll(cfg.RootDir)
	rootDir := cfg.RootDir

	// make sure config is set properly
	data, err := os.ReadFile(filepath.Join(rootDir, config.DefaultConfigDir, config.DefaultConfigFileName))
	require.Nil(err)

	assertValidConfig(t, string(data))

	// TODO: make sure the cfg returned and testconfig are the same!
	baseConfig := config.DefaultBaseConfig()
	ensureFiles(t, rootDir, config.DefaultDataDir, baseConfig.Genesis, baseConfig.PrivValidatorKey, baseConfig.PrivValidatorState)
}

func TestWriteConfigFileScalerMaxQueueSizeRoundTrip(t *testing.T) {
	zero, seven := 0, 7
	cfg := config.DefaultConfig()
	cfg.P2P.LibP2PConfig.Scaler.MaxQueueSize = 123
	cfg.P2P.LibP2PConfig.Scaler.Overrides = []config.LibP2PScalerOverride{
		{Reactor: "UNSET", MinWorkers: 1, MaxWorkers: 2},
		{Reactor: "UNBOUNDED", MinWorkers: 1, MaxWorkers: 2, MaxQueueSize: &zero},
		{Reactor: "CAPPED", MinWorkers: 1, MaxWorkers: 2, MaxQueueSize: &seven},
	}

	configFile := filepath.Join(t.TempDir(), "config.toml")
	config.WriteConfigFile(configFile, cfg)

	v := viper.New()
	v.SetConfigFile(configFile)
	require.NoError(t, v.ReadInConfig())

	reloaded := config.DefaultConfig()
	require.NoError(t, v.Unmarshal(reloaded))

	scaler := reloaded.P2P.LibP2PConfig.Scaler
	require.Equal(t, 123, scaler.MaxQueueSize)
	require.Len(t, scaler.Overrides, 3)
	require.Nil(t, scaler.Overrides[0].MaxQueueSize, "unset override must stay unset to inherit the global cap")
	require.NotNil(t, scaler.Overrides[1].MaxQueueSize)
	require.Equal(t, 0, *scaler.Overrides[1].MaxQueueSize)
	require.NotNil(t, scaler.Overrides[2].MaxQueueSize)
	require.Equal(t, 7, *scaler.Overrides[2].MaxQueueSize)
}

func assertValidConfig(t *testing.T, configFile string) {
	t.Helper()
	// list of words we expect in the config
	elems := []string{
		"moniker",
		"seeds",
		"proxy_app",
		"create_empty_blocks",
		"peer",
		"timeout",
		"broadcast",
		"send",
		"addr",
		"wal",
		"propose",
		"max",
		"genesis",
	}
	for _, e := range elems {
		assert.Contains(t, configFile, e)
	}
}
