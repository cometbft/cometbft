package config

import (
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/internal/confix"
)

// DiffCommand creates a new command for comparing configuration files.
func DiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [target-version] <config-path>",
		Short: "Outputs all config values that are different from the default.",
		Long:  "This command compares the configuration file with the defaults and outputs any differences.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var configPath string
			if len(args) > 1 {
				configPath = args[1]
			} else {
				configPath = defaultConfigPath(cmd)
			}

			targetVersion := args[0]
			if _, ok := confix.Migrations[targetVersion]; !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, slices.Collect(maps.Keys(confix.Migrations)))
			}

			targetVersionFile, err := confix.LoadLocalConfig(targetVersion + ".toml")
			if err != nil {
				return fmt.Errorf("failed to load internal config: %w", err)
			}

			rawFile, err := confix.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			diff := confix.DiffValues(rawFile, targetVersionFile)
			if len(diff) == 0 {
				fmt.Print("All config values are the same as the defaults.\n")
			}

			fmt.Print("The following config values are different from the defaults:\n")

			confix.PrintDiff(cmd.OutOrStdout(), diff)
			return nil
		},
	}

	return cmd
}
