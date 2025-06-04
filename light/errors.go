package light

import (
	"errors"
	"fmt"
	"time"

	cmtbytes "github.com/cometbft/cometbft/v2/libs/bytes"
	cmtmath "github.com/cometbft/cometbft/v2/libs/math"
	"github.com/cometbft/cometbft/v2/light/provider"
	"github.com/cometbft/cometbft/v2/types"
)

var (

	// ErrFailedHeaderCrossReferencing is returned when the detector was not able to cross reference the header
	// with any of the connected witnesses.
	ErrFailedHeaderCrossReferencing = errors.New("all witnesses have either not responded, don't have the " +
		"blocks or sent invalid blocks. You should look to change your witnesses " +
		"or review the light client's logs for more information")
	// ErrLightClientAttack is returned when the light client has detected an attempt
	// to verify a false header and has sent the evidence to either a witness or primary.
	ErrLightClientAttack = errors.New(`attempted attack detected.
Light client received valid conflicting header from witness.
Unable to verify header. Evidence has been sent to both providers.
Check logs for full evidence and trace`)

	// ErrNoWitnesses means that there are not enough witnesses connected to
	// continue running the light client.
	ErrNoWitnesses               = errors.New("no witnesses connected. please reset light client")
	ErrNilOrSinglePrimaryTrace   = errors.New("nil or single block primary trace")
	ErrHeaderHeightAdjacent      = errors.New("headers must be non adjacent in height")
	ErrHeaderHeightNotAdjacent   = errors.New("headers must be adjacent in height")
	ErrNegativeOrZeroPeriod      = errors.New("negative or zero period")
	ErrNegativeHeight            = errors.New("negative height")
	ErrNegativeOrZeroHeight      = errors.New("negative or zero height")
	ErrInvalidBlockTime          = errors.New("expected traceblock to have a lesser time than the target block")
	ErrRemoveStoredBlocksRefused = errors.New("refused to remove the stored light blocks despite hashes mismatch")
	ErrNoHeadersExist            = errors.New("no headers exist")
	ErrNilHeader                 = errors.New("nil header")
	ErrEmptyTrustedStore         = errors.New("trusted store is empty")
)

// ErrOldHeaderExpired means the old (trusted) header has expired according to
// the given trustingPeriod and current time. If so, the light client must be
// reset subjectively.
type ErrOldHeaderExpired struct {
	At  time.Time
	Now time.Time
}

func (e ErrOldHeaderExpired) Error() string {
	return fmt.Sprintf("old header has expired at %v (now: %v)", e.At, e.Now)
}

type ErrTargetBlockHeightLessThanTrusted struct {
	Target  int64
	Trusted int64
}

func (e ErrTargetBlockHeightLessThanTrusted) Error() string {
	return fmt.Sprintf("target block has a height lower than the trusted height (%d < %d)", e.Target, e.Trusted)
}

type ErrHeaderHeightNotMonotonic struct {
	GotHeight int64
	OldHeight int64
}

func (e ErrHeaderHeightNotMonotonic) Error() string {
	return fmt.Sprintf("expected new header height %d to be greater than one of old header %d", e.GotHeight, e.OldHeight)
}

type ErrHeaderTimeNotMonotonic struct {
	GotTime time.Time
	OldTime time.Time
}

func (e ErrHeaderTimeNotMonotonic) Error() string {
	return fmt.Sprintf("expected new header time %v to be after old header time %v", e.GotTime, e.OldTime)
}

type ErrHeaderTimeExceedMaxClockDrift struct {
	Ti    time.Time
	Now   time.Time
	Drift time.Duration
}

func (e ErrHeaderTimeExceedMaxClockDrift) Error() string {
	return fmt.Sprintf("new header has a time from the future %v (now: %v; max clock drift: %v)", e.Ti, e.Now, e.Drift)
}

type ErrUnverifiedHeight struct {
	Height int64
}

func (e ErrUnverifiedHeight) Error() string {
	return fmt.Sprintf("unverified header/valset requested (latest: %d)", e.Height)
}

type ErrInvalidTrustLevel struct {
	Level cmtmath.Fraction
}

func (e ErrInvalidTrustLevel) Error() string {
	return fmt.Sprintf("trustLevel must be within [1/3, 1], given %v", e.Level)
}

type ErrValidatorsMismatch struct {
	HeaderHash     cmtbytes.HexBytes
	ValidatorsHash cmtbytes.HexBytes
	Height         int64
}

func (e ErrValidatorsMismatch) Error() string {
	return fmt.Sprintf("expected new header validators (%X) to match those that were supplied (%X) at height %d", e.HeaderHash, e.ValidatorsHash, e.Height)
}

type ErrValidatorHashMismatch struct {
	TrustedHash   cmtbytes.HexBytes
	ValidatorHash cmtbytes.HexBytes
}

func (e ErrValidatorHashMismatch) Error() string {
	return fmt.Sprintf("expected old header next validators (%X) to match those from new header (%X)", e.TrustedHash, e.ValidatorHash)
}

