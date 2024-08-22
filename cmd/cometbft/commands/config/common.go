package config

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/cmd/cometbft/commands"
	cfg "github.com/cometbft/cometbft/config"
)

func defaultConfigPath(cmd *cobra.Command) string {
	home, err := commands.ConfigHome(cmd)
	if err != nil {
		return ""
	}
	return filepath.Join(home, cfg.DefaultConfigDir, cfg.DefaultConfigFileName)
}
