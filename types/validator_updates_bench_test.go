package types

import (
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

// BenchmarkValidatorSetUpdateBatches includes the copy made by state.updateState
// before applying an ABCI validator update batch to a 200-validator set.
func BenchmarkValidatorSetUpdateBatches(b *testing.B) {
	const validators = 200
	initial := make([]*Validator, validators)
	for i := range initial {
		initial[i] = NewValidator(ed25519.GenPrivKeyFromSecret([]byte(fmt.Sprintf("existing-%d", i))).PubKey(), int64(1000+i*13%400))
	}
	base := NewValidatorSet(initial)
	for _, kind := range []string{"updates", "additions", "removals"} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%s/N=%d", kind, n), func(b *testing.B) {
				changes := make([]*Validator, n)
				for i := range changes {
					if kind == "additions" {
						changes[i] = NewValidator(ed25519.GenPrivKeyFromSecret([]byte(fmt.Sprintf("added-%d", i))).PubKey(), int64(500+i*73%1000))
					} else {
						changes[i] = initial[(i*13+17)%validators].Copy()
						changes[i].VotingPower = int64(500 + i*73%1000)
						if kind == "removals" {
							changes[i].VotingPower = 0
						}
					}
				}
				b.ReportAllocs()
				for b.Loop() {
					next := base.Copy()
					if err := next.UpdateWithChangeSet(changes); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
