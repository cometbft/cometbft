package main

import (
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
)

// execute executes a shell command.
func exec(args ...string) error {
	_, err := execOutput(args...)
	return err
}

func execOutput(args ...string) ([]byte, error) {
	//nolint:gosec // G204: Subprocess launched with a potential tainted input or cmd arguments
	cmd := osexec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	switch err := err.(type) {
	case nil:
		return out, nil
	case *osexec.ExitError:
		return nil, fmt.Errorf("failed to run %q:\n%v", args, string(out))
	default:
		return nil, err
	}
}

// execVerbose executes a shell command while displaying its output.
func execVerbose(args ...string) error {
	//nolint:gosec // G204: Subprocess launched with a potential tainted input or cmd arguments
	cmd := osexec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// execCompose runs a Docker Compose command for a testnet.
func execCompose(dir string, args ...string) error {
	return exec(append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

func execComposeOutput(dir string, args ...string) ([]byte, error) {
	return execOutput(append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

// execComposeVerbose runs a Docker Compose command for a testnet and displays its output.
func execComposeVerbose(dir string, args ...string) error {
	return execVerbose(append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

// execDocker runs a Docker command.
func execDocker(args ...string) error {
	return exec(append([]string{"docker"}, args...)...)
}
