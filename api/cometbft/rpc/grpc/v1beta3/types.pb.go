// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cometbft/rpc/grpc/v1beta3/types.proto

package v1beta3

import (
	context "context"
	fmt "fmt"
	v1beta3 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta3"
	v1beta1 "github.com/cometbft/cometbft/api/cometbft/rpc/grpc/v1beta1"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
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

// ResponseBroadcastTx is a response of broadcasting the transaction.
type ResponseBroadcastTx struct {
	CheckTx  *v1beta3.ResponseCheckTx `protobuf:"bytes,1,opt,name=check_tx,json=checkTx,proto3" json:"check_tx,omitempty"`
	TxResult *v1beta3.ExecTxResult    `protobuf:"bytes,2,opt,name=tx_result,json=txResult,proto3" json:"tx_result,omitempty"`
}

func (m *ResponseBroadcastTx) Reset()         { *m = ResponseBroadcastTx{} }
func (m *ResponseBroadcastTx) String() string { return proto.CompactTextString(m) }
func (*ResponseBroadcastTx) ProtoMessage()    {}
func (*ResponseBroadcastTx) Descriptor() ([]byte, []int) {
	return fileDescriptor_e521bcdb5edbf680, []int{0}
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

func (m *ResponseBroadcastTx) GetCheckTx() *v1beta3.ResponseCheckTx {
	if m != nil {
		return m.CheckTx
	}
	return nil
}

func (m *ResponseBroadcastTx) GetTxResult() *v1beta3.ExecTxResult {
	if m != nil {
		return m.TxResult
	}
	return nil
}

func init() {
	proto.RegisterType((*ResponseBroadcastTx)(nil), "cometbft.rpc.grpc.v1beta3.ResponseBroadcastTx")
}

func init() {
	proto.RegisterFile("cometbft/rpc/grpc/v1beta3/types.proto", fileDescriptor_e521bcdb5edbf680)
}

var fileDescriptor_e521bcdb5edbf680 = []byte{
	// 308 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4d, 0xce, 0xcf, 0x4d,
	0x2d, 0x49, 0x4a, 0x2b, 0xd1, 0x2f, 0x2a, 0x48, 0xd6, 0x4f, 0x07, 0x11, 0x65, 0x86, 0x49, 0xa9,
	0x25, 0x89, 0xc6, 0xfa, 0x25, 0x95, 0x05, 0xa9, 0xc5, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42,
	0x92, 0x30, 0x65, 0x7a, 0x45, 0x05, 0xc9, 0x7a, 0x20, 0x65, 0x7a, 0x50, 0x65, 0x52, 0x38, 0x4d,
	0x30, 0x44, 0x36, 0x41, 0x4a, 0x11, 0xae, 0x2c, 0x31, 0x29, 0x39, 0x13, 0x9b, 0x25, 0x4a, 0xb3,
	0x18, 0xb9, 0x84, 0x83, 0x52, 0x8b, 0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x9d, 0x8a, 0xf2, 0x13, 0x53,
	0x92, 0x13, 0x8b, 0x4b, 0x42, 0x2a, 0x84, 0x1c, 0xb9, 0x38, 0x92, 0x33, 0x52, 0x93, 0xb3, 0xe3,
	0x4b, 0x2a, 0x24, 0x18, 0x15, 0x18, 0x35, 0xb8, 0x8d, 0xd4, 0xf4, 0xe0, 0xee, 0x01, 0x99, 0x06,
	0x73, 0x8b, 0x1e, 0x4c, 0xb7, 0x33, 0x48, 0x79, 0x48, 0x45, 0x10, 0x7b, 0x32, 0x84, 0x21, 0xe4,
	0xc0, 0xc5, 0x59, 0x52, 0x11, 0x5f, 0x94, 0x5a, 0x5c, 0x9a, 0x53, 0x22, 0xc1, 0x04, 0x36, 0x43,
	0x19, 0x87, 0x19, 0xae, 0x15, 0xa9, 0xc9, 0x21, 0x15, 0x41, 0x60, 0xa5, 0x41, 0x1c, 0x25, 0x50,
	0x96, 0xd1, 0x55, 0x46, 0x2e, 0x1e, 0xb8, 0xa3, 0x1c, 0x03, 0x3c, 0x85, 0xc2, 0xb9, 0x58, 0x02,
	0x32, 0xf3, 0xd2, 0x85, 0x90, 0xdc, 0x82, 0x16, 0x36, 0x86, 0x7a, 0x41, 0xa9, 0x85, 0xa5, 0xa9,
	0xc5, 0x25, 0x20, 0x75, 0x52, 0xea, 0x78, 0xd5, 0x41, 0xdc, 0x0d, 0x36, 0x30, 0x87, 0x8b, 0x1b,
	0xd9, 0xf7, 0xba, 0x84, 0xcd, 0x47, 0x52, 0x2e, 0xa5, 0x87, 0x53, 0x39, 0x22, 0x78, 0x90, 0xd4,
	0x3b, 0x85, 0x9c, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x13,
	0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43, 0x94, 0x55, 0x7a, 0x66,
	0x49, 0x46, 0x69, 0x12, 0xc8, 0x3c, 0x7d, 0x78, 0xe4, 0x21, 0x62, 0xb1, 0x20, 0x53, 0x1f, 0x67,
	0xda, 0x49, 0x62, 0x03, 0xc7, 0xa8, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xc2, 0xcf, 0xef, 0x1f,
	0x5f, 0x02, 0x00, 0x00,
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
	// Ping the connection.
	Ping(ctx context.Context, in *v1beta1.RequestPing, opts ...grpc.CallOption) (*v1beta1.ResponsePing, error)
	// BroadcastTx broadcasts a transaction.
	BroadcastTx(ctx context.Context, in *v1beta1.RequestBroadcastTx, opts ...grpc.CallOption) (*ResponseBroadcastTx, error)
}

type broadcastAPIClient struct {
	cc grpc1.ClientConn
}

func NewBroadcastAPIClient(cc grpc1.ClientConn) BroadcastAPIClient {
	return &broadcastAPIClient{cc}
}

func (c *broadcastAPIClient) Ping(ctx context.Context, in *v1beta1.RequestPing, opts ...grpc.CallOption) (*v1beta1.ResponsePing, error) {
	out := new(v1beta1.ResponsePing)
	err := c.cc.Invoke(ctx, "/cometbft.rpc.grpc.v1beta3.BroadcastAPI/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *broadcastAPIClient) BroadcastTx(ctx context.Context, in *v1beta1.RequestBroadcastTx, opts ...grpc.CallOption) (*ResponseBroadcastTx, error) {
	out := new(ResponseBroadcastTx)
	err := c.cc.Invoke(ctx, "/cometbft.rpc.grpc.v1beta3.BroadcastAPI/BroadcastTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BroadcastAPIServer is the server API for BroadcastAPI service.
type BroadcastAPIServer interface {
	// Ping the connection.
	Ping(context.Context, *v1beta1.RequestPing) (*v1beta1.ResponsePing, error)
	// BroadcastTx broadcasts a transaction.
	BroadcastTx(context.Context, *v1beta1.RequestBroadcastTx) (*ResponseBroadcastTx, error)
}

// UnimplementedBroadcastAPIServer can be embedded to have forward compatible implementations.
type UnimplementedBroadcastAPIServer struct {
}

func (*UnimplementedBroadcastAPIServer) Ping(ctx context.Context, req *v1beta1.RequestPing) (*v1beta1.ResponsePing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedBroadcastAPIServer) BroadcastTx(ctx context.Context, req *v1beta1.RequestBroadcastTx) (*ResponseBroadcastTx, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadcastTx not implemented")
}

func RegisterBroadcastAPIServer(s grpc1.Server, srv BroadcastAPIServer) {
	s.RegisterService(&_BroadcastAPI_serviceDesc, srv)
}

func _BroadcastAPI_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1beta1.RequestPing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastAPIServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cometbft.rpc.grpc.v1beta3.BroadcastAPI/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastAPIServer).Ping(ctx, req.(*v1beta1.RequestPing))
	}
	return interceptor(ctx, in, info, handler)
}

func _BroadcastAPI_BroadcastTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1beta1.RequestBroadcastTx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BroadcastAPIServer).BroadcastTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cometbft.rpc.grpc.v1beta3.BroadcastAPI/BroadcastTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BroadcastAPIServer).BroadcastTx(ctx, req.(*v1beta1.RequestBroadcastTx))
	}
	return interceptor(ctx, in, info, handler)
}

var BroadcastAPI_serviceDesc = _BroadcastAPI_serviceDesc
var _BroadcastAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cometbft.rpc.grpc.v1beta3.BroadcastAPI",
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
	Metadata: "cometbft/rpc/grpc/v1beta3/types.proto",
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
	if m.TxResult != nil {
		{
			size, err := m.TxResult.MarshalToSizedBuffer(dAtA[:i])
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
	if m.TxResult != nil {
		l = m.TxResult.Size()
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
				m.CheckTx = &v1beta3.ResponseCheckTx{}
			}
			if err := m.CheckTx.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TxResult", wireType)
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
			if m.TxResult == nil {
				m.TxResult = &v1beta3.ExecTxResult{}
			}
			if err := m.TxResult.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
