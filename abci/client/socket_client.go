package abcicli

import (
	"bufio"
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/libs/timer"
)

const (
	// reqQueueSize is the buffer size for the request queue.
	// This allows up to 256 pending requests before blocking.
	reqQueueSize = 256 // TODO make configurable
	// flushThrottleMS is the maximum time to wait before auto-flushing
	// the request buffer to the server.
	flushThrottleMS = 20 // Don't wait longer than...
)

// socketClient is the client side implementation of the Tendermint
// Socket Protocol (TSP). It is used by CometBFT to communicate with
// an out-of-process ABCI application running the socketServer.
//
// This implementation is goroutine-safe. All calls are serialized to the server
// through a buffered queue. The socketClient tracks responses and expects them
// to respect the order of the requests sent, ensuring proper request-response
// matching for reliable communication.
type socketClient struct {
	service.BaseService

	addr        string
	mustConnect bool
	conn        net.Conn

	// reqQueue buffers requests before they are sent to the server
	reqQueue chan *ReqRes
	// flushTimer controls automatic flushing of the request buffer
	flushTimer *timer.ThrottleTimer

	mtx sync.Mutex
	err error
	// reqSent tracks requests that have been sent and are waiting for responses
	reqSent *list.List
	// resCb is the global callback function called for all responses
	resCb func(*types.Request, *types.Response)
}

var _ Client = (*socketClient)(nil)

// NewSocketClient creates a new socket client that connects to the specified address.
// If mustConnect is true, the client will return an error immediately if it fails
// to connect. If false, it will continue retrying the connection in the background.
func NewSocketClient(addr string, mustConnect bool) Client {
	cli := &socketClient{
		reqQueue:    make(chan *ReqRes, reqQueueSize),
		flushTimer:  timer.NewThrottleTimer("socketClient", flushThrottleMS),
		mustConnect: mustConnect,

		addr:    addr,
		reqSent: list.New(),
		resCb:   nil,
	}
	cli.BaseService = *service.NewBaseService(nil, "socketClient", cli)
	return cli
}

// OnStart implements Service by establishing a connection to the server and
// starting the request sending and response receiving goroutines.
func (cli *socketClient) OnStart() error {
	var (
		err  error
		conn net.Conn
	)

	// Connection retry loop - attempts to establish socket connection
	for {
		conn, err = cmtnet.Connect(cli.addr)
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.socketClient failed to connect to %v.  Retrying after %vs...",
				cli.addr, dialRetryIntervalSeconds), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue
		}
		cli.conn = conn

		// Start the request sending and response receiving goroutines
		go cli.sendRequestsRoutine(conn)
		go cli.recvResponseRoutine(conn)

		return nil
	}
}

// OnStop implements Service by closing the connection and cleaning up all
// pending requests and timers.
func (cli *socketClient) OnStop() {
	if cli.conn != nil {
		cli.conn.Close()
	}

	cli.flushQueue()
	cli.flushTimer.Stop()
}

