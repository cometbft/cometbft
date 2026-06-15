package pkcs11

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePIN_Inline(t *testing.T) {
	pin, err := resolvePIN(Config{PIN: "1234"})
	require.NoError(t, err)
	require.Equal(t, "1234", pin)
}

func TestResolvePIN_Env(t *testing.T) {
	t.Setenv("COMETKMS_TEST_PIN", "envpin")
	pin, err := resolvePIN(Config{PINEnv: "COMETKMS_TEST_PIN"})
	require.NoError(t, err)
	require.Equal(t, "envpin", pin)
}

func TestResolvePIN_EnvEmpty(t *testing.T) {
	t.Setenv("COMETKMS_TEST_PIN", "")
	_, err := resolvePIN(Config{PINEnv: "COMETKMS_TEST_PIN"})
	require.Error(t, err)
}

func TestResolvePIN_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pin")
	require.NoError(t, os.WriteFile(path, []byte("filepin\n"), 0o600))
	pin, err := resolvePIN(Config{PINFile: path})
	require.NoError(t, err)
	require.Equal(t, "filepin", pin) // trailing newline trimmed
}

func TestResolvePIN_FileEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pin")
	require.NoError(t, os.WriteFile(path, []byte("\n"), 0o600))
	_, err := resolvePIN(Config{PINFile: path})
	require.Error(t, err)
}

func TestResolvePIN_FileMissing(t *testing.T) {
	_, err := resolvePIN(Config{PINFile: filepath.Join(t.TempDir(), "nope")})
	require.Error(t, err)
}

func TestResolvePIN_NoSource(t *testing.T) {
	_, err := resolvePIN(Config{})
	require.Error(t, err)
}
