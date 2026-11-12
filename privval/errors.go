package privval

import (
	"errors"
	"fmt"
)

// EndpointTimeoutError occurs when endpoint times out.
type EndpointTimeoutError struct{}

// Implement the net.Error interface.
func (e EndpointTimeoutError) Error() string   { return "endpoint connection timed out" }
func (e EndpointTimeoutError) Timeout() bool   { return true }
func (e EndpointTimeoutError) Temporary() bool { return true }

// Socket errors.
var (
	ErrConnectionTimeout = EndpointTimeoutError{}
	ErrNoConnection      = errors.New("endpoint is not connected")
	ErrReadTimeout       = errors.New("endpoint read timed out")
	ErrWriteTimeout      = errors.New("endpoint write timed out")
)

// Signing state errors. These are returned by FilePV when a request to sign a
// vote or proposal would regress the last signed height/round/step, or when it
// conflicts with previously signed data for the same HRS. Callers can match
// them with errors.Is instead of comparing error strings.
var (
	// ErrHeightRegression is returned when the height to sign is lower than the
	// last signed height.
	ErrHeightRegression = errors.New("height regression")
	// ErrRoundRegression is returned when the round to sign is lower than the
	// last signed round at the same height.
	ErrRoundRegression = errors.New("round regression")
	// ErrStepRegression is returned when the step to sign is lower than the
	// last signed step at the same height and round.
	ErrStepRegression = errors.New("step regression")
	// ErrNoSignBytes is returned when the last sign state matches the requested
	// HRS but holds no SignBytes to reuse.
	ErrNoSignBytes = errors.New("no SignBytes found")
	// ErrConflictingData is returned when a vote or proposal for an
	// already-signed HRS differs from the previously signed data in a way that
	// is not a benign timestamp-only difference. Signing it would be a
	// double-sign.
	ErrConflictingData = errors.New("conflicting data")
	// ErrUnexpectedVoteExtension is returned when a vote carries an extension
	// but is not a non-nil precommit, the only case where extensions are
	// allowed.
	ErrUnexpectedVoteExtension = errors.New("unexpected vote extension - extensions are only allowed in non-nil precommits")
)

// RemoteSignerError allows (remote) validators to include meaningful error
// descriptions in their reply.
type RemoteSignerError struct {
	// TODO(ismail): create an enum of known errors
	Code        int
	Description string
}

func (e *RemoteSignerError) Error() string {
	return fmt.Sprintf("signerEndpoint returned error #%d: %s", e.Code, e.Description)
}
