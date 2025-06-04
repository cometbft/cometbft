//go:build bls12381

package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	e2e "github.com/cometbft/cometbft/v2/test/e2e/pkg"
)

// TestGenerator tests that only valid manifests are generated.
func TestGenerator(t *testing.T) {
	cfg := &generateConfig{
		randSource: rand.New(rand.NewSource(randomSeed)),
	}
	manifests, err := Generate(cfg)
	require.NoError(t, err)

	for idx, m := range manifests {
		t.Run(fmt.Sprintf("Case%04d", idx), func(t *testing.T) {
			infra, err := e2e.NewDockerInfrastructureData(m)
			require.NoError(t, err)
			_, err = e2e.NewTestnetFromManifest(m, filepath.Join(t.TempDir(), fmt.Sprintf("Case%04d", idx)), infra, "")
			require.NoError(t, err)
		})
	}
}
