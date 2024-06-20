package types

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	cmttime "github.com/cometbft/cometbft/types/time"
)

<<<<<<< HEAD
// ValidateTime does a basic time validation ensuring time does not drift too
// much: +/- one year.
// TODO: reduce this to eg 1 day
// NOTE: DO NOT USE in ValidateBasic methods in this package. This function
// can only be used for real time validation, like on proposals and votes
// in the consensus. If consensus is stuck, and rounds increase for more than a day,
// having only a 1-day band here could break things...
// Can't use for validating blocks because we may be syncing years worth of history.
func ValidateTime(t time.Time) error {
	var (
		now     = cmttime.Now()
		oneYear = 8766 * time.Hour
	)
	if t.Before(now.Add(-oneYear)) || t.After(now.Add(oneYear)) {
		return fmt.Errorf("time drifted too much. Expected: -1 < %v < 1 year", now)
=======
const batchVerifyThreshold = 2

func shouldBatchVerify(vals *ValidatorSet, commit *Commit) bool {
	return len(commit.Signatures) >= batchVerifyThreshold &&
		batch.SupportsBatchVerifier(vals.GetProposer().PubKey) &&
		vals.AllKeysHaveSameType()
}

// VerifyCommit verifies +2/3 of the set had signed the given commit.
//
// It checks all the signatures! While it's safe to exit as soon as we have
// 2/3+ signatures, doing so would impact incentivization logic in the ABCI
// application that depends on the LastCommitInfo sent in FinalizeBlock, which
// includes which validators signed. For instance, Gaia incentivizes proposers
// with a bonus for including more than +2/3 of the signatures.
func VerifyCommit(chainID string, vals *ValidatorSet, blockID BlockID,
	height int64, commit *Commit,
) error {
	// run a basic validation of the arguments
	if err := verifyBasicValsAndCommit(vals, commit, height, blockID); err != nil {
		return err
>>>>>>> 35401c321 (fix(types): DO NOT batch verify if vals keys (#3196))
	}
	return nil
}

// ValidateHash returns an error if the hash is not empty, but its
// size != tmhash.Size.
func ValidateHash(h []byte) error {
	if len(h) > 0 && len(h) != tmhash.Size {
		return fmt.Errorf("expected size to be %d bytes, got %d bytes",
			tmhash.Size,
			len(h),
		)
	}
	return nil
}
