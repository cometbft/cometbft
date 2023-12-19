package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	v2 "github.com/cometbft/cometbft/cmd/cometbft/commands/migrate_db/v2"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"
)

var migrateDBTargetVersion uint

func init() {
	MigrateDBCmd.Flags().UintVar(&migrateDBTargetVersion, "target_version", 2, "version to migrate to")
}

var MigrateDBCmd = &cobra.Command{
	Use:     "migrate-db",
	Aliases: []string{"migrate_db"},
	Short:   "Migrate the database to a new version",
	Long: `
A database migration is performed when its schema changes in
a breaking way. To use a new version of CometBFT, you
sometimes will need to migrate the data from an old to a new
format. After performing an upgrade, you won't be able to run
an old version of CometBFT.

IMPORTANT: please backup your database before running this
command! You can do so by stopping the node and copying the
"data" directory.

After the migration is complete, it's recommended to run:

cometbft experimental-compact-goleveldb

to compact the database IF you're using goleveldb.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := ParseConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		fmt.Print("Did you backup your data? Y/N: ")

		// Read user input
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		switch strings.ToUpper(strings.TrimSpace(answer)) {
		case "Y":
			// TODO: read the current version from the database
			// https://github.com/cometbft/cometbft/issues/1822
			targetVersion := migrateDBTargetVersion
			fmt.Printf("Migrating databases to version %d...\n", targetVersion)
			if err := migrateDBs(targetVersion, config); err != nil {
				return fmt.Errorf("error migrating databases: %w", err)
			}
		case "N":
			fmt.Println("Please consider backing up your data before proceeding.")
			return nil
		default:
			return fmt.Errorf("invalid response. please enter Y or N")
		}
		return nil
	},
}

func migrateDBs(targetVersion uint, config *cfg.Config) error {
	// blockstore
	blockStoreDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return err
	}
	defer blockStoreDB.Close()
	if err := migrateBlockStoreDB(blockStoreDB, targetVersion); err != nil {
		return fmt.Errorf("blockstore: %w", err)
	}

	// state
	stateDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "state", Config: config})
	if err != nil {
		return err
	}
	defer stateDB.Close()
	if err := migrateStateDB(stateDB, targetVersion); err != nil {
		return fmt.Errorf("state: %w", err)
	}

	// evidence
	evidenceDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "evidence", Config: config})
	if err != nil {
		return err
	}
	defer evidenceDB.Close()
	if err := migrateEvidenceDB(evidenceDB, targetVersion); err != nil {
		return fmt.Errorf("evidence: %w", err)
	}

	lightClientDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "light-client-db", Config: config})
	if err != nil {
		return err
	}
	defer lightClientDB.Close()
	if err := migrateLightClientDB(lightClientDB, targetVersion); err != nil {
		return fmt.Errorf("light client db: %w", err)
	}

	return nil
}

func migrateBlockStoreDB(db dbm.DB, targetVersion uint) error {
	switch targetVersion {
	case 2:
		return v2.MigrateBlockStore(db)
	default:
		return fmt.Errorf("unsupported target version: %d", targetVersion)
	}
}

func migrateStateDB(db dbm.DB, targetVersion uint) error {
	switch targetVersion {
	case 2:
		return v2.MigrateStateDB(db)
	default:
		return fmt.Errorf("unsupported target version: %d", targetVersion)
	}
}

func migrateEvidenceDB(db dbm.DB, targetVersion uint) error {
	switch targetVersion {
	case 2:
		return v2.MigrateEvidenceDB(db)
	default:
		return fmt.Errorf("unsupported target version: %d", targetVersion)
	}
}

func migrateLightClientDB(db dbm.DB, targetVersion uint) error {
	switch targetVersion {
	case 2:
		return v2.MigrateLightClientDB(db)
	default:
		return fmt.Errorf("unsupported target version: %d", targetVersion)
	}
}
