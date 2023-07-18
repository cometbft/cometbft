package proxy

import (
	"fmt"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/abci/types"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	e2e "github.com/cometbft/cometbft/test/e2e/app"
)

//go:generate ../scripts/mockery_generate.sh ClientCreator

// ClientCreator creates new ABCI clients based on the intended use of the client.
type ClientCreator interface {
	// NewABCIConsensusClient creates an ABCI client for handling
	// consensus-related queries.
	NewABCIConsensusClient() (abcicli.Client, error)
	// NewABCIMempoolClient creates an ABCI client for handling mempool-related
	// queries.
	NewABCIMempoolClient() (abcicli.Client, error)
	// NewABCIQueryClient creates an ABCI client for handling
	// query/info-related queries.
	NewABCIQueryClient() (abcicli.Client, error)
	// NewABCISnapshotClient creates an ABCI client for handling
	// snapshot-related queries.
	NewABCISnapshotClient() (abcicli.Client, error)
}

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type localClientCreator struct {
	mtx *cmtsync.Mutex
	app types.Application
}

// NewLocalClientCreator returns a [ClientCreator] for the given app, which
// will be running locally.
//
// Maintains a single mutex over all new clients created with NewABCIClient.
func NewLocalClientCreator(app types.Application) ClientCreator {
	return &localClientCreator{
		mtx: new(cmtsync.Mutex),
		app: app,
	}
}

// NewABCIConsensusClient implements ClientCreator.
func (l *localClientCreator) NewABCIConsensusClient() (abcicli.Client, error) {
	return l.newABCIClient()
}

// NewABCIMempoolClient implements ClientCreator.
func (l *localClientCreator) NewABCIMempoolClient() (abcicli.Client, error) {
	return l.newABCIClient()
}

// NewABCIQueryClient implements ClientCreator.
func (l *localClientCreator) NewABCIQueryClient() (abcicli.Client, error) {
	return l.newABCIClient()
}

// NewABCISnapshotClient implements ClientCreator.
func (l *localClientCreator) NewABCISnapshotClient() (abcicli.Client, error) {
	return l.newABCIClient()
}

func (l *localClientCreator) newABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(l.mtx, l.app), nil
}

//-------------------------------------------------------------------------
// connection-synchronized local client uses a mutex per "connection" on an
// in-process app

type connSyncLocalClientCreator struct {
	app types.Application
}

// NewConnSyncLocalClientCreator returns a [ClientCreator] for the given app,
// which will be running locally.
//
// Returned clients are "connection-synchronized", meaning every client created
// will have its own mutex, allowing interaction across clients to take place
// in parallel, but every call on a particular client is effectively
// serialized.
func NewConnSyncLocalClientCreator(app types.Application) ClientCreator {
	return &connSyncLocalClientCreator{
		app: app,
	}
}

// NewABCIConsensusClient implements ClientCreator.
func (c *connSyncLocalClientCreator) NewABCIConsensusClient() (abcicli.Client, error) {
	return c.newABCIClient()
}

// NewABCIMempoolClient implements ClientCreator.
func (c *connSyncLocalClientCreator) NewABCIMempoolClient() (abcicli.Client, error) {
	return c.newABCIClient()
}

// NewABCIQueryClient implements ClientCreator.
func (c *connSyncLocalClientCreator) NewABCIQueryClient() (abcicli.Client, error) {
	return c.newABCIClient()
}

// NewABCISnapshotClient implements ClientCreator.
func (c *connSyncLocalClientCreator) NewABCISnapshotClient() (abcicli.Client, error) {
	return c.newABCIClient()
}

func (c *connSyncLocalClientCreator) newABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(nil, c.app), nil
}

//---------------------------------------------------------------
// remote proxy opens new connections to an external app process

type remoteClientCreator struct {
	addr        string
	transport   string
	mustConnect bool
}

// NewRemoteClientCreator returns a ClientCreator for the given address (e.g.
// "192.168.0.1") and transport (e.g. "tcp"). Set mustConnect to true if you
// want the client to connect before reporting success.
func NewRemoteClientCreator(addr, transport string, mustConnect bool) ClientCreator {
	return &remoteClientCreator{
		addr:        addr,
		transport:   transport,
		mustConnect: mustConnect,
	}
}

// NewABCIConsensusClient implements ClientCreator.
func (r *remoteClientCreator) NewABCIConsensusClient() (abcicli.Client, error) {
	return r.newABCIClient()
}

// NewABCIMempoolClient implements ClientCreator.
func (r *remoteClientCreator) NewABCIMempoolClient() (abcicli.Client, error) {
	return r.newABCIClient()
}

// NewABCIQueryClient implements ClientCreator.
func (r *remoteClientCreator) NewABCIQueryClient() (abcicli.Client, error) {
	return r.newABCIClient()
}

// NewABCISnapshotClient implements ClientCreator.
func (r *remoteClientCreator) NewABCISnapshotClient() (abcicli.Client, error) {
	return r.newABCIClient()
}

func (r *remoteClientCreator) newABCIClient() (abcicli.Client, error) {
	remoteApp, err := abcicli.NewClient(r.addr, r.transport, r.mustConnect)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	return remoteApp, nil
}

// DefaultClientCreator returns a default [ClientCreator], which will create a
// local client if addr is one of "kvstore", "persistent_kvstore", "e2e",
// "noop".
//
// Otherwise a remote client will be created.
func DefaultClientCreator(addr, transport, dbDir string) ClientCreator {
	switch addr {
	case "kvstore":
		return NewLocalClientCreator(kvstore.NewInMemoryApplication())
	case "persistent_kvstore":
		return NewLocalClientCreator(kvstore.NewPersistentApplication(dbDir))
	case "e2e":
		app, err := e2e.NewApplication(e2e.DefaultConfig(dbDir))
		if err != nil {
			panic(err)
		}
		return NewLocalClientCreator(app)
	case "noop":
		return NewLocalClientCreator(types.NewBaseApplication())
	default:
		mustConnect := false // loop retrying
		return NewRemoteClientCreator(addr, transport, mustConnect)
	}
}
