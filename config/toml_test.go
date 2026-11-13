package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/BurntSushi/toml"
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

// Regression test for https://github.com/cometbft/cometbft/issues/5991:
// bootstrap_peers must render on its own line, not inside the preceding comment.
func TestBootstrapPeersRendered(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tmpDir := t.TempDir()
		config.EnsureRoot(tmpDir)

		data, err := os.ReadFile(filepath.Join(tmpDir, config.DefaultConfigDir, config.DefaultConfigFileName))
		require.NoError(t, err)

		assert.Regexp(t, regexp.MustCompile(`(?m)^bootstrap_peers = \[\]$`), string(data))
		assertBootstrapPeersParsed(t, data, 0)
	})

	t.Run("with peers", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.P2P.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "192.0.2.1:26656", ID: "peer-1"},
			{Host: "192.0.2.2:26656", ID: "peer-2", Persistent: true},
		}

		path := filepath.Join(t.TempDir(), config.DefaultConfigFileName)
		config.WriteConfigFile(path, cfg)

		data, err := os.ReadFile(path)
		require.NoError(t, err)

		assert.Regexp(t, regexp.MustCompile(`(?m)^bootstrap_peers = \[$`), string(data))
		assertBootstrapPeersParsed(t, data, 2)
	})
}

func assertBootstrapPeersParsed(t *testing.T, data []byte, expected int) {
	t.Helper()

	var parsed struct {
		P2P struct {
			LibP2P struct {
				BootstrapPeers []map[string]any `toml:"bootstrap_peers"`
			} `toml:"libp2p"`
		} `toml:"p2p"`
	}
	md, err := toml.Decode(string(data), &parsed)
	require.NoError(t, err)
	require.True(t, md.IsDefined("p2p", "libp2p", "bootstrap_peers"), "bootstrap_peers is missing from the parsed config")
	assert.Len(t, parsed.P2P.LibP2P.BootstrapPeers, expected)
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