// Error returns any error that caused the client to stop unexpectedly.
// This includes connection errors, protocol errors, or other failures.
func (cli *socketClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

//----------------------------------------

// SetResponseCallback sets a global callback function that will be executed
// for each response received from the server. This callback is called for
// all successful responses, including internally generated flush responses.
//
// NOTE: callback may receive internally generated flush responses.
func (cli *socketClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

// CheckTxAsync performs an asynchronous CheckTx operation by queuing the
// request and returning immediately. The response will be available through
// the returned ReqRes object.
func (cli *socketClient) CheckTxAsync(ctx context.Context, req *types.RequestCheckTx) (*ReqRes, error) {
	return cli.queueRequest(ctx, types.ToRequestCheckTx(req))
}

//----------------------------------------

// sendRequestsRoutine is a goroutine that continuously sends requests from
// the queue to the server. It handles buffering, flushing, and error conditions.
func (cli *socketClient) sendRequestsRoutine(conn io.Writer) {
	w := bufio.NewWriter(conn)
	for {
		select {
		case reqres := <-cli.reqQueue:
			// N.B. We must track the request before sending it, otherwise the
			// server may reply before we track it, and the receiver will fail for an
			// unsolicited reply.
			cli.trackRequest(reqres)

			err := types.WriteMessage(reqres.Request, w)
			if err != nil {
				cli.stopForError(fmt.Errorf("write to buffer: %w", err))
				return
			}

			// If it's a flush request, immediately flush the buffer to the server
			if _, ok := reqres.Request.Value.(*types.Request_Flush); ok {
				err = w.Flush()
				if err != nil {
					cli.stopForError(fmt.Errorf("flush buffer: %w", err))
					return
				}
			}
		case <-cli.flushTimer.Ch: // Auto-flush timer expired
			select {
			case cli.reqQueue <- NewReqRes(types.ToRequestFlush()):
			default:
				// Queue is full, skip this flush attempt
			}
		case <-cli.Quit():
			return
		}
	}
}

// recvResponseRoutine is a goroutine that continuously reads responses from
// the server and processes them. It handles message parsing and error conditions.
func (cli *socketClient) recvResponseRoutine(conn io.Reader) {
	r := bufio.NewReader(conn)
	for {
		if !cli.IsRunning() {
			return
		}

		res := &types.Response{}
		err := types.ReadMessage(r, res)
		if err != nil {
			cli.stopForError(fmt.Errorf("read message: %w", err))
			return
		}

		switch r := res.Value.(type) {
		case *types.Response_Exception: // Application responded with an error
			// XXX After setting cli.err, release waiters (e.g. reqres.Done())
			cli.stopForError(errors.New(r.Exception.Error))
			return
		default:
			// Process the normal response and match it with the corresponding request
			err := cli.didRecvResponse(res)
			if err != nil {
				cli.stopForError(err)
				return
			}
		}
	}
}

// trackRequest adds a request to the list of sent requests waiting for responses.
// This is used to match incoming responses with their corresponding requests.
func (cli *socketClient) trackRequest(reqres *ReqRes) {
	// N.B. We must NOT hold the client state lock while checking this, or we
	// may deadlock with shutdown.
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.reqSent.PushBack(reqres)
}

// didRecvResponse processes a received response by matching it with the
// corresponding request and notifying any registered callbacks.
func (cli *socketClient) didRecvResponse(res *types.Response) error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Get the first pending request (FIFO order)
	next := cli.reqSent.Front()
	if next == nil {
		return fmt.Errorf("unexpected response %T when no call was made", res.Value)
	}

	reqres := next.Value.(*ReqRes)
	// Verify that the response type matches the request type
	if !resMatchesReq(reqres.Request, res) {
		return fmt.Errorf("unexpected response %T to the request %T", res.Value, reqres.Request.Value)
	}

	reqres.Response = res
	reqres.Done()            // Release any goroutines waiting for this response
	cli.reqSent.Remove(next) // Remove the completed request from the list

	// Notify global response callback if set
	if cli.resCb != nil {
		cli.resCb(reqres.Request, res)
	}

	// Notify request-specific callback if set
	//
	// NOTE: It is possible this callback isn't set on the reqres object at this
	// point, in which case it will be called later when it is set.
	reqres.InvokeCallback()

	return nil
}

//----------------------------------------

// Flush sends a flush request to the server and waits for the response.
// This ensures all pending requests are transmitted immediately.
func (cli *socketClient) Flush(ctx context.Context) error {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestFlush())
	if err != nil {
		return err
	}
	reqRes.Wait()
	return nil
}

func (cli *socketClient) Echo(ctx context.Context, msg string) (*types.ResponseEcho, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestEcho(msg))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetEcho(), cli.Error()
}

