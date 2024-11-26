package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/libs/log"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra/digitalocean"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra/docker"
)

const randomSeed = 2308084734268

var logger = log.NewLoggerWithColor(os.Stdout, false)

func main() {
	NewCLI().Run()
}

// CLI is the Cobra-based command-line interface.
type CLI struct {
	root     *cobra.Command
	testnet  *e2e.Testnet
	preserve bool
	infp     infra.Provider
}

// NewCLI sets up the CLI.
func NewCLI() *CLI {
	cli := &CLI{}
	cli.root = &cobra.Command{
		Use:           "runner",
		Short:         "End-to-end test runner",
		SilenceUsage:  true,
		SilenceErrors: true, // we'll output them ourselves in Run()
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			m, err := e2e.LoadManifest(file)
			if err != nil {
				return err
			}

			inft, err := cmd.Flags().GetString("infrastructure-type")
			if err != nil {
				return err
			}

			var ifd e2e.InfrastructureData
			switch inft {
			case "docker":
				var err error
				ifd, err = e2e.NewDockerInfrastructureData(m)
				if err != nil {
					return err
				}
			case "digital-ocean":
				p, err := cmd.Flags().GetString("infrastructure-data")
				if err != nil {
					return err
				}
				if p == "" {
					return errors.New("'--infrastructure-data' must be set when using the 'digital-ocean' infrastructure-type")
				}
				ifd, err = e2e.InfrastructureDataFromFile(p)
				if err != nil {
					return fmt.Errorf("parsing infrastructure data: %s", err)
				}
			default:
				return fmt.Errorf("unknown infrastructure type '%s'", inft)
			}

			testnetDir, err := cmd.Flags().GetString("testnet-dir")
			if err != nil {
				return err
			}

			testnet, err := e2e.LoadTestnet(file, ifd, testnetDir)
			if err != nil {
				return fmt.Errorf("loading testnet: %s", err)
			}

			cli.testnet = testnet
			switch inft {
			case "docker":
				cli.infp = &docker.Provider{
					ProviderData: infra.ProviderData{
						Testnet:            testnet,
						InfrastructureData: ifd,
					},
				}
			case "digital-ocean":
				cli.infp = &digitalocean.Provider{
					ProviderData: infra.ProviderData{
						Testnet:            testnet,
						InfrastructureData: ifd,
					},
				}
			default:
				return fmt.Errorf("bad infrastructure type: %s", inft)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := Cleanup(cli.testnet); err != nil {
				return err
			}
			if err := Setup(cli.testnet, cli.infp); err != nil {
				return err
			}

			r := rand.New(rand.NewSource(randomSeed)) //nolint: gosec

			chLoadResult := make(chan error)
			ctx, loadCancel := context.WithCancel(context.Background())
			defer loadCancel()
			go func() {
				err := Load(ctx, cli.testnet, false)
				if err != nil {
					logger.Error(fmt.Sprintf("Transaction load failed: %v", err.Error()))
				}
				chLoadResult <- err
			}()

			if err := Start(cmd.Context(), cli.testnet, cli.infp); err != nil {
				return err
			}

			if err := Wait(cmd.Context(), cli.testnet, 5); err != nil { // allow some txs to go through
				return err
			}

			if cli.testnet.HasPerturbations() {
				if err := Perturb(cmd.Context(), cli.testnet, cli.infp); err != nil {
					return err
				}
				if err := Wait(cmd.Context(), cli.testnet, 5); err != nil { // allow some txs to go through
					return err
				}
			}

			if cli.testnet.Evidence > 0 {
				if err := InjectEvidence(ctx, r, cli.testnet, cli.testnet.Evidence); err != nil {
					return err
				}
				if err := Wait(cmd.Context(), cli.testnet, 5); err != nil { // ensure chain progress
					return err
				}
			}

			loadCancel()
			if err := <-chLoadResult; err != nil {
				return err
			}
			if err := Wait(cmd.Context(), cli.testnet, 5); err != nil { // wait for network to settle before tests
				return err
			}
			if err := Test(cli.testnet, cli.infp.GetInfrastructureData()); err != nil {
				return err
			}
			if !cli.preserve {
				if err := Cleanup(cli.testnet); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cli.root.PersistentFlags().StringP("file", "f", "", "Testnet TOML manifest")
	_ = cli.root.MarkPersistentFlagRequired("file")

	cli.root.PersistentFlags().StringP("testnet-dir", "d", "", "Set the directory for the testnet files generated during setup")

	cli.root.PersistentFlags().StringP("infrastructure-type", "", "docker", "Backing infrastructure used to run the testnet. Either 'digital-ocean' or 'docker'")

	cli.root.PersistentFlags().StringP("infrastructure-data", "", "", "path to the json file containing the infrastructure data. Only used if the 'infrastructure-type' is set to a value other than 'docker'")

	cli.root.Flags().BoolVarP(&cli.preserve, "preserve", "p", false,
		"Preserves the running of the test net after tests are completed")

	cli.root.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Generates the testnet directory and configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Setup(cli.testnet, cli.infp)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts the testnet, waiting for nodes to become available",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := os.Stat(cli.testnet.Dir)
			if errors.Is(err, fs.ErrNotExist) {
				err = Setup(cli.testnet, cli.infp)
			}
			if err != nil {
				return err
			}
			return Start(cmd.Context(), cli.testnet, cli.infp)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "perturb",
		Short: "Perturbs the testnet, e.g. by restarting or disconnecting nodes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return Perturb(cmd.Context(), cli.testnet, cli.infp)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "wait",
		Short: "Waits for a few blocks to be produced and all nodes to catch up",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return Wait(cmd.Context(), cli.testnet, 5)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stops the testnet",
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.Info("Stopping testnet")
			return cli.infp.StopTestnet(context.Background())
		},
	})

	loadCmd := &cobra.Command{
		Use:   "load",
		Short: "Generates transaction load until the command is canceled.",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			useInternalIP, err := cmd.Flags().GetBool("internal-ip")
			if err != nil {
				return err
			}
			if loadRate, err := cmd.Flags().GetInt("rate"); err != nil {
				return err
			} else if loadRate > 0 {
				cli.testnet.LoadTxBatchSize = loadRate
			}
			if loadSize, err := cmd.Flags().GetInt("size"); err != nil {
				return err
			} else if loadSize > 0 {
				cli.testnet.LoadTxSizeBytes = loadSize
			}
			if loadConnections, err := cmd.Flags().GetInt("conn"); err != nil {
				return err
			} else if loadConnections > 0 {
				cli.testnet.LoadTxConnections = loadConnections
			}
			if loadTime, err := cmd.Flags().GetInt("time"); err != nil {
				return err
			} else if loadTime > 0 {
				cli.testnet.LoadMaxSeconds = loadTime
			}
			if loadTargetNodes, err := cmd.Flags().GetStringSlice("nodes"); err != nil {
				return err
			} else if len(loadTargetNodes) > 0 {
				cli.testnet.LoadTargetNodes = loadTargetNodes
			}
			if duplicateTxsToN, err := cmd.Flags().GetInt("duplicate-num-nodes"); err != nil {
				return err
			} else if duplicateTxsToN > 0 {
				cli.testnet.LoadDuplicateTxs = duplicateTxsToN
			}
			if err = cli.testnet.Validate(); err != nil {
				return err
			}

			return Load(context.Background(), cli.testnet, useInternalIP)
		},
	}
	loadCmd.PersistentFlags().IntP("rate", "r", -1,
		"Number of transactions generate each second on all connections). Overwrites manifest option load_tx_batch_size.")
	loadCmd.PersistentFlags().IntP("size", "s", -1,
		"Transaction size in bytes. Overwrites manifest option load_tx_size_bytes.")
	loadCmd.PersistentFlags().IntP("conn", "c", -1,
		"Number of connections to open at each target node simultaneously. Overwrites manifest option load_tx_connections.")
	loadCmd.PersistentFlags().IntP("time", "t", -1,
		"Maximum duration (in seconds) of the load test. Overwrites manifest option load_max_seconds.")
	loadCmd.PersistentFlags().StringSliceP("nodes", "n", nil,
		"Comma-separated list of node names to send load to. Manifest option send_no_load will be ignored.")
	loadCmd.PersistentFlags().BoolP("internal-ip", "i", false,
		"Use nodes' internal IP addresses when sending transaction load. For running from inside a DO private network.")
	loadCmd.PersistentFlags().IntP("duplicate-num-nodes", "", 0,
		"Number of nodes that will receive the same transactions")

	cli.root.AddCommand(loadCmd)

	cli.root.AddCommand(&cobra.Command{
		Use:   "evidence [amount]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Generates and broadcasts evidence to a random node",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount := 1

			if len(args) == 1 {
				amount, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}

			return InjectEvidence(
				cmd.Context(),
				rand.New(rand.NewSource(randomSeed)), //nolint: gosec
				cli.testnet,
				amount,
			)
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Runs test cases against a running testnet",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Test(cli.testnet, cli.infp.GetInfrastructureData())
		},
	})

	monitorCmd := cobra.Command{
		Use:     "monitor",
		Aliases: []string{"mon"},
		Short:   "Manage monitoring services such as Prometheus, Grafana, ElasticSearch, etc.",
		Long: "Manage monitoring services such as Prometheus, Grafana, ElasticSearch, etc.\n" +
			"First run 'setup' to generate a Prometheus config file.",
	}
	monitorStartCmd := cobra.Command{
		Use:     "start",
		Aliases: []string{"up"},
		Short:   "Start monitoring services.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := os.Stat(PrometheusConfigFile)
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("file %s not found", PrometheusConfigFile)
			}
			if err := docker.ExecComposeVerbose(cmd.Context(), "monitoring", "up", "-d"); err != nil {
				return err
			}
			logger.Info("Grafana: http://localhost:3000 ; Prometheus: http://localhost:9090")
			return nil
		},
	}
	monitorStopCmd := cobra.Command{
		Use:     "stop",
		Aliases: []string{"down"},
		Short:   "Stop monitoring services.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := os.Stat(PrometheusConfigFile)
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			logger.Info("Shutting down monitoring services.")
			if err := docker.ExecComposeVerbose(cmd.Context(), "monitoring", "down"); err != nil {
				return err
			}
			// Remove prometheus config only when there is no testnet.
			if _, err := os.Stat(cli.testnet.Dir); errors.Is(err, fs.ErrNotExist) {
				if err := os.RemoveAll(PrometheusConfigFile); err != nil {
					return err
				}
			}
			return nil
		},
	}
	monitorCmd.AddCommand(&monitorStartCmd)
	monitorCmd.AddCommand(&monitorStopCmd)
	cli.root.AddCommand(&monitorCmd)

	cli.root.AddCommand(&cobra.Command{
		Use:     "cleanup",
		Aliases: []string{"clean"},
		Short:   "Removes the testnet directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Alert if monitoring services are still running.
			outBytes, err := docker.ExecComposeOutput(cmd.Context(), "monitoring", "ps", "--services", "--filter", "status=running")
			out := strings.TrimSpace(string(outBytes))
			if err == nil && len(out) != 0 {
				logger.Info("Monitoring services are still running:\n" + out)
			}
			return Cleanup(cli.testnet)
		},
	})

	var splitLogs bool
	logCmd := &cobra.Command{
		Use:   "logs",
		Short: "Shows the testnet logs. Use `--split` to split logs into separate files",
		RunE: func(cmd *cobra.Command, _ []string) error {
			splitLogs, _ = cmd.Flags().GetBool("split")
			if splitLogs {
				for _, node := range cli.testnet.Nodes {
					fmt.Println("Log for", node.Name)
					err := docker.ExecComposeVerbose(context.Background(), cli.testnet.Dir, "logs", node.Name)
					if err != nil {
						return err
					}
				}
				return nil
			}
			return docker.ExecComposeVerbose(context.Background(), cli.testnet.Dir, "logs")
		},
	}
	logCmd.PersistentFlags().BoolVar(&splitLogs, "split", false, "outputs separate logs for each container")
	cli.root.AddCommand(logCmd)

	cli.root.AddCommand(&cobra.Command{
		Use:   "tail",
		Short: "Tails the testnet logs",
		RunE: func(_ *cobra.Command, _ []string) error {
			return docker.ExecComposeVerbose(context.Background(), cli.testnet.Dir, "logs", "--follow")
		},
	})

	cli.root.AddCommand(&cobra.Command{
		Use:   "benchmark",
		Short: "Benchmarks testnet",
		Long: `Benchmarks the following metrics:
	Mean Block Interval
	Standard Deviation
	Min Block Interval
	Max Block Interval
over a 100 block sampling period.

Does not run any perturbations.
		`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := Cleanup(cli.testnet); err != nil {
				return err
			}
			if err := Setup(cli.testnet, cli.infp); err != nil {
				return err
			}

			chLoadResult := make(chan error)
			ctx, loadCancel := context.WithCancel(cmd.Context())
			defer loadCancel()
			go func() {
				err := Load(ctx, cli.testnet, false)
				if err != nil {
					logger.Error(fmt.Sprintf("Transaction load errored: %v", err.Error()))
				}
				chLoadResult <- err
			}()

			if err := Start(cmd.Context(), cli.testnet, cli.infp); err != nil {
				return err
			}

			if err := Wait(cmd.Context(), cli.testnet, 5); err != nil { // allow some txs to go through
				return err
			}

			// we benchmark performance over the next 100 blocks
			if err := Benchmark(cmd.Context(), cli.testnet, 100); err != nil {
				return err
			}

			loadCancel()
			if err := <-chLoadResult; err != nil {
				return err
			}

			return Cleanup(cli.testnet)
		},
	})

	return cli
}

// Run runs the CLI.
func (cli *CLI) Run() {
	if err := cli.root.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
