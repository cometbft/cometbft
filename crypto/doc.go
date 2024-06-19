// crypto is a customized/convenience cryptography package for CometBFT.
//
// It wraps select functionality of equivalent functions in the
// Go standard library, for easy usage with our libraries.
//
// Keys:
//
// All key generation functions return an instance of the PrivKey interface
// which implements methods:
//
//	type PrivKey interface {
//		Bytes() []byte
//		Sign(msg []byte) ([]byte, error)
//		PubKey() PubKey
//		Equals(other PrivKey) bool
//		Type() string
//	}
//
// From the above method we can retrieve the public key if needed:
//
//	privKey, err := ed25519.GenPrivKey()
//	if err != nil {
//		panic(err)
//	}
//	pubKey := privKey.PubKey()
//
// TODO: Add more docs in here
package crypto
