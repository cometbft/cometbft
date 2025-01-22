package schema

import "github.com/cometbft/cometbft/pkg/trace"

const (
	ABCITable = "abci"
)

// ABCIUpdate is an enum that represents the different types of ABCI
// trace data.
type ABCIUpdate string

const (
	PrepareProposalStart ABCIUpdate = "prepare_proposal_start"
	PrepareProposalEnd   ABCIUpdate = "prepare_proposal_end"
	ProcessProposalStart ABCIUpdate = "process_proposal_start"
	ProcessProposalEnd   ABCIUpdate = "process_proposal_end"
	CommitStart          ABCIUpdate = "commit_start"
	CommitEnd            ABCIUpdate = "commit_end"
)

// ABCI describes schema for the "abci" table.
type ABCI struct {
	TraceType string `json:"trace"`
	Height    int64  `json:"height"`
	Round     int32  `json:"round"`
}

// Table returns the table name for the ABCI struct and fulfills the
// trace.Entry interface.
func (m ABCI) Table() string {
	return ABCITable
}

// WriteABCI writes a trace for an ABCI method.
func WriteABCI(client trace.Tracer, traceType ABCIUpdate, height int64, round int32) {
	client.Write(ABCI{
		TraceType: string(traceType),
		Height:    height,
		Round:     round,
	})
}
