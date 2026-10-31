package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWithArgsRestoresEnv(t *testing.T) {
	const unset, set, copied = "CMTCLITEST_UNSET", "CMTCLITEST_SET", "CMTCLITEST_COPIED"
	require.NoError(t, os.Unsetenv(unset))
	require.NoError(t, os.Unsetenv(copied))
	t.Setenv(set, "before")

	cmd := Executor{Command: &cobra.Command{Use: "noop", RunE: func(*cobra.Command, []string) error {
		// variables the command sets itself must not leak either
		return os.Setenv(copied, "x")
	}}}
	err := RunWithArgs(cmd, []string{"noop"}, map[string]string{unset: "x", set: "y"})
	require.NoError(t, err)

	_, ok := os.LookupEnv(unset)
	assert.False(t, ok, "a variable that was unset before must be unset again")
	_, ok = os.LookupEnv(copied)
	assert.False(t, ok, "a variable set by the command must be removed")
	assert.Equal(t, "before", os.Getenv(set))
}
