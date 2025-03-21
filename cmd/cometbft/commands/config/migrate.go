package config

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/internal/confix"
)

var (
	FlagStdOut       bool
	FlagVerbose      bool
	FlagSkipValidate bool
)

func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] <config-path>",
		Short: "Migrate configuration file to the specified version",
		Long: `Migrate the contents of the configuration to the specified version.
The output is written in-place unless --stdout is provided.
In case of any error in updating the file, no output is written.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var configPath string
			if len(args) > 1 {
				configPath = args[1]
			} else {
				configPath = defaultConfigPath(cmd)
			}

			targetVersion := args[0]
			plan, ok := confix.Migrations[targetVersion]
			if !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, slices.Collect(maps.Keys(confix.Migrations)))
			}

			rawFile, err := confix.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()
			if FlagVerbose {
				ctx = confix.WithLogWriter(ctx, cmd.ErrOrStderr())
			}

			outputPath := configPath
			if FlagStdOut {
				outputPath = ""
			}

			if err := confix.Upgrade(ctx, plan(rawFile, targetVersion), configPath, outputPath, FlagSkipValidate); err != nil {
				return fmt.Errorf("failed to migrate config: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")
	cmd.Flags().BoolVar(&FlagVerbose, "verbose", false, "log changes to stderr")
	cmd.Flags().BoolVar(&FlagSkipValidate, "skip-validate", false, "skip configuration validation (allows to migrate unknown configurations)")

	return cmd
}
