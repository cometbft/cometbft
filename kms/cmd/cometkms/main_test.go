package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitCreatesConfigAndIdentity(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, runInit(home))

	_, err := os.Stat(filepath.Join(home, "cometkms.toml"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(home, "identity.json"))
	require.NoError(t, err)
}
