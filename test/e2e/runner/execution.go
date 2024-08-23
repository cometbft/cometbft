package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra/docker"
)

func Save(testnet *e2e.Testnet) error {
	logger.Info("saving execution", "msg", "saving e2e network execution information")
	// Fetch and save the execution logs
	now := time.Now()
	timestamp := now.Format("20060102_150405")
	executionFolder := filepath.Join("networks_executions", testnet.Name, timestamp)
	if err := os.MkdirAll(executionFolder, 0o755); err != nil {
		logger.Error("error saving execution", "msg", "error creating executions folder", "err", err.Error())
		return err
	}
	for _, node := range testnet.Nodes {
		// Pause the container to capture the logs
		_, err := docker.ExecComposeOutput(context.Background(), testnet.Dir, "pause", node.Name)
		if err != nil {
			logger.Error("error saving execution", "msg", "error pausing container", "node", node.Name, "err", err.Error())
			return err
		}

		// Get the logs from the Docker container
		data, err := docker.ExecComposeOutput(context.Background(), testnet.Dir, "logs", node.Name)
		if err != nil {
			logger.Error("error saving execution", "msg", "error getting logs from container", "node", node.Name, "err", err.Error())
			return err
		}

		// Create a file to write the processed lines
		nodeFolder := filepath.Join(executionFolder, node.Name)
		if err := os.MkdirAll(nodeFolder, 0o755); err != nil {
			logger.Error("error saving execution", "msg", "error creating node folder", "err", err.Error())
			return err
		}

		logFile := filepath.Join(nodeFolder, "docker.log")
		outputFile, err := os.Create(logFile)
		if err != nil {
			logger.Error("error saving execution", "msg", "error creating log file", "file", logFile, "err", err.Error())
			return err
		}
		defer outputFile.Close()

		// Create a buffered writer for efficient writing
		writer := bufio.NewWriter(outputFile)

		// Create a new Scanner to read the data line by line
		scanner := bufio.NewScanner(bytes.NewReader(data))

		// Iterate over each line
		for scanner.Scan() {
			// Get the current line
			line := scanner.Text()
			// Split the log line by the first occurrence of '|'
			parts := strings.SplitN(line, "|", 2)
			// Check if the split was successful and there are at least two parts
			if len(parts) == 2 {
				strippedLine := strings.TrimSpace(parts[1])
				// Write the stripped line to the file
				_, err := writer.WriteString(strippedLine + "\n")
				if err != nil {
					logger.Error("error saving execution", "msg", "error writing to log file", "file", logFile, "err", err.Error())
					return err
				}
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Error("error saving execution", "msg", "error scanning log file", "file", logFile, "err", err.Error())
			return err
		}

		err = writer.Flush()
		if err != nil {
			logger.Error("error saving execution", "msg", "error flushing log file", "file", logFile, "err", err.Error())
			return err
		}

		// Save manifest file
		if err := copyFile(testnet.File, executionFolder); err != nil {
			logger.Error("error saving execution", "msg", "error copying manifest file", "file", testnet.File, "err", err.Error())
			return err
		}
	}

	logger.Info("saved execution", "msg", "finished saving execution information", "path", executionFolder)

	return nil
}

// copyFile copies a file from a source to a destination location.
func copyFile(source string, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	manifestFile := filepath.Join(dest, "manifest.toml")
	destFile, err := os.Create(manifestFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the content from source file to destination file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
