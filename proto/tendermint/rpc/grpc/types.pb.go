// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/rpc/grpc/types.proto

package coregrpc

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	types "github.com/cometbft/cometbft/abci/types"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type RequestPing struct {
}

func (m *RequestPing) Reset()         { *m = RequestPing{} }
func (m *RequestPing) String() string { return proto.CompactTextString(m) }
func (*RequestPing) ProtoMessage()    {}
func (*RequestPing) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ffff5682c662b95, []int{0}
}
func (m *RequestPing) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RequestPing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RequestPing.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RequestPing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequestPing.Merge(m, src)
}
func (m *RequestPing) XXX_Size() int {
	return m.Size()
}
func (m *RequestPing) XXX_DiscardUnknown() {
	xxx_messageInfo_RequestPing.DiscardUnknown(m)
}

var xxx_messageInfo_RequestPing proto.InternalMessageInfo

type RequestBroadcastTx struct {
	Tx []byte `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
}

func (m *RequestBroadcastTx) Reset()         { *m = RequestBroadcastTx{} }
func (m *RequestBroadcastTx) String() string { return proto.CompactTextString(m) }
func (*RequestBroadcastTx) ProtoMessage()    {}
func (*RequestBroadcastTx) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ffff5682c662b95, []int{1}
}
func (m *RequestBroadcastTx) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RequestBroadcastTx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RequestBroadcastTx.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RequestBroadcastTx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequestBroadcastTx.Merge(m, src)
}
func (m *RequestBroadcastTx) XXX_Size() int {
	return m.Size()
}
func (m *RequestBroadcastTx) XXX_DiscardUnknown() {
	xxx_messageInfo_RequestBroadcastTx.DiscardUnknown(m)
}

var xxx_messageInfo_RequestBroadcastTx proto.InternalMessageInfo

func (m *RequestBroadcastTx) GetTx() []byte {
	if m != nil {
		return m.Tx
	}
	return nil
}

type ResponsePing struct {
}

func (m *ResponsePing) Reset()         { *m = ResponsePing{} }
func (m *ResponsePing) String() string { return proto.CompactTextString(m) }
func (*ResponsePing) ProtoMessage()    {}
func (*ResponsePing) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ffff5682c662b95, []int{2}
}
func (m *ResponsePing) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ResponsePing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ResponsePing.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ResponsePing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponsePing.Merge(m, src)
}
func (m *ResponsePing) XXX_Size() int {
	return m.Size()
}
func (m *ResponsePing) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponsePing.DiscardUnknown(m)
}

var xxx_messageInfo_ResponsePing proto.InternalMessageInfo

type ResponseBroadcastTx struct {
	CheckTx   *types.ResponseCheckTx   `protobuf:"bytes,1,opt,name=check_tx,json=checkTx,proto3" json:"check_tx,omitempty"`
	DeliverTx *types.ResponseDeliverTx `protobuf:"bytes,2,opt,name=deliver_tx,json=deliverTx,proto3" json:"deliver_tx,omitempty"`
}

func (m *ResponseBroadcastTx) Reset()         { *m = ResponseBroadcastTx{} }
func (m *ResponseBroadcastTx) String() string { return proto.CompactTextString(m) }
func (*ResponseBroadcastTx) ProtoMessage()    {}
func (*ResponseBroadcastTx) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ffff5682c662b95, []int{3}
}
func (m *ResponseBroadcastTx) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ResponseBroadcastTx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ResponseBroadcastTx.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ResponseBroadcastTx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseBroadcastTx.Merge(m, src)
}
func (m *ResponseBroadcastTx) XXX_Size() int {
	return m.Size()
}
func (m *ResponseBroadcastTx) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseBroadcastTx.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseBroadcastTx proto.InternalMessageInfo

func (m *ResponseBroadcastTx) GetCheckTx() *types.ResponseCheckTx {
	if m != nil {
		return m.CheckTx
	}
	return nil
}

func (m *ResponseBroadcastTx) GetDeliverTx() *types.ResponseDeliverTx {
	if m != nil {
		return m.DeliverTx
	}
	return nil
}

func init() {
	proto.RegisterType((*RequestPing)(nil), "tendermint.rpc.grpc.RequestPing")
	proto.RegisterType((*RequestBroadcastTx)(nil), "tendermint.rpc.grpc.RequestBroadcastTx")
	proto.RegisterType((*ResponsePing)(nil), "tendermint.rpc.grpc.ResponsePing")
	proto.RegisterType((*ResponseBroadcastTx)(nil), "tendermint.rpc.grpc.ResponseBroadcastTx")
}

func init() { proto.RegisterFile("tendermint/rpc/grpc/types.proto", fileDescriptor_0ffff5682c662b95) }

var fileDescriptor_0ffff5682c662b95 = []byte{
	// 316 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2f, 0x49, 0xcd, 0x4b,
	0x49, 0x2d, 0xca, 0xcd, 0xcc, 0x2b, 0xd1, 0x2f, 0x2a, 0x48, 0xd6, 0x4f, 0x07, 0x11, 0x25, 0x95,
	0x05, 0xa9, 0xc5, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0xc2, 0x08, 0x05, 0x7a, 0x45, 0x05,
	0xc9, 0x7a, 0x20, 0x05, 0x52, 0xd2, 0x48, 0xba, 0x12, 0x93, 0x92, 0x33, 0x91, 0x75, 0x28, 0xf1,
	0x72, 0x71, 0x07, 0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x04, 0x64, 0xe6, 0xa5, 0x2b, 0xa9, 0x70,
	0x09, 0x41, 0xb9, 0x4e, 0x45, 0xf9, 0x89, 0x29, 0xc9, 0x89, 0xc5, 0x25, 0x21, 0x15, 0x42, 0x7c,
	0x5c, 0x4c, 0x25, 0x15, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x3c, 0x41, 0x4c, 0x25, 0x15, 0x4a, 0x7c,
	0x5c, 0x3c, 0x41, 0xa9, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x60, 0x5d, 0x53, 0x19, 0xb9, 0x84,
	0x61, 0x02, 0xc8, 0xfa, 0xac, 0xb9, 0x38, 0x92, 0x33, 0x52, 0x93, 0xb3, 0xe3, 0xa1, 0xba, 0xb9,
	0x8d, 0x14, 0xf4, 0x90, 0x5c, 0x08, 0x72, 0x8c, 0x1e, 0x4c, 0x9f, 0x33, 0x48, 0x61, 0x48, 0x45,
	0x10, 0x7b, 0x32, 0x84, 0x21, 0xe4, 0xc8, 0xc5, 0x95, 0x92, 0x9a, 0x93, 0x59, 0x96, 0x5a, 0x04,
	0xd2, 0xce, 0x04, 0xd6, 0xae, 0x84, 0x53, 0xbb, 0x0b, 0x44, 0x69, 0x48, 0x45, 0x10, 0x67, 0x0a,
	0x8c, 0x69, 0xb4, 0x97, 0x91, 0x8b, 0x07, 0xee, 0x1e, 0xc7, 0x00, 0x4f, 0x21, 0x6f, 0x2e, 0x16,
	0x90, 0x83, 0x85, 0x50, 0x9c, 0x01, 0x0b, 0x28, 0x3d, 0xa4, 0x80, 0x90, 0x52, 0xc4, 0xa1, 0x02,
	0xe1, 0x6b, 0xa1, 0x04, 0x2e, 0x6e, 0x64, 0xcf, 0xaa, 0xe3, 0x33, 0x13, 0x49, 0xa1, 0x94, 0x06,
	0x5e, 0xa3, 0x91, 0x54, 0x3a, 0xf9, 0x9c, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83,
	0x47, 0x72, 0x8c, 0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43,
	0x94, 0x51, 0x7a, 0x66, 0x49, 0x46, 0x69, 0x92, 0x5e, 0x72, 0x7e, 0xae, 0x3e, 0x52, 0xf4, 0x62,
	0x49, 0x1f, 0xd6, 0xc9, 0xf9, 0x45, 0xa9, 0x20, 0x46, 0x12, 0x1b, 0x38, 0xc6, 0x8d, 0x01, 0x01,
	0x00, 0x00, 0xff, 0xff, 0xf6, 0x4b, 0x02, 0xd8, 0x46, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BroadcastAPIClient is the client API for BroadcastAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BroadcastAPIClient interface {
	Ping(ctx context.Context, in *RequestPing, opts ...grpc.CallOption) (*ResponsePing, error)
	BroadcastTx(ctx context.Context, in *RequestBroadcastTx, opts ...grpc.CallOption) (*ResponseBroadcastTx, error)
}

type broadcastAPIClient struct {
	cc grpc1.ClientConn
}

func NewBroadcastAPIClient(cc grpc1.ClientConn) BroadcastAPIClient {
	return &broadcastAPIClient{cc}
}

func (c *broadcastAPIClient) Ping(ctx context.Context, in *RequestPing, opts ...grpc.CallOption) (*ResponsePing, error) {
	out := new(ResponsePing)
	err := c.cc.Invoke(ctx, "/tendermint.rpc.grpc.BroadcastAPI/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *broadcastAPIClient) BroadcastTx(ctx context.Context, in *RequestBroadcastTx, opts ...grpc.CallOption) (*ResponseBroadcastTx, error) {
	out := new(ResponseBroadcastTx)
	err := c.cc.Invoke(ctx, "/tendermint.rpc.grpc.BroadcastAPI/BroadcastTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BroadcastAPIServer is the server API for BroadcastAPI service.
type BroadcastAPIServer interface {
	Ping(context.Context, *RequestPing) (*ResponsePing, error)
	BroadcastTx(context.Context, *RequestBroadcastTx) (*ResponseBroadcastTx, error)
}

// UnimplementedBroadcastAPIServer can be embedded to have forward compatible implementations.
type UnimplementedBroadcastAPIServer struct {
}

func (*UnimplementedBroadcastAPIServer) Ping(ctx context.Context, req *RequestPing) (*ResponsePing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedBroadcastAPIServer) BroadcastTx(ctx context.Context, req *RequestBroadcastTx) (*ResponseBroadcastTx, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadcastTx not implemented")
}

func RegisterBroadcastAPIServer(s grpc1.Server, srv BroadcastAPIServer) {
	s.RegisterService(&_BroadcastAPI_serviceDesc, srv)
}

func _BroadcastAPI_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestPing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastAPIServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tendermint.rpc.grpc.BroadcastAPI/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastAPIServer).Ping(ctx, req.(*RequestPing))
	}
	return interceptor(ctx, in, info, handler)
}

func _BroadcastAPI_BroadcastTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestBroadcastTx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastAPIServer).BroadcastTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tendermint.rpc.grpc.BroadcastAPI/BroadcastTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastAPIServer).BroadcastTx(ctx, req.(*RequestBroadcastTx))
	}
	return interceptor(ctx, in, info, handler)
}

var _BroadcastAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tendermint.rpc.grpc.BroadcastAPI",
	HandlerType: (*BroadcastAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _BroadcastAPI_Ping_Handler,
		},
		{
			MethodName: "BroadcastTx",
			Handler:    _BroadcastAPI_BroadcastTx_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tendermint/rpc/grpc/types.proto",
}

func (m *RequestPing) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RequestPing) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RequestPing) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *RequestBroadcastTx) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RequestBroadcastTx) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RequestBroadcastTx) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Tx) > 0 {
		i -= len(m.Tx)
		copy(dAtA[i:], m.Tx)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Tx)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ResponsePing) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ResponsePing) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ResponsePing) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *ResponseBroadcastTx) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ResponseBroadcastTx) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ResponseBroadcastTx) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DeliverTx != nil {
		{
			size, err := m.DeliverTx.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.CheckTx != nil {
		{
			size, err := m.CheckTx.MarshalToSizedBuffer(dAtA[:i])
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
func (m *RequestPing) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *RequestBroadcastTx) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Tx)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *ResponsePing) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *ResponseBroadcastTx) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CheckTx != nil {
		l = m.CheckTx.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	if m.DeliverTx != nil {
		l = m.DeliverTx.Size()
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
func (m *RequestPing) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: RequestPing: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RequestPing: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *RequestBroadcastTx) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: RequestBroadcastTx: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RequestBroadcastTx: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tx", wireType)
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
			m.Tx = append(m.Tx[:0], dAtA[iNdEx:postIndex]...)
			if m.Tx == nil {
				m.Tx = []byte{}
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
func (m *ResponsePing) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ResponsePing: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ResponsePing: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *ResponseBroadcastTx) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ResponseBroadcastTx: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ResponseBroadcastTx: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CheckTx", wireType)
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
			if m.CheckTx == nil {
				m.CheckTx = &types.ResponseCheckTx{}
			}
			if err := m.CheckTx.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeliverTx", wireType)
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
			if m.DeliverTx == nil {
				m.DeliverTx = &types.ResponseDeliverTx{}
			}
			if err := m.DeliverTx.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
