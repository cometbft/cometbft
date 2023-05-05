package digitalocean

import (
	"context"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
)

var _ infra.Provider = (*Provider)(nil)

// Provider implements a DigitalOcean-backed infrastructure provider.
type Provider struct {
	Testnet            *e2e.Testnet
	InfrastructureData e2e.InfrastructureData
}

// Noop currently. Setup is performed externally to the e2e test tool.
func (p *Provider) Setup() error {
	return nil
}

func (p Provider) StartNodes(_ context.Context, nodes ...*e2e.Node) error {
	//TODO Not implemented (next PR)
	return nil
}
func (p Provider) StopTestnet(_ context.Context) error {
	//TODO Not implemented (next PR)
	return nil
}
