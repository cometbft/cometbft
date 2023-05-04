package digitalocean

import (
	"context"
	"fmt"
	"net"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
	e2essh "github.com/cometbft/cometbft/test/e2e/pkg/ssh"
	"golang.org/x/crypto/ssh"
)

const (
	sshPort     = 22
	testappName = "testappd"
)

var _ infra.Provider = (*Provider)(nil)

// Provider implements a DigitalOcean-backed infrastructure provider.
type Provider struct {
	Testnet            *e2e.Testnet
	InfrastructureData e2e.InfrastructureData
	SSHConfig          *ssh.ClientConfig
}

// Noop currently. Setup is performed externally to the e2e test tool.
func (p *Provider) Setup() error {
	return nil
}

// Noop currently. Node creation is currently performed externally to the e2e test tool.
func (p Provider) CreateNode(ctx context.Context, n *e2e.Node) error {
	return nil
}
func (p Provider) StartComet(ctx context.Context, n *e2e.Node) error {
	return e2essh.Exec(p.SSHConfig, fmt.Sprintf("%s:%d", n.ExternalIP, sshPort), fmt.Sprintf("systemctl start %s", testappName))
}
func (p Provider) TerminateComet(ctx context.Context, n *e2e.Node) error {
	return e2essh.Exec(p.SSHConfig, fmt.Sprintf("%s:%d", n.ExternalIP, sshPort), fmt.Sprintf("systemctl -s SIGTERM %s", testappName))
}
func (p Provider) KillComet(ctx context.Context, n *e2e.Node) error {
	return e2essh.Exec(p.SSHConfig, fmt.Sprintf("%s:%d", n.ExternalIP, sshPort), fmt.Sprintf("systemctl -s SIGKILL %s", testappName))
}
func (p Provider) GetReachableIP(ctx context.Context, n *e2e.Node) net.IP {
	return n.InternalIP
}
