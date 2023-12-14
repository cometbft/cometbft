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

const ymlSystemd = "systemd-action.yml"

func (p Provider) StartNodes(ctx context.Context, nodes ...*e2e.Node) error {
	nodeIPs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}
	if err := p.writePlaybook(ymlSystemd, true); err != nil {
		return err
	}

	return execAnsible(ctx, p.Testnet.Dir, ymlSystemd, nodeIPs)
}
func (p Provider) StopTestnet(ctx context.Context) error {
	nodeIPs := make([]string, len(p.Testnet.Nodes))
	for i, n := range p.Testnet.Nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}

	if err := p.writePlaybook(ymlSystemd, false); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, ymlSystemd, nodeIPs)
}

func (p Provider) writePlaybook(yaml string, starting bool) error {
	playbook := ansibleSystemdBytes(starting)
	//nolint: gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	err := os.WriteFile(filepath.Join(p.Testnet.Dir, yaml), []byte(playbook), 0o644)
	if err != nil {
		return err
	}
	return nil
}

// file as bytes to be written out to disk.
// ansibleStartBytes generates an Ansible playbook to start the network
func ansibleSystemdBytes(starting bool) string {
	startStop := "stopped"
	if starting {
		startStop = "started"
	}
	playbook := fmt.Sprintf(`- name: start/stop testapp
  hosts: all
  gather_facts: yes
  vars:
    ansible_host_key_checking: false

  tasks:
<<<<<<< HEAD
  - name: operate on the systemd-unit
    ansible.builtin.systemd:
=======
`

func ansibleAddTask(playbook, name, contents string) string {
	return playbook + "  - name: " + name + "\n" + contents + "\n"
}

func ansibleAddCopyTask(playbook, name, src, dest string) string {
	copyTask := fmt.Sprintf("    ansible.builtin.copy:\n"+
		"      src: %s\n"+
		"      dest: %s\n",
		src, dest)
	return ansibleAddTask(playbook, name, copyTask)
}

func ansibleAddSystemdTask(playbook string, starting bool) string {
	startStop := "stopped"
	if starting {
		startStop = "started"
	}
	// testappd is the name of the daemon running the node in the ansible scripts in the qa-infra repo.
	contents := fmt.Sprintf(`    ansible.builtin.systemd:
>>>>>>> 61d508ee2 (chore(docs): explain the effect of `timeout_propose` (#1798))
      name: testappd
      state: %s
      enabled: yes`, startStop)
	return playbook
}

// ExecCompose runs a Docker Compose command for a testnet.
func execAnsible(ctx context.Context, dir, playbook string, nodeIPs []string, args ...string) error {
	playbook = filepath.Join(dir, playbook)
	return exec.CommandVerbose(ctx, append(
		[]string{"ansible-playbook", playbook, "-f", "50", "-u", "root", "--inventory", strings.Join(nodeIPs, ",") + ","},
		args...)...)
}
