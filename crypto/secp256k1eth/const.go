package secp256k1eth

const (
	// PrivKeySize defines the length of the PrivKey byte array.
	PrivKeySize = 32
	// PubKeySize, in Ethereum format, is comprised of 65 bytes for two field elements (x and y)
	// and a prefix byte.
	// Only uncompressed public keys are supported, so the the prefix byte is always set
	// to 0x04 to indicate "uncompressed".
	PubKeySize = 65
	// SignatureLength is the size of the ECDSA signature.
	SignatureLength = 65
	// KeyType is the string constant for Ethereum-compatible Secp256k1.
	KeyType = "secp256k1eth"
	// Secp256k1-Eth private key name.
	PrivKeyName = "cometbft/PrivKeySecp256k1eth"
	// Secp256k1-Eth public key name.
	PubKeyName = "cometbft/PubKeySecp256k1eth"
)
