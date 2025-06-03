package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/v2/crypto/ed25519"
	kt "github.com/cometbft/cometbft/v2/internal/keytypes"
	cmtjson "github.com/cometbft/cometbft/v2/libs/json"
	"github.com/cometbft/cometbft/v2/privval"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.
var GenValidatorCmd = &cobra.Command{
	Use:     "gen-validator",
	Aliases: []string{"gen_validator"},
	Short:   "Generate new validator keypair",
	Long:    `Generate new validator keypair using an optional key-type (default: "ed25519").`,
	RunE:    genValidator,
}

func init() {
	GenValidatorCmd.Flags().StringVarP(&keyType, "key-type", "k", ed25519.KeyType, fmt.Sprintf("private key type (one of %s)", kt.SupportedKeyTypesStr()))
}

func genValidator(*cobra.Command, []string) error {
	pv, err := privval.GenFilePV("", "", genPrivKeyFromFlag)
	if err != nil {
		return fmt.Errorf("cannot generate file pv: %w", err)
	}
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		return fmt.Errorf("failed to marshal private validator: %w", err)
	}
	fmt.Printf(`%v
`, string(jsbz))
	return nil
}
