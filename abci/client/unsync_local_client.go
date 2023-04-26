package abcicli

// This is effectively a copy of the fully unsynchronized local client found in
// https://github.com/tendermint/tendermint/pull/9660. Note that this was reverted and replaced by a version
// that has a mutex per instance of the client in https://github.com/tendermint/tendermint/pull/9830.

import (
	"context"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

var _ Client = (*unsyncLocalClient)(nil)

type unsyncLocalClient struct {
	service.BaseService

	types.Application

	// This mutex is exclusively used to protect the callback.
	mtx cmtsync.Mutex
	Callback
}

var _ Client = (*localClient)(nil)

// NewUnsyncLocalClient creates an unsynchronized local client, which will be
// directly calling the methods of the given app.
//
// Unlike NewLocalClient, it does not hold a mutex around the application, so
// it is up to the application to manage its synchronization properly.
func NewUnsyncLocalClient(app types.Application) Client {
	cli := &unsyncLocalClient{
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "unsyncLocalClient", cli)
	return cli
}

func (app *unsyncLocalClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

func (app *unsyncLocalClient) CheckTxAsync(ctx context.Context, req *types.RequestCheckTx) (*ReqRes, error) {
	res, err := app.Application.CheckTx(ctx, req)
	if err != nil {
		return nil, err
	}
	return app.callback(
		types.ToRequestCheckTx(req),
		types.ToResponseCheckTx(res),
	), nil
}

func (app *unsyncLocalClient) callback(req *types.Request, res *types.Response) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	app.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.callbackInvoked = true
	return rr
}

//-------------------------------------------------------

func (app *unsyncLocalClient) Error() error {
	return nil
}

func (app *unsyncLocalClient) Flush(context.Context) error {
	return nil
}

func (app *unsyncLocalClient) Echo(_ context.Context, msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *unsyncLocalClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	return app.Application.Info(ctx, req)
}

func (app *unsyncLocalClient) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	return app.Application.CheckTx(ctx, req)
}

func (app *unsyncLocalClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	return app.Application.Query(ctx, req)
}

func (app *unsyncLocalClient) Commit(ctx context.Context, req *types.RequestCommit) (*types.ResponseCommit, error) {
	return app.Application.Commit(ctx, req)
}

func (app *unsyncLocalClient) InitChain(ctx context.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	return app.Application.InitChain(ctx, req)
}

func (app *unsyncLocalClient) ListSnapshots(ctx context.Context, req *types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	return app.Application.ListSnapshots(ctx, req)
}

func (app *unsyncLocalClient) OfferSnapshot(ctx context.Context, req *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	return app.Application.OfferSnapshot(ctx, req)
}

func (app *unsyncLocalClient) LoadSnapshotChunk(ctx context.Context,
	req *types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error) {
	return app.Application.LoadSnapshotChunk(ctx, req)
}

func (app *unsyncLocalClient) ApplySnapshotChunk(ctx context.Context,
	req *types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error) {
	return app.Application.ApplySnapshotChunk(ctx, req)
}

func (app *unsyncLocalClient) PrepareProposal(ctx context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	return app.Application.PrepareProposal(ctx, req)
}

func (app *unsyncLocalClient) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	return app.Application.ProcessProposal(ctx, req)
}

func (app *unsyncLocalClient) ExtendVote(ctx context.Context, req *types.RequestExtendVote) (*types.ResponseExtendVote, error) {
	return app.Application.ExtendVote(ctx, req)
}

func (app *unsyncLocalClient) VerifyVoteExtension(ctx context.Context, req *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error) {
	return app.Application.VerifyVoteExtension(ctx, req)
}

func (app *unsyncLocalClient) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	return app.Application.FinalizeBlock(ctx, req)
}
