package infra

import (
	"context"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

// Provider defines an API for manipulating the infrastructure of a
// specific set of testnet infrastructure.
type Provider interface {

	// Setup generates any necessary configuration for the infrastructure
	// provider during testnet setup.
	Setup() error

	// Starts the nodes passed as parameter. A nodes MUST NOT
	// be started twice before calling StopTestnet
	// If no nodes are passed, start the whole network
	StartNodes(context.Context, ...*e2e.Node) error

	// Stops the whole network
	StopTestnet(context.Context) error
}
