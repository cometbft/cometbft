package blocksync

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	bcproto "github.com/cometbft/cometbft/proto/tendermint/blocksync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	mpmocks "github.com/cometbft/cometbft/mempool/mocks"
	"github.com/cometbft/cometbft/p2p"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

var config *cfg.Config

func randGenesisDoc(numValidators int, randPower bool, minPower int64) (*types.GenesisDoc, []types.PrivValidator) {
	validators := make([]types.GenesisValidator, numValidators)
	privValidators := make([]types.PrivValidator, numValidators)
	for i := 0; i < numValidators; i++ {
		val, privVal := types.RandValidator(randPower, minPower)
		validators[i] = types.GenesisValidator{
			PubKey: val.PubKey,
			Power:  val.VotingPower,
		}
		privValidators[i] = privVal
	}
	sort.Sort(types.PrivValidatorsByAddress(privValidators))

	consPar := types.DefaultConsensusParams()
	consPar.ABCI.VoteExtensionsEnableHeight = 1
	return &types.GenesisDoc{
		GenesisTime:     cmttime.Now(),
		ChainID:         test.DefaultTestChainID,
		Validators:      validators,
		ConsensusParams: consPar,
	}, privValidators
}

type ReactorPair struct {
	reactor *ByzantineReactor
	app     proxy.AppConns
}

