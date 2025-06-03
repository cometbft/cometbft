package confix_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cometbft/cometbft/v2/internal/confix"
)

func mustReadConfig(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	return f
}

func TestCheckValid(t *testing.T) {
	err := confix.CheckValid(mustReadConfig(t, "data/v0.34.toml"))
	assert.NoError(t, err)

	err = confix.CheckValid(mustReadConfig(t, "data/v0.37.toml"))
	assert.NoError(t, err)

	err = confix.CheckValid(mustReadConfig(t, "data/v0.38.toml"))
	assert.NoError(t, err)

	err = confix.CheckValid(mustReadConfig(t, "data/v1.0.toml"))
	assert.NoError(t, err)
}
