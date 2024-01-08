package digitalocean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// Setup generates the file mapping IPs to zones, used for emulating latencies.
func (p *Provider) Setup() error {
	err := infra.GenerateIPZonesTable(p.Testnet.Nodes, p.IPZonesFilePath(), false)
	if err != nil {
		return err
	}

	return nil
}

var ymlPlaybookSeq int

func getNextPlaybookFilename() string {
	const ymlPlaybookAction = "playbook-action"
	ymlPlaybookSeq++
	return ymlPlaybookAction + strconv.Itoa(ymlPlaybookSeq) + ".yml"
}

func (p Provider) StartNodes(ctx context.Context, nodes ...*e2e.Node) error {
	nodeIPs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}
	playbook := ansibleSystemdBytes(true)
	playbookFile := getNextPlaybookFilename()
	if err := p.writePlaybook(playbookFile, playbook); err != nil {
		return err
	}

	return execAnsible(ctx, p.Testnet.Dir, playbookFile, nodeIPs)
}

// SetLatency prepares and executes the latency-setter script in the given node.
func (p Provider) SetLatency(ctx context.Context, node *e2e.Node) error {
	// Directory in the DigitalOcean node that contains all latency files.
	remoteDir := "/root/cometbft/test/e2e/latency/"

	playbook := "- name: e2e custom playbook\n" +
		"  hosts: all\n" +
		"  tasks:\n"

	// Add task to copy the necessary files to the node.
	playbook = ansibleAddCopyTask(playbook, "copy zones file to node", filepath.Base(p.IPZonesFilePath()), remoteDir)

	// Add task to execute latency-setter script in the node.
	cmd := fmt.Sprintf("%s set %s %s eth0",
		filepath.Join(remoteDir, "latency-setter.py"),
		filepath.Join(remoteDir, filepath.Base(p.IPZonesFilePath())),
		filepath.Join(remoteDir, "aws-latencies.csv"),
	)
	playbook = ansibleAddShellTasks(playbook, "execute latency setter script", cmd)

	// Execute playbook
	playbookFile := getNextPlaybookFilename()
	if err := p.writePlaybook(playbookFile, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, playbookFile, []string{node.ExternalIP.String()})
}

func (p Provider) StopTestnet(ctx context.Context) error {
	nodeIPs := make([]string, len(p.Testnet.Nodes))
	for i, n := range p.Testnet.Nodes {
		nodeIPs[i] = n.ExternalIP.String()
	}

	playbook := ansibleSystemdBytes(false)
	playbookFile := getNextPlaybookFilename()
	if err := p.writePlaybook(playbookFile, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, playbookFile, nodeIPs)
}

func (p Provider) Disconnect(ctx context.Context, _ string, ip string) error {
	playbook := ansiblePerturbConnectionBytes(true)
	playbookFile := getNextPlaybookFilename()
	if err := p.writePlaybook(playbookFile, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, playbookFile, []string{ip})
}

func (p Provider) Reconnect(ctx context.Context, _ string, ip string) error {
	playbook := ansiblePerturbConnectionBytes(false)
	playbookFile := getNextPlaybookFilename()
	if err := p.writePlaybook(playbookFile, playbook); err != nil {
		return err
	}
	return execAnsible(ctx, p.Testnet.Dir, playbookFile, []string{ip})
}

func (p Provider) CheckUpgraded(_ context.Context, node *e2e.Node) (string, bool, error) {
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

const basePlaybook = `- name: e2e custom playbook
  hosts: all
  gather_facts: yes
  vars:
    ansible_host_key_checking: false

  tasks:
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
// ansibleStartBytes generates an Ansible playbook to start the network.
func ansibleSystemdBytes(starting bool) string {
	return ansibleAddSystemdTask(basePlaybook, starting)
}

func ansiblePerturbConnectionBytes(disconnect bool) string {
	disconnecting := "reconnect"
	op := "-D"
	if disconnect {
		disconnecting = "disconnect"
		op = "-A"
	}
	playbook := basePlaybook
	for _, dir := range []string{"INPUT", "OUTPUT"} {
		playbook = ansibleAddShellTasks(playbook, disconnecting+" node",
			fmt.Sprintf("iptables %s %s -p tcp --dport 26656 -j DROP", op, dir))
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
