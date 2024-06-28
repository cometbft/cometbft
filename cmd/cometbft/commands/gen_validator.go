package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/crypto/sr25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
)

var keyType string

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
	GenValidatorCmd.Flags().StringVarP(&keyType, "key-type", "k", ed25519.KeyType, "private key type")
}

func genValidator(*cobra.Command, []string) error {
	var pk crypto.PrivKey
	switch keyType {
	case secp256k1.KeyType:
		pk = secp256k1.GenPrivKey()
	case sr25519.KeyType:
		pk = sr25519.GenPrivKey()
	case bls12381.KeyType:
		var err error
		pk, err = bls12381.GenPrivKey()
		if err != nil {
			return fmt.Errorf("failed to generate BLS key: %w", err)
		}
	default:
		pk = ed25519.GenPrivKey()
	}
	pv := privval.NewFilePV(pk, "", "")
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		return fmt.Errorf("failed to marshal private validator: %w", err)
	}
	fmt.Printf(`%v
`, string(jsbz))
	return nil
}
