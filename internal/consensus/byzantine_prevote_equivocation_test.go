package consensus

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcons "github.com/cometbft/cometbft/api/cometbft/consensus/v1"
	"github.com/cometbft/cometbft/internal/evidence"
	sm "github.com/cometbft/cometbft/internal/state"
	"github.com/cometbft/cometbft/internal/store"
	cmtsync "github.com/cometbft/cometbft/internal/sync"
	"github.com/cometbft/cometbft/libs/log"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Byzantine node sends two different prevotes (nil and blockID) to the same validator
// TestByzantinePrevoteEquivocation tests the scenario where a Byzantine node sends two different prevotes (nil and blockID) to the same validator.
func TestByzantinePrevoteEquivocation(t *testing.T) {
	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define constants for the number of validators, the Byzantine node, and the height at which to prevote
	const nValidators = 4
	const byzantineNode = 0
	const prevoteHeight = int64(2)
	testName := "consensus_byzantine_test"
	appFunc := newKVStore

	// Generate a random genesis document and private validators
	genDoc, privVals := randGenesisDoc(nValidators, false, 30, nil)
	css, err := initializeValidators(t, nValidators, genDoc, privVals, testName, appFunc)
	require.NoError(t, err)
	// Initialize the reactors for each of the validators
	reactors, blocksSubs, eventBuses, err := initializeReactors(nValidators, css)
	require.NoError(t, err)

	// Make connected switches and start all reactors
	p2p.MakeConnectedSwitches(config.P2P, nValidators, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("CONSENSUS", reactors[i])
		s.SetLogger(reactors[i].conS.Logger.With("module", "p2p"))
		return s
	}, p2p.Connect2Switches)

	// Create byzantine validator
	bcs := css[byzantineNode]

	// Alter prevote so that the byzantine node double votes when height is 2
	alterPrevoteForByzantineNode(t, bcs, prevoteHeight, reactors, byzantineNode)
	introduceLazyProposer(t, css[1], ctx)

	// Start the consensus reactors
	for i := 0; i < nValidators; i++ {
		s := reactors[i].conS.GetState()
		reactors[i].SwitchToConsensus(s, false)
	}
	defer stopConsensusNet(log.TestingLogger(), reactors, eventBuses)

	// Evidence should be submitted and committed at the third height but
	// we will check the first six just in case
	evidenceFromEachValidator := make([]types.Evidence, nValidators)

	// Wait for all validators to commit evidence
	wg := new(sync.WaitGroup)
	for i := 0; i < nValidators; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for msg := range blocksSubs[i].Out() {
				block := msg.Data().(types.EventDataNewBlock).Block
				if len(block.Evidence.Evidence) != 0 {
					evidenceFromEachValidator[i] = block.Evidence.Evidence[0]
					return
				}
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	pubkey, err := bcs.privValidator.GetPubKey()
	require.NoError(t, err)

	const timeout = 180 * time.Second // Increase timeout to 180 seconds (this is a temporary measure)

	// Check if evidence was committed
	select {
	case <-done:
		for idx, ev := range evidenceFromEachValidator {
			if assert.NotNil(t, ev, idx) {
				ev, ok := ev.(*types.DuplicateVoteEvidence)
				assert.True(t, ok)
				assert.Equal(t, pubkey.Address(), ev.VoteA.ValidatorAddress)
				assert.Equal(t, prevoteHeight, ev.Height())
			}
		}
	case <-time.After(timeout):
		t.Logf("Evidence from each validator: %v", evidenceFromEachValidator) // Log evidence
		t.Fatalf("Timed out waiting for validators to commit evidence")
	}
}

// HELPER FUNCTIONS FOR  TESTBYZANTINEPREVOTEEQUIVOCATION

