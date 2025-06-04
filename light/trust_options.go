package light

import (
	"time"

	"github.com/cometbft/cometbft/v2/crypto/tmhash"
)

// TrustOptions are the trust parameters needed when a new light client
// connects to the network or when an existing light client that has been
// offline for longer than the trusting period connects to the network.
//
// The expectation is the user will get this information from a trusted source
// like a validator, a friend, or a secure website. A more user friendly
// solution with trust tradeoffs is that we establish an https based protocol
// with a default end point that populates this information. Also an on-chain
// registry of roots-of-trust (e.g. on the Cosmos Hub) seems likely in the
// future.
type TrustOptions struct {
	// tp: trusting period.
	//
	// Should be significantly less than the unbonding period (e.g. unbonding
	// period = 3 weeks, trusting period = 2 weeks).
	//
	// More specifically, trusting period + time needed to check headers + time
	// needed to report and punish misbehavior should be less than the unbonding
	// period.
	Period time.Duration

	// Header's Height and Hash must both be provided to force the trusting of a
	// particular header.
	Height int64
	Hash   []byte
}

// ValidateBasic performs basic validation.
func (opts TrustOptions) ValidateBasic() error {
	if opts.Period <= 0 {
		return ErrNegativeOrZeroPeriod
	}
	if opts.Height <= 0 {
		return ErrNegativeOrZeroHeight
	}
	if len(opts.Hash) != tmhash.Size {
		return ErrInvalidHashSize{Expected: tmhash.Size, Actual: len(opts.Hash)}
	}
	return nil
}