func (cli *socketClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestInfo(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetInfo(), cli.Error()
}

func (cli *socketClient) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestCheckTx(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetCheckTx(), cli.Error()
}

func (cli *socketClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestQuery(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetQuery(), cli.Error()
}

func (cli *socketClient) Commit(ctx context.Context, _ *types.RequestCommit) (*types.ResponseCommit, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestCommit())
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetCommit(), cli.Error()
}

func (cli *socketClient) InitChain(ctx context.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestInitChain(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetInitChain(), cli.Error()
}

func (cli *socketClient) ListSnapshots(ctx context.Context, req *types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestListSnapshots(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetListSnapshots(), cli.Error()
}

func (cli *socketClient) OfferSnapshot(ctx context.Context, req *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestOfferSnapshot(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetOfferSnapshot(), cli.Error()
}

func (cli *socketClient) LoadSnapshotChunk(ctx context.Context, req *types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestLoadSnapshotChunk(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetLoadSnapshotChunk(), cli.Error()
}

func (cli *socketClient) ApplySnapshotChunk(ctx context.Context, req *types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestApplySnapshotChunk(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetApplySnapshotChunk(), cli.Error()
}

func (cli *socketClient) PrepareProposal(ctx context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestPrepareProposal(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetPrepareProposal(), cli.Error()
}

func (cli *socketClient) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestProcessProposal(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetProcessProposal(), cli.Error()
}

func (cli *socketClient) ExtendVote(ctx context.Context, req *types.RequestExtendVote) (*types.ResponseExtendVote, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestExtendVote(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetExtendVote(), cli.Error()
}

func (cli *socketClient) VerifyVoteExtension(ctx context.Context, req *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestVerifyVoteExtension(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetVerifyVoteExtension(), cli.Error()
}

func (cli *socketClient) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestFinalizeBlock(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetFinalizeBlock(), cli.Error()
}

// queueRequest adds a request to the send queue and manages the auto-flush timer.
// It returns a ReqRes object that can be used to wait for the response.
func (cli *socketClient) queueRequest(ctx context.Context, req *types.Request) (*ReqRes, error) {
	reqres := NewReqRes(req)

	// TODO: set cli.err if reqQueue times out
	select {
	case cli.reqQueue <- reqres:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Manage auto-flush timer based on request type
	switch req.Value.(type) {
	case *types.Request_Flush:
		// Flush requests disable auto-flush since they handle it explicitly
		cli.flushTimer.Unset()
	default:
		// Other requests enable auto-flush to ensure timely transmission
		cli.flushTimer.Set()
	}

	return reqres, nil
}

// flushQueue marks all pending requests as complete and discards them.
// This is called during shutdown to clean up any remaining requests.
func (cli *socketClient) flushQueue() {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Mark all in-flight messages as resolved (they will get cli.Error())
	for req := cli.reqSent.Front(); req != nil; req = req.Next() {
		reqres := req.Value.(*ReqRes)
		reqres.Done()
	}

	// Mark all queued messages as resolved
LOOP:
	for {
		select {
		case reqres := <-cli.reqQueue:
			reqres.Done()
		default:
			break LOOP
		}
	}
}

//----------------------------------------

// resMatchesReq verifies that a response type matches the corresponding request type.
// This ensures proper request-response pairing in the socket protocol.
func resMatchesReq(req *types.Request, res *types.Response) (ok bool) {
	switch req.Value.(type) {
	case *types.Request_Echo:
		_, ok = res.Value.(*types.Response_Echo)
	case *types.Request_Flush:
		_, ok = res.Value.(*types.Response_Flush)
	case *types.Request_Info:
		_, ok = res.Value.(*types.Response_Info)
	case *types.Request_CheckTx:
		_, ok = res.Value.(*types.Response_CheckTx)
	case *types.Request_Commit:
		_, ok = res.Value.(*types.Response_Commit)
	case *types.Request_Query:
		_, ok = res.Value.(*types.Response_Query)
	case *types.Request_InitChain:
		_, ok = res.Value.(*types.Response_InitChain)
	case *types.Request_ApplySnapshotChunk:
		_, ok = res.Value.(*types.Response_ApplySnapshotChunk)
	case *types.Request_LoadSnapshotChunk:
		_, ok = res.Value.(*types.Response_LoadSnapshotChunk)
	case *types.Request_ListSnapshots:
		_, ok = res.Value.(*types.Response_ListSnapshots)
	case *types.Request_OfferSnapshot:
		_, ok = res.Value.(*types.Response_OfferSnapshot)
	case *types.Request_ExtendVote:
		_, ok = res.Value.(*types.Response_ExtendVote)
	case *types.Request_VerifyVoteExtension:
		_, ok = res.Value.(*types.Response_VerifyVoteExtension)
	case *types.Request_PrepareProposal:
		_, ok = res.Value.(*types.Response_PrepareProposal)
	case *types.Request_ProcessProposal:
		_, ok = res.Value.(*types.Response_ProcessProposal)
	case *types.Request_FinalizeBlock:
		_, ok = res.Value.(*types.Response_FinalizeBlock)
	}
	return ok
}

// stopForError stops the client due to an error and records the error state.
// This is called when connection errors, protocol errors, or other failures occur.
func (cli *socketClient) stopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.socketClient for error: %v", err.Error()))
	if err := cli.Stop(); err != nil {
		cli.Logger.Error("Error stopping abci.socketClient", "err", err)
	}
}
