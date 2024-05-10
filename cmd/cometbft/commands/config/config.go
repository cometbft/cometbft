package config

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/libs/log"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

// ConfigCommand contains all the confix commands
// These command can be used to interactively update a config value.
func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Utilities for managing configuration",
	}

	cmd.AddCommand(
		MigrateCommand(),
		DiffCommand(),
		GetCommand(),
		SetCommand(),
		ViewCommand(),
	)

	return cmd
}
