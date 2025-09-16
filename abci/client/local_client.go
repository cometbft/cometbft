package abcicli

import (
	"context"

	types "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// localClient is a local implementation of the ABCI Client interface that wraps
// an ABCI Application. It provides thread-safe access to the application by
// serializing all ABCI calls through a global mutex.
//
// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	service.BaseService

	mtx *cmtsync.Mutex
	types.Application
	Callback
}

var _ Client = (*localClient)(nil)

// NewLocalClient creates a local ABCI client that wraps an ABCI Application.
// This client is used when CometBFT and the application run in the same process,
// eliminating the need for network communication. The client ensures thread-safe
// access to the application by serializing all ABCI calls through a global mutex.
// If mtx is nil, a new mutex will be created.
func NewLocalClient(mtx *cmtsync.Mutex, app types.Application) Client {
	if mtx == nil {
		mtx = new(cmtsync.Mutex)
	}
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

// SetResponseCallback sets the callback function that will be invoked for all
// ABCI responses. This is thread-safe and can be called concurrently.
func (app *localClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// CheckTxAsync performs an asynchronous CheckTx operation by calling the
// underlying application and immediately invoking the callback with the result.
// This method is thread-safe and serializes access to the application.
func (app *localClient) CheckTxAsync(ctx context.Context, req *types.RequestCheckTx) (*ReqRes, error) {
	app.mtx.Lock()
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

// callback invokes the global callback function and creates a ReqRes object
// for the request-response pair. The callback is marked as invoked to prevent
// duplicate executions.
func (app *localClient) callback(req *types.Request, res *types.Response) *ReqRes {
	app.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.callbackInvoked = true
	return rr
}

// newLocalReqRes creates a new ReqRes object for local client operations.
// Unlike network clients, the response is immediately available since there
// is no network communication involved.
func newLocalReqRes(req *types.Request, res *types.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.Response = res
	return reqRes
}

//-------------------------------------------------------

// Error returns nil since local clients do not have connection errors.
// Network-related errors are not applicable to in-process communication.
func (app *localClient) Error() error {
	return nil
}

// Flush is a no-op for local clients since there is no network buffer to flush.
// This method exists to satisfy the Client interface.
func (app *localClient) Flush(context.Context) error {
	return nil
}

// Echo returns the same message that was sent, implementing a simple echo service.
// This is used for connection testing in network clients but is trivial for local clients.
func (app *localClient) Echo(_ context.Context, msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *localClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Info(ctx, req)
}

func (app *localClient) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.CheckTx(ctx, req)
}

func (app *localClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Query(ctx, req)
}

func (app *localClient) Commit(ctx context.Context, req *types.RequestCommit) (*types.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Commit(ctx, req)
}

func (app *localClient) InitChain(ctx context.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	app.mtx.Lock()
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
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.PrepareProposal(ctx, req)
}

func (app *localClient) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	app.mtx.Lock()
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
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.FinalizeBlock(ctx, req)
}
