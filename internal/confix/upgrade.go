package confix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/creachadair/atomicfile"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/transform"
	"github.com/spf13/viper"

	"github.com/cometbft/cometbft/v2/config"
)

// Upgrade reads the configuration file at configPath and applies any
// transformations necessary to Upgrade it to the current version. If this
// succeeds, the transformed output is written to outputPath. As a special
// case, if outputPath == "" the output is written to stdout.
//
// It is safe if outputPath == inputPath. If a regular file outputPath already
// exists, it is overwritten. In case of error, the output is not written.
//
// Upgrade is a convenience wrapper for calls to LoadConfig, ApplyFixes, and
// CheckValid. If the caller requires more control over the behavior of the
// Upgrade, call those functions directly.
func Upgrade(ctx context.Context, plan transform.Plan, configPath, outputPath string, skipValidate bool) error {
	if configPath == "" {
		return errors.New("empty input configuration path")
	}

	doc, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// transforms doc and reports whether it succeeded.
	if err := plan.Apply(ctx, doc); err != nil {
		return fmt.Errorf("updating %q: %w", configPath, err)
	}

	var buf bytes.Buffer
	if err := tomledit.Format(&buf, doc); err != nil {
		return fmt.Errorf("formatting config: %w", err)
	}

	// allow to skip validation
	if !skipValidate {
		// verify that file is valid after applying fixes
		if err := CheckValid(buf.Bytes()); err != nil {
			return fmt.Errorf("updated config is invalid: %w", err)
		}
	}

	if outputPath == "" {
		_, err = os.Stdout.Write(buf.Bytes())
	} else {
		err = atomicfile.WriteData(outputPath, buf.Bytes(), 0o600)
	}

	return err
}

// CheckValid checks whether the specified config appears to be a valid CometBFT config file.
// It tries to unmarshal the config into both the server and client config structs.
func CheckValid(data []byte) error {
	v := viper.New()
	v.SetConfigType("toml")

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal as config: %w", err)
	}

	return nil
}
