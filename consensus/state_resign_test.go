package consensus

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/internal/test"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// TestDecideProposalSameHRSResign covers a proposer asked to propose again
// at a height, round and step it already signed a proposal for, as after a
// restart before that block committed. FilePV then returns the earlier
// signature together with the earlier timestamp (the sign bytes differ
// only by timestamp). The proposal sent must carry that timestamp too:
// with the new timestamp and the old signature it does not verify, and
// the node rejects its own proposal ("invalid proposal signature"),
// losing the round.
func TestDecideProposalSameHRSResign(t *testing.T) {
	config := test.ResetTestRoot("consensus_resign_test")
	defer os.RemoveAll(config.RootDir)

	pv := loadPrivValidator(config)
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)
	state, err := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:         test.DefaultTestChainID,
		GenesisTime:     cmttime.Now(),
		ConsensusParams: test.ConsensusParams(),
		Validators:      []types.GenesisValidator{{PubKey: pubKey, Power: 10}},
	})
	require.NoError(t, err)
	cs := newStateWithConfig(config, state, pv, kvstore.NewInMemoryApplication())

	// The same block both times, as a restarted proposer re-proposes it.
	block, err := cs.createProposalBlock(context.Background())
	require.NoError(t, err)
	parts, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	cs.ValidBlock, cs.ValidBlockParts = block, parts

	propose := func() *types.Proposal {
		cs.decideProposal(cs.Height, cs.Round)
		var prop *types.Proposal
		for {
			select {
			case mi := <-cs.internalMsgQueue:
				if pm, ok := mi.Msg.(*ProposalMessage); ok {
					prop = pm.Proposal
				}
			default:
				require.NotNil(t, prop, "no proposal sent")
				return prop
			}
		}
	}

	first := propose()
	time.Sleep(5 * time.Millisecond) // a later timestamp for the second proposal
	second := propose()

	require.Equal(t, first.Signature, second.Signature, "FilePV should reuse the signature at the same HRS")
	require.True(t,
		pubKey.VerifySignature(types.ProposalSignBytes(cs.state.ChainID, second.ToProto()), second.Signature),
		"re-proposal at the same HRS carries a signature that does not verify (timestamp %v, signed timestamp %v)",
		second.Timestamp, first.Timestamp)
	require.Equal(t, first.Timestamp, second.Timestamp)
}
