// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/Jim3Things/CloudChamber/pkg/protos/common/common.proto

package common

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type BladeCapacity struct {
	Cores                float32                  `protobuf:"fixed32,1,opt,name=cores,proto3" json:"cores,omitempty"`
	MemoryInMb           int64                    `protobuf:"varint,2,opt,name=memory_in_mb,json=memoryInMb,proto3" json:"memory_in_mb,omitempty"`
	DiskInGb             int64                    `protobuf:"varint,3,opt,name=disk_in_gb,json=diskInGb,proto3" json:"disk_in_gb,omitempty"`
	NetworkBandwidth     int64                    `protobuf:"varint,4,opt,name=network_bandwidth,json=networkBandwidth,proto3" json:"network_bandwidth,omitempty"`
	Arch                 string                   `protobuf:"bytes,5,opt,name=arch,proto3" json:"arch,omitempty"`
	Gpus                 *BladeCapacityGpuDetails `protobuf:"bytes,6,opt,name=gpus,proto3" json:"gpus,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *BladeCapacity) Reset()         { *m = BladeCapacity{} }
func (m *BladeCapacity) String() string { return proto.CompactTextString(m) }
func (*BladeCapacity) ProtoMessage()    {}
func (*BladeCapacity) Descriptor() ([]byte, []int) {
	return fileDescriptor_c431e74178074209, []int{0}
}

func (m *BladeCapacity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BladeCapacity.Unmarshal(m, b)
}
func (m *BladeCapacity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BladeCapacity.Marshal(b, m, deterministic)
}
func (m *BladeCapacity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BladeCapacity.Merge(m, src)
}
func (m *BladeCapacity) XXX_Size() int {
	return xxx_messageInfo_BladeCapacity.Size(m)
}
func (m *BladeCapacity) XXX_DiscardUnknown() {
	xxx_messageInfo_BladeCapacity.DiscardUnknown(m)
}

var xxx_messageInfo_BladeCapacity proto.InternalMessageInfo

func (m *BladeCapacity) GetCores() float32 {
	if m != nil {
		return m.Cores
	}
	return 0
}

func (m *BladeCapacity) GetMemoryInMb() int64 {
	if m != nil {
		return m.MemoryInMb
	}
	return 0
}

func (m *BladeCapacity) GetDiskInGb() int64 {
	if m != nil {
		return m.DiskInGb
	}
	return 0
}

func (m *BladeCapacity) GetNetworkBandwidth() int64 {
	if m != nil {
		return m.NetworkBandwidth
	}
	return 0
}

func (m *BladeCapacity) GetArch() string {
	if m != nil {
		return m.Arch
	}
	return ""
}

func (m *BladeCapacity) GetGpus() *BladeCapacityGpuDetails {
	if m != nil {
		return m.Gpus
	}
	return nil
}

// GPUs may not be present at all on a blade, so if they are
// not, then these fields are missing.
type BladeCapacityGpuDetails struct {
	Units                int32    `protobuf:"varint,1,opt,name=units,proto3" json:"units,omitempty"`
	Arch                 string   `protobuf:"bytes,2,opt,name=arch,proto3" json:"arch,omitempty"`
	Present              bool     `protobuf:"varint,3,opt,name=present,proto3" json:"present,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BladeCapacityGpuDetails) Reset()         { *m = BladeCapacityGpuDetails{} }
func (m *BladeCapacityGpuDetails) String() string { return proto.CompactTextString(m) }
func (*BladeCapacityGpuDetails) ProtoMessage()    {}
func (*BladeCapacityGpuDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_c431e74178074209, []int{0, 0}
}

func (m *BladeCapacityGpuDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BladeCapacityGpuDetails.Unmarshal(m, b)
}
func (m *BladeCapacityGpuDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BladeCapacityGpuDetails.Marshal(b, m, deterministic)
}
func (m *BladeCapacityGpuDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BladeCapacityGpuDetails.Merge(m, src)
}
func (m *BladeCapacityGpuDetails) XXX_Size() int {
	return xxx_messageInfo_BladeCapacityGpuDetails.Size(m)
}
func (m *BladeCapacityGpuDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_BladeCapacityGpuDetails.DiscardUnknown(m)
}

var xxx_messageInfo_BladeCapacityGpuDetails proto.InternalMessageInfo

func (m *BladeCapacityGpuDetails) GetUnits() int32 {
	if m != nil {
		return m.Units
	}
	return 0
}

func (m *BladeCapacityGpuDetails) GetArch() string {
	if m != nil {
		return m.Arch
	}
	return ""
}

func (m *BladeCapacityGpuDetails) GetPresent() bool {
	if m != nil {
		return m.Present
	}
	return false
}

type Timestamp struct {
	Ticks                int64    `protobuf:"varint,1,opt,name=ticks,proto3" json:"ticks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Timestamp) Reset()         { *m = Timestamp{} }
func (m *Timestamp) String() string { return proto.CompactTextString(m) }
func (*Timestamp) ProtoMessage()    {}
func (*Timestamp) Descriptor() ([]byte, []int) {
	return fileDescriptor_c431e74178074209, []int{1}
}

func (m *Timestamp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Timestamp.Unmarshal(m, b)
}
func (m *Timestamp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Timestamp.Marshal(b, m, deterministic)
}
func (m *Timestamp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Timestamp.Merge(m, src)
}
func (m *Timestamp) XXX_Size() int {
	return xxx_messageInfo_Timestamp.Size(m)
}
func (m *Timestamp) XXX_DiscardUnknown() {
	xxx_messageInfo_Timestamp.DiscardUnknown(m)
}

var xxx_messageInfo_Timestamp proto.InternalMessageInfo

func (m *Timestamp) GetTicks() int64 {
	if m != nil {
		return m.Ticks
	}
	return 0
}

func init() {
	proto.RegisterType((*BladeCapacity)(nil), "common.blade_capacity")
	proto.RegisterType((*BladeCapacityGpuDetails)(nil), "common.blade_capacity.gpu_details")
	proto.RegisterType((*Timestamp)(nil), "common.timestamp")
}

func init() {
	proto.RegisterFile("github.com/Jim3Things/CloudChamber/pkg/protos/common/common.proto", fileDescriptor_c431e74178074209)
}

var fileDescriptor_c431e74178074209 = []byte{
	// 315 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0x3d, 0x6b, 0xeb, 0x30,
	0x14, 0x86, 0xb1, 0xf3, 0x71, 0x13, 0xe5, 0x72, 0xb9, 0x15, 0x1d, 0x44, 0xe9, 0xe0, 0x66, 0x32,
	0x14, 0x6c, 0x68, 0x4a, 0xf6, 0x26, 0x43, 0x49, 0xa1, 0x43, 0x45, 0xa7, 0x2e, 0x46, 0x92, 0x85,
	0x2d, 0x1c, 0x7d, 0x20, 0xc9, 0x84, 0xfc, 0xad, 0xfe, 0xc2, 0x62, 0x29, 0x29, 0xe9, 0xda, 0x49,
	0x7a, 0x9e, 0x73, 0x38, 0x47, 0x2f, 0x02, 0x4f, 0x8d, 0xf0, 0x6d, 0x4f, 0x0b, 0xa6, 0x65, 0xf9,
	0x22, 0xe4, 0xea, 0xbd, 0x15, 0xaa, 0x71, 0xe5, 0x76, 0xaf, 0xfb, 0x7a, 0xdb, 0x12, 0x49, 0xb9,
	0x2d, 0x4d, 0xd7, 0x94, 0xc6, 0x6a, 0xaf, 0x5d, 0xc9, 0xb4, 0x94, 0x5a, 0x9d, 0x8e, 0x22, 0x48,
	0x38, 0x8d, 0xb4, 0xfc, 0x4c, 0xc1, 0x3f, 0xba, 0x27, 0x35, 0xaf, 0x18, 0x31, 0x84, 0x09, 0x7f,
	0x84, 0xd7, 0x60, 0xc2, 0xb4, 0xe5, 0x0e, 0x25, 0x59, 0x92, 0xa7, 0x38, 0x02, 0xcc, 0xc0, 0x5f,
	0xc9, 0xa5, 0xb6, 0xc7, 0x4a, 0xa8, 0x4a, 0x52, 0x94, 0x66, 0x49, 0x3e, 0xc2, 0x20, 0xba, 0x9d,
	0x7a, 0xa5, 0xf0, 0x16, 0x80, 0x5a, 0xb8, 0x6e, 0xa8, 0x37, 0x14, 0x8d, 0x42, 0x7d, 0x36, 0x98,
	0x9d, 0x7a, 0xa6, 0xf0, 0x1e, 0x5c, 0x29, 0xee, 0x0f, 0xda, 0x76, 0x15, 0x25, 0xaa, 0x3e, 0x88,
	0xda, 0xb7, 0x68, 0x1c, 0x9a, 0xfe, 0x9f, 0x0a, 0x9b, 0xb3, 0x87, 0x10, 0x8c, 0x89, 0x65, 0x2d,
	0x9a, 0x64, 0x49, 0x3e, 0xc7, 0xe1, 0x0e, 0xd7, 0x60, 0xdc, 0x98, 0xde, 0xa1, 0x69, 0x96, 0xe4,
	0x8b, 0x87, 0x65, 0x71, 0x8a, 0xf3, 0xf3, 0xf1, 0x45, 0x63, 0xfa, 0xaa, 0xe6, 0x9e, 0x88, 0xbd,
	0xc3, 0xa1, 0xff, 0xe6, 0x0d, 0x2c, 0x2e, 0xe4, 0x90, 0xae, 0x57, 0xc2, 0xc7, 0x74, 0x13, 0x1c,
	0xe1, 0x7b, 0x61, 0x7a, 0xb1, 0x10, 0x81, 0x3f, 0xc6, 0x72, 0xc7, 0x95, 0x0f, 0x61, 0x66, 0xf8,
	0x8c, 0xcb, 0x3b, 0x30, 0xf7, 0x42, 0x72, 0xe7, 0x89, 0x34, 0xc3, 0x40, 0x2f, 0x58, 0x17, 0x07,
	0x8e, 0x70, 0x84, 0xcd, 0xfa, 0xe3, 0xf1, 0x37, 0x9f, 0x44, 0xa7, 0x01, 0x57, 0x5f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x9b, 0xe6, 0xfd, 0xa5, 0xe3, 0x01, 0x00, 0x00,
}