func newReactor(
	t *testing.T,
	logger log.Logger,
	genDoc *types.GenesisDoc,
	privVals []types.PrivValidator,
	maxBlockHeight int64,
	incorrectData ...int64,
) ReactorPair {
	if len(privVals) != 1 {
		panic("only support one validator")
	}
	var incorrectBlock int64 = 0
	if len(incorrectData) > 0 {
		incorrectBlock = incorrectData[0]
	}

	app := abci.NewBaseApplication()
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := proxy.NewAppConns(cc, proxy.NopMetrics())
	err := proxyApp.Start()
	if err != nil {
		panic(fmt.Errorf("error start app: %w", err))
	}

	blockDB := dbm.NewMemDB()
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	blockStore := store.NewBlockStore(blockDB)

	state, err := stateStore.LoadFromDBOrGenesisDoc(genDoc)
	if err != nil {
		panic(fmt.Errorf("error constructing state from genesis file: %w", err))
	}

	mp := &mpmocks.Mempool{}
	mp.On("Lock").Return()
	mp.On("Unlock").Return()
	mp.On("FlushAppConn", mock.Anything).Return(nil)
	mp.On("Update",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil)

	// Make the Reactor itself.
	// NOTE we have to create and commit the blocks first because
	// pool.height is determined from the store.
	blockSync := true
	db := dbm.NewMemDB()
	stateStore = sm.NewStore(db, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	blockExec := sm.NewBlockExecutor(stateStore, log.TestingLogger(), proxyApp.Consensus(),
		mp, sm.EmptyEvidencePool{}, blockStore)
	if err = stateStore.Save(state); err != nil {
		panic(err)
	}

	// The commit we are building for the current height.
	seenExtCommit := &types.ExtendedCommit{}

	pubKey, err := privVals[0].GetPubKey()
	if err != nil {
		panic(err)
	}
	addr := pubKey.Address()
	idx, _ := state.Validators.GetByAddress(addr)

	// let's add some blocks in
	for blockHeight := int64(1); blockHeight <= maxBlockHeight; blockHeight++ {
		voteExtensionIsEnabled := genDoc.ConsensusParams.ABCI.VoteExtensionsEnabled(blockHeight)

		lastExtCommit := seenExtCommit.Clone()

		thisBlock := state.MakeBlock(blockHeight, nil, lastExtCommit.ToCommit(), nil, state.Validators.Proposer.Address)

		thisParts, err := thisBlock.MakePartSet(types.BlockPartSizeBytes)
		require.NoError(t, err)
		blockID := types.BlockID{Hash: thisBlock.Hash(), PartSetHeader: thisParts.Header()}

		// Simulate a commit for the current height
		vote, err := types.MakeVote(
			privVals[0],
			thisBlock.Header.ChainID,
			idx,
			thisBlock.Header.Height,
			0,
			cmtproto.PrecommitType,
			blockID,
			time.Now(),
		)
		if err != nil {
			panic(err)
		}
		seenExtCommit = &types.ExtendedCommit{
			Height:             vote.Height,
			Round:              vote.Round,
			BlockID:            blockID,
			ExtendedSignatures: []types.ExtendedCommitSig{vote.ExtendedCommitSig()},
		}

		state, err = blockExec.ApplyBlock(state, blockID, thisBlock)
		if err != nil {
			panic(fmt.Errorf("error apply block: %w", err))
		}

		saveCorrectVoteExtensions := blockHeight != incorrectBlock
		if saveCorrectVoteExtensions == voteExtensionIsEnabled {
			blockStore.SaveBlockWithExtendedCommit(thisBlock, thisParts, seenExtCommit)
		} else {
			blockStore.SaveBlock(thisBlock, thisParts, seenExtCommit.ToCommit())
		}
	}

	bcReactor := NewByzantineReactor(incorrectBlock, NewReactor(state.Copy(), blockExec, blockStore, blockSync, NopMetrics(), 0))
	bcReactor.SetLogger(logger.With("module", "blocksync"))

	return ReactorPair{bcReactor, proxyApp}
}

func TestNoBlockResponse(t *testing.T) {
	config = test.ResetTestRoot("blocksync_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	maxBlockHeight := int64(65)

	reactorPairs := make([]ReactorPair, 2)

	reactorPairs[0] = newReactor(t, log.TestingLogger(), genDoc, privVals, maxBlockHeight)
	reactorPairs[1] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)

	p2p.MakeConnectedSwitches(config.P2P, 2, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("BLOCKSYNC", reactorPairs[i].reactor)
		return s
	}, p2p.Connect2Switches)

	defer func() {
		for _, r := range reactorPairs {
			err := r.reactor.Stop()
			require.NoError(t, err)
			err = r.app.Stop()
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		height   int64
		existent bool
	}{
		{maxBlockHeight + 2, false},
		{10, true},
		{1, true},
		{100, false},
	}

	for {
		if reactorPairs[1].reactor.pool.IsCaughtUp() {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	assert.Equal(t, maxBlockHeight, reactorPairs[0].reactor.store.Height())

	for _, tt := range tests {
		block := reactorPairs[1].reactor.store.LoadBlock(tt.height)
		if tt.existent {
			assert.True(t, block != nil)
		} else {
			assert.True(t, block == nil)
		}
	}
}

// NOTE: This is too hard to test without
// an easy way to add test peer to switch
// or without significant refactoring of the module.
// Alternatively we could actually dial a TCP conn but
// that seems extreme.
func TestBadBlockStopsPeer(t *testing.T) {
	config = test.ResetTestRoot("blocksync_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	maxBlockHeight := int64(148)

	// Other chain needs a different validator set
	otherGenDoc, otherPrivVals := randGenesisDoc(1, false, 30)
	otherChain := newReactor(t, log.TestingLogger(), otherGenDoc, otherPrivVals, maxBlockHeight)

	defer func() {
		err := otherChain.reactor.Stop()
		require.Error(t, err)
		err = otherChain.app.Stop()
		require.NoError(t, err)
	}()

	reactorPairs := make([]ReactorPair, 4)

	reactorPairs[0] = newReactor(t, log.TestingLogger(), genDoc, privVals, maxBlockHeight)
	reactorPairs[1] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)
	reactorPairs[2] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)
	reactorPairs[3] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)

	switches := p2p.MakeConnectedSwitches(config.P2P, 4, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("BLOCKSYNC", reactorPairs[i].reactor)
		return s
	}, p2p.Connect2Switches)

	defer func() {
		for _, r := range reactorPairs {
			err := r.reactor.Stop()
			require.NoError(t, err)

			err = r.app.Stop()
			require.NoError(t, err)
		}
	}()

	for {
		time.Sleep(1 * time.Second)
		caughtUp := true
		for _, r := range reactorPairs {
			if !r.reactor.pool.IsCaughtUp() {
				caughtUp = false
			}
		}
		if caughtUp {
			break
		}
	}

	// at this time, reactors[0-3] is the newest
	assert.Equal(t, 3, reactorPairs[1].reactor.Switch.Peers().Size())

	// Mark reactorPairs[3] as an invalid peer. Fiddling with .store without a mutex is a data
	// race, but can't be easily avoided.
	reactorPairs[3].reactor.store = otherChain.reactor.store

	lastReactorPair := newReactor(t, log.TestingLogger(), genDoc, privVals, 0)
	reactorPairs = append(reactorPairs, lastReactorPair)

	switches = append(switches, p2p.MakeConnectedSwitches(config.P2P, 1, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("BLOCKSYNC", reactorPairs[len(reactorPairs)-1].reactor)
		return s
	}, p2p.Connect2Switches)...)

	for i := 0; i < len(reactorPairs)-1; i++ {
		p2p.Connect2Switches(switches, i, len(reactorPairs)-1)
	}

	for {
		if lastReactorPair.reactor.pool.IsCaughtUp() || lastReactorPair.reactor.Switch.Peers().Size() == 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	assert.True(t, lastReactorPair.reactor.Switch.Peers().Size() < len(reactorPairs)-1)
}

