package node

import (
	"fmt"
	"os"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/types"
)

func TestLoadStateFromDBOrGenesisDocProviderWithConfig(t *testing.T) {
	cfg := test.ResetTestRoot(t.Name())

	cfg.DBBackend = string(dbm.GoLevelDBBackend)
	_, stateDB, err := initDBs(cfg, config.DefaultDBProvider)
	if err != nil {
		t.Fatalf("state DB setup: %s", err)
	}

	genDocProvider := func() (ChecksummedGenesisDoc, error) {
		genDocJSON, err := os.ReadFile(cfg.GenesisFile())
		if err != nil {
			formatStr := "reading genesis file: %s"
			return ChecksummedGenesisDoc{}, fmt.Errorf(formatStr, err)
		}

		genDoc, err := types.GenesisDocFromJSON(genDocJSON)
		if err != nil {
			formatStr := "parsing genesis file: %s"
			return ChecksummedGenesisDoc{}, fmt.Errorf(formatStr, err)
		}

		return ChecksummedGenesisDoc{GenesisDoc: genDoc}, nil
	}

	_, _, err = LoadStateFromDBOrGenesisDocProviderWithConfig(
		stateDB,
		genDocProvider,
		cfg.Storage.GenesisHash,
		nil,
	)
	if err != nil {
		t.Errorf("test failed with error: %s", err)
	}
}
