// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	abcicli "github.com/cometbft/cometbft/abci/client"

	log "github.com/cometbft/cometbft/libs/log"

	mock "github.com/stretchr/testify/mock"

	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// ApplySnapshotChunk provides a mock function with given fields: ctx, req
func (_m *Client) ApplySnapshotChunk(ctx context.Context, req *v1.ApplySnapshotChunkRequest) (*v1.ApplySnapshotChunkResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for ApplySnapshotChunk")
	}

	var r0 *v1.ApplySnapshotChunkResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ApplySnapshotChunkRequest) (*v1.ApplySnapshotChunkResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ApplySnapshotChunkRequest) *v1.ApplySnapshotChunkResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ApplySnapshotChunkResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.ApplySnapshotChunkRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CheckTx provides a mock function with given fields: ctx, req
func (_m *Client) CheckTx(ctx context.Context, req *v1.CheckTxRequest) (*v1.CheckTxResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for CheckTx")
	}

	var r0 *v1.CheckTxResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CheckTxRequest) (*v1.CheckTxResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CheckTxRequest) *v1.CheckTxResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.CheckTxResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.CheckTxRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CheckTxAsync provides a mock function with given fields: ctx, req
func (_m *Client) CheckTxAsync(ctx context.Context, req *v1.CheckTxRequest) (*abcicli.ReqRes, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for CheckTxAsync")
	}

	var r0 *abcicli.ReqRes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CheckTxRequest) (*abcicli.ReqRes, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CheckTxRequest) *abcicli.ReqRes); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcicli.ReqRes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.CheckTxRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Commit provides a mock function with given fields: ctx, req
func (_m *Client) Commit(ctx context.Context, req *v1.CommitRequest) (*v1.CommitResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 *v1.CommitResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CommitRequest) (*v1.CommitResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.CommitRequest) *v1.CommitResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.CommitResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.CommitRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Echo provides a mock function with given fields: ctx, echo
func (_m *Client) Echo(ctx context.Context, echo string) (*v1.EchoResponse, error) {
	ret := _m.Called(ctx, echo)

	if len(ret) == 0 {
		panic("no return value specified for Echo")
	}

	var r0 *v1.EchoResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1.EchoResponse, error)); ok {
		return rf(ctx, echo)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1.EchoResponse); ok {
		r0 = rf(ctx, echo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.EchoResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, echo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Error provides a mock function with given fields:
func (_m *Client) Error() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Error")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExtendVote provides a mock function with given fields: ctx, req
func (_m *Client) ExtendVote(ctx context.Context, req *v1.ExtendVoteRequest) (*v1.ExtendVoteResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for ExtendVote")
	}

	var r0 *v1.ExtendVoteResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ExtendVoteRequest) (*v1.ExtendVoteResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ExtendVoteRequest) *v1.ExtendVoteResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ExtendVoteResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.ExtendVoteRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FinalizeBlock provides a mock function with given fields: ctx, req
func (_m *Client) FinalizeBlock(ctx context.Context, req *v1.FinalizeBlockRequest) (*v1.FinalizeBlockResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for FinalizeBlock")
	}

	var r0 *v1.FinalizeBlockResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.FinalizeBlockRequest) (*v1.FinalizeBlockResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.FinalizeBlockRequest) *v1.FinalizeBlockResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.FinalizeBlockResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.FinalizeBlockRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Flush provides a mock function with given fields: ctx
func (_m *Client) Flush(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Flush")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Info provides a mock function with given fields: ctx, req
func (_m *Client) Info(ctx context.Context, req *v1.InfoRequest) (*v1.InfoResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for Info")
	}

	var r0 *v1.InfoResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.InfoRequest) (*v1.InfoResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.InfoRequest) *v1.InfoResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.InfoResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.InfoRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InitChain provides a mock function with given fields: ctx, req
func (_m *Client) InitChain(ctx context.Context, req *v1.InitChainRequest) (*v1.InitChainResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for InitChain")
	}

	var r0 *v1.InitChainResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.InitChainRequest) (*v1.InitChainResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.InitChainRequest) *v1.InitChainResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.InitChainResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.InitChainRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsRunning provides a mock function with given fields:
