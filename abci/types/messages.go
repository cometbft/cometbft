package types

import (
	"io"

	"github.com/cosmos/gogoproto/proto"

	v3 "github.com/cometbft/cometbft/api/cometbft/abci/v3"
	"github.com/cometbft/cometbft/libs/protoio"
)

const (
	maxMsgSize = 104857600 // 100MB
)

// WriteMessage writes a varint length-delimited protobuf message.
func WriteMessage(msg proto.Message, w io.Writer) error {
	protoWriter := protoio.NewDelimitedWriter(w)
	_, err := protoWriter.WriteMsg(msg)
	return err
}

// ReadMessage reads a varint length-delimited protobuf message.
func ReadMessage(r io.Reader, msg proto.Message) error {
	_, err := protoio.NewDelimitedReader(r, maxMsgSize).ReadMsg(msg)
	return err
}

//----------------------------------------

func ToRequestEcho(message string) *Request {
	return &Request{
		Value: &v3.Request_Echo{&RequestEcho{Message: message}},
	}
}

func ToRequestFlush() *Request {
	return &Request{
		Value: &v3.Request_Flush{&RequestFlush{}},
	}
}

func ToRequestInfo(req *RequestInfo) *Request {
	return &Request{
		Value: &v3.Request_Info{req},
	}
}

func ToRequestCheckTx(req *RequestCheckTx) *Request {
	return &Request{
		Value: &v3.Request_CheckTx{req},
	}
}

func ToRequestCommit() *Request {
	return &Request{
		Value: &v3.Request_Commit{&RequestCommit{}},
	}
}

func ToRequestQuery(req *RequestQuery) *Request {
	return &Request{
		Value: &v3.Request_Query{req},
	}
}

func ToRequestInitChain(req *RequestInitChain) *Request {
	return &Request{
		Value: &v3.Request_InitChain{req},
	}
}

func ToRequestListSnapshots(req *RequestListSnapshots) *Request {
	return &Request{
		Value: &v3.Request_ListSnapshots{req},
	}
}

func ToRequestOfferSnapshot(req *RequestOfferSnapshot) *Request {
	return &Request{
		Value: &v3.Request_OfferSnapshot{req},
	}
}

func ToRequestLoadSnapshotChunk(req *RequestLoadSnapshotChunk) *Request {
	return &Request{
		Value: &v3.Request_LoadSnapshotChunk{req},
	}
}

func ToRequestApplySnapshotChunk(req *RequestApplySnapshotChunk) *Request {
	return &Request{
		Value: &v3.Request_ApplySnapshotChunk{req},
	}
}

func ToRequestPrepareProposal(req *RequestPrepareProposal) *Request {
	return &Request{
		Value: &v3.Request_PrepareProposal{req},
	}
}

func ToRequestProcessProposal(req *RequestProcessProposal) *Request {
	return &Request{
		Value: &v3.Request_ProcessProposal{req},
	}
}

func ToRequestExtendVote(req *RequestExtendVote) *Request {
	return &Request{
		Value: &v3.Request_ExtendVote{req},
	}
}

func ToRequestVerifyVoteExtension(req *RequestVerifyVoteExtension) *Request {
	return &Request{
		Value: &v3.Request_VerifyVoteExtension{req},
	}
}

func ToRequestFinalizeBlock(req *RequestFinalizeBlock) *Request {
	return &Request{
		Value: &v3.Request_FinalizeBlock{req},
	}
}

//----------------------------------------

func ToResponseException(errStr string) *Response {
	return &Response{
		Value: &v3.Response_Exception{&ResponseException{Error: errStr}},
	}
}

func ToResponseEcho(message string) *Response {
	return &Response{
		Value: &v3.Response_Echo{&ResponseEcho{Message: message}},
	}
}

func ToResponseFlush() *Response {
	return &Response{
		Value: &v3.Response_Flush{&ResponseFlush{}},
	}
}

func ToResponseInfo(res *ResponseInfo) *Response {
	return &Response{
		Value: &v3.Response_Info{res},
	}
}

func ToResponseCheckTx(res *ResponseCheckTx) *Response {
	return &Response{
		Value: &v3.Response_CheckTx{res},
	}
}

func ToResponseCommit(res *ResponseCommit) *Response {
	return &Response{
		Value: &v3.Response_Commit{res},
	}
}

func ToResponseQuery(res *ResponseQuery) *Response {
	return &Response{
		Value: &v3.Response_Query{res},
	}
}

func ToResponseInitChain(res *ResponseInitChain) *Response {
	return &Response{
		Value: &v3.Response_InitChain{res},
	}
}

func ToResponseListSnapshots(res *ResponseListSnapshots) *Response {
	return &Response{
		Value: &v3.Response_ListSnapshots{res},
	}
}

func ToResponseOfferSnapshot(res *ResponseOfferSnapshot) *Response {
	return &Response{
		Value: &v3.Response_OfferSnapshot{res},
	}
}

func ToResponseLoadSnapshotChunk(res *ResponseLoadSnapshotChunk) *Response {
	return &Response{
		Value: &v3.Response_LoadSnapshotChunk{res},
	}
}

func ToResponseApplySnapshotChunk(res *ResponseApplySnapshotChunk) *Response {
	return &Response{
		Value: &v3.Response_ApplySnapshotChunk{res},
	}
}

func ToResponsePrepareProposal(res *ResponsePrepareProposal) *Response {
	return &Response{
		Value: &v3.Response_PrepareProposal{res},
	}
}

func ToResponseProcessProposal(res *ResponseProcessProposal) *Response {
	return &Response{
		Value: &v3.Response_ProcessProposal{res},
	}
}

func ToResponseExtendVote(res *ResponseExtendVote) *Response {
	return &Response{
		Value: &v3.Response_ExtendVote{res},
	}
}

func ToResponseVerifyVoteExtension(res *ResponseVerifyVoteExtension) *Response {
	return &Response{
		Value: &v3.Response_VerifyVoteExtension{res},
	}
}

func ToResponseFinalizeBlock(res *ResponseFinalizeBlock) *Response {
	return &Response{
		Value: &v3.Response_FinalizeBlock{res},
	}
}