func TestCheckSwitchToConsensusLastHeightZero(t *testing.T) {
	const maxBlockHeight = int64(45)

	config = test.ResetTestRoot("blocksync_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	reactorPairs := make([]ReactorPair, 1, 2)
	reactorPairs[0] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)
	reactorPairs[0].reactor.switchToConsensusMs = 50
	defer func() {
		for _, r := range reactorPairs {
			err := r.reactor.Stop()
			require.NoError(t, err)
			err = r.app.Stop()
			require.NoError(t, err)
		}
	}()

	reactorPairs = append(reactorPairs, newReactor(t, log.TestingLogger(), genDoc, privVals, maxBlockHeight))

	var switches []*p2p.Switch
	for _, r := range reactorPairs {
		switches = append(switches, p2p.MakeConnectedSwitches(config.P2P, 1, func(i int, s *p2p.Switch) *p2p.Switch {
			s.AddReactor("BLOCKSYNC", r.reactor)
			return s
		}, p2p.Connect2Switches)...)
	}

	time.Sleep(60 * time.Millisecond)

	// Connect both switches
	p2p.Connect2Switches(switches, 0, 1)

	startTime := time.Now()
	for {
		time.Sleep(20 * time.Millisecond)
		caughtUp := true
		for _, r := range reactorPairs {
			if !r.reactor.pool.IsCaughtUp() {
				caughtUp = false
				break
			}
		}
		if caughtUp {
			break
		}
		if time.Since(startTime) > 90*time.Second {
			msg := "timeout: reactors didn't catch up;"
			for i, r := range reactorPairs {
				h, p, lr := r.reactor.pool.GetStatus()
				c := r.reactor.pool.IsCaughtUp()
				msg += fmt.Sprintf(" reactor#%d (h %d, p %d, lr %d, c %t);", i, h, p, lr, c)
			}
			require.Fail(t, msg)
		}
	}

	// -1 because of "-1" in IsCaughtUp
	// -1 pool.height points to the _next_ height
	// -1 because we measure height of block store
	const maxDiff = 3
	for _, r := range reactorPairs {
		assert.GreaterOrEqual(t, r.reactor.store.Height(), maxBlockHeight-maxDiff)
	}
}

func ExtendedCommitNetworkHelper(t *testing.T, maxBlockHeight int64, enableVoteExtensionAt int64, invalidBlockHeightAt int64) {
	config = test.ResetTestRoot("blocksync_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)
	genDoc.ConsensusParams.ABCI.VoteExtensionsEnableHeight = enableVoteExtensionAt

	reactorPairs := make([]ReactorPair, 1, 2)
	reactorPairs[0] = newReactor(t, log.TestingLogger(), genDoc, privVals, 0)
	reactorPairs[0].reactor.switchToConsensusMs = 50
	defer func() {
		for _, r := range reactorPairs {
			err := r.reactor.Stop()
			require.NoError(t, err)
			err = r.app.Stop()
			require.NoError(t, err)
		}
	}()

	reactorPairs = append(reactorPairs, newReactor(t, log.TestingLogger(), genDoc, privVals, maxBlockHeight, invalidBlockHeightAt))

	var switches []*p2p.Switch
	for _, r := range reactorPairs {
		switches = append(switches, p2p.MakeConnectedSwitches(config.P2P, 1, func(i int, s *p2p.Switch) *p2p.Switch {
			s.AddReactor("BLOCKSYNC", r.reactor)
			return s
		}, p2p.Connect2Switches)...)
	}

	time.Sleep(60 * time.Millisecond)

	// Connect both switches
	p2p.Connect2Switches(switches, 0, 1)

	startTime := time.Now()
	for {
		time.Sleep(20 * time.Millisecond)
		// The reactor can never catch up, because at one point it disconnects.
		require.False(t, reactorPairs[0].reactor.pool.IsCaughtUp(), "node caught up when it should not have")
		// After 5 seconds, the test should have executed.
		if time.Since(startTime) > 5*time.Second {
			assert.Equal(t, 0, reactorPairs[0].reactor.Switch.Peers().Size(), "node should have disconnected but didn't")
			assert.Equal(t, 0, reactorPairs[1].reactor.Switch.Peers().Size(), "node should have disconnected but didn't")
			break
		}
	}
}

// TestCheckExtendedCommitExtra tests when VoteExtension is disabled but an ExtendedVote is present in the block.
func TestCheckExtendedCommitExtra(t *testing.T) {
	const maxBlockHeight = 10
	const enableVoteExtension = 5
	const invalidBlockHeight = 3

	ExtendedCommitNetworkHelper(t, maxBlockHeight, enableVoteExtension, invalidBlockHeight)
}

// TestCheckExtendedCommitMissing tests when VoteExtension is enabled but the ExtendedVote is missing from the block.
func TestCheckExtendedCommitMissing(t *testing.T) {
	const maxBlockHeight = 10
	const enableVoteExtension = 5
	const invalidBlockHeight = 8

	ExtendedCommitNetworkHelper(t, maxBlockHeight, enableVoteExtension, invalidBlockHeight)
}

// ByzantineReactor is a blockstore reactor implementation where a corrupted block can be sent to a peer.
// The corruption is that the block contains extended commit signatures when vote extensions are disabled or
// it has no extended commit signatures while vote extensions are enabled.
// If the corrupted block height is set to 0, the reactor behaves as normal.
type ByzantineReactor struct {
	*Reactor
	corruptedBlock int64
}

func NewByzantineReactor(invalidBlock int64, conR *Reactor) *ByzantineReactor {
	return &ByzantineReactor{
		Reactor:        conR,
		corruptedBlock: invalidBlock,
	}
}

// respondToPeer (overridden method) loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
// Byzantine modification: if corruptedBlock is set, send the wrong Block.
func (bcR *ByzantineReactor) respondToPeer(msg *bcproto.BlockRequest, src p2p.Peer) (queued bool) {
	block := bcR.store.LoadBlock(msg.Height)
	if block == nil {
		bcR.Logger.Info("Peer asking for a block we don't have", "src", src, "height", msg.Height)
		return src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message:   &bcproto.NoBlockResponse{Height: msg.Height},
		})
	}

	state, err := bcR.blockExec.Store().Load()
	if err != nil {
		bcR.Logger.Error("loading state", "err", err)
		return false
	}
	var extCommit *types.ExtendedCommit
	voteExtensionEnabled := state.ConsensusParams.ABCI.VoteExtensionsEnabled(msg.Height)
	incorrectBlock := bcR.corruptedBlock == msg.Height
	if voteExtensionEnabled && !incorrectBlock || !voteExtensionEnabled && incorrectBlock {
		extCommit = bcR.store.LoadBlockExtendedCommit(msg.Height)
		if extCommit == nil {
			bcR.Logger.Error("found block in store with no extended commit", "block", block)
			return false
		}
	}

	bl, err := block.ToProto()
	if err != nil {
		bcR.Logger.Error("could not convert msg to protobuf", "err", err)
		return false
	}

	return src.TrySend(p2p.Envelope{
		ChannelID: BlocksyncChannel,
		Message: &bcproto.BlockResponse{
			Block:     bl,
			ExtCommit: extCommit.ToProto(),
		},
	})
}