type ErrBlockHashMismatch struct {
	TraceBlockHash  cmtbytes.HexBytes
	SourceBlockHash cmtbytes.HexBytes
}

func (e ErrBlockHashMismatch) Error() string {
	return fmt.Sprintf("trusted block is different to the source's first block (%X = %X)", e.TraceBlockHash, e.SourceBlockHash)
}

type ErrHeaderHashMismatch struct {
	Expected cmtbytes.HexBytes
	Actual   cmtbytes.HexBytes
}

func (e ErrHeaderHashMismatch) Error() string {
	return fmt.Sprintf("expected header's hash %X, but got %X", e.Expected, e.Actual)
}

type ErrExistingHeaderHashMismatch struct {
	Existing cmtbytes.HexBytes
	New      cmtbytes.HexBytes
}

func (e ErrExistingHeaderHashMismatch) Error() string {
	return fmt.Sprintf("existing trusted header %X does not match newHeader %X", e.Existing, e.New)
}

type ErrLightHeaderHashMismatch struct {
	Existing cmtbytes.HexBytes
	New      cmtbytes.HexBytes
}

func (e ErrLightHeaderHashMismatch) Error() string {
	return fmt.Sprintf("light block header %X does not match newHeader %X", e.Existing, e.New)
}

type ErrInvalidHashSize struct {
	Expected int
	Actual   int
}

func (e ErrInvalidHashSize) Error() string {
	return fmt.Sprintf("expected hash size to be %d bytes, got %d bytes", e.Expected, e.Actual)
}

type ErrUnexpectedChainID struct {
	Index    int
	Witness  provider.Provider
	Actual   string
	Expected string
}

func (e ErrUnexpectedChainID) Error() string {
	return fmt.Sprintf("witness #%d: %v is on another chain %s, expected %s", e.Index, e.Witness, e.Actual, e.Expected)
}

// ErrNewValSetCantBeTrusted means the new validator set cannot be trusted
// because < 1/3rd (+trustLevel+) of the old validator set has signed.
type ErrNewValSetCantBeTrusted struct {
	Reason types.ErrNotEnoughVotingPowerSigned
}

func (e ErrNewValSetCantBeTrusted) Error() string {
	return fmt.Sprintf("can't trust new val set: %v", e.Reason)
}

// ErrInvalidHeader means the header either failed the basic validation or
// commit is not signed by 2/3+.
type ErrInvalidHeader struct {
	Reason error
}

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("invalid header: %v", e.Reason)
}

func (e ErrInvalidHeader) Unwrap() error {
	return e.Reason
}

type ErrVerifySkipping struct {
	Err error
}

func (e ErrVerifySkipping) Error() string {
	return fmt.Sprintf("verifySkipping of conflicting header failed: %v", e.Err)
}

func (e ErrVerifySkipping) Unwrap() error {
	return e.Err
}

type ErrExamineTrace struct {
	Err error
}

func (e ErrExamineTrace) Error() string {
	return fmt.Sprintf("failed to examine trace: %v", e.Err)
}

func (e ErrExamineTrace) Unwrap() error {
	return e.Err
}

type ErrHeaderValidateBasic struct {
	Err error
}

func (e ErrHeaderValidateBasic) Error() string {
	return fmt.Sprintf("untrustedHeader.ValidateBasic failed: %v", e.Err)
}

func (e ErrHeaderValidateBasic) Unwrap() error {
	return e.Err
}

type ErrInvalidTrustOptions struct {
	Err error
}

func (e ErrInvalidTrustOptions) Error() string {
	return fmt.Sprintf("invalid TrustOptions: %v", e.Err)
}

func (e ErrInvalidTrustOptions) Unwrap() error {
	return e.Err
}

type ErrGetTrustedBlock struct {
	Err error
}

func (e ErrGetTrustedBlock) Error() string {
	return fmt.Sprintf("can't get last trusted light block: %v", e.Err)
}

func (e ErrGetTrustedBlock) Unwrap() error {
	return e.Err
}

type ErrGetTrustedBlockHeight struct {
	Err error
}

func (e ErrGetTrustedBlockHeight) Error() string {
	return fmt.Sprintf("can't get last trusted light block height: %v", e.Err)
}

func (e ErrGetTrustedBlockHeight) Unwrap() error {
	return e.Err
}

type ErrCleanup struct {
	Err error
}

func (e ErrCleanup) Error() string {
	return fmt.Sprintf("failed to cleanup: %v", e.Err)
}

func (e ErrCleanup) Unwrap() error {
	return e.Err
}

type ErrGetBlock struct {
	Err error
}

func (e ErrGetBlock) Error() string {
	return fmt.Sprintf("failed to retrieve light block from primary to verify against: %v", e.Err)
}

func (e ErrGetBlock) Unwrap() error {
	return e.Err
}

type ErrGetFirstBlock struct {
	Err error
}

