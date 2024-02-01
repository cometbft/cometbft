package types

// Config struct for app
type Config map[string]CustomNode

type CustomNode struct {
	Host string `json:"host"`
	Path string `json:"path"`
}
