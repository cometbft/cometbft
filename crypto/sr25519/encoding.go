package sr25519

import cmtjson "github.com/cometbft/cometbft/libs/json"

const (
	// Deprecated: This key type will be removed, do not use.
	PrivKeyName = "tendermint/PrivKeySr25519"

	// Deprecated: This key type will be removed, do not use.
	PubKeyName = "tendermint/PubKeySr25519"
)

func init() {
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}
