package node

import (
	"fmt"
	"os"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStateFromDBOrGenesisDocProviderWithConfig(t *testing.T) {
	cfg := test.ResetTestRoot(t.Name())
	cfg.DBBackend = string(dbm.GoLevelDBBackend)

	_, stateDB, err := initDBs(cfg, config.DefaultDBProvider)
	require.NoErrorf(t, err, "state DB setup: %s", err)

	genDocProviderFunc := func(sha256Checksum []byte) GenesisDocProvider {

		return func() (ChecksummedGenesisDoc, error) {
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

			checksummedGenesisDoc := ChecksummedGenesisDoc{
				GenesisDoc:     genDoc,
				Sha256Checksum: sha256Checksum,
			}

			return checksummedGenesisDoc, nil
		}
	}

	t.Run("NilGenesisChecksum", func(t *testing.T) {
		genDocProvider := genDocProviderFunc(nil)

		_, _, err = LoadStateFromDBOrGenesisDocProviderWithConfig(
			stateDB,
			genDocProvider,
			cfg.Storage.GenesisHash,
			nil,
		)
		require.Error(t, err)
		require.ErrorAs(t, err, &ErrSaveGenesisDocHash{})

		wantErr := "failed to save genesis doc hash to db: value cannot be nil"
		assert.EqualError(t, err, wantErr)
	})

	t.Run("ShorterGenesisChecksum", func(t *testing.T) {
		genDocProvider := genDocProviderFunc([]byte("shorter"))

		_, _, err = LoadStateFromDBOrGenesisDocProviderWithConfig(
			stateDB,
			genDocProvider,
			cfg.Storage.GenesisHash,
			nil,
		)
		assert.Error(t, err)
	})
}