// Receive implements Reactor by handling 4 types of messages (look below).
// Copied unchanged from reactor.go so the correct respondToPeer is called.
func (bcR *ByzantineReactor) Receive(e p2p.Envelope) { //nolint: dupl
	if err := ValidateMsg(e.Message); err != nil {
		bcR.Logger.Error("Peer sent us invalid msg", "peer", e.Src, "msg", e.Message, "err", err)
		bcR.Switch.StopPeerForError(e.Src, err)
		return
	}

	bcR.Logger.Debug("Receive", "e.Src", e.Src, "chID", e.ChannelID, "msg", e.Message)

	switch msg := e.Message.(type) {
	case *bcproto.BlockRequest:
		bcR.respondToPeer(msg, e.Src)
	case *bcproto.BlockResponse:
		bi, err := types.BlockFromProto(msg.Block)
		if err != nil {
			bcR.Logger.Error("Peer sent us invalid block", "peer", e.Src, "msg", e.Message, "err", err)
			bcR.Switch.StopPeerForError(e.Src, err)
			return
		}
		var extCommit *types.ExtendedCommit
		if msg.ExtCommit != nil {
			var err error
			extCommit, err = types.ExtendedCommitFromProto(msg.ExtCommit)
			if err != nil {
				bcR.Logger.Error("failed to convert extended commit from proto",
					"peer", e.Src,
					"err", err)
				bcR.Switch.StopPeerForError(e.Src, err)
				return
			}
		}

		if err := bcR.pool.AddBlock(e.Src.ID(), bi, extCommit, msg.Block.Size()); err != nil {
			bcR.Logger.Error("failed to add block", "peer", e.Src, "err", err)
		}
	case *bcproto.StatusRequest:
		// Send peer our state.
		e.Src.TrySend(p2p.Envelope{
			ChannelID: BlocksyncChannel,
			Message: &bcproto.StatusResponse{
				Height: bcR.store.Height(),
				Base:   bcR.store.Base(),
			},
		})
	case *bcproto.StatusResponse:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerRange(e.Src.ID(), msg.Base, msg.Height)
	case *bcproto.NoBlockResponse:
		bcR.Logger.Debug("Peer does not have requested block", "peer", e.Src, "height", msg.Height)
		bcR.pool.RedoRequestFrom(msg.Height, e.Src.ID())
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}
