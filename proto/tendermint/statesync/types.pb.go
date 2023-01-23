// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/statesync/types.proto

package statesync

import (
	fmt "fmt"
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

type Message struct {
	// Types that are valid to be assigned to Sum:
        //
	//	*Message_SnapshotsRequest
	//	*Message_SnapshotsResponse
	//	*Message_ChunkRequest
	//	*Message_ChunkResponse
	Sum isMessage_Sum `protobuf_oneof:"sum"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_a1c2869546ca7914, []int{0}
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

type Message_SnapshotsRequest struct {
	SnapshotsRequest *SnapshotsRequest `protobuf:"bytes,1,opt,name=snapshots_request,json=snapshotsRequest,proto3,oneof" json:"snapshots_request,omitempty"`
}
type Message_SnapshotsResponse struct {
	SnapshotsResponse *SnapshotsResponse `protobuf:"bytes,2,opt,name=snapshots_response,json=snapshotsResponse,proto3,oneof" json:"snapshots_response,omitempty"`
}
type Message_ChunkRequest struct {
	ChunkRequest *ChunkRequest `protobuf:"bytes,3,opt,name=chunk_request,json=chunkRequest,proto3,oneof" json:"chunk_request,omitempty"`
}
type Message_ChunkResponse struct {
	ChunkResponse *ChunkResponse `protobuf:"bytes,4,opt,name=chunk_response,json=chunkResponse,proto3,oneof" json:"chunk_response,omitempty"`
}

func (*Message_SnapshotsRequest) isMessage_Sum()  {}
func (*Message_SnapshotsResponse) isMessage_Sum() {}
func (*Message_ChunkRequest) isMessage_Sum()      {}
func (*Message_ChunkResponse) isMessage_Sum()     {}

func (m *Message) GetSum() isMessage_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *Message) GetSnapshotsRequest() *SnapshotsRequest {
	if x, ok := m.GetSum().(*Message_SnapshotsRequest); ok {
		return x.SnapshotsRequest
	}
	return nil
}

func (m *Message) GetSnapshotsResponse() *SnapshotsResponse {
	if x, ok := m.GetSum().(*Message_SnapshotsResponse); ok {
		return x.SnapshotsResponse
	}
	return nil
}

func (m *Message) GetChunkRequest() *ChunkRequest {
	if x, ok := m.GetSum().(*Message_ChunkRequest); ok {
		return x.ChunkRequest
	}
	return nil
}

func (m *Message) GetChunkResponse() *ChunkResponse {
	if x, ok := m.GetSum().(*Message_ChunkResponse); ok {
		return x.ChunkResponse
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Message) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Message_SnapshotsRequest)(nil),
		(*Message_SnapshotsResponse)(nil),
		(*Message_ChunkRequest)(nil),
		(*Message_ChunkResponse)(nil),
	}
}

type SnapshotsRequest struct {
}

func (m *SnapshotsRequest) Reset()         { *m = SnapshotsRequest{} }
func (m *SnapshotsRequest) String() string { return proto.CompactTextString(m) }
func (*SnapshotsRequest) ProtoMessage()    {}
func (*SnapshotsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a1c2869546ca7914, []int{1}
}
func (m *SnapshotsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SnapshotsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SnapshotsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SnapshotsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotsRequest.Merge(m, src)
}
func (m *SnapshotsRequest) XXX_Size() int {
	return m.Size()
}
func (m *SnapshotsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotsRequest proto.InternalMessageInfo

type SnapshotsResponse struct {
	Height   uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Format   uint32 `protobuf:"varint,2,opt,name=format,proto3" json:"format,omitempty"`
	Chunks   uint32 `protobuf:"varint,3,opt,name=chunks,proto3" json:"chunks,omitempty"`
	Hash     []byte `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	Metadata []byte `protobuf:"bytes,5,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (m *SnapshotsResponse) Reset()         { *m = SnapshotsResponse{} }
func (m *SnapshotsResponse) String() string { return proto.CompactTextString(m) }
func (*SnapshotsResponse) ProtoMessage()    {}
func (*SnapshotsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a1c2869546ca7914, []int{2}
}
func (m *SnapshotsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SnapshotsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SnapshotsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SnapshotsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotsResponse.Merge(m, src)
}
func (m *SnapshotsResponse) XXX_Size() int {
	return m.Size()
}
func (m *SnapshotsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotsResponse proto.InternalMessageInfo

func (m *SnapshotsResponse) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *SnapshotsResponse) GetFormat() uint32 {
	if m != nil {
		return m.Format
	}
	return 0
}

func (m *SnapshotsResponse) GetChunks() uint32 {
	if m != nil {
		return m.Chunks
	}
	return 0
}

func (m *SnapshotsResponse) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *SnapshotsResponse) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

type ChunkRequest struct {
	Height uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Format uint32 `protobuf:"varint,2,opt,name=format,proto3" json:"format,omitempty"`
	Index  uint32 `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
}

func (m *ChunkRequest) Reset()         { *m = ChunkRequest{} }
func (m *ChunkRequest) String() string { return proto.CompactTextString(m) }
func (*ChunkRequest) ProtoMessage()    {}
func (*ChunkRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a1c2869546ca7914, []int{3}
}
func (m *ChunkRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ChunkRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ChunkRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ChunkRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChunkRequest.Merge(m, src)
}
func (m *ChunkRequest) XXX_Size() int {
	return m.Size()
}
func (m *ChunkRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ChunkRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ChunkRequest proto.InternalMessageInfo

func (m *ChunkRequest) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ChunkRequest) GetFormat() uint32 {
	if m != nil {
		return m.Format
	}
	return 0
}

func (m *ChunkRequest) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

type ChunkResponse struct {
	Height  uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Format  uint32 `protobuf:"varint,2,opt,name=format,proto3" json:"format,omitempty"`
	Index   uint32 `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
	Chunk   []byte `protobuf:"bytes,4,opt,name=chunk,proto3" json:"chunk,omitempty"`
	Missing bool   `protobuf:"varint,5,opt,name=missing,proto3" json:"missing,omitempty"`
}

func (m *ChunkResponse) Reset()         { *m = ChunkResponse{} }
func (m *ChunkResponse) String() string { return proto.CompactTextString(m) }
func (*ChunkResponse) ProtoMessage()    {}
func (*ChunkResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a1c2869546ca7914, []int{4}
}
func (m *ChunkResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ChunkResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ChunkResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ChunkResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChunkResponse.Merge(m, src)
}
func (m *ChunkResponse) XXX_Size() int {
	return m.Size()
}
func (m *ChunkResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ChunkResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ChunkResponse proto.InternalMessageInfo

func (m *ChunkResponse) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ChunkResponse) GetFormat() uint32 {
	if m != nil {
		return m.Format
	}
	return 0
}

func (m *ChunkResponse) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *ChunkResponse) GetChunk() []byte {
	if m != nil {
		return m.Chunk
	}
	return nil
}

func (m *ChunkResponse) GetMissing() bool {
	if m != nil {
		return m.Missing
	}
	return false
}

func init() {
	proto.RegisterType((*Message)(nil), "tendermint.statesync.Message")
	proto.RegisterType((*SnapshotsRequest)(nil), "tendermint.statesync.SnapshotsRequest")
	proto.RegisterType((*SnapshotsResponse)(nil), "tendermint.statesync.SnapshotsResponse")
	proto.RegisterType((*ChunkRequest)(nil), "tendermint.statesync.ChunkRequest")
	proto.RegisterType((*ChunkResponse)(nil), "tendermint.statesync.ChunkResponse")
}

func init() { proto.RegisterFile("tendermint/statesync/types.proto", fileDescriptor_a1c2869546ca7914) }

var fileDescriptor_a1c2869546ca7914 = []byte{
	// 393 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xcd, 0x6a, 0xdb, 0x40,
	0x18, 0x94, 0xfc, 0xcf, 0x57, 0xab, 0xd8, 0x8b, 0x29, 0xa2, 0x07, 0x61, 0x54, 0x68, 0x7b, 0x92,
	0xa0, 0x3d, 0xf6, 0xe6, 0x5e, 0x5c, 0x68, 0x2f, 0xdb, 0x18, 0x42, 0x2e, 0x61, 0x2d, 0x6f, 0x24,
	0x11, 0xb4, 0x52, 0xf4, 0xad, 0x20, 0x7e, 0x80, 0x9c, 0x72, 0xc9, 0x63, 0xe5, 0xe8, 0x63, 0xc8,
	0x29, 0xd8, 0x2f, 0x12, 0xb4, 0x92, 0x65, 0xc5, 0x31, 0x09, 0x81, 0xdc, 0x76, 0xc6, 0xe3, 0xd1,
	0xcc, 0xc0, 0x07, 0x63, 0xc9, 0xc5, 0x82, 0xa7, 0x51, 0x28, 0xa4, 0x8b, 0x92, 0x49, 0x8e, 0x4b,
	0xe1, 0xb9, 0x72, 0x99, 0x70, 0x74, 0x92, 0x34, 0x96, 0x31, 0x19, 0xed, 0x14, 0x4e, 0xa5, 0xb0,
	0xef, 0x1b, 0xd0, 0xfd, 0xc7, 0x11, 0x99, 0xcf, 0xc9, 0x0c, 0x86, 0x28, 0x58, 0x82, 0x41, 0x2c,
	0xf1, 0x34, 0xe5, 0x17, 0x19, 0x47, 0x69, 0xea, 0x63, 0xfd, 0xfb, 0x87, 0x1f, 0x5f, 0x9d, 0x43,
	0xff, 0x76, 0xfe, 0x6f, 0xe5, 0xb4, 0x50, 0x4f, 0x35, 0x3a, 0xc0, 0x3d, 0x8e, 0x1c, 0x03, 0xa9,
	0xdb, 0x62, 0x12, 0x0b, 0xe4, 0x66, 0x43, 0xf9, 0x7e, 0x7b, 0xd5, 0xb7, 0x90, 0x4f, 0x35, 0x3a,
	0xc4, 0x7d, 0x92, 0xfc, 0x01, 0xc3, 0x0b, 0x32, 0x71, 0x5e, 0x85, 0x6d, 0x2a, 0x53, 0xfb, 0xb0,
	0xe9, 0xef, 0x5c, 0xba, 0x0b, 0xda, 0xf7, 0x6a, 0x98, 0xfc, 0x85, 0x8f, 0x5b, 0xab, 0x32, 0x60,
	0x4b, 0x79, 0x7d, 0x79, 0xd1, 0xab, 0x0a, 0x67, 0x78, 0x75, 0x62, 0xd2, 0x86, 0x26, 0x66, 0x91,
	0x4d, 0x60, 0xb0, 0xbf, 0x90, 0x7d, 0xad, 0xc3, 0xf0, 0x59, 0x3d, 0xf2, 0x09, 0x3a, 0x01, 0x0f,
	0xfd, 0xa0, 0xd8, 0xbb, 0x45, 0x4b, 0x94, 0xf3, 0x67, 0x71, 0x1a, 0x31, 0xa9, 0xf6, 0x32, 0x68,
	0x89, 0x72, 0x5e, 0x7d, 0x11, 0x55, 0x65, 0x83, 0x96, 0x88, 0x10, 0x68, 0x05, 0x0c, 0x03, 0x15,
	0xbe, 0x4f, 0xd5, 0x9b, 0x7c, 0x86, 0x5e, 0xc4, 0x25, 0x5b, 0x30, 0xc9, 0xcc, 0xb6, 0xe2, 0x2b,
	0x6c, 0x1f, 0x41, 0xbf, 0x3e, 0xcb, 0x9b, 0x73, 0x8c, 0xa0, 0x1d, 0x8a, 0x05, 0xbf, 0x2c, 0x63,
	0x14, 0xc0, 0xbe, 0xd2, 0xc1, 0x78, 0xb2, 0xd0, 0xfb, 0xf8, 0xe6, 0xac, 0xea, 0x59, 0xd6, 0x2b,
	0x00, 0x31, 0xa1, 0x1b, 0x85, 0x88, 0xa1, 0xf0, 0x55, 0xbd, 0x1e, 0xdd, 0xc2, 0xc9, 0xec, 0x76,
	0x6d, 0xe9, 0xab, 0xb5, 0xa5, 0x3f, 0xac, 0x2d, 0xfd, 0x66, 0x63, 0x69, 0xab, 0x8d, 0xa5, 0xdd,
	0x6d, 0x2c, 0xed, 0xe4, 0x97, 0x1f, 0xca, 0x20, 0x9b, 0x3b, 0x5e, 0x1c, 0xb9, 0xb5, 0xcb, 0xa9,
	0x3d, 0xd5, 0xd1, 0xb8, 0x87, 0xae, 0x6a, 0xde, 0x51, 0xbf, 0xfd, 0x7c, 0x0c, 0x00, 0x00, 0xff,
	0xff, 0xcc, 0x16, 0xc2, 0x8b, 0x74, 0x03, 0x00, 0x00,
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

func (m *Message_SnapshotsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_SnapshotsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.SnapshotsRequest != nil {
		{
			size, err := m.SnapshotsRequest.MarshalToSizedBuffer(dAtA[:i])
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
func (m *Message_SnapshotsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_SnapshotsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.SnapshotsResponse != nil {
		{
			size, err := m.SnapshotsResponse.MarshalToSizedBuffer(dAtA[:i])
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
func (m *Message_ChunkRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_ChunkRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ChunkRequest != nil {
		{
			size, err := m.ChunkRequest.MarshalToSizedBuffer(dAtA[:i])
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
func (m *Message_ChunkResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_ChunkResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ChunkResponse != nil {
		{
			size, err := m.ChunkResponse.MarshalToSizedBuffer(dAtA[:i])
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
func (m *SnapshotsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SnapshotsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SnapshotsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *SnapshotsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SnapshotsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SnapshotsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0x22
	}
	if m.Chunks != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Chunks))
		i--
		dAtA[i] = 0x18
	}
	if m.Format != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Format))
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

func (m *ChunkRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ChunkRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ChunkRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Index != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x18
	}
	if m.Format != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Format))
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

func (m *ChunkResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ChunkResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ChunkResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Missing {
		i--
		if m.Missing {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if len(m.Chunk) > 0 {
		i -= len(m.Chunk)
		copy(dAtA[i:], m.Chunk)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Chunk)))
		i--
		dAtA[i] = 0x22
	}
	if m.Index != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Index))
		i--
		dAtA[i] = 0x18
	}
	if m.Format != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.Format))
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

func (m *Message_SnapshotsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SnapshotsRequest != nil {
		l = m.SnapshotsRequest.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_SnapshotsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SnapshotsResponse != nil {
		l = m.SnapshotsResponse.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_ChunkRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ChunkRequest != nil {
		l = m.ChunkRequest.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_ChunkResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ChunkResponse != nil {
		l = m.ChunkResponse.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *SnapshotsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *SnapshotsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	if m.Format != 0 {
		n += 1 + sovTypes(uint64(m.Format))
	}
	if m.Chunks != 0 {
		n += 1 + sovTypes(uint64(m.Chunks))
	}
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *ChunkRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	if m.Format != 0 {
		n += 1 + sovTypes(uint64(m.Format))
	}
	if m.Index != 0 {
		n += 1 + sovTypes(uint64(m.Index))
	}
	return n
}

func (m *ChunkResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovTypes(uint64(m.Height))
	}
	if m.Format != 0 {
		n += 1 + sovTypes(uint64(m.Format))
	}
	if m.Index != 0 {
		n += 1 + sovTypes(uint64(m.Index))
	}
	l = len(m.Chunk)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	if m.Missing {
		n += 2
	}
	return n
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
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
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotsRequest", wireType)
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
			v := &SnapshotsRequest{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_SnapshotsRequest{v}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotsResponse", wireType)
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
			v := &SnapshotsResponse{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_SnapshotsResponse{v}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChunkRequest", wireType)
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
			v := &ChunkRequest{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_ChunkRequest{v}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChunkResponse", wireType)
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
			v := &ChunkResponse{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_ChunkResponse{v}
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
func (m *SnapshotsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SnapshotsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SnapshotsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *SnapshotsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SnapshotsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SnapshotsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
				m.Height |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Format", wireType)
			}
			m.Format = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Format |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Chunks", wireType)
			}
			m.Chunks = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Chunks |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = append(m.Hash[:0], dAtA[iNdEx:postIndex]...)
			if m.Hash == nil {
				m.Hash = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metadata = append(m.Metadata[:0], dAtA[iNdEx:postIndex]...)
			if m.Metadata == nil {
				m.Metadata = []byte{}
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
func (m *ChunkRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ChunkRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ChunkRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
				m.Height |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Format", wireType)
			}
			m.Format = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Format |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= uint32(b&0x7F) << shift
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
func (m *ChunkResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ChunkResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ChunkResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
				m.Height |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Format", wireType)
			}
			m.Format = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Format |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Chunk", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Chunk = append(m.Chunk[:0], dAtA[iNdEx:postIndex]...)
			if m.Chunk == nil {
				m.Chunk = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Missing", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Missing = bool(v != 0)
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
