// Package custom is a crypto library extension on top of the ed25519 implementation. Import this package (with _) to
// override the default ed25519 implementation with a custom one.
package custom

import (
	"github.com/cometbft/cometbft/crypto/custom/indicator"
)

// This init has to run _before_ the init function in ed25519 so it can indicate if the default privkey/pubkey names
// should be registered or not for amino-encoding. This is ensured by the Golang compiler as it loads init functions
// across packages in lexicographical order and custom < ed25519.
func init() {
	// Indicate to the ed25519 implementation that the keys and functions will be overwritten.
	indicator.SetCustomized()
}
