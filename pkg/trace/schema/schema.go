package schema

import (
	"strings"

	"github.com/cometbft/cometbft/config"
)

func init() {
	config.DefaultTracingTables = strings.Join(AllTables(), ",")
}

func AllTables() []string {
	tables := []string{}
	tables = append(tables, MempoolTables()...)
	tables = append(tables, ConsensusTables()...)
	tables = append(tables, P2PTables()...)
	tables = append(tables, ABCITable)
	return tables
}

const (
	Broadcast = "broadcast"
)

type TransferType int

const (
	Download TransferType = iota
	Upload
)

func (t TransferType) String() string {
	switch t {
	case Download:
		return "download"
	case Upload:
		return "upload"
	default:
		return "unknown"
	}
}
