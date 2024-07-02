// Package txhash provides a way to change the hash function used for
// transaction hashing.
//
// The default hash function is SHA256.
// You can change the hash function by calling txhash.Set.
//
// You can also change how a checksum is converted to a string by calling
// txhash.SetFmtHash.
package txhash
