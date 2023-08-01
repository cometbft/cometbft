package config

import (
	"errors"
	"fmt"
)

var (
	// ErrNotEnoughRpcServers is returned if the number of rpc servers is less than two
	ErrNotEnoughRpcServers = errors.New("at least two rpc_servers entries is required")

	// ErrEmptyRpcServerEntry is returned when an empty string corresponding to RPC entry is found during validate basic
	ErrEmptyRpcServerEntry = errors.New("found empty rpc_servers entry")

	// ErrInsufficientDiscoveryTime is returned when snapshot discovery time is less than 5 seconds
	ErrInsufficientDiscoveryTime = errors.New("discovery time must be at least five seconds")

	// ErrInsufficientChunkRequestTimeout is returned when timeout for re-requesting a chunk is less than 5 seconds
	ErrInsufficientChunkRequestTimeout = errors.New("chunk_request_timeout must be at least 5 seconds")
)

// ErrInSection is returned if validate basic does not pass for any underlying config service.
type ErrInSection struct {
	Err     error
	Section string
}

func (e ErrInSection) Error() string {
	return fmt.Sprintf("error in [%s] section: %s", e.Section, e.Err.Error())
}

func (e ErrInSection) Unwrap() error {
	return e.Err
}
