package types

import (
	"io"

	"github.com/cosmos/gogoproto/proto"

	pb "github.com/cometbft/cometbft/api/cometbft/abci/v1beta4"
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
		Value: &pb.Request_Echo{Echo: &RequestEcho{Message: message}},
	}
}

func ToRequestFlush() *Request {
	return &Request{
		Value: &pb.Request_Flush{Flush: &RequestFlush{}},
	}
}

func ToRequestInfo(req *RequestInfo) *Request {
	return &Request{
		Value: &pb.Request_Info{Info: req},
	}
}

func ToRequestCheckTx(req *RequestCheckTx) *Request {
	return &Request{
		Value: &pb.Request_CheckTx{CheckTx: req},
	}
}

func ToRequestCommit() *Request {
	return &Request{
		Value: &pb.Request_Commit{Commit: &RequestCommit{}},
	}
}

func ToRequestQuery(req *RequestQuery) *Request {
	return &Request{
		Value: &pb.Request_Query{Query: req},
	}
}

func ToRequestInitChain(req *RequestInitChain) *Request {
	return &Request{
		Value: &pb.Request_InitChain{InitChain: req},
	}
}

func ToRequestListSnapshots(req *RequestListSnapshots) *Request {
	return &Request{
		Value: &pb.Request_ListSnapshots{ListSnapshots: req},
	}
}

func ToRequestOfferSnapshot(req *RequestOfferSnapshot) *Request {
	return &Request{
		Value: &pb.Request_OfferSnapshot{OfferSnapshot: req},
	}
}

func ToRequestLoadSnapshotChunk(req *RequestLoadSnapshotChunk) *Request {
	return &Request{
		Value: &pb.Request_LoadSnapshotChunk{LoadSnapshotChunk: req},
	}
}

func ToRequestApplySnapshotChunk(req *RequestApplySnapshotChunk) *Request {
	return &Request{
		Value: &pb.Request_ApplySnapshotChunk{ApplySnapshotChunk: req},
	}
}

func ToRequestPrepareProposal(req *RequestPrepareProposal) *Request {
	return &Request{
		Value: &pb.Request_PrepareProposal{PrepareProposal: req},
	}
}

func ToRequestProcessProposal(req *RequestProcessProposal) *Request {
	return &Request{
		Value: &pb.Request_ProcessProposal{ProcessProposal: req},
	}
}

func ToRequestExtendVote(req *RequestExtendVote) *Request {
	return &Request{
		Value: &pb.Request_ExtendVote{ExtendVote: req},
	}
}

func ToRequestVerifyVoteExtension(req *RequestVerifyVoteExtension) *Request {
	return &Request{
		Value: &pb.Request_VerifyVoteExtension{VerifyVoteExtension: req},
	}
}

func ToRequestFinalizeBlock(req *RequestFinalizeBlock) *Request {
	return &Request{
		Value: &pb.Request_FinalizeBlock{FinalizeBlock: req},
	}
}

//----------------------------------------

func ToResponseException(errStr string) *Response {
	return &Response{
		Value: &pb.Response_Exception{Exception: &ResponseException{Error: errStr}},
	}
}

func ToResponseEcho(message string) *Response {
	return &Response{
		Value: &pb.Response_Echo{Echo: &ResponseEcho{Message: message}},
	}
}

func ToResponseFlush() *Response {
	return &Response{
		Value: &pb.Response_Flush{Flush: &ResponseFlush{}},
	}
}

func ToResponseInfo(res *ResponseInfo) *Response {
	return &Response{
		Value: &pb.Response_Info{Info: res},
	}
}

func ToResponseCheckTx(res *ResponseCheckTx) *Response {
	return &Response{
		Value: &pb.Response_CheckTx{CheckTx: res},
	}
}

func ToResponseCommit(res *ResponseCommit) *Response {
	return &Response{
		Value: &pb.Response_Commit{Commit: res},
	}
}

func ToResponseQuery(res *ResponseQuery) *Response {
	return &Response{
		Value: &pb.Response_Query{Query: res},
	}
}

func ToResponseInitChain(res *ResponseInitChain) *Response {
	return &Response{
		Value: &pb.Response_InitChain{InitChain: res},
	}
}

func ToResponseListSnapshots(res *ResponseListSnapshots) *Response {
	return &Response{
		Value: &pb.Response_ListSnapshots{ListSnapshots: res},
	}
}

func ToResponseOfferSnapshot(res *ResponseOfferSnapshot) *Response {
	return &Response{
		Value: &pb.Response_OfferSnapshot{OfferSnapshot: res},
	}
}

func ToResponseLoadSnapshotChunk(res *ResponseLoadSnapshotChunk) *Response {
	return &Response{
		Value: &pb.Response_LoadSnapshotChunk{LoadSnapshotChunk: res},
	}
}

func ToResponseApplySnapshotChunk(res *ResponseApplySnapshotChunk) *Response {
	return &Response{
		Value: &pb.Response_ApplySnapshotChunk{ApplySnapshotChunk: res},
	}
}

func ToResponsePrepareProposal(res *ResponsePrepareProposal) *Response {
	return &Response{
		Value: &pb.Response_PrepareProposal{PrepareProposal: res},
	}
}

func ToResponseProcessProposal(res *ResponseProcessProposal) *Response {
	return &Response{
		Value: &pb.Response_ProcessProposal{ProcessProposal: res},
	}
}

func ToResponseExtendVote(res *ResponseExtendVote) *Response {
	return &Response{
		Value: &pb.Response_ExtendVote{ExtendVote: res},
	}
}

func ToResponseVerifyVoteExtension(res *ResponseVerifyVoteExtension) *Response {
	return &Response{
		Value: &pb.Response_VerifyVoteExtension{VerifyVoteExtension: res},
	}
}

func ToResponseFinalizeBlock(res *ResponseFinalizeBlock) *Response {
	return &Response{
		Value: &pb.Response_FinalizeBlock{FinalizeBlock: res},
	}
}
