// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cometbft/services/version/v1/version_service.proto

package v1

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
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

func init() {
	proto.RegisterFile("cometbft/services/version/v1/version_service.proto", fileDescriptor_054267f78f0fa7a9)
}

var fileDescriptor_054267f78f0fa7a9 = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0x4a, 0xce, 0xcf, 0x4d,
	0x2d, 0x49, 0x4a, 0x2b, 0xd1, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0x2d, 0xd6, 0x2f, 0x4b,
	0x2d, 0x2a, 0xce, 0xcc, 0xcf, 0xd3, 0x2f, 0x33, 0x84, 0x31, 0xe3, 0xa1, 0x72, 0x7a, 0x05, 0x45,
	0xf9, 0x25, 0xf9, 0x42, 0x32, 0x30, 0x3d, 0x7a, 0x30, 0x3d, 0x7a, 0x50, 0x85, 0x7a, 0x65, 0x86,
	0x52, 0x5a, 0xc4, 0x98, 0x08, 0x31, 0xc9, 0xa8, 0x91, 0x91, 0x8b, 0x2f, 0x0c, 0x22, 0x12, 0x0c,
	0x51, 0x2c, 0x94, 0xcf, 0xc5, 0xe5, 0x9e, 0x5a, 0x02, 0x15, 0x14, 0xd2, 0xd7, 0xc3, 0x67, 0x97,
	0x1e, 0x42, 0x65, 0x50, 0x6a, 0x61, 0x69, 0x6a, 0x71, 0x89, 0x94, 0x01, 0xf1, 0x1a, 0x8a, 0x0b,
	0xf2, 0xf3, 0x8a, 0x53, 0x9d, 0xc2, 0x4f, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1,
	0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21,
	0xca, 0x36, 0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x09, 0x64, 0xa2, 0x3e, 0xdc, 0x53, 0x70, 0x46, 0x62,
	0x41, 0xa6, 0x3e, 0x3e, 0xaf, 0x26, 0xb1, 0x81, 0xfd, 0x68, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff,
	0xbb, 0x08, 0xe5, 0x2b, 0x63, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// VersionServiceClient is the client API for VersionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type VersionServiceClient interface {
	// GetVersion retrieves version information about the node and the protocols
	// it implements.
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error)
}

type versionServiceClient struct {
	cc grpc1.ClientConn
}

func NewVersionServiceClient(cc grpc1.ClientConn) VersionServiceClient {
	return &versionServiceClient{cc}
}

func (c *versionServiceClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, "/cometbft.services.version.v1.VersionService/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VersionServiceServer is the server API for VersionService service.
type VersionServiceServer interface {
	// GetVersion retrieves version information about the node and the protocols
	// it implements.
	GetVersion(context.Context, *GetVersionRequest) (*GetVersionResponse, error)
}

// UnimplementedVersionServiceServer can be embedded to have forward compatible implementations.
type UnimplementedVersionServiceServer struct {
}

func (*UnimplementedVersionServiceServer) GetVersion(ctx context.Context, req *GetVersionRequest) (*GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}

func RegisterVersionServiceServer(s grpc1.Server, srv VersionServiceServer) {
	s.RegisterService(&_VersionService_serviceDesc, srv)
}

func _VersionService_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionServiceServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cometbft.services.version.v1.VersionService/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionServiceServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var VersionService_serviceDesc = _VersionService_serviceDesc
var _VersionService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cometbft.services.version.v1.VersionService",
	HandlerType: (*VersionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVersion",
			Handler:    _VersionService_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cometbft/services/version/v1/version_service.proto",
}
