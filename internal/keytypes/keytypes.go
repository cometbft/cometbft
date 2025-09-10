package keytypes

import (
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
)

var keyTypes map[string]func() (crypto.PrivKey, error)

func init() {
	keyTypes = map[string]func() (crypto.PrivKey, error){
		ed25519.KeyType: func() (crypto.PrivKey, error) { //nolint: unparam
			return ed25519.GenPrivKey(), nil
		},
		secp256k1.KeyType: func() (crypto.PrivKey, error) { //nolint: unparam
			return secp256k1.GenPrivKey(), nil
		},
	}
}

func GenPrivKey(keyType string) (crypto.PrivKey, error) {
	genF, ok := keyTypes[keyType]
	if !ok {
		return nil, fmt.Errorf("unsupported key type: %q", keyType)
	}
	return genF()
}

func SupportedKeyTypesStr() string {
	keyTypesSlice := make([]string, 0, len(keyTypes))
	for k := range keyTypes {
		keyTypesSlice = append(keyTypesSlice, fmt.Sprintf("%q", k))
	}
	return strings.Join(keyTypesSlice, ", ")
}

func ListSupportedKeyTypes() []string {
	keyTypesSlice := make([]string, 0, len(keyTypes))
	for k := range keyTypes {
		keyTypesSlice = append(keyTypesSlice, k)
	}
	return keyTypesSlice
}
