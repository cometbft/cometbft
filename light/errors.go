package light

import (
	"errors"
	"fmt"
	"time"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	"github.com/cometbft/cometbft/types"
)

var (

	// ErrFailedHeaderCrossReferencing is returned when the detector was not able to cross reference the header
	// with any of the connected witnesses.
	ErrFailedHeaderCrossReferencing = errors.New("all witnesses have either not responded, don't have the " +
		" blocks or sent invalid blocks. You should look to change your witnesses" +
		"  or review the light client's logs for more information")
	// ErrLightClientAttack is returned when the light client has detected an attempt
	// to verify a false header and has sent the evidence to either a witness or primary.
	ErrLightClientAttack = errors.New(`attempted attack detected.
Light client received valid conflicting header from witness.
Unable to verify header. Evidence has been sent to both providers.
Check logs for full evidence and trace`,
	)

	// ErrNoWitnesses means that there are not enough witnesses connected to
	// continue running the light client.
	ErrNoWitnesses             = errors.New("no witnesses connected. please reset light client")
	ErrNilOrSinglePrimaryTrace = errors.New("nil or single block primary trace")
	ErrHeaderHeightAdjacent    = errors.New("headers must be non adjacent in height")
	ErrHeaderHeightNotAdjacent = errors.New("headers must be adjacent in height")
	ErrNegativeOrZeroPeriod    = errors.New("negative or zero period")
	ErrNegativeOrZeroHeight    = errors.New("zero or negative height")
	ErrBlockTimeSanityCheck    = errors.New("sanity check failed: expected traceblock to have a lesser time than the target block")
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

type ErrBlockHeightCmp struct {
	GetHeight  int64
	WantHeight int64
}

func (e ErrBlockHeightCmp) Error() string {
	return fmt.Sprintf("target block has a height lower than the trusted height (%d < %d)", e.GetHeight, e.WantHeight)
}

type ErrHeaderHeightCmp struct {
	WantHeight int64
	GetHeight  int64
}

func (e ErrHeaderHeightCmp) Error() string {
	return fmt.Sprintf("expected new header height %d to be greater than one of old header %d", e.WantHeight, e.GetHeight)
}

type ErrHeaderTimeCmp struct {
	WantTime time.Time
	GetTime  time.Time
}

func (e ErrHeaderTimeCmp) Error() string {
	return fmt.Sprintf("expected new header time %v to be after old header time %v", e.WantTime, e.GetTime)
}

type ErrHeaderTimeExceedMaxClockDrift struct {
	Ti    time.Time
	Now   time.Time
	Drift time.Duration
}

func (e ErrHeaderTimeExceedMaxClockDrift) Error() string {
	return fmt.Sprintf("new header has a time from the future %v (now: %v; max clock drift: %v)", e.Ti, e.Now, e.Drift)
}

type ErrInvalidTrustLevel struct {
	Level cmtmath.Fraction
}

func (e ErrInvalidTrustLevel) Error() string {
	return fmt.Sprintf("trustLevel must be within [1/3, 1], given %v", e.Level)
}

type ErrHeaderValidatorHashAtGivenHeightMismatch struct {
	VH     cmtbytes.HexBytes
	SH     []byte
	Height int64
}

func (e ErrHeaderValidatorHashAtGivenHeightMismatch) Error() string {
	return fmt.Sprintf("expected new header validators (%X) to match those that were supplied (%X) at height %d", e.VH, e.SH, e.Height)
}

type ErrValidatorHashMismatch struct {
	TH  cmtbytes.HexBytes
	UTH cmtbytes.HexBytes
}

func (e ErrValidatorHashMismatch) Error() string {
	return fmt.Sprintf("expected old header next validators (%X) to match those from new header (%X)", e.TH, e.UTH)
}

type ErrBlockHashMismatch struct {
	TH cmtbytes.HexBytes
	SH cmtbytes.HexBytes
}

func (e ErrBlockHashMismatch) Error() string {
	return fmt.Sprintf("trusted block is different to the source's first block (%X = %X)", e.TH, e.SH)
}

type ErrHashSizeMismatch struct {
	Want int
	Get  int
}

func (e ErrHashSizeMismatch) Error() string {
	return fmt.Sprintf("expected hash size to be %d bytes, got %d bytes", e.Want, e.Get)
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

// ----------------------------- INTERNAL ERRORS ---------------------------------

// ErrConflictingHeaders is thrown when two conflicting headers are discovered.
type errConflictingHeaders struct {
	Block        *types.LightBlock
	WitnessIndex int
}

func (e errConflictingHeaders) Error() string {
	return fmt.Sprintf(
		"header hash (%X) from witness (%d) does not match primary",
		e.Block.Hash(), e.WitnessIndex)
}

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
