package types

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/internal/protoio"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	cmttime "github.com/cometbft/cometbft/types/time"
)

var (
	testProposal *Proposal
	pbp          *cmtproto.Proposal
)

func init() {
	stamp, err := time.Parse(TimeFormat, "2018-02-11T07:09:22.765Z")
	if err != nil {
		panic(err)
	}
	testProposal = &Proposal{
		Height: 12345,
		Round:  23456,
		BlockID: BlockID{
			Hash:          []byte("--June_15_2020_amino_was_removed"),
			PartSetHeader: PartSetHeader{Total: 111, Hash: []byte("--June_15_2020_amino_was_removed")},
		},
		POLRound:  -1,
		Timestamp: stamp,
	}
	pbp = testProposal.ToProto()
}

func TestProposalSignable(t *testing.T) {
	chainID := "test_chain_id"
	signBytes := ProposalSignBytes(chainID, pbp)
	pb := CanonicalizeProposal(chainID, pbp)

	expected, err := protoio.MarshalDelimited(&pb)
	require.NoError(t, err)
	require.Equal(t, expected, signBytes, "Got unexpected sign bytes for Proposal")
}

func TestProposalString(t *testing.T) {
	str := testProposal.String()
	expected := `Proposal{12345/23456 (2D2D4A756E655F31355F323032305F616D696E6F5F7761735F72656D6F766564:111:2D2D4A756E65, -1) 000000000000 @ 2018-02-11T07:09:22.765Z}` //nolint:lll // ignore line length for tests
	if str != expected {
		t.Errorf("got unexpected string for Proposal. Expected:\n%v\nGot:\n%v", expected, str)
	}
}

func TestProposalVerifySignature(t *testing.T) {
	privVal := NewMockPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	prop := NewProposal(
		4, 2, 2,
		BlockID{cmtrand.Bytes(tmhash.Size), PartSetHeader{777, cmtrand.Bytes(tmhash.Size)}}, cmttime.Now())
	p := prop.ToProto()
	signBytes := ProposalSignBytes("test_chain_id", p)

	// sign it
	err = privVal.SignProposal("test_chain_id", p)
	require.NoError(t, err)
	prop.Signature = p.Signature

	// verify the same proposal
	valid := pubKey.VerifySignature(signBytes, prop.Signature)
	require.True(t, valid)

	// serialize, deserialize and verify again....
	newProp := new(cmtproto.Proposal)
	pb := prop.ToProto()

	bs, err := proto.Marshal(pb)
	require.NoError(t, err)

	err = proto.Unmarshal(bs, newProp)
	require.NoError(t, err)

	np, err := ProposalFromProto(newProp)
	require.NoError(t, err)

	// verify the transmitted proposal
	newSignBytes := ProposalSignBytes("test_chain_id", pb)
	require.Equal(t, string(signBytes), string(newSignBytes))
	valid = pubKey.VerifySignature(newSignBytes, np.Signature)
	require.True(t, valid)
}

func BenchmarkProposalWriteSignBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProposalSignBytes("test_chain_id", pbp)
	}
}

func BenchmarkProposalSign(b *testing.B) {
	privVal := NewMockPV()
	for i := 0; i < b.N; i++ {
		err := privVal.SignProposal("test_chain_id", pbp)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkProposalVerifySignature(b *testing.B) {
	privVal := NewMockPV()
	err := privVal.SignProposal("test_chain_id", pbp)
	require.NoError(b, err)
	pubKey, err := privVal.GetPubKey()
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		pubKey.VerifySignature(ProposalSignBytes("test_chain_id", pbp), testProposal.Signature)
	}
}

func TestProposalValidateBasic(t *testing.T) {
	privVal := NewMockPV()
	blockID := makeBlockID(tmhash.Sum([]byte("blockhash")), math.MaxInt32, tmhash.Sum([]byte("partshash")))
	location, err := time.LoadLocation("CET")
	require.NoError(t, err)

	testCases := []struct {
		testName         string
		malleateProposal func(*Proposal)
		expectErr        bool
	}{
		{"Good Proposal", func(p *Proposal) {}, false},
		{"Invalid Type", func(p *Proposal) { p.Type = PrecommitType }, true},
		{"Invalid Height", func(p *Proposal) { p.Height = -1 }, true},
		{"Zero Height", func(p *Proposal) { p.Height = 0 }, true},
		{"Invalid Round", func(p *Proposal) { p.Round = -1 }, true},
		{"Invalid POLRound", func(p *Proposal) { p.POLRound = -2 }, true},
		{"POLRound == Round", func(p *Proposal) { p.POLRound = p.Round }, true},
		{"Invalid BlockId", func(p *Proposal) {
			p.BlockID = BlockID{[]byte{1, 2, 3}, PartSetHeader{111, []byte("blockparts")}}
		}, true},
		{"Invalid Signature", func(p *Proposal) {
			p.Signature = make([]byte, 0)
		}, true},
		{"Small Signature", func(p *Proposal) {
			p.Signature = make([]byte, MaxSignatureSize-1)
		}, false},
		{"Too big Signature", func(p *Proposal) {
			p.Signature = make([]byte, MaxSignatureSize+1)
		}, true},
		{"Non canonical time", func(p *Proposal) {
			p.Timestamp = time.Now().In(location)
		}, true},
		{"Not rounded time", func(p *Proposal) {
			p.Timestamp = time.Now()
		}, true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			prop := NewProposal(
				4, 2, 1,
				blockID, cmttime.Now())
			p := prop.ToProto()
			err := privVal.SignProposal("test_chain_id", p)
			prop.Signature = p.Signature
			require.NoError(t, err)

			tc.malleateProposal(prop)
			err = prop.ValidateBasic()
			errMessage := fmt.Sprintf("Validate Basic had an unexpected error: %v", err)
			assert.Equal(t, tc.expectErr, prop.ValidateBasic() != nil, errMessage)
		})
	}
}

