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

	CreateNode(context.Context, *e2e.Node) error
	StartComet(context.Context, *e2e.Node) error
	KillComet(context.Context, *e2e.Node) error
	TerminateComet(context.Context, *e2e.Node) error
}
