package blocksync

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

func TestReactorCombined(t *testing.T) {
	t.Run("ingestsBlock", func(t *testing.T) {
		// ARRANGE
		ts := newCombinedModeTestSuite(t, "blocksync_ingest_block")

		// Given two reactors for two different nodes
		// Provider has +2 blocks in the state
		var (
			provider = newReactor(t, ts.logger, ts.genDoc, ts.privVals, 4)
			follower = newReactor(t, ts.logger, ts.genDoc, ts.privVals, 2)
		)

		follower.reactor.combinedModeEnabled = true
		follower.reactor.intervalStatusUpdate = combinedModeInternalStatusUpdate
		provider.reactor.intervalStatusUpdate = combinedModeInternalStatusUpdate

		ts.blockIngestor.SetOnIngest(func(vb consensus.IngestCandidate) (error, bool) {
			ts.logger.Info("mock: receive block", "height", vb.Height())
			return nil, false
		})

		// Given two switches
		switches := p2p.MakeConnectedSwitches(ts.config.P2P, 2, func(i int, s *p2p.Switch) *p2p.Switch {
			switch i {
			case 0:
				s.AddReactor("BLOCKSYNC", provider.reactor)
			case 1:
				s.AddReactor("BLOCKSYNC", follower.reactor)
				s.AddReactor("CONSENSUS", ts.blockIngestor)
			}

			return s
		}, p2p.Connect2Switches)

		// Given the follower switch
		followerSwitch := switches[1]

		// Ensure the system is running
		require.True(t, followerSwitch.IsRunning())
		require.True(t, follower.reactor.IsRunning())

		// ACT
		// Wait for the block ingestor to receive the block
		// - Due to commit verification, we need N+1 blocks to verify N blocks
		// - So diff of 2 blocks yields to one block to be ingested
		check := func() bool {
			return len(ts.blockIngestor.Requests()) == 1
		}

		// ASSERT
		require.Eventually(t, check, 10*time.Second, 100*time.Millisecond)

		// check block
		block := ts.blockIngestor.Requests()[0]
		require.Equal(t, int64(3), block.Height())

		// ensure the pool progressed
		require.Equal(t, int64(4), follower.reactor.pool.Height())
	})

	t.Run("alreadyIngested", func(t *testing.T) {
		t.Skip()
		// ARRANGE
		ts := newCombinedModeTestSuite(t, "blocksync_ingest_block")

		// Given two reactors for two different nodes
		// Provider has +2 blocks in the state
		var (
			provider = newReactor(t, ts.logger, ts.genDoc, ts.privVals, 4)
			follower = newReactor(t, ts.logger, ts.genDoc, ts.privVals, 2)
		)

		follower.reactor.combinedModeEnabled = true
		follower.reactor.intervalStatusUpdate = combinedModeInternalStatusUpdate
		provider.reactor.intervalStatusUpdate = combinedModeInternalStatusUpdate

		ts.blockIngestor.SetOnIngest(func(vb consensus.IngestCandidate) (error, bool) {
			ts.logger.Info("mock: receive block", "height", vb.Height())
			return consensus.ErrAlreadyIncluded, false
		})

		// Given two switches
		switches := p2p.MakeConnectedSwitches(ts.config.P2P, 2, func(i int, s *p2p.Switch) *p2p.Switch {
			switch i {
			case 0:
				s.AddReactor("BLOCKSYNC", provider.reactor)
			case 1:
				s.AddReactor("BLOCKSYNC", follower.reactor)
				s.AddReactor("CONSENSUS", ts.blockIngestor)
			}

			return s
		}, p2p.Connect2Switches)

		// Given the follower switch
		followerSwitch := switches[1]

		// Ensure the system is running
		require.True(t, followerSwitch.IsRunning())
		require.True(t, follower.reactor.IsRunning())

		// ACT
		// Wait for the block ingestor to receive the block
		// - Due to commit verification, we need N+1 blocks to verify N blocks
		// - So diff of 2 blocks yields to one block to be ingested
		check := func() bool {
			return len(ts.blockIngestor.Requests()) == 1
		}

		// ASSERT
		require.Eventually(t, check, 10*time.Second, 100*time.Millisecond)

		block := ts.blockIngestor.Requests()[0]
		require.Equal(t, int64(3), block.Height())

		// ensure the pool progressed (even being a noop)
		require.Equal(t, int64(4), follower.reactor.pool.Height())
	})
}

type combinedModeTestSuite struct {
	t             *testing.T
	blockIngestor *blockIngestorMock
	config        *cfg.Config
	genDoc        *types.GenesisDoc
	privVals      []types.PrivValidator
	logger        log.Logger
}

func newCombinedModeTestSuite(t *testing.T, name string) *combinedModeTestSuite {
	logger := log.TestingLogger()

	config := test.ResetTestRoot(name)
	t.Cleanup(func() { _ = os.RemoveAll(config.RootDir) })

	genDoc, privVals := randGenesisDoc(1, false, 30)

	return &combinedModeTestSuite{
		t:             t,
		blockIngestor: newBlockIngestorMock(t),
		config:        config,
		genDoc:        genDoc,
		privVals:      privVals,
		logger:        logger,
	}
}

type blockIngestorMock struct {
	*p2p.BaseReactor

	t           *testing.T
	mu          sync.Mutex
	onIngest    func(consensus.IngestCandidate) (error, bool)
	storedCalls []consensus.IngestCandidate
}

func newBlockIngestorMock(t *testing.T) *blockIngestorMock {
	m := &blockIngestorMock{
		t: t,
	}
	m.BaseReactor = p2p.NewBaseReactor("consensus-ingestor-mock", m)
	return m
}

func (m *blockIngestorMock) IngestVerifiedBlock(vb consensus.IngestCandidate) (error, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.storedCalls = append(m.storedCalls, vb)

	if m.onIngest == nil {
		return nil, false
	}

	return m.onIngest(vb)
}

func (m *blockIngestorMock) SetOnIngest(onIngest func(vb consensus.IngestCandidate) (error, bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.onIngest = onIngest
}

func (m *blockIngestorMock) Requests() []consensus.IngestCandidate {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]consensus.IngestCandidate, len(m.storedCalls))
	copy(out, m.storedCalls)
	return out
}
