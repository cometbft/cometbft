package identity_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/identity"
)

func TestLoadOrGenIsStable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "identity.json")

	k1, err := identity.LoadOrGen(path)
	require.NoError(t, err)
	require.NotNil(t, k1)

	k2, err := identity.LoadOrGen(path)
	require.NoError(t, err)
	require.True(t, k1.PubKey().Equals(k2.PubKey()), "reload must yield same key")
}