func (_m *Client) IsRunning() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsRunning")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ListSnapshots provides a mock function with given fields: ctx, req
func (_m *Client) ListSnapshots(ctx context.Context, req *v1.ListSnapshotsRequest) (*v1.ListSnapshotsResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for ListSnapshots")
	}

	var r0 *v1.ListSnapshotsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ListSnapshotsRequest) (*v1.ListSnapshotsResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ListSnapshotsRequest) *v1.ListSnapshotsResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ListSnapshotsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.ListSnapshotsRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadSnapshotChunk provides a mock function with given fields: ctx, req
func (_m *Client) LoadSnapshotChunk(ctx context.Context, req *v1.LoadSnapshotChunkRequest) (*v1.LoadSnapshotChunkResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for LoadSnapshotChunk")
	}

	var r0 *v1.LoadSnapshotChunkResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.LoadSnapshotChunkRequest) (*v1.LoadSnapshotChunkResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.LoadSnapshotChunkRequest) *v1.LoadSnapshotChunkResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.LoadSnapshotChunkResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.LoadSnapshotChunkRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OfferSnapshot provides a mock function with given fields: ctx, req
func (_m *Client) OfferSnapshot(ctx context.Context, req *v1.OfferSnapshotRequest) (*v1.OfferSnapshotResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for OfferSnapshot")
	}

	var r0 *v1.OfferSnapshotResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.OfferSnapshotRequest) (*v1.OfferSnapshotResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.OfferSnapshotRequest) *v1.OfferSnapshotResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.OfferSnapshotResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.OfferSnapshotRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OnReset provides a mock function with given fields:
func (_m *Client) OnReset() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OnReset")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStart provides a mock function with given fields:
func (_m *Client) OnStart() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OnStart")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStop provides a mock function with given fields:
func (_m *Client) OnStop() {
	_m.Called()
}

// PrepareProposal provides a mock function with given fields: ctx, req
func (_m *Client) PrepareProposal(ctx context.Context, req *v1.PrepareProposalRequest) (*v1.PrepareProposalResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for PrepareProposal")
	}

	var r0 *v1.PrepareProposalResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.PrepareProposalRequest) (*v1.PrepareProposalResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.PrepareProposalRequest) *v1.PrepareProposalResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.PrepareProposalResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.PrepareProposalRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProcessProposal provides a mock function with given fields: ctx, req
func (_m *Client) ProcessProposal(ctx context.Context, req *v1.ProcessProposalRequest) (*v1.ProcessProposalResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for ProcessProposal")
	}

	var r0 *v1.ProcessProposalResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ProcessProposalRequest) (*v1.ProcessProposalResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ProcessProposalRequest) *v1.ProcessProposalResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ProcessProposalResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.ProcessProposalRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Query provides a mock function with given fields: ctx, req
func (_m *Client) Query(ctx context.Context, req *v1.QueryRequest) (*v1.QueryResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for Query")
	}

	var r0 *v1.QueryResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.QueryRequest) (*v1.QueryResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.QueryRequest) *v1.QueryResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.QueryResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.QueryRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Quit provides a mock function with given fields:
func (_m *Client) Quit() <-chan struct{} {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Quit")
	}

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Reset provides a mock function with given fields:
func (_m *Client) Reset() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Reset")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetLogger provides a mock function with given fields: l
func (_m *Client) SetLogger(l log.Logger) {
	_m.Called(l)
}

// SetResponseCallback provides a mock function with given fields: cb
func (_m *Client) SetResponseCallback(cb abcicli.Callback) {
	_m.Called(cb)
}

// Start provides a mock function with given fields:
func (_m *Client) Start() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *Client) Stop() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *Client) String() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for String")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// VerifyVoteExtension provides a mock function with given fields: ctx, req
func (_m *Client) VerifyVoteExtension(ctx context.Context, req *v1.VerifyVoteExtensionRequest) (*v1.VerifyVoteExtensionResponse, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for VerifyVoteExtension")
	}

	var r0 *v1.VerifyVoteExtensionResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VerifyVoteExtensionRequest) (*v1.VerifyVoteExtensionResponse, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.VerifyVoteExtensionRequest) *v1.VerifyVoteExtensionResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.VerifyVoteExtensionResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.VerifyVoteExtensionRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
