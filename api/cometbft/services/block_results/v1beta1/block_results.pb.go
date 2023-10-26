// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cometbft/services/block_results/v1beta1/block_results.proto

package v1beta1

import (
	fmt "fmt"
	v1beta1 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta1"
	v1beta2 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta2"
	v1beta3 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta3"
	v1beta31 "github.com/cometbft/cometbft/api/cometbft/types/v1beta3"
	proto "github.com/cosmos/gogoproto/proto"
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

type GetBlockResultsRequest struct {
	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *GetBlockResultsRequest) Reset()         { *m = GetBlockResultsRequest{} }
func (m *GetBlockResultsRequest) String() string { return proto.CompactTextString(m) }
func (*GetBlockResultsRequest) ProtoMessage()    {}
func (*GetBlockResultsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_12a641024ad7106e, []int{0}
}
func (m *GetBlockResultsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetBlockResultsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetBlockResultsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetBlockResultsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlockResultsRequest.Merge(m, src)
}
func (m *GetBlockResultsRequest) XXX_Size() int {
	return m.Size()
}
func (m *GetBlockResultsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlockResultsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlockResultsRequest proto.InternalMessageInfo

func (m *GetBlockResultsRequest) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type GetLatestBlockResultsRequest struct {
}

func (m *GetLatestBlockResultsRequest) Reset()         { *m = GetLatestBlockResultsRequest{} }
func (m *GetLatestBlockResultsRequest) String() string { return proto.CompactTextString(m) }
func (*GetLatestBlockResultsRequest) ProtoMessage()    {}
func (*GetLatestBlockResultsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_12a641024ad7106e, []int{1}
}
func (m *GetLatestBlockResultsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetLatestBlockResultsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetLatestBlockResultsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetLatestBlockResultsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetLatestBlockResultsRequest.Merge(m, src)
}
func (m *GetLatestBlockResultsRequest) XXX_Size() int {
	return m.Size()
}
func (m *GetLatestBlockResultsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetLatestBlockResultsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetLatestBlockResultsRequest proto.InternalMessageInfo

type GetBlockResultsResponse struct {
	Height                int64                      `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	TxResults             []*v1beta3.ExecTxResult    `protobuf:"bytes,2,rep,name=tx_results,json=txResults,proto3" json:"tx_results,omitempty"`
	FinalizeBlockEvents   []*v1beta2.Event           `protobuf:"bytes,3,rep,name=finalize_block_events,json=finalizeBlockEvents,proto3" json:"finalize_block_events,omitempty"`
	ValidatorUpdates      []*v1beta1.ValidatorUpdate `protobuf:"bytes,4,rep,name=validator_updates,json=validatorUpdates,proto3" json:"validator_updates,omitempty"`
	ConsensusParamUpdates *v1beta31.ConsensusParams  `protobuf:"bytes,5,opt,name=consensus_param_updates,json=consensusParamUpdates,proto3" json:"consensus_param_updates,omitempty"`
	AppHash               []byte                     `protobuf:"bytes,6,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
}

func (m *GetBlockResultsResponse) Reset()         { *m = GetBlockResultsResponse{} }
func (m *GetBlockResultsResponse) String() string { return proto.CompactTextString(m) }
func (*GetBlockResultsResponse) ProtoMessage()    {}
func (*GetBlockResultsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_12a641024ad7106e, []int{2}
}
func (m *GetBlockResultsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetBlockResultsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetBlockResultsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetBlockResultsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlockResultsResponse.Merge(m, src)
}
func (m *GetBlockResultsResponse) XXX_Size() int {
	return m.Size()
}
func (m *GetBlockResultsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlockResultsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlockResultsResponse proto.InternalMessageInfo

func (m *GetBlockResultsResponse) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *GetBlockResultsResponse) GetTxResults() []*v1beta3.ExecTxResult {
	if m != nil {
		return m.TxResults
	}
	return nil
}

func (m *GetBlockResultsResponse) GetFinalizeBlockEvents() []*v1beta2.Event {
	if m != nil {
		return m.FinalizeBlockEvents
	}
	return nil
}

func (m *GetBlockResultsResponse) GetValidatorUpdates() []*v1beta1.ValidatorUpdate {
	if m != nil {
		return m.ValidatorUpdates
	}
	return nil
}

func (m *GetBlockResultsResponse) GetConsensusParamUpdates() *v1beta31.ConsensusParams {
	if m != nil {
		return m.ConsensusParamUpdates
	}
	return nil
}

func (m *GetBlockResultsResponse) GetAppHash() []byte {
	if m != nil {
		return m.AppHash
	}
	return nil
}

type GetLatestBlockResultsResponse struct {
	Height                int64                      `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	TxResults             []*v1beta3.ExecTxResult    `protobuf:"bytes,2,rep,name=tx_results,json=txResults,proto3" json:"tx_results,omitempty"`
	FinalizeBlockEvents   []*v1beta2.Event           `protobuf:"bytes,3,rep,name=finalize_block_events,json=finalizeBlockEvents,proto3" json:"finalize_block_events,omitempty"`
	ValidatorUpdates      []*v1beta1.ValidatorUpdate `protobuf:"bytes,4,rep,name=validator_updates,json=validatorUpdates,proto3" json:"validator_updates,omitempty"`
	ConsensusParamUpdates *v1beta31.ConsensusParams  `protobuf:"bytes,5,opt,name=consensus_param_updates,json=consensusParamUpdates,proto3" json:"consensus_param_updates,omitempty"`
	AppHash               []byte                     `protobuf:"bytes,6,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
}

func (m *GetLatestBlockResultsResponse) Reset()         { *m = GetLatestBlockResultsResponse{} }
func (m *GetLatestBlockResultsResponse) String() string { return proto.CompactTextString(m) }
func (*GetLatestBlockResultsResponse) ProtoMessage()    {}
func (*GetLatestBlockResultsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_12a641024ad7106e, []int{3}
}
func (m *GetLatestBlockResultsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetLatestBlockResultsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetLatestBlockResultsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GetLatestBlockResultsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetLatestBlockResultsResponse.Merge(m, src)
}
func (m *GetLatestBlockResultsResponse) XXX_Size() int {
	return m.Size()
}
func (m *GetLatestBlockResultsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetLatestBlockResultsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetLatestBlockResultsResponse proto.InternalMessageInfo

func (m *GetLatestBlockResultsResponse) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *GetLatestBlockResultsResponse) GetTxResults() []*v1beta3.ExecTxResult {
	if m != nil {
		return m.TxResults
	}
	return nil
}

func (m *GetLatestBlockResultsResponse) GetFinalizeBlockEvents() []*v1beta2.Event {
	if m != nil {
		return m.FinalizeBlockEvents
	}
	return nil
}

func (m *GetLatestBlockResultsResponse) GetValidatorUpdates() []*v1beta1.ValidatorUpdate {
	if m != nil {
		return m.ValidatorUpdates
	}
	return nil
}

func (m *GetLatestBlockResultsResponse) GetConsensusParamUpdates() *v1beta31.ConsensusParams {
	if m != nil {
		return m.ConsensusParamUpdates
	}
	return nil
}

func (m *GetLatestBlockResultsResponse) GetAppHash() []byte {
	if m != nil {
		return m.AppHash
	}
	return nil
}

func init() {
	proto.RegisterType((*GetBlockResultsRequest)(nil), "cometbft.services.block_results.v1beta1.GetBlockResultsRequest")
	proto.RegisterType((*GetLatestBlockResultsRequest)(nil), "cometbft.services.block_results.v1beta1.GetLatestBlockResultsRequest")
	proto.RegisterType((*GetBlockResultsResponse)(nil), "cometbft.services.block_results.v1beta1.GetBlockResultsResponse")
	proto.RegisterType((*GetLatestBlockResultsResponse)(nil), "cometbft.services.block_results.v1beta1.GetLatestBlockResultsResponse")
}

func init() {
	proto.RegisterFile("cometbft/services/block_results/v1beta1/block_results.proto", fileDescriptor_12a641024ad7106e)
}

var fileDescriptor_12a641024ad7106e = []byte{
	// 444 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xec, 0x94, 0xbf, 0x6e, 0xd4, 0x40,
	0x10, 0xc6, 0xcf, 0x18, 0x0e, 0xd8, 0x50, 0x80, 0x51, 0x12, 0x83, 0x82, 0x75, 0x5c, 0x24, 0x72,
	0x95, 0xcd, 0xdd, 0x95, 0x74, 0x87, 0xa2, 0xa4, 0xa0, 0x88, 0x96, 0x3f, 0x05, 0x8d, 0xb5, 0x76,
	0x26, 0xf1, 0x8a, 0x8b, 0xbd, 0x78, 0xc6, 0xd6, 0xc1, 0x53, 0x50, 0x52, 0xf1, 0x3c, 0x94, 0x29,
	0x29, 0xd1, 0xdd, 0x8b, 0x20, 0xaf, 0xbd, 0x8e, 0x2c, 0x1c, 0x41, 0x49, 0x41, 0x67, 0x8f, 0xbe,
	0xef, 0xb7, 0x33, 0xf3, 0x69, 0x97, 0xbd, 0x88, 0xb3, 0x0b, 0xa0, 0xe8, 0x8c, 0x02, 0x84, 0xbc,
	0x94, 0x31, 0x60, 0x10, 0x2d, 0xb3, 0xf8, 0x43, 0x98, 0x03, 0x16, 0x4b, 0xc2, 0xa0, 0x9c, 0x46,
	0x40, 0x62, 0xda, 0xad, 0xfa, 0x2a, 0xcf, 0x28, 0x73, 0x0e, 0x8c, 0xd9, 0x37, 0x66, 0xbf, 0x2b,
	0x6b, 0xcc, 0x8f, 0x9f, 0xb6, 0xa7, 0x88, 0x28, 0x96, 0x2d, 0x93, 0x3e, 0x29, 0x68, 0x58, 0xfd,
	0x92, 0xd9, 0x9f, 0x25, 0xf3, 0x8e, 0x64, 0xbf, 0x95, 0xe8, 0x6a, 0xab, 0x51, 0x22, 0x17, 0x17,
	0x8d, 0x68, 0xfc, 0x9c, 0xed, 0x1c, 0x01, 0x2d, 0xaa, 0x4e, 0x79, 0xdd, 0x28, 0x87, 0x8f, 0x05,
	0x20, 0x39, 0x3b, 0x6c, 0x98, 0x80, 0x3c, 0x4f, 0xc8, 0xb5, 0x46, 0xd6, 0xc4, 0xe6, 0xcd, 0xdf,
	0xd8, 0x63, 0x7b, 0x47, 0x40, 0xaf, 0x04, 0x01, 0xf6, 0xf9, 0xc6, 0x5f, 0x6d, 0xb6, 0xfb, 0x1b,
	0x12, 0x55, 0x96, 0x22, 0x5c, 0xc7, 0x74, 0x16, 0x8c, 0xd1, 0xca, 0x6c, 0xca, 0xbd, 0x31, 0xb2,
	0x27, 0x5b, 0xb3, 0x7d, 0xbf, 0xdd, 0x68, 0x35, 0x62, 0xb3, 0xbf, 0xb9, 0x7f, 0xb8, 0x82, 0xf8,
	0xcd, 0xaa, 0x26, 0xf3, 0xbb, 0xd4, 0x7c, 0xa1, 0x73, 0xc2, 0xb6, 0xcf, 0x64, 0x2a, 0x96, 0xf2,
	0x33, 0x84, 0xf5, 0xe6, 0xa1, 0x84, 0x94, 0xd0, 0xb5, 0x35, 0x6e, 0xaf, 0x17, 0x37, 0xf3, 0x0f,
	0x2b, 0x11, 0x7f, 0x68, 0xac, 0xba, 0x6d, 0x5d, 0x43, 0xe7, 0x35, 0x7b, 0x50, 0x8a, 0xa5, 0x3c,
	0x15, 0x94, 0xe5, 0x61, 0xa1, 0x4e, 0xab, 0x91, 0xdd, 0x9b, 0x9a, 0xf6, 0xac, 0x97, 0x36, 0xf5,
	0xdf, 0x19, 0xfd, 0x5b, 0x2d, 0xe7, 0xf7, 0xcb, 0x6e, 0x01, 0x9d, 0x90, 0xed, 0xc6, 0xd5, 0x2e,
	0x52, 0x2c, 0x30, 0xd4, 0x51, 0xb4, 0xe8, 0x5b, 0x23, 0x6b, 0xb2, 0x35, 0x3b, 0xb8, 0x42, 0xd7,
	0x69, 0x9a, 0xc1, 0x5f, 0x1a, 0xdb, 0x89, 0x0e, 0x90, 0x6f, 0xc7, 0x9d, 0x82, 0x39, 0xe0, 0x11,
	0xbb, 0x23, 0x94, 0x0a, 0x13, 0x81, 0x89, 0x3b, 0x1c, 0x59, 0x93, 0x7b, 0xfc, 0xb6, 0x50, 0xea,
	0x58, 0x60, 0x32, 0xfe, 0x66, 0xb3, 0x27, 0xd7, 0x64, 0xf7, 0x3f, 0xa0, 0x7f, 0x20, 0xa0, 0x45,
	0xf4, 0x7d, 0xed, 0x59, 0x97, 0x6b, 0xcf, 0xfa, 0xb9, 0xf6, 0xac, 0x2f, 0x1b, 0x6f, 0x70, 0xb9,
	0xf1, 0x06, 0x3f, 0x36, 0xde, 0xe0, 0xfd, 0xf1, 0xb9, 0xa4, 0xa4, 0x88, 0xaa, 0xa3, 0x83, 0xf6,
	0x5e, 0x5f, 0xbd, 0x01, 0x4a, 0x06, 0x7f, 0xf9, 0x78, 0x45, 0x43, 0x7d, 0xf1, 0xe7, 0xbf, 0x02,
	0x00, 0x00, 0xff, 0xff, 0xc3, 0xe2, 0x59, 0xb2, 0xee, 0x04, 0x00, 0x00,
}

func (m *GetBlockResultsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetBlockResultsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetBlockResultsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintBlockResults(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *GetLatestBlockResultsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetLatestBlockResultsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetLatestBlockResultsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *GetBlockResultsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetBlockResultsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetBlockResultsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AppHash) > 0 {
		i -= len(m.AppHash)
		copy(dAtA[i:], m.AppHash)
		i = encodeVarintBlockResults(dAtA, i, uint64(len(m.AppHash)))
		i--
		dAtA[i] = 0x32
	}
	if m.ConsensusParamUpdates != nil {
		{
			size, err := m.ConsensusParamUpdates.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintBlockResults(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if len(m.ValidatorUpdates) > 0 {
		for iNdEx := len(m.ValidatorUpdates) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorUpdates[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.FinalizeBlockEvents) > 0 {
		for iNdEx := len(m.FinalizeBlockEvents) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FinalizeBlockEvents[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.TxResults) > 0 {
		for iNdEx := len(m.TxResults) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TxResults[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Height != 0 {
		i = encodeVarintBlockResults(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *GetLatestBlockResultsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetLatestBlockResultsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GetLatestBlockResultsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AppHash) > 0 {
		i -= len(m.AppHash)
		copy(dAtA[i:], m.AppHash)
		i = encodeVarintBlockResults(dAtA, i, uint64(len(m.AppHash)))
		i--
		dAtA[i] = 0x32
	}
	if m.ConsensusParamUpdates != nil {
		{
			size, err := m.ConsensusParamUpdates.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintBlockResults(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if len(m.ValidatorUpdates) > 0 {
		for iNdEx := len(m.ValidatorUpdates) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorUpdates[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.FinalizeBlockEvents) > 0 {
		for iNdEx := len(m.FinalizeBlockEvents) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FinalizeBlockEvents[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.TxResults) > 0 {
		for iNdEx := len(m.TxResults) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TxResults[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintBlockResults(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Height != 0 {
		i = encodeVarintBlockResults(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintBlockResults(dAtA []byte, offset int, v uint64) int {
	offset -= sovBlockResults(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GetBlockResultsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovBlockResults(uint64(m.Height))
	}
	return n
}

func (m *GetLatestBlockResultsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *GetBlockResultsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovBlockResults(uint64(m.Height))
	}
	if len(m.TxResults) > 0 {
		for _, e := range m.TxResults {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if len(m.FinalizeBlockEvents) > 0 {
		for _, e := range m.FinalizeBlockEvents {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if len(m.ValidatorUpdates) > 0 {
		for _, e := range m.ValidatorUpdates {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if m.ConsensusParamUpdates != nil {
		l = m.ConsensusParamUpdates.Size()
		n += 1 + l + sovBlockResults(uint64(l))
	}
	l = len(m.AppHash)
	if l > 0 {
		n += 1 + l + sovBlockResults(uint64(l))
	}
	return n
}

func (m *GetLatestBlockResultsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Height != 0 {
		n += 1 + sovBlockResults(uint64(m.Height))
	}
	if len(m.TxResults) > 0 {
		for _, e := range m.TxResults {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if len(m.FinalizeBlockEvents) > 0 {
		for _, e := range m.FinalizeBlockEvents {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if len(m.ValidatorUpdates) > 0 {
		for _, e := range m.ValidatorUpdates {
			l = e.Size()
			n += 1 + l + sovBlockResults(uint64(l))
		}
	}
	if m.ConsensusParamUpdates != nil {
		l = m.ConsensusParamUpdates.Size()
		n += 1 + l + sovBlockResults(uint64(l))
	}
	l = len(m.AppHash)
	if l > 0 {
		n += 1 + l + sovBlockResults(uint64(l))
	}
	return n
}

func sovBlockResults(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozBlockResults(x uint64) (n int) {
	return sovBlockResults(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GetBlockResultsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBlockResults
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
			return fmt.Errorf("proto: GetBlockResultsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetBlockResultsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
			skippy, err := skipBlockResults(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBlockResults
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
func (m *GetLatestBlockResultsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBlockResults
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
			return fmt.Errorf("proto: GetLatestBlockResultsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetLatestBlockResultsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipBlockResults(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBlockResults
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
func (m *GetBlockResultsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBlockResults
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
			return fmt.Errorf("proto: GetBlockResultsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetBlockResultsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TxResults", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TxResults = append(m.TxResults, &v1beta3.ExecTxResult{})
			if err := m.TxResults[len(m.TxResults)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinalizeBlockEvents", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FinalizeBlockEvents = append(m.FinalizeBlockEvents, &v1beta2.Event{})
			if err := m.FinalizeBlockEvents[len(m.FinalizeBlockEvents)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorUpdates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorUpdates = append(m.ValidatorUpdates, &v1beta1.ValidatorUpdate{})
			if err := m.ValidatorUpdates[len(m.ValidatorUpdates)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsensusParamUpdates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ConsensusParamUpdates == nil {
				m.ConsensusParamUpdates = &v1beta31.ConsensusParams{}
			}
			if err := m.ConsensusParamUpdates.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppHash = append(m.AppHash[:0], dAtA[iNdEx:postIndex]...)
			if m.AppHash == nil {
				m.AppHash = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBlockResults(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBlockResults
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
func (m *GetLatestBlockResultsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBlockResults
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
			return fmt.Errorf("proto: GetLatestBlockResultsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetLatestBlockResultsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TxResults", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TxResults = append(m.TxResults, &v1beta3.ExecTxResult{})
			if err := m.TxResults[len(m.TxResults)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinalizeBlockEvents", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FinalizeBlockEvents = append(m.FinalizeBlockEvents, &v1beta2.Event{})
			if err := m.FinalizeBlockEvents[len(m.FinalizeBlockEvents)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorUpdates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorUpdates = append(m.ValidatorUpdates, &v1beta1.ValidatorUpdate{})
			if err := m.ValidatorUpdates[len(m.ValidatorUpdates)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsensusParamUpdates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ConsensusParamUpdates == nil {
				m.ConsensusParamUpdates = &v1beta31.ConsensusParams{}
			}
			if err := m.ConsensusParamUpdates.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBlockResults
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
				return ErrInvalidLengthBlockResults
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthBlockResults
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppHash = append(m.AppHash[:0], dAtA[iNdEx:postIndex]...)
			if m.AppHash == nil {
				m.AppHash = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBlockResults(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthBlockResults
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
func skipBlockResults(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBlockResults
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
					return 0, ErrIntOverflowBlockResults
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
					return 0, ErrIntOverflowBlockResults
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
				return 0, ErrInvalidLengthBlockResults
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupBlockResults
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthBlockResults
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthBlockResults        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBlockResults          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupBlockResults = fmt.Errorf("proto: unexpected end of group")
)
