// Command cometkms is an external remote signer for CometBFT validators.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/kms/internal/app"
	"github.com/cometbft/cometbft/kms/internal/config"
	"github.com/cometbft/cometbft/kms/internal/identity"
	"github.com/cometbft/cometbft/kms/internal/version"
)

const defaultConfigTemplate = `# cometkms configuration

[[chain]]
id = "my-chain-1"
# state_file defaults to <home>/state/<id>.json if omitted

[[validator]]
chain_id = "my-chain-1"
addr = "tcp://127.0.0.1:26659"
identity_key = "identity.json"

[[providers.softsign]]
chain_ids = ["my-chain-1"]
key_file = "priv_validator_key.json"
`

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{Use: "cometkms", Short: "External remote signer for CometBFT validators"}
	root.AddCommand(versionCmd(), initCmd(), startCmd())
	return root
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the cometkms version",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(version.String())
			return nil
		},
	}
}

func initCmd() *cobra.Command {
	var home string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a config file and generate an identity key",
		RunE:  func(_ *cobra.Command, _ []string) error { return runInit(home) },
	}
	cmd.Flags().StringVar(&home, "home", ".", "config/home directory")
	return cmd
}

func runInit(home string) error {
	if err := os.MkdirAll(home, 0o700); err != nil {
		return err
	}
	cfgPath := filepath.Join(home, "cometkms.toml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := os.WriteFile(cfgPath, []byte(defaultConfigTemplate), 0o600); err != nil {
			return err
		}
	}
	if _, err := identity.LoadOrGen(filepath.Join(home, "identity.json")); err != nil {
		return err
	}
	fmt.Printf("initialized cometkms in %s\n", home)
	return nil
}

func startCmd() *cobra.Command {
	var cfgPath, home string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Connect to validators and serve signing requests",
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			if err := cfg.Validate(home); err != nil {
				return err
			}

			mgr, err := app.Build(cfg, logger)
			if err != nil {
				return err
			}
			if err := mgr.Start(); err != nil {
				return err
			}
			defer mgr.Stop()

			logger.Info("cometkms started; press Ctrl-C to stop")
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			logger.Info("cometkms shutting down")
			return nil
		},
	}
	cmd.Flags().StringVarP(&cfgPath, "config", "c", "cometkms.toml", "path to config file")
	cmd.Flags().StringVar(&home, "home", ".", "home directory for relative paths/state")
	return cmd
}