// initializeValidators sets up a network of validators for testing.
// It creates nValidators number of validators, initializes their state,
// and returns a slice of these states.
// genDoc is the genesis document used to initialize the state.
// privVals is a slice of private validators.
// testName is the name of the test for logging purposes.
// appFunc is a function that returns an ABCI application used in consensus state.
func initializeValidators(t *testing.T, nValidators int, genDoc *types.GenesisDoc, privVals []types.PrivValidator, testName string, appFunc func() abci.Application) ([]*State, error) {
	css := make([]*State, nValidators)

	for i := 0; i < nValidators; i++ {
		// Create logger for each validator
		logger := consensusLogger().With("test", testName, "validator", i)
		logger.Info("Initializing validator", "index", i)

		// Each state needs its own db
		stateDB := dbm.NewMemDB()

		// Create a new state store for each validator
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: false,
		})

		// Load state from DB or genesis document
		state, err := stateStore.LoadFromDBOrGenesisDoc(genDoc)
		require.NoError(t, err)

		// Reset config for each validator
		thisConfig := ResetConfig(fmt.Sprintf("%s_%d", testName, i))

		// Clean up config root directory after test
		defer os.RemoveAll(thisConfig.RootDir)

		// Ensure directory for write-ahead log exists
		ensureDir(path.Dir(thisConfig.Consensus.WalFile()), 0o700)

		// Initialize application
		app := appFunc()

		// Initialize validators from the state
		vals := types.TM2PB.ValidatorUpdates(state.Validators)

		// Initialize chain with validators
		_, err = app.InitChain(context.Background(), &abci.InitChainRequest{Validators: vals})
		if err != nil {
			logger.Error("Failed to initialize chain", "error", err)
			return nil, err
		}

		// Initialize block DB and store
		blockDB := dbm.NewMemDB()
		blockStore := store.NewBlockStore(blockDB)

		// Create a new mutex for each validator
		mtx := new(cmtsync.Mutex)

		// Create proxy app connections for consensus and mempool
		proxyAppConnCon := proxy.NewAppConnConsensus(abcicli.NewLocalClient(mtx, app), proxy.NopMetrics())
		proxyAppConnMem := proxy.NewAppConnMempool(abcicli.NewLocalClient(mtx, app), proxy.NopMetrics())

		// Initialize mempool
		mempool := mempl.NewCListMempool(config.Mempool,
			proxyAppConnMem,
			state.LastBlockHeight,
			mempl.WithPreCheck(sm.TxPreCheck(state)),
			mempl.WithPostCheck(sm.TxPostCheck(state)))

		// Enable transactions if necessary
		if thisConfig.Consensus.WaitForTxs() {
			mempool.EnableTxsAvailable()
		}

		// Initialize evidence pool
		evidenceDB := dbm.NewMemDB()
		evpool, err := evidence.NewPool(evidenceDB, stateStore, blockStore)
		if err != nil {
			logger.Error("Failed to initialize evidence pool", "error", err)
			return nil, err
		}
		evpool.SetLogger(logger.With("module", "evidence"))

		// Initialize state
		blockExec := sm.NewBlockExecutor(stateStore, log.TestingLogger(), proxyAppConnCon, mempool, evpool, blockStore)
		cs := NewState(thisConfig.Consensus, state, blockExec, blockStore, mempool, evpool)
		cs.SetLogger(cs.Logger)

		// Set private validator
		pv := privVals[i]
		cs.SetPrivValidator(pv)

		// Initialize event bus
		eventBus := types.NewEventBus()
		eventBus.SetLogger(log.TestingLogger().With("module", "events"))
		err = eventBus.Start()
		if err != nil {
			logger.Error("Failed to start event bus", "error", err)
			return nil, err
		}
		// Set the event bus
		cs.SetEventBus(eventBus)

		// Create a new ticker function for the consensus state
		tickerFunc := newMockTickerFunc(true)

		// Set the timeout ticker
		cs.SetTimeoutTicker(tickerFunc())
		cs.SetLogger(logger)

		// Add the consensus state to the slice
		css[i] = cs
	}

	// Return the slice of consensus states
	return css, nil
}