func TestProposalProtoBuf(t *testing.T) {
	proposal := NewProposal(1, 2, 3, makeBlockID([]byte("hash"), 2, []byte("part_set_hash")), cmttime.Now())
	proposal.Signature = []byte("sig")
	proposal2 := NewProposal(1, 2, 3, BlockID{}, cmttime.Now())

	testCases := []struct {
		msg     string
		p1      *Proposal
		expPass bool
	}{
		{"success", proposal, true},
		{"success", proposal2, false}, // blcokID cannot be empty
		{"empty proposal failure validatebasic", &Proposal{}, false},
		{"nil proposal", nil, false},
	}
	for _, tc := range testCases {
		protoProposal := tc.p1.ToProto()

		p, err := ProposalFromProto(protoProposal)
		if tc.expPass {
			require.NoError(t, err)
			require.Equal(t, tc.p1, p, tc.msg)
		} else {
			require.Error(t, err)
		}
	}
}

func TestIsTimely(t *testing.T) {
	genesisTime, err := time.Parse(time.RFC3339, "2019-03-13T23:00:00Z")
	require.NoError(t, err)
	testCases := []struct {
		name           string
		proposalHeight int64
		proposalTime   time.Time
		recvTime       time.Time
		round          int32
		precision      time.Duration
		msgDelay       time.Duration
		expectTimely   bool
	}{
		// proposalTime - precision <= localTime <= proposalTime + (msgDelay * (1.1)^round) + precision
		{
			// Checking that the following inequality evaluates to true:
			// 0 - 2 <= 1 <= 0 + 1 + 2
			name:           "basic timely",
			proposalHeight: 2,
			proposalTime:   genesisTime,
			recvTime:       genesisTime.Add(1 * time.Nanosecond),
			round:          0,
			precision:      time.Nanosecond * 2,
			msgDelay:       time.Nanosecond,
			expectTimely:   true,
		},
		{
			// Checking that the following inequality evaluates to false:
			// 0 - 2 <= 4 <= 0 + 1 + 2
			name:           "local time too large",
			proposalHeight: 2,
			proposalTime:   genesisTime,
			recvTime:       genesisTime.Add(4 * time.Nanosecond),
			round:          0,
			precision:      time.Nanosecond * 2,
			msgDelay:       time.Nanosecond,
			expectTimely:   false,
		},
		{
			// Checking that the following inequality evaluates to false:
			// 4 - 2 <= 0 <= 4 + 2 + 1
			name:           "proposal time too large",
			proposalHeight: 2,
			proposalTime:   genesisTime.Add(4 * time.Nanosecond),
			recvTime:       genesisTime,
			round:          0,
			precision:      time.Nanosecond * 2,
			msgDelay:       time.Nanosecond,
			expectTimely:   false,
		},
		{
			/*
				With adaptive synchrony params after 4 rounds, MSGDELAY(4) = 1.4641 * 2000ns
				Recv time must not exceed proposalTime + 2928ns (msgDelay) + 50ns (prcision) = 2978ns
			*/
			name:           "proposal time too large for round",
			proposalHeight: 2,
			proposalTime:   genesisTime,
			recvTime:       genesisTime.Add(2980 * time.Nanosecond),
			round:          4,
			precision:      time.Nanosecond * 50,
			msgDelay:       time.Nanosecond * 2000,
			expectTimely:   false,
		},
		{
			name:           "proposal timely for round",
			proposalHeight: 2,
			proposalTime:   genesisTime,
			recvTime:       genesisTime.Add(2970 * time.Nanosecond),
			round:          4,
			precision:      time.Nanosecond * 50,
			msgDelay:       time.Nanosecond * 2000,
			expectTimely:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p := Proposal{
				Height:    testCase.proposalHeight,
				Timestamp: testCase.proposalTime,
				Round:     testCase.round,
			}

			sp := AdaptiveSynchronyParams(
				testCase.precision,
				testCase.msgDelay,
				testCase.round,
			)

			ti := p.IsTimely(testCase.recvTime, sp)
			assert.Equal(t, testCase.expectTimely, ti)
		})
	}
}
