package infra

import (
	"bytes"
	"html/template"
	"os"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

// Generate file with table mapping IP addresses to geographical zone for latencies.
func GenerateIPZonesTable(nodes []*e2e.Node, zonesPath string) error {
	// Generate file with table mapping IP addresses to geographical zone for latencies.
	zonesTable, err := zonesTableBytes(nodes)
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

func zonesTableBytes(nodes []*e2e.Node) ([]byte, error) {
	tmpl, err := template.New("zones").Parse(`Node,IP,Zone
{{- range . }}
{{- if .Zone }}
{{ .Name }},{{ .InternalIP }},{{ .Zone }}
{{- end }}
{{- end }}`)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nodes)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
