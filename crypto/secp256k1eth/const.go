package secp256k1eth

const (
	// PrivKeyName is the amino route for the private key.
	PrivKeyName = "tendermint/PrivKeySecp256k1eth"
	// PubKeyName is the amino route for the public key.
	PubKeyName = "tendermint/PubKeySecp256k1eth"

	// KeyType is the string identifier for the Ethereum-compatible secp256k1
	// algorithm.
	KeyType = "secp256k1eth"
	// PrivKeySize is the size, in bytes, of a private key (a 32-byte scalar).
	PrivKeySize = 32
	// PubKeySize is the compressed SEC1 public key length (1 parity byte + 32
	// byte X coordinate).
	PubKeySize = 33
	// SignatureSize is the go-ethereum signature length: [R || S || V].
	SignatureSize = 65
)
