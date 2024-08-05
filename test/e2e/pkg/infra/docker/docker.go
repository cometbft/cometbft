package docker

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"text/template"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/exec"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
)

var _ infra.Provider = (*Provider)(nil)

// Provider implements a docker-compose backed infrastructure provider.
type Provider struct {
	infra.ProviderData
}

// Setup generates the docker-compose file and write it to disk, erroring if
// any of these operations fail.
func (p *Provider) Setup() error {
	compose, err := dockerComposeBytes(p.Testnet)
	if err != nil {
		return err
	}
	//nolint: gosec
	// G306: Expect WriteFile permissions to be 0600 or less
	err = os.WriteFile(filepath.Join(p.Testnet.Dir, "docker-compose.yml"), compose, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func (p Provider) StartNodes(ctx context.Context, nodes ...*e2e.Node) error {
	nodeNames := make([]string, len(nodes))
	for i, n := range nodes {
		nodeNames[i] = n.Name
	}
	return ExecCompose(ctx, p.Testnet.Dir, append([]string{"up", "-d"}, nodeNames...)...)
}

func (p Provider) StopTestnet(ctx context.Context) error {
	return ExecCompose(ctx, p.Testnet.Dir, "down")
}

// dockerComposeBytes generates a Docker Compose config file for a testnet and returns the
// file as bytes to be written out to disk.
func dockerComposeBytes(testnet *e2e.Testnet) ([]byte, error) {
	// Must use version 2 Docker Compose format, to support IPv6.
	tmpl, err := template.New("docker-compose").Parse(`version: '2.4'
networks:
  {{ .Name }}:
    labels:
      e2e: true
    driver: bridge
{{- if .IPv6 }}
    enable_ipv6: true
{{- end }}
    ipam:
      driver: default
      config:
      - subnet: {{ .IP }}

services:
{{- range .Nodes }}
  {{ .Name }}:
    labels:
      e2e: true
    container_name: {{ .Name }}
    image: {{ .Version }}
{{- if or (eq .ABCIProtocol "builtin") (eq .ABCIProtocol "builtin_connsync") }}
    entrypoint: /usr/bin/entrypoint-builtin
{{- end }}
    init: true
    ports:
    - 26656
    - {{ if .ProxyPort }}{{ .ProxyPort }}:{{ end }}26657
{{- if .PrometheusProxyPort }}
    - {{ .PrometheusProxyPort }}:26660
{{- end }}
    - 6060
    volumes:
    - ./{{ .Name }}:/cometbft
    - ./{{ .Name }}:/tendermint
    networks:
      {{ $.Name }}:
        ipv{{ if $.IPv6 }}6{{ else }}4{{ end}}_address: {{ .InternalIP }}
{{- if ne .Version $.UpgradeVersion}}

  {{ .Name }}_u:
    labels:
      e2e: true
    container_name: {{ .Name }}_u
    image: {{ $.UpgradeVersion }}
{{- if or (eq .ABCIProtocol "builtin") (eq .ABCIProtocol "builtin_connsync") }}
    entrypoint: /usr/bin/entrypoint-builtin
{{- end }}
    init: true
    ports:
    - 26656
    - {{ if .ProxyPort }}{{ .ProxyPort }}:{{ end }}26657
{{- if .PrometheusProxyPort }}
    - {{ .PrometheusProxyPort }}:26660
{{- end }}
    - 6060
    volumes:
    - ./{{ .Name }}:/cometbft
    - ./{{ .Name }}:/tendermint
    networks:
      {{ $.Name }}:
        ipv{{ if $.IPv6 }}6{{ else }}4{{ end}}_address: {{ .InternalIP }}
{{- end }}

{{end}}`)
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
func ExecCompose(ctx context.Context, dir string, args ...string) error {
	return exec.Command(ctx, append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

// ExecCompose runs a Docker Compose command for a testnet and returns the command's output.
func ExecComposeOutput(ctx context.Context, dir string, args ...string) ([]byte, error) {
	return exec.CommandOutput(ctx, append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

// ExecComposeVerbose runs a Docker Compose command for a testnet and displays its output.
func ExecComposeVerbose(ctx context.Context, dir string, args ...string) error {
	return exec.CommandVerbose(ctx, append(
		[]string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml")},
		args...)...)
}

// Exec runs a Docker command.
func Exec(ctx context.Context, args ...string) error {
	return exec.Command(ctx, append([]string{"docker"}, args...)...)
}