func (e ErrGetFirstBlock) Error() string {
	return fmt.Sprintf("can't get first light block: %v", e.Err)
}

func (e ErrGetFirstBlock) Unwrap() error {
	return e.Err
}

type ErrGetFirstBlockHeight struct {
	Err error
}

func (e ErrGetFirstBlockHeight) Error() string {
	return fmt.Sprintf("can't get first light block height: %v", e.Err)
}

func (e ErrGetFirstBlockHeight) Unwrap() error {
	return e.Err
}

type ErrInvalidCommit struct {
	Err error
}

func (e ErrInvalidCommit) Error() string {
	return fmt.Sprintf("invalid commit: %v", e.Err)
}

func (e ErrInvalidCommit) Unwrap() error {
	return e.Err
}

type ErrGetLastTrustedHeight struct {
	Err error
}

func (e ErrGetLastTrustedHeight) Error() string {
	return fmt.Sprintf("can't get last trusted height: %v", e.Err)
}

func (e ErrGetLastTrustedHeight) Unwrap() error {
	return e.Err
}

type ErrPrune struct {
	Err error
}

func (e ErrPrune) Error() string {
	return fmt.Sprintf("prune: %v", e.Err)
}

func (e ErrPrune) Unwrap() error {
	return e.Err
}

type ErrSaveTrustedHeader struct {
	Err error
}

func (e ErrSaveTrustedHeader) Error() string {
	return fmt.Sprintf("failed to save trusted header: %v", e.Err)
}

func (e ErrSaveTrustedHeader) Unwrap() error {
	return e.Err
}

type ErrCleanupAfter struct {
	Height int64
	Err    error
}

func (e ErrCleanupAfter) Error() string {
	return fmt.Sprintf("cleanup after height %d failed: %v", e.Height, e.Err)
}

func (e ErrCleanupAfter) Unwrap() error {
	return e.Err
}

type ErrGetSignedHeaderBeforeHeight struct {
	Height int64
	Err    error
}

func (e ErrGetSignedHeaderBeforeHeight) Error() string {
	return fmt.Sprintf("can't get signed header before height %d: %v", e.Height, e.Err)
}

func (e ErrGetSignedHeaderBeforeHeight) Unwrap() error {
	return e.Err
}

type ErrGetHeaderBeforeHeight struct {
	Height int64
	Err    error
}

func (e ErrGetHeaderBeforeHeight) Error() string {
	return fmt.Sprintf("failed to get header before %d: %v", e.Height, e.Err)
}

func (e ErrGetHeaderBeforeHeight) Unwrap() error {
	return e.Err
}

type ErrGetHeaderAtHeight struct {
	Height int64
	Err    error
}

func (e ErrGetHeaderAtHeight) Error() string {
	return fmt.Sprintf("failed to obtain the header at height #%d: %v", e.Height, e.Err)
}

func (e ErrGetHeaderAtHeight) Unwrap() error {
	return e.Err
}

// ErrVerificationFailed means either sequential or skipping verification has
// failed to verify from header #1 to header #2 due to some reason.
type ErrVerificationFailed struct {
	From   int64
	To     int64
	Reason error
}

// Unwrap returns underlying reason.
func (e ErrVerificationFailed) Unwrap() error {
	return e.Reason
}

func (e ErrVerificationFailed) Error() string {
	return fmt.Sprintf("verify from #%d to #%d failed: %v", e.From, e.To, e.Reason)
}

// ErrConflictingHeaders is thrown when two conflicting headers are discovered.
type ErrConflictingHeaders struct {
	Block        *types.LightBlock
	WitnessIndex int
}

func (e ErrConflictingHeaders) Error() string {
	return fmt.Sprintf(
		"header hash (%X) from witness (%d) does not match primary",
		e.Block.Hash(), e.WitnessIndex)
}

// ErrProposerPrioritiesDiverge is thrown when two conflicting headers are
// discovered, but the error is non-attributable comparing to ErrConflictingHeaders.
// The difference is in validator set proposer priorities, which may change
// with every round of consensus.
type ErrProposerPrioritiesDiverge struct {
	WitnessHash  []byte
	WitnessIndex int
	PrimaryHash  []byte
}

func (e ErrProposerPrioritiesDiverge) Error() string {
	return fmt.Sprintf(
		"validator set's proposer priority hashes do not match: witness[%d]=%X, primary=%X",
		e.WitnessIndex, e.WitnessHash, e.PrimaryHash)
}

// ----------------------------- INTERNAL ERRORS ---------------------------------

// errBadWitness is returned when the witness either does not respond or
// responds with an invalid header.
type errBadWitness struct {
	Reason       error
	WitnessIndex int
}

func (e errBadWitness) Error() string {
	return fmt.Sprintf("Witness %d returned error: %s", e.WitnessIndex, e.Reason.Error())
}

func (e errBadWitness) Unwrap() error {
	return e.Reason
}

var errNoDivergence = errors.New(
	"sanity check failed: no divergence between the original trace and the provider's new trace",
)
