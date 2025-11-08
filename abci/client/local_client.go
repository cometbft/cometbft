package abcicli

import (
	"context"
	"sync"
	"time"

	types "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

var (
	// globalMetrics is shared across all local clients to avoid duplicate registration
	globalMetrics     *Metrics
	globalMetricsOnce sync.Once
)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	service.BaseService

	metrics *Metrics
	mtx     *cmtsync.Mutex
	types.Application
	Callback
}

var _ Client = (*localClient)(nil)

// NewLocalClient creates a local client, which wraps the application interface that
// Tendermint as the client will call to the application as the server. The only
// difference, is that the local client has a global mutex which enforces serialization
// of all the ABCI calls from Tendermint to the Application.
func NewLocalClient(mtx *cmtsync.Mutex, app types.Application) Client {
	if mtx == nil {
		mtx = new(cmtsync.Mutex)
	}

	// Use sync.Once to ensure metrics are only registered once across all local clients
	globalMetricsOnce.Do(func() {
		globalMetrics = PrometheusMetrics("cometbft")
	})

	cli := &localClient{
		metrics:     globalMetrics,
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

func (app *localClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

func (app *localClient) CheckTxAsync(ctx context.Context, req *types.RequestCheckTx) (*ReqRes, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "check_tx_async").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	res, err := app.Application.CheckTx(ctx, req)
	if err != nil {
		return nil, err
	}
	return app.callback(
		types.ToRequestCheckTx(req),
		types.ToResponseCheckTx(res),
	), nil
}

func (app *localClient) callback(req *types.Request, res *types.Response) *ReqRes {
	app.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.callbackInvoked = true
	return rr
}

func newLocalReqRes(req *types.Request, res *types.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	return reqRes
}

//-------------------------------------------------------

func (app *localClient) Error() error {
	return nil
}

func (app *localClient) Flush(context.Context) error {
	return nil
}

func (app *localClient) Echo(_ context.Context, msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *localClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "info").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.Info(ctx, req)
}

func (app *localClient) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "check_tx").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.CheckTx(ctx, req)
}

func (app *localClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "query").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.Query(ctx, req)
}

func (app *localClient) Commit(ctx context.Context, req *types.RequestCommit) (*types.ResponseCommit, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "commit").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.Commit(ctx, req)
}

func (app *localClient) InitChain(ctx context.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "init_chain").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.InitChain(ctx, req)
}

func (app *localClient) ListSnapshots(ctx context.Context, req *types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.ListSnapshots(ctx, req)
}

func (app *localClient) OfferSnapshot(ctx context.Context, req *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.OfferSnapshot(ctx, req)
}

func (app *localClient) LoadSnapshotChunk(ctx context.Context,
	req *types.RequestLoadSnapshotChunk,
) (*types.ResponseLoadSnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.LoadSnapshotChunk(ctx, req)
}

func (app *localClient) ApplySnapshotChunk(ctx context.Context,
	req *types.RequestApplySnapshotChunk,
) (*types.ResponseApplySnapshotChunk, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.ApplySnapshotChunk(ctx, req)
}

func (app *localClient) PrepareProposal(ctx context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "prepare_proposal").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.PrepareProposal(ctx, req)
}

func (app *localClient) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "process_proposal").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.ProcessProposal(ctx, req)
}

func (app *localClient) ExtendVote(ctx context.Context, req *types.RequestExtendVote) (*types.ResponseExtendVote, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.ExtendVote(ctx, req)
}

func (app *localClient) VerifyVoteExtension(ctx context.Context, req *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.VerifyVoteExtension(ctx, req)
}

func (app *localClient) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	start := time.Now()
	app.mtx.Lock()
	app.metrics.LockWaitSeconds.With("method", "finalize_block").Observe(time.Since(start).Seconds())
	defer app.mtx.Unlock()

	return app.Application.FinalizeBlock(ctx, req)
}
