package confix

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/creachadair/tomledit"
)

//go:embed data
var data embed.FS

// LoadLocalConfig loads and parses the TOML document from confix data.
func LoadLocalConfig(fileName string) (*tomledit.Document, error) {
	f, err := data.Open(filepath.Join("data", fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w. This file should have been included in confix", err)
	}
	defer f.Close()

	return tomledit.Parse(f)
}

// LoadConfig loads and parses the TOML document from path.
func LoadConfig(path string) (*tomledit.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer f.Close()

	return tomledit.Parse(f)
}
