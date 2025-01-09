package main

import (
	"context"
	"fmt"
	"os"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/exec"
)

// Test runs test cases under tests.
func Test(testnet *e2e.Testnet, ifd *e2e.InfrastructureData) error {
	err := os.Setenv("E2E_MANIFEST", testnet.File)
	if err != nil {
		return err
	}
	err = os.Setenv("E2E_TESTNET_DIR", testnet.Dir)
	if err != nil {
		return err
	}
	if p := ifd.Path; p != "" {
		err = os.Setenv("INFRASTRUCTURE_FILE", p)
		if err != nil {
			return err
		}
	}
	err = os.Setenv("INFRASTRUCTURE_TYPE", ifd.Provider)
	if err != nil {
		return err
	}

	cmd := []string{"go", "test", "-tags", "bls12381,secp256k1eth", "-count", "1"}
	verbose := os.Getenv("VERBOSE")
	if verbose == "1" {
		cmd = append(cmd, "-v")
	}
	cmd = append(cmd, "./tests/...")

	tests := "all tests"
	runTest := os.Getenv("RUN_TEST")
	if len(runTest) != 0 {
		cmd = append(cmd, "-run", runTest)
		tests = fmt.Sprintf("%q", runTest)
	}

	logger.Info(fmt.Sprintf("Running %s in ./tests/...", tests))

	return exec.CommandVerbose(context.Background(), cmd...)
}
