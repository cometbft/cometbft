package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	cmtos "github.com/cometbft/cometbft/v2/internal/os"
	cmtjson "github.com/cometbft/cometbft/v2/libs/json"
	"github.com/cometbft/cometbft/v2/privval"
)

// ShowValidatorCmd adds capabilities for showing the validator info.
var ShowValidatorCmd = &cobra.Command{
	Use:     "show-validator",
	Aliases: []string{"show_validator"},
	Short:   "Show this node's validator info",
	RunE:    showValidator,
}

func showValidator(*cobra.Command, []string) error {
	keyFilePath := config.PrivValidatorKeyFile()
	if !cmtos.FileExists(keyFilePath) {
		return fmt.Errorf("private validator file %s does not exist", keyFilePath)
	}

	pv := privval.LoadFilePV(keyFilePath, config.PrivValidatorStateFile())

	pubKey, err := pv.GetPubKey()
	if err != nil {
		return fmt.Errorf("can't get pubkey: %w", err)
	}

	bz, err := cmtjson.Marshal(pubKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private validator pubkey: %w", err)
	}

	fmt.Println(string(bz))
	return nil
}
