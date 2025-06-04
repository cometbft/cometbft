package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/cometbft/cometbft/v2/config"
	"github.com/cometbft/cometbft/v2/libs/cli"
	cmtflags "github.com/cometbft/cometbft/v2/libs/cli/flags"
	"github.com/cometbft/cometbft/v2/libs/log"
)

var (
	config = cfg.DefaultConfig()
	logger = log.NewLogger(os.Stdout)
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", config.LogLevel, "log level")
}

func ConfigHome(cmd *cobra.Command) (string, error) {
	var home string
	switch {
	case os.Getenv("CMTHOME") != "":
		home = os.Getenv("CMTHOME")
	case os.Getenv("TMHOME") != "":
		// XXX: Deprecated.
		home = os.Getenv("TMHOME")
	default:
		var err error
		// Default: $HOME/.cometbft
		home, err = cmd.Flags().GetString(cli.HomeFlag)
		if err != nil {
			return "", err
		}
	}

	return home, nil
}

// ParseConfig retrieves the default environment configuration,
// sets up the CometBFT root and ensures that the root exists.
func ParseConfig(cmd *cobra.Command) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	if os.Getenv("TMHOME") != "" {
		logger.Error("Deprecated environment variable TMHOME identified. CMTHOME should be used instead.")
	}
	home, err := ConfigHome(cmd)
	if err != nil {
		return nil, err
	}
	conf.RootDir = home

	conf.SetRoot(conf.RootDir)
	cfg.EnsureRoot(conf.RootDir)
	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}
	if warnings := conf.CheckDeprecated(); len(warnings) > 0 {
		for _, warning := range warnings {
			logger.Warn("deprecated usage found in configuration file", "usage", warning)
		}
	}
	return conf, nil
}

// RootCmd is the root command for CometBFT core.
var RootCmd = &cobra.Command{
	Use:   "cometbft",
	Short: "BFT state machine replication for applications in any programming languages",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		config, err = ParseConfig(cmd)
		if err != nil {
			return err
		}

		for _, possibleMisconfiguration := range config.PossibleMisconfigurations() {
			logger.Info(possibleMisconfiguration)
		}

		if config.LogFormat == cfg.LogFormatJSON {
			logger = log.NewJSONLogger(os.Stdout)
		} else if !config.LogColors {
			logger = log.NewLoggerWithColor(os.Stdout, false)
		}

		logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)
		if err != nil {
			return err
		}

		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}

		return nil
	},
}
