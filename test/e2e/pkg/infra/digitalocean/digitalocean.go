package digitalocean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/exec"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
)

var _ infra.Provider = (*Provider)(nil)

// Provider implements a DigitalOcean-backed infrastructure provider.
type Provider struct {
	infra.ProviderData
}

// Noop currently. Setup is performed externally to the e2e test tool.
func (p *Provider) Setup() error {
	return nil
}

const (
	ymlSystemd = "systemd-action.yml"
	ymlConnect = "connect-action.yml"
)

func (p Provider) StartNodes(ctx context.Context, nodes ...*e2e.Node) error {
	nodeIPs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}
	playbook := ansibleSystemdBytes(true)
	if err := p.writePlaybook(ymlSystemd, playbook); err != nil {
		return err
	}

	return execAnsible(ctx, p.Testnet.Dir, ymlSystemd, nodeIPs)
}
func (p Provider) StopTestnet(ctx context.Context) error {
	nodeIPs := make([]string, len(p.Testnet.Nodes))
	for i, n := range p.Testnet.Nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}

	playbook := ansibleSystemdBytes(false)
	if err := p.writePlaybook(ymlSystemd, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, ymlSystemd, nodeIPs)
}
func (p Provider) Connect(ctx context.Context, _ string, ip string) error {
	playbook := ansiblePerturbConnectionBytes(false)
	if err := p.writePlaybook(ymlConnect, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, ymlConnect, []string{ip})
}
func (p Provider) Disconnect(ctx context.Context, _ string, ip string) error {
	playbook := ansiblePerturbConnectionBytes(true)
	if err := p.writePlaybook(ymlConnect, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, ymlConnect, []string{ip})
}

func (p Provider) CheckUpgraded(ctx context.Context, node *e2e.Node) (string, bool, error) {
	// Upgrade not supported yet by DO provider
	return node.Name, false, nil
}

func (p Provider) writePlaybook(yaml, playbook string) error {
	//nolint: gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	err := os.WriteFile(filepath.Join(p.Testnet.Dir, yaml), []byte(playbook), 0o644)
	if err != nil {
		return err
	}
	return nil
}

const basePlaybook = `- name: start/stop testapp
  hosts: all
  gather_facts: yes
  vars:
    ansible_host_key_checking: false

  tasks:
`

func ansibleAddTask(playbook, name, contents string) string {
	return playbook + "  - name: " + name + "\n" + contents
}

func ansibleAddSystemdTask(playbook string, starting bool) string {
	startStop := "stopped"
	if starting {
		startStop = "started"
	}
	contents := fmt.Sprintf(`    ansible.builtin.systemd:
      name: testappd
      state: %s
      enabled: yes`, startStop)

	return ansibleAddTask(playbook, "operate on the systemd-unit", contents)
}

func ansibleAddShellTasks(playbook, name string, shells ...string) string {
	for _, shell := range shells {
		contents := fmt.Sprintf("    shell: \"%s\"\n", shell)
		playbook = ansibleAddTask(playbook, name, contents)
	}
	return playbook
}

// file as bytes to be written out to disk.
// ansibleStartBytes generates an Ansible playbook to start the network
func ansibleSystemdBytes(starting bool) string {
	return ansibleAddSystemdTask(basePlaybook, starting)
}

func ansiblePerturbConnectionBytes(disconnect bool) string {
	disconnecting := "disconnect"
	op := "-A"
	if disconnect {
		disconnecting = "reconnect"
		op = "-D"
	}
	playbook := basePlaybook
	for _, dir := range []string{"INPUT", "OUTPUT"} {
		playbook = ansibleAddShellTasks(playbook, disconnecting+" node",
			"iptables %s %s -p tcp --destination-port 26656 -j REJECT --reject-with tcp-reset", op, dir)
	}
	return playbook
}

// ExecCompose runs a Docker Compose command for a testnet.
func execAnsible(ctx context.Context, dir, playbook string, nodeIPs []string, args ...string) error {
	playbook = filepath.Join(dir, playbook)
	return exec.CommandVerbose(ctx, append(
		[]string{"ansible-playbook", playbook, "-f", "50", "-u", "root", "--inventory", strings.Join(nodeIPs, ",") + ","},
		args...)...)
}
