package types

// Config struct for app
type CustomNodeConfig map[string]CustomNode

type CustomNode struct {
	Host string `json:"host"`
	Path string `json:"path"`
}
