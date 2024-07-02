package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func ViewCommand() *cobra.Command {
	flagOutputFormat := "output-format"

	cmd := &cobra.Command{
		Use:   "view [config]",
		Short: "View the config file",
		Long:  "View the config file. The [config] is an optional absolute path to the config file (default: `~/.cometbft/config/config.toml`)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var filename string
			if len(args) > 0 {
				filename = args[0]
			} else {
				filename = defaultConfigPath(cmd)
			}

			file, err := os.ReadFile(filename)
			if err != nil {
				return err
			}

			if format, _ := cmd.Flags().GetString(flagOutputFormat); format == "toml" {
				cmd.Println(string(file))
				return nil
			}

			var v any
			if err := toml.Unmarshal(file, &v); err != nil {
				return fmt.Errorf("failed to decode config file: %w", err)
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(v)
		},
	}

	// output flag
	cmd.Flags().String(flagOutputFormat, "toml", "Output format (json|toml)")

	return cmd
}
