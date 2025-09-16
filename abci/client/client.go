package abcicli

import (
	"context"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

//go:generate ../../scripts/mockery_generate.sh Client

// Client defines the interface for an ABCI (Application Blockchain Interface) client.
// ABCI is the interface between CometBFT and the application, allowing the application
// to process transactions and maintain state.
//
// NOTE: these are client errors, eg. ABCI socket connectivity issues.
// Application-related errors are reflected in response via ABCI error codes
// and (potentially) error response.
type Client interface {
	service.Service
	types.Application

	// TODO: remove as each method now returns an error
	Error() error
	// TODO: remove as this is not implemented
	Flush(context.Context) error
	Echo(context.Context, string) (*types.ResponseEcho, error)

	// FIXME: All other operations are run synchronously and rely
	// on the caller to dictate concurrency (i.e. run a go routine),
	// with the exception of `CheckTxAsync` which we maintain
	// for the v0 mempool. We should explore refactoring the
	// mempool to remove this vestige behavior.
	SetResponseCallback(Callback)
	CheckTxAsync(context.Context, *types.RequestCheckTx) (*ReqRes, error)
}

//----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// Supported transport types are "socket" (Unix domain socket or TCP) and "grpc".
// It returns an error if the transport is not supported.
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		client = NewSocketClient(addr, mustConnect)
	case "grpc":
		client = NewGRPCClient(addr, mustConnect)
	default:
		err = fmt.Errorf("unknown abci transport %s", transport)
	}
	return
}

type Callback func(*types.Request, *types.Response)

// ReqRes represents a request-response pair for asynchronous ABCI operations.
// It provides synchronization mechanisms and callback support for handling
// responses when they become available.
type ReqRes struct {
	*types.Request
	*sync.WaitGroup
	*types.Response // Not set atomically, so be sure to use WaitGroup.

	mtx cmtsync.Mutex

	// callbackInvoked as a variable to track if the callback was already
	// invoked during the regular execution of the request. This variable
	// allows clients to set the callback simultaneously without potentially
	// invoking the callback twice by accident, once when 'SetCallback' is
	// called and once during the normal request.
	callbackInvoked bool
	cb              func(*types.Response) // A single callback that may be set.
}

func NewReqRes(req *types.Request) *ReqRes {
	return &ReqRes{
		Request:   req,
		WaitGroup: waitGroup1(),
		Response:  nil,

		callbackInvoked: false,
		cb:              nil,
	}
}

// SetCallback sets the callback function for this request-response pair.
// If the response is already available, the callback will be invoked immediately.
// Note: only one callback is supported per ReqRes instance.
func (r *ReqRes) SetCallback(cb func(res *types.Response)) {
	r.mtx.Lock()

	if r.callbackInvoked {
		r.mtx.Unlock()
		cb(r.Response)
		return
	}

	r.cb = cb
	r.mtx.Unlock()
}

// InvokeCallback invokes a thread-safe execution of the configured callback
// if one is set. This method marks the callback as invoked to prevent
// duplicate executions.
func (r *ReqRes) InvokeCallback() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.cb != nil {
		r.cb(r.Response)
	}
	r.callbackInvoked = true
}

// GetCallback returns the configured callback function, which may be nil.
// Note: it is not safe to concurrently call this method when the request
// is marked as done and SetCallback is called, as this could invoke the
// callback twice and create a race condition.
//
// ref: https://github.com/tendermint/tendermint/issues/5439
func (r *ReqRes) GetCallback() func(*types.Response) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.cb
}

// waitGroup1 creates a new WaitGroup with a count of 1, ready for use
// in ReqRes to wait for a single response.
func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return
}
