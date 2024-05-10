package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/cmd/cometbft/commands"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func ViewCommand() *cobra.Command {
	flagOutputFormat := "output-format"

	cmd := &cobra.Command{
		Use:   "view [config]",
		Short: "View the config file",
		Long:  "View the config file. The [config] is an absolute path to the config file (default: `~/.cometbft/config/config.toml`)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			if filename == "" {
				home, err := commands.ConfigHome(cmd)
				if err != nil {
					return err
				}
				filename = filepath.Join(home, cfg.DefaultConfigDir, cfg.DefaultConfigFileName)
			}

			file, err := os.ReadFile(filename)
			if err != nil {
				return err
			}

			if format, _ := cmd.Flags().GetString(flagOutputFormat); format == "toml" {
				cmd.Println(string(file))
				return nil
			}

			var v interface{}
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
