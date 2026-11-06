package node

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	proxymocks "github.com/cometbft/cometbft/proxy/mocks"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/statesync"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

// stubStateProvider is a no-op statesync.StateProvider. It is never actually
// invoked in this test (no peer ever supplies a snapshot), it only needs to
// be non-nil so that performStateSync skips constructing a real
// light-client-backed StateProvider (which requires >= 2 configured RPC
// servers).
type stubStateProvider struct{}

func (stubStateProvider) AppHash(context.Context, uint64) ([]byte, error) { return nil, nil }
func (stubStateProvider) Commit(context.Context, uint64) (*types.Commit, error) {
	return nil, nil
}
func (stubStateProvider) State(context.Context, uint64) (sm.State, error) {
	return sm.State{}, nil
}

// stubBlockSyncReactor is a minimal p2p.Reactor that also implements the
// unexported blocksyncEnabler interface required by performStateSync, so
// that the type assertion in performStateSync succeeds without needing a
// real blocksync.Reactor.
type stubBlockSyncReactor struct {
	p2p.BaseReactor
}

func newStubBlockSyncReactor() *stubBlockSyncReactor {
	r := &stubBlockSyncReactor{}
	r.BaseReactor = *p2p.NewBaseReactor("StubBlockSync", r)
	return r
}

func (*stubBlockSyncReactor) Enable(sm.State) error { return nil }

// TestPerformStateSyncStopsWithReactor checks that the goroutine started by
// performStateSync ends when the state sync reactor is stopped (which is what
// Node.OnStop does through the switch) and that Node.OnStop can wait for it
// before it closes the stores.
func TestPerformStateSyncStopsWithReactor(t *testing.T) {
	scfg := cfg.DefaultStateSyncConfig()
	scfg.Enable = true
	scfg.DiscoveryTime = 5 * time.Second // minimum config.ValidateBasic() accepts

	reactor := statesync.NewReactor(*scfg, &proxymocks.AppConnSnapshot{}, &proxymocks.AppConnQuery{}, statesync.NopMetrics())
	reactor.SetLogger(log.TestingLogger())

	sw := p2p.MakeSwitch(cfg.DefaultP2PConfig(), 0, func(_ int, sw *p2p.Switch) *p2p.Switch { return sw })
	reactor.SetSwitch(sw)
	require.NoError(t, reactor.Start())

	nodeCfg := cfg.DefaultConfig()
	nodeCfg.StateSync = scfg

	n := &Node{
		config:            nodeCfg,
		stateSyncReactor:  reactor,
		stateSyncProvider: stubStateProvider{},
		stateSyncGenesis:  sm.State{},
		bcReactor:         newStubBlockSyncReactor(),
		stateStore:        sm.NewStore(dbm.NewMemDB(), sm.StoreOptions{}),
		blockStore:        store.NewBlockStore(dbm.NewMemDB()),
	}

	require.NoError(t, n.performStateSync())
	require.NoError(t, reactor.Stop())

	done := make(chan struct{})
	go func() {
		n.stateSyncWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("state sync goroutine is still running after the state sync reactor was stopped")
	}
}
