package crypto_test

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto"
)

func ExampleSha256() {
	sum := crypto.Sha256([]byte("This is CometBFT"))
	fmt.Printf("%x\n", sum)
	// Output:
	// f91afb642f3d1c87c17eb01aae5cb65c242dfdbe7cf1066cc260f4ce5d33b94e
}
