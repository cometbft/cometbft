package infra

import (
	"bytes"
	"os"
	"text/template"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

// GenerateIPZonesTable generates a file with a table mapping IP addresses to geographical zone for latencies.
func GenerateIPZonesTable(nodes []*e2e.Node, zonesPath string, useInternalIP bool) error {
	// Generate file with table mapping IP addresses to geographical zone for latencies.
	zonesTable, err := zonesTableBytes(nodes, useInternalIP)
	if err != nil {
		return err
	}
	//nolint: gosec // G306: Expect WriteFile permissions to be 0600 or less
	err = os.WriteFile(zonesPath, zonesTable, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func zonesTableBytes(nodes []*e2e.Node, useInternalIP bool) ([]byte, error) {
	tmpl, err := template.New("zones").Parse(`Node,IP,Zone
{{- range .Nodes }}
{{- if .Zone }}
{{ .Name }},{{ if $.UseInternalIP }}{{ .InternalIP }}{{ else }}{{ .ExternalIP }}{{ end }},{{ .Zone }}
{{- end }}
{{- end }}`)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Nodes         []*e2e.Node
		UseInternalIP bool
	}{
		Nodes:         nodes,
		UseInternalIP: useInternalIP,
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
