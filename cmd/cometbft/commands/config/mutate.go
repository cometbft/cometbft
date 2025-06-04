package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/v2/internal/confix"
)

// SetCommand returns a CLI command to interactively update an application config value.
func SetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [config] [key] [value]",
		Short: "Set a config value",
		Long:  "Set a config value. The [config] is an optional absolute path to the config file (default: `~/.cometbft/config/config.toml`)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				filename, inputValue string
				key                  []string
			)
			switch len(args) {
			case 2:
				{
					filename = defaultConfigPath(cmd)
					// parse key e.g mempool.size -> [mempool, size]
					key = strings.Split(args[0], ".")
					inputValue = args[1]
				}
			case 3:
				{
					filename, inputValue = args[0], args[2]
					key = strings.Split(args[1], ".")
				}
			default:
				return errors.New("expected 2 or 3 arguments")
			}

			plan := transform.Plan{
				{
					Desc: fmt.Sprintf("update %q=%q in %s", key, inputValue, filename),
					T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
						results := doc.Find(key...)
						if len(results) == 0 {
							return fmt.Errorf("key %q not found", key)
						} else if len(results) > 1 {
							return fmt.Errorf("key %q is ambiguous", key)
						}

						value, err := parser.ParseValue(inputValue)
						if err != nil {
							value = parser.MustValue(`"` + inputValue + `"`)
						}

						if ok := transform.InsertMapping(results[0].Section, &parser.KeyValue{
							Block: results[0].Block,
							Name:  results[0].Name,
							Value: value,
						}, true); !ok {
							return errors.New("failed to set value")
						}

						return nil
					}),
				},
			}

			outputPath := filename
			if FlagStdOut {
				outputPath = ""
			}

			ctx := cmd.Context()
			if FlagVerbose {
				ctx = confix.WithLogWriter(ctx, cmd.ErrOrStderr())
			}

			return confix.Upgrade(ctx, plan, filename, outputPath, FlagSkipValidate)
		},
	}

	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")
	cmd.Flags().BoolVarP(&FlagVerbose, "verbose", "v", false, "log changes to stderr")
	cmd.Flags().BoolVarP(&FlagSkipValidate, "skip-validate", "s", false, "skip configuration validation (allows to mutate unknown configurations)")

	return cmd
}

// GetCommand returns a CLI command to interactively get an application config value.
func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [config] [key]",
		Short: "Get a config value",
		Long:  "Get a config value. The [config] is an optional absolute path to the config file (default: `~/.cometbft/config/config.toml`)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				filename, key string
				keys          []string
			)
			switch len(args) {
			case 1:
				{
					filename = defaultConfigPath(cmd)
					// parse key e.g mempool.size -> [mempool, size]
					key = args[0]
					keys = strings.Split(key, ".")
				}
			case 2:
				{
					filename = args[0]
					key = args[1]
					keys = strings.Split(key, ".")
				}
			default:
				return errors.New("expected 1 or 2 arguments")
			}

			doc, err := confix.LoadConfig(filename)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			results := doc.Find(keys...)
			if len(results) == 0 {
				return fmt.Errorf("key %q not found", key)
			} else if len(results) > 1 {
				return fmt.Errorf("key %q is ambiguous", key)
			}

			fmt.Printf("%s\n", results[0].Value.String())
			return nil
		},
	}

	return cmd
}
