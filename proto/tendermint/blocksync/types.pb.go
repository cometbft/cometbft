// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/blocksync/types.proto

package blocksync

import (
	fmt "fmt"
	types "github.com/cometbft/cometbft/proto/tendermint/types"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// BlockRequest requests a block for a specific height
type BlockRequest struct {
	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *BlockRequest) Reset()         { *m = BlockRequest{} }
func (m *BlockRequest) String() string { return proto.CompactTextString(m) }
func (*BlockRequest) ProtoMessage()    {}
func (*BlockRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{0}
}
func (m *BlockRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BlockRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BlockRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BlockRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockRequest.Merge(m, src)
}
func (m *BlockRequest) XXX_Size() int {
	return m.Size()
}
func (m *BlockRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockRequest proto.InternalMessageInfo

func (m *BlockRequest) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

// NoBlockResponse informs the node that the peer does not have block at the requested height
type NoBlockResponse struct {
	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *NoBlockResponse) Reset()         { *m = NoBlockResponse{} }
func (m *NoBlockResponse) String() string { return proto.CompactTextString(m) }
func (*NoBlockResponse) ProtoMessage()    {}
func (*NoBlockResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{1}
}
func (m *NoBlockResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NoBlockResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NoBlockResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NoBlockResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NoBlockResponse.Merge(m, src)
}
func (m *NoBlockResponse) XXX_Size() int {
	return m.Size()
}
func (m *NoBlockResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_NoBlockResponse.DiscardUnknown(m)
}

var xxx_messageInfo_NoBlockResponse proto.InternalMessageInfo

func (m *NoBlockResponse) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

// BlockResponse returns block to the requested
type BlockResponse struct {
	Block *types.Block `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
}

func (m *BlockResponse) Reset()         { *m = BlockResponse{} }
func (m *BlockResponse) String() string { return proto.CompactTextString(m) }
func (*BlockResponse) ProtoMessage()    {}
func (*BlockResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{2}
}
func (m *BlockResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BlockResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BlockResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BlockResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockResponse.Merge(m, src)
}
func (m *BlockResponse) XXX_Size() int {
	return m.Size()
}
func (m *BlockResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockResponse proto.InternalMessageInfo

func (m *BlockResponse) GetBlock() *types.Block {
	if m != nil {
		return m.Block
	}
	return nil
}

// StatusRequest requests the status of a peer.
type StatusRequest struct {
}

func (m *StatusRequest) Reset()         { *m = StatusRequest{} }
func (m *StatusRequest) String() string { return proto.CompactTextString(m) }
func (*StatusRequest) ProtoMessage()    {}
func (*StatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{3}
}
func (m *StatusRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StatusRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StatusRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StatusRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusRequest.Merge(m, src)
}
func (m *StatusRequest) XXX_Size() int {
	return m.Size()
}
func (m *StatusRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StatusRequest proto.InternalMessageInfo

// StatusResponse is a peer response to inform their status.
type StatusResponse struct {
	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Base   int64 `protobuf:"varint,2,opt,name=base,proto3" json:"base,omitempty"`
}

func (m *StatusResponse) Reset()         { *m = StatusResponse{} }
func (m *StatusResponse) String() string { return proto.CompactTextString(m) }
func (*StatusResponse) ProtoMessage()    {}
func (*StatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{4}
}
func (m *StatusResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StatusResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusResponse.Merge(m, src)
}
func (m *StatusResponse) XXX_Size() int {
	return m.Size()
}
func (m *StatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StatusResponse proto.InternalMessageInfo

func (m *StatusResponse) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *StatusResponse) GetBase() int64 {
	if m != nil {
		return m.Base
	}
	return 0
}

type Message struct {
	// Types that are valid to be assigned to Sum:
	//	*Message_BlockRequest
	//	*Message_NoBlockResponse
	//	*Message_BlockResponse
	//	*Message_StatusRequest
	//	*Message_StatusResponse
	Sum isMessage_Sum `protobuf_oneof:"sum"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_19b397c236e0fa07, []int{5}
}
func (m *Message) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Message.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return m.Size()
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

type isMessage_Sum interface {
	isMessage_Sum()
	MarshalTo([]byte) (int, error)
	Size() int
}

type Message_BlockRequest struct {
	BlockRequest *BlockRequest `protobuf:"bytes,1,opt,name=block_request,json=blockRequest,proto3,oneof" json:"block_request,omitempty"`
}
type Message_NoBlockResponse struct {
	NoBlockResponse *NoBlockResponse `protobuf:"bytes,2,opt,name=no_block_response,json=noBlockResponse,proto3,oneof" json:"no_block_response,omitempty"`
}
type Message_BlockResponse struct {
	BlockResponse *BlockResponse `protobuf:"bytes,3,opt,name=block_response,json=blockResponse,proto3,oneof" json:"block_response,omitempty"`
}
type Message_StatusRequest struct {
	StatusRequest *StatusRequest `protobuf:"bytes,4,opt,name=status_request,json=statusRequest,proto3,oneof" json:"status_request,omitempty"`
}
type Message_StatusResponse struct {
	StatusResponse *StatusResponse `protobuf:"bytes,5,opt,name=status_response,json=statusResponse,proto3,oneof" json:"status_response,omitempty"`
}

func (*Message_BlockRequest) isMessage_Sum()    {}
func (*Message_NoBlockResponse) isMessage_Sum() {}
func (*Message_BlockResponse) isMessage_Sum()   {}
func (*Message_StatusRequest) isMessage_Sum()   {}
func (*Message_StatusResponse) isMessage_Sum()  {}

func (m *Message) GetSum() isMessage_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *Message) GetBlockRequest() *BlockRequest {
	if x, ok := m.GetSum().(*Message_BlockRequest); ok {
		return x.BlockRequest
	}
	return nil
}

func (m *Message) GetNoBlockResponse() *NoBlockResponse {
	if x, ok := m.GetSum().(*Message_NoBlockResponse); ok {
		return x.NoBlockResponse
	}
	return nil
}

func (m *Message) GetBlockResponse() *BlockResponse {
	if x, ok := m.GetSum().(*Message_BlockResponse); ok {
		return x.BlockResponse
	}
	return nil
}

func (m *Message) GetStatusRequest() *StatusRequest {
	if x, ok := m.GetSum().(*Message_StatusRequest); ok {
		return x.StatusRequest
	}
	return nil
}

func (m *Message) GetStatusResponse() *StatusResponse {
	if x, ok := m.GetSum().(*Message_StatusResponse); ok {
		return x.StatusResponse
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Message) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Message_BlockRequest)(nil),
		(*Message_NoBlockResponse)(nil),
		(*Message_BlockResponse)(nil),
		(*Message_StatusRequest)(nil),
		(*Message_StatusResponse)(nil),
	}
}

