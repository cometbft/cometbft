// Package txhash provides a way to change the hash function used for
// transaction hashing.
//
// The default hash function is SHA256.
// You can change the hash function by calling txhash.Set.
//
// You can also change how a checksum is converted to a string by calling
// txhash.SetFmtHash.
//
// WARNING: Currently, CometBFT does NOT support changing the hash function
// after a chain has been started. A chain is expected to agree on the hash
// function to be used before genesis time, and stick to it. If a chain changes
// the hash function after it was started (e.g., as part of a coordinated
// upgrade) many things will stop working (block validation, block sync, light
// clients, etc).
package txhash
