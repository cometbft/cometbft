package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	migrateDBTargetVersion uint
	db_dir                 string
)

func init() {
	MigrateDBCmd.Flags().UintVar(&migrateDBTargetVersion, "target_version", 2, "version to migrate to")
	MigrateDBCmd.Flags().StringVar(&db_dir, "db_dir", "~/.cometbft/data", "path to the database directory")
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
`,
	RunE: func(*cobra.Command, []string) error {
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
			fmt.Printf("Migrating database to version %d...\n", targetVersion)
			if err := migrateDB(targetVersion); err != nil {
				return fmt.Errorf("error migrating database: %w", err)
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

func migrateDB(targetVersion uint) error {
	blockdb := filepath.Join(dbDir, "blockstore.db")
	state := filepath.Join(dbDir, "state.db")
	wal := filepath.Join(dbDir, "cs.wal")
	evidence := filepath.Join(dbDir, "evidence.db")
	txIndex := filepath.Join(dbDir, "tx_index.db")

	return nil
}
