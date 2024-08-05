// crypto is a customized/convenience cryptography package for supporting
// CometBFT.

// It wraps select functionality of equivalent functions in the
// Go standard library, for easy usage with our libraries.

// Keys:

// All key generation functions return an instance of the PrivKey interface
<<<<<<< HEAD
// which implements methods

//     AssertIsPrivKeyInner()
//     Bytes() []byte
//     Sign(msg []byte) Signature
//     PubKey() PubKey
//     Equals(PrivKey) bool
//     Wrap() PrivKey

// From the above method we can:
// a) Retrieve the public key if needed

//     pubKey := key.PubKey()

// For example:
//     privKey, err := ed25519.GenPrivKey()
//     if err != nil {
// 	...
//     }
//     pubKey := privKey.PubKey()
//     ...
//     // And then you can use the private and public key
//     doSomething(privKey, pubKey)

// We also provide hashing wrappers around algorithms:

// Sha256
//     sum := crypto.Sha256([]byte("This is CometBFT"))
//     fmt.Printf("%x\n", sum)

=======
// which implements methods:
//
//	type PrivKey interface {
//		Bytes() []byte
//		Sign(msg []byte) ([]byte, error)
//		PubKey() PubKey
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
// The resulting public key is an instance of the PubKey interface:
//
//	type PubKey interface {
//		Address() Address
//		Bytes() []byte
//		VerifySignature(msg []byte, sig []byte) bool
//		Type() string
//	}
>>>>>>> bfa7aa85d (feat(crypto)!: remove PubKey#Equals and PrivKey#Equals (#3606))
package crypto

// TODO: Add more docs in here