func initializeReactors(nValidators int, css []*State) ([]*Reactor, []types.Subscription, []*types.EventBus, error) {
	reactors := make([]*Reactor, nValidators)
	blocksSubs := make([]types.Subscription, 0)
	eventBuses := make([]*types.EventBus, nValidators)

	for i := 0; i < nValidators; i++ {
		reactors[i] = NewReactor(css[i], true) // so we don't start the consensus states
		reactors[i].SetLogger(css[i].Logger)

		// eventBus is already started with the cs
		eventBuses[i] = css[i].eventBus
		reactors[i].SetEventBus(eventBuses[i])

		// Subscribe to new block events
		blocksSub, err := eventBuses[i].Subscribe(context.Background(), testSubscriber, types.EventQueryNewBlock, 100)
		if err != nil {
			return nil, nil, nil, err
		}
		blocksSubs = append(blocksSubs, blocksSub)

		// Simulate handle initChain in handshake if last block height is 0
		if css[i].state.LastBlockHeight == 0 {
			err = css[i].blockExec.Store().Save(css[i].state)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	return reactors, blocksSubs, eventBuses, nil
}

func alterPrevoteForByzantineNode(t *testing.T, bcs *State, prevoteHeight int64, reactors []*Reactor, byzantineNode int) {
	bcs.doPrevote = func(height int64, round int32) {
		if height == prevoteHeight {
			bcs.Logger.Info("Sending two votes")
			prevote1, err := bcs.signVote(types.PrevoteType, bcs.ProposalBlock.Hash(), bcs.ProposalBlockParts.Header(), nil)
			require.NoError(t, err)
			prevote2, err := bcs.signVote(types.PrevoteType, nil, types.PartSetHeader{}, nil)
			require.NoError(t, err)
			peerList := reactors[byzantineNode].Switch.Peers().List()
			bcs.Logger.Info("Getting peer list", "peers", peerList)
			for i, peer := range peerList {
				if i < len(peerList)/2 {
					bcs.Logger.Info("Signed and pushed vote", "vote", prevote1, "peer", peer)
					peer.Send(p2p.Envelope{
						Message:   &cmtcons.Vote{Vote: prevote1.ToProto()},
						ChannelID: VoteChannel,
					})
				} else {
					bcs.Logger.Info("Signed and pushed vote", "vote", prevote2, "peer", peer)
					peer.Send(p2p.Envelope{
						Message:   &cmtcons.Vote{Vote: prevote2.ToProto()},
						ChannelID: VoteChannel,
					})
				}
			}
		} else {
			bcs.Logger.Info("Behaving normally")
			bcs.defaultDoPrevote(height, round)
		}
	}
}

// introduceLazyProposer modifies the decideProposal function of a given State instance
// to simulate a lazy proposer. The lazy proposer proposes a condensed commit and signs the proposal.
// This function is used in testing scenarios to simulate different behaviors of proposers.
func introduceLazyProposer(t *testing.T, lazyProposer *State, ctx context.Context) {
	// Overwrite the decideProposal function
	lazyProposer.decideProposal = func(height int64, round int32) {
		// Log the action of the lazy proposer
		lazyProposer.Logger.Info("Lazy Proposer proposing condensed commit")

		// Panic if the private validator is not set
		if lazyProposer.privValidator == nil {
			panic("entered createProposalBlock with privValidator being nil")
		}

		var extCommit *types.ExtendedCommit
		switch {
		// If the height is the initial height, create an empty ExtendedCommit
		case lazyProposer.Height == lazyProposer.state.InitialHeight:
			extCommit = &types.ExtendedCommit{}
		// If the last commit has a two-thirds majority, create an ExtendedCommit with vote extensions
		case lazyProposer.LastCommit.HasTwoThirdsMajority():
			veHeightParam := types.ABCIParams{VoteExtensionsEnableHeight: height}
			extCommit = lazyProposer.LastCommit.MakeExtendedCommit(veHeightParam)
		// If the private validator public key is not set, log an error and return
		default:
			lazyProposer.Logger.Error("enterPropose", "error", ErrPubKeyIsNotSet)
			return
		}

		// Set the last signature in the ExtendedCommit to absent
		extCommit.ExtendedSignatures[len(extCommit.ExtendedSignatures)-1] = types.NewExtendedCommitSigAbsent()

		// If the private validator public key is not set, log an error and return
		if lazyProposer.privValidatorPubKey == nil {
			lazyProposer.Logger.Error(fmt.Sprintf("enterPropose: %v", ErrPubKeyIsNotSet))
			return
		}
		// Get the address of the proposer
		proposerAddr := lazyProposer.privValidatorPubKey.Address()

		// Create a proposal block
		block, err := lazyProposer.blockExec.CreateProposalBlock(
			ctx, lazyProposer.Height, lazyProposer.state, extCommit, proposerAddr)
		require.NoError(t, err)

		// Create a part set from the block
		blockParts, err := block.MakePartSet(types.BlockPartSizeBytes)
		require.NoError(t, err)

		// Flush the write-ahead log to disk
		if err := lazyProposer.wal.FlushAndSync(); err != nil {
			lazyProposer.Logger.Error("Error flushing to disk")
		}

		// Create a block ID for the proposal
		propBlockID := types.BlockID{Hash: block.Hash(), PartSetHeader: blockParts.Header()}
		// Create a new proposal
		proposal := types.NewProposal(height, round, lazyProposer.ValidRound, propBlockID)
		p := proposal.ToProto()

		// Sign the proposal
		if err := lazyProposer.privValidator.SignProposal(lazyProposer.state.ChainID, p); err == nil {
			proposal.Signature = p.Signature

			// Send the proposal message
			lazyProposer.sendInternalMessage(msgInfo{&ProposalMessage{proposal}, ""})

			// Send all block part messages
			for i := 0; i < int(blockParts.Total()); i++ {
				part := blockParts.GetPart(i)
				lazyProposer.sendInternalMessage(msgInfo{&BlockPartMessage{lazyProposer.Height, lazyProposer.Round, part}, ""})
			}

			// Log the signed proposal
			lazyProposer.Logger.Info("Signed proposal", "height", height, "round", round, "proposal", proposal)
			lazyProposer.Logger.Debug(fmt.Sprintf("Signed proposal block: %v", block))
		} else if !lazyProposer.replayMode {
			// Log an error if signing the proposal failed
			lazyProposer.Logger.Error("enterPropose: Error signing proposal", "height", height, "round", round, "err", err)
		}
	}
}