func init() {
	proto.RegisterType((*BlockRequest)(nil), "tendermint.blocksync.BlockRequest")
	proto.RegisterType((*NoBlockResponse)(nil), "tendermint.blocksync.NoBlockResponse")
	proto.RegisterType((*BlockResponse)(nil), "tendermint.blocksync.BlockResponse")
	proto.RegisterType((*StatusRequest)(nil), "tendermint.blocksync.StatusRequest")
	proto.RegisterType((*StatusResponse)(nil), "tendermint.blocksync.StatusResponse")
	proto.RegisterType((*Message)(nil), "tendermint.blocksync.Message")
}

func init() { proto.RegisterFile("tendermint/blocksync/types.proto", fileDescriptor_19b397c236e0fa07) }

var fileDescriptor_19b397c236e0fa07 = []byte{
	// 370 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0xcd, 0x4e, 0xc2, 0x40,
	0x14, 0x85, 0x5b, 0x0b, 0x98, 0x5c, 0x28, 0x8d, 0x8d, 0x51, 0x62, 0x4c, 0x43, 0xea, 0x4f, 0x74,
	0x61, 0x9b, 0xe0, 0xc2, 0x8d, 0x71, 0xc1, 0x0a, 0x13, 0x7f, 0x62, 0xd9, 0xb9, 0x21, 0x4c, 0x1d,
	0x81, 0x68, 0x3b, 0xc8, 0x9d, 0x2e, 0x78, 0x0b, 0x5f, 0xc0, 0xf7, 0x71, 0xc9, 0xd2, 0xa5, 0x81,
	0x17, 0x31, 0xcc, 0x94, 0x52, 0x9a, 0xda, 0xdd, 0xcc, 0xed, 0xb9, 0xdf, 0x39, 0x3d, 0x69, 0xa1,
	0xc9, 0x69, 0xf8, 0x42, 0x27, 0xc1, 0x28, 0xe4, 0x2e, 0x79, 0x67, 0xfe, 0x1b, 0x4e, 0x43, 0xdf,
	0xe5, 0xd3, 0x31, 0x45, 0x67, 0x3c, 0x61, 0x9c, 0x99, 0xbb, 0x6b, 0x85, 0x93, 0x28, 0x0e, 0x0e,
	0x53, 0x7b, 0x42, 0x2d, 0xb7, 0xe5, 0x8e, 0x7d, 0x0a, 0xb5, 0xf6, 0xf2, 0xea, 0xd1, 0x8f, 0x88,
	0x22, 0x37, 0xf7, 0xa0, 0x32, 0xa4, 0xa3, 0xc1, 0x90, 0x37, 0xd4, 0xa6, 0x7a, 0xa6, 0x79, 0xf1,
	0xcd, 0x3e, 0x07, 0xe3, 0x81, 0xc5, 0x4a, 0x1c, 0xb3, 0x10, 0xe9, 0xbf, 0xd2, 0x1b, 0xd0, 0x37,
	0x85, 0x17, 0x50, 0x16, 0x96, 0x42, 0x57, 0x6d, 0xed, 0x3b, 0xa9, 0x9c, 0x32, 0xbf, 0xd4, 0x4b,
	0x95, 0x6d, 0x80, 0xde, 0xe5, 0x7d, 0x1e, 0x61, 0x9c, 0xc9, 0xbe, 0x86, 0xfa, 0x6a, 0x50, 0x6c,
	0x6d, 0x9a, 0x50, 0x22, 0x7d, 0xa4, 0x8d, 0x2d, 0x31, 0x15, 0x67, 0xfb, 0x4b, 0x83, 0xed, 0x7b,
	0x8a, 0xd8, 0x1f, 0x50, 0xf3, 0x16, 0x74, 0xe1, 0xd1, 0x9b, 0x48, 0x74, 0x9c, 0xc8, 0x76, 0xf2,
	0x9a, 0x73, 0xd2, 0xc5, 0x74, 0x14, 0xaf, 0x46, 0xd2, 0x45, 0x75, 0x61, 0x27, 0x64, 0xbd, 0x15,
	0x4d, 0xe6, 0x12, 0xbe, 0xd5, 0xd6, 0x49, 0x3e, 0x2e, 0xd3, 0x5f, 0x47, 0xf1, 0x8c, 0x30, 0x53,
	0xe9, 0x1d, 0xd4, 0x33, 0x44, 0x4d, 0x10, 0x8f, 0x0a, 0x03, 0x26, 0x3c, 0x9d, 0x64, 0x69, 0x28,
	0x7a, 0x4b, 0x5e, 0xb7, 0x54, 0x44, 0xdb, 0x28, 0x7d, 0x49, 0xc3, 0xf4, 0xc0, 0x7c, 0x04, 0x23,
	0xa1, 0xc5, 0xe1, 0xca, 0x02, 0x77, 0x5c, 0x8c, 0x4b, 0xd2, 0xd5, 0x71, 0x63, 0xd2, 0x2e, 0x83,
	0x86, 0x51, 0xd0, 0x7e, 0xfa, 0x9e, 0x5b, 0xea, 0x6c, 0x6e, 0xa9, 0xbf, 0x73, 0x4b, 0xfd, 0x5c,
	0x58, 0xca, 0x6c, 0x61, 0x29, 0x3f, 0x0b, 0x4b, 0x79, 0xbe, 0x1a, 0x8c, 0xf8, 0x30, 0x22, 0x8e,
	0xcf, 0x02, 0xd7, 0x67, 0x01, 0xe5, 0xe4, 0x95, 0xaf, 0x0f, 0xe2, 0xfb, 0x75, 0xf3, 0x7e, 0x0a,
	0x52, 0x11, 0xcf, 0x2e, 0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0x77, 0x16, 0x29, 0xcc, 0x33, 0x03,
	0x00, 0x00,
}

