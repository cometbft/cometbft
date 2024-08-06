package commands

import (
	"fmt"

	"github.com/spf13/cobra"

<<<<<<< HEAD
=======
	"github.com/cometbft/cometbft/crypto/ed25519"
	kt "github.com/cometbft/cometbft/internal/keytypes"
>>>>>>> bd06fecb6 (feat(privval)!: add flag `key-type` to all relevant CometBFT commands and thread it through the code (#3517))
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.
var GenValidatorCmd = &cobra.Command{
	Use:     "gen-validator",
	Aliases: []string{"gen_validator"},
	Short:   "Generate new validator keypair",
	Run:     genValidator,
}

<<<<<<< HEAD
func genValidator(*cobra.Command, []string) {
	pv := privval.GenFilePV("", "")
=======
func init() {
	GenValidatorCmd.Flags().StringVarP(&keyType, "key-type", "k", ed25519.KeyType, fmt.Sprintf("private key type (one of %s)", kt.SupportedKeyTypesStr()))
}

func genValidator(*cobra.Command, []string) error {
	pv, err := privval.GenFilePV("", "", genPrivKeyFromFlag)
	if err != nil {
		return fmt.Errorf("cannot generate file pv: %w", err)
	}
>>>>>>> bd06fecb6 (feat(privval)!: add flag `key-type` to all relevant CometBFT commands and thread it through the code (#3517))
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		panic(err)
	}
	fmt.Printf(`%v
`, string(jsbz))
}
