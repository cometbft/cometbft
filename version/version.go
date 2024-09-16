package version

const (
	// TMCoreSemVer is the used as the fallback version of CometBFT Core
	// when not using git describe. It is formatted with semantic versioning.
	TMCoreSemVer = "0.34.35"
	// ABCISemVer is the semantic version of the ABCI library
	ABCISemVer = "0.17.0"

	ABCIVersion = ABCISemVer
)

var (
	// P2PProtocol versions all p2p behaviour and msgs.
	// This includes proposer selection.
	P2PProtocol uint64 = 8

	// BlockProtocol versions all block data structures and processing.
	// This includes validity of blocks and state updates.
	BlockProtocol uint64 = 11
)

// TMGitCommitHash uses git rev-parse HEAD to find commit hash which is helpful
// for the engineering team when working with the cometbft binary. See Makefile
var TMGitCommitHash = ""