func (m *BlockRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BlockRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *NoBlockResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NoBlockResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NoBlockResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *BlockResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BlockResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Block != nil {
		{
			size, err := m.Block.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *StatusRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StatusRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *StatusResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StatusResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Base != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Base))
		i--
		dAtA[i] = 0x10
	}
	if m.Height != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Message) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Sum != nil {
		{
			size := m.Sum.Size()
			i -= size
			if _, err := m.Sum.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *Message_BlockRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_BlockRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.BlockRequest != nil {
		{
			size, err := m.BlockRequest.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func (m *Message_NoBlockResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_NoBlockResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.NoBlockResponse != nil {
		{
			size, err := m.NoBlockResponse.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}
func (m *Message_BlockResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_BlockResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.BlockResponse != nil {
		{
			size, err := m.BlockResponse.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	return len(dAtA) - i, nil
}
func (m *Message_StatusRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_StatusRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.StatusRequest != nil {
		{
			size, err := m.StatusRequest.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	return len(dAtA) - i, nil
}
func (m *Message_StatusResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_StatusResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.StatusResponse != nil {
		{
			size, err := m.StatusResponse.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	return len(dAtA) - i, nil
}
func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	offset -= sovTypes(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *BlockRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	return n
}

func (m *NoBlockResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	return n
}

func (m *BlockResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Block != nil {
		l = m.Block.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *StatusRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *StatusResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	if m.Base != 0 {
		n += 1 + sovTypes(uint64(m.Base))
	}
	return n
}

func (m *Message) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Sum != nil {
		n += m.Sum.Size()
	}
	return n
}

func (m *Message_BlockRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BlockRequest != nil {
		l = m.BlockRequest.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_NoBlockResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.NoBlockResponse != nil {
		l = m.NoBlockResponse.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_BlockResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BlockResponse != nil {
		l = m.BlockResponse.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_StatusRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.StatusRequest != nil {
		l = m.StatusRequest.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_StatusResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.StatusResponse != nil {
		l = m.StatusResponse.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BlockRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *NoBlockResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NoBlockResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NoBlockResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Block", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Block == nil {
				m.Block = &types.Block{}
			}
			if err := m.Block.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StatusRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StatusRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StatusResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StatusResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Base", wireType)
			}
			m.Base = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Base |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Message) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockRequest", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &BlockRequest{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_BlockRequest{v}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NoBlockResponse", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &NoBlockResponse{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_NoBlockResponse{v}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockResponse", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &BlockResponse{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_BlockResponse{v}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StatusRequest", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &StatusRequest{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_StatusRequest{v}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StatusResponse", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &StatusResponse{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_StatusResponse{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTypes
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTypes
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTypes
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTypes        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTypes          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTypes = fmt.Errorf("proto: unexpected end of group")
)
