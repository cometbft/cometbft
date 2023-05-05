package digitalocean

import (
	"bytes"
	"context"
	"html/template"
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

func (p Provider) StartNodes(ctx context.Context, nodes ...*e2e.Node) error {
	nodeNames := make([]string, len(nodes))
	for i, n := range nodes {
		nodeNames[i] = n.Name
	}
	const yml = "start-network.yml"
	if err := p.writePlaybook(yml); err != nil {
		return err
	}

	return execAnsible(ctx, p.Testnet.Dir, yml, "--limit", strings.Join(nodeNames, ","))
}
func (p Provider) StopTestnet(_ context.Context) error {
	//TODO Not implemented (next PR)
	return nil
}

func (p Provider) writePlaybook(yaml string) error {
	playbook, err := ansibleStartBytes(p.Testnet)
	if err != nil {
		return err
	}
	//nolint: gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	err = os.WriteFile(filepath.Join(p.Testnet.Dir, yaml), playbook, 0o644)
	if err != nil {
		return err
	}
	return nil
}

// file as bytes to be written out to disk.
// ansibleStartBytes generates an Ansible playbook to start the network
func ansibleStartBytes(testnet *e2e.Testnet) ([]byte, error) {
	tmpl, err := template.New("ansible-start").Parse(`- name: start testapp
  hosts: validators
  gather_facts: yes
  vars:
    ansible_host_key_checking: false

  tasks:
  - name: start the systemd-unit
    ansible.builtin.systemd:
      name: testappd
      state: started
      enabled: yes`)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, testnet)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExecCompose runs a Docker Compose command for a testnet.
func execAnsible(ctx context.Context, dir, playbook string, args ...string) error {
	playbook = filepath.Join(dir, playbook)
	hostsFile := filepath.Join(dir, "hosts")
	return exec.Command(ctx, append(
		[]string{"ansible-playbook", playbook, "-f", "50", "-u", "root", "-i", hostsFile},
		args...)...)
}
