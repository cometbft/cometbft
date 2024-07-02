package config

import (
	"github.com/spf13/cobra"
)

// Command contains all the confix commands
// These command can be used to interactively update a config value.
func Command() *cobra.Command {
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
