// Code generated by protoc-gen-go. DO NOT EDIT.
// proto/beacon/rpc/v1/services.proto is a deprecated file.

package ethereum_beacon_rpc_v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/empty"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ValidatorRole int32 // Deprecated: Do not use.
const (
	ValidatorRole_UNKNOWN    ValidatorRole = 0
	ValidatorRole_ATTESTER   ValidatorRole = 1
	ValidatorRole_PROPOSER   ValidatorRole = 2
	ValidatorRole_AGGREGATOR ValidatorRole = 3
)

var ValidatorRole_name = map[int32]string{
	0: "UNKNOWN",
	1: "ATTESTER",
	2: "PROPOSER",
	3: "AGGREGATOR",
}

var ValidatorRole_value = map[string]int32{
	"UNKNOWN":    0,
	"ATTESTER":   1,
	"PROPOSER":   2,
	"AGGREGATOR": 3,
}

func (x ValidatorRole) String() string {
	return proto.EnumName(ValidatorRole_name, int32(x))
}

func (ValidatorRole) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_9eb4e94b85965285, []int{0}
}

// Deprecated: Do not use.
type ExitedValidatorsRequest struct {
	PublicKeys           [][]byte `protobuf:"bytes,1,rep,name=public_keys,json=publicKeys,proto3" json:"public_keys,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExitedValidatorsRequest) Reset()         { *m = ExitedValidatorsRequest{} }
func (m *ExitedValidatorsRequest) String() string { return proto.CompactTextString(m) }
func (*ExitedValidatorsRequest) ProtoMessage()    {}
func (*ExitedValidatorsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9eb4e94b85965285, []int{0}
}

func (m *ExitedValidatorsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitedValidatorsRequest.Unmarshal(m, b)
}
func (m *ExitedValidatorsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitedValidatorsRequest.Marshal(b, m, deterministic)
}
func (m *ExitedValidatorsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitedValidatorsRequest.Merge(m, src)
}
func (m *ExitedValidatorsRequest) XXX_Size() int {
	return xxx_messageInfo_ExitedValidatorsRequest.Size(m)
}
func (m *ExitedValidatorsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitedValidatorsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ExitedValidatorsRequest proto.InternalMessageInfo

func (m *ExitedValidatorsRequest) GetPublicKeys() [][]byte {
	if m != nil {
		return m.PublicKeys
	}
	return nil
}

// Deprecated: Do not use.
type ExitedValidatorsResponse struct {
	PublicKeys           [][]byte `protobuf:"bytes,1,rep,name=public_keys,json=publicKeys,proto3" json:"public_keys,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExitedValidatorsResponse) Reset()         { *m = ExitedValidatorsResponse{} }
func (m *ExitedValidatorsResponse) String() string { return proto.CompactTextString(m) }
func (*ExitedValidatorsResponse) ProtoMessage()    {}
func (*ExitedValidatorsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9eb4e94b85965285, []int{1}
}

func (m *ExitedValidatorsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitedValidatorsResponse.Unmarshal(m, b)
}
func (m *ExitedValidatorsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitedValidatorsResponse.Marshal(b, m, deterministic)
}
func (m *ExitedValidatorsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitedValidatorsResponse.Merge(m, src)
}
func (m *ExitedValidatorsResponse) XXX_Size() int {
	return xxx_messageInfo_ExitedValidatorsResponse.Size(m)
}
func (m *ExitedValidatorsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitedValidatorsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ExitedValidatorsResponse proto.InternalMessageInfo

func (m *ExitedValidatorsResponse) GetPublicKeys() [][]byte {
	if m != nil {
		return m.PublicKeys
	}
	return nil
}

func init() {
	proto.RegisterEnum("ethereum.beacon.rpc.v1.ValidatorRole", ValidatorRole_name, ValidatorRole_value)
	proto.RegisterType((*ExitedValidatorsRequest)(nil), "ethereum.beacon.rpc.v1.ExitedValidatorsRequest")
	proto.RegisterType((*ExitedValidatorsResponse)(nil), "ethereum.beacon.rpc.v1.ExitedValidatorsResponse")
}

func init() {
	proto.RegisterFile("proto/beacon/rpc/v1/services.proto", fileDescriptor_9eb4e94b85965285)
}

var fileDescriptor_9eb4e94b85965285 = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x8f, 0xb1, 0x4b, 0xc3, 0x40,
	0x14, 0xc6, 0xbd, 0x04, 0x54, 0x5e, 0xab, 0x84, 0x1b, 0x34, 0xd4, 0x41, 0xc9, 0x24, 0x0e, 0x39,
	0x8a, 0x9b, 0x83, 0x12, 0xe1, 0xc8, 0x50, 0x49, 0xca, 0x35, 0xea, 0x28, 0xc9, 0xf5, 0x59, 0x83,
	0x69, 0xef, 0xbc, 0xbb, 0x04, 0xf3, 0x9f, 0xf9, 0xe7, 0x89, 0x09, 0xb8, 0xb8, 0x38, 0xbe, 0xf7,
	0xfd, 0x7e, 0x1f, 0x7c, 0x10, 0x69, 0xa3, 0x9c, 0x62, 0x15, 0x96, 0x52, 0xed, 0x98, 0xd1, 0x92,
	0x75, 0x73, 0x66, 0xd1, 0x74, 0xb5, 0x44, 0x1b, 0x0f, 0x21, 0x3d, 0x41, 0xf7, 0x86, 0x06, 0xdb,
	0x6d, 0x3c, 0x62, 0xb1, 0xd1, 0x32, 0xee, 0xe6, 0xb3, 0xb3, 0x8d, 0x52, 0x9b, 0x06, 0xd9, 0x40,
	0x55, 0xed, 0x2b, 0xc3, 0xad, 0x76, 0xfd, 0x28, 0x45, 0xb7, 0x70, 0xca, 0x3f, 0x6b, 0x87, 0xeb,
	0xa7, 0xb2, 0xa9, 0xd7, 0xa5, 0x53, 0xc6, 0x0a, 0xfc, 0x68, 0xd1, 0x3a, 0x7a, 0x0e, 0x13, 0xdd,
	0x56, 0x4d, 0x2d, 0x5f, 0xde, 0xb1, 0xb7, 0x21, 0xb9, 0xf0, 0x2f, 0xa7, 0x02, 0xc6, 0xd7, 0x02,
	0x7b, 0x7b, 0xe3, 0x85, 0x24, 0xba, 0x83, 0xf0, 0xaf, 0x6f, 0xb5, 0xda, 0x59, 0xfc, 0x57, 0xc1,
	0xd5, 0x03, 0x1c, 0xfd, 0xaa, 0x42, 0x35, 0x48, 0x27, 0x70, 0xf0, 0x98, 0x2d, 0xb2, 0xfc, 0x39,
	0x0b, 0xf6, 0xe8, 0x14, 0x0e, 0x93, 0xa2, 0xe0, 0xab, 0x82, 0x8b, 0x80, 0xfc, 0x5c, 0x4b, 0x91,
	0x2f, 0xf3, 0x15, 0x17, 0x81, 0x47, 0x8f, 0x01, 0x92, 0x34, 0x15, 0x3c, 0x4d, 0x8a, 0x5c, 0x04,
	0xfe, 0xcc, 0x0b, 0xc9, 0xbd, 0xff, 0x45, 0x48, 0xb5, 0x3f, 0x4c, 0xbb, 0xfe, 0x0e, 0x00, 0x00,
	0xff, 0xff, 0x15, 0x01, 0x10, 0xf7, 0x35, 0x01, 0x00, 0x00,
}
