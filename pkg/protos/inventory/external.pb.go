// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/Jim3Things/CloudChamber/pkg/protos/inventory/external.proto

package inventory

import (
	fmt "fmt"
	common "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

type External struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *External) Reset()         { *m = External{} }
func (m *External) String() string { return proto.CompactTextString(m) }
func (*External) ProtoMessage()    {}
func (*External) Descriptor() ([]byte, []int) {
	return fileDescriptor_687eec8f588ec561, []int{0}
}

func (m *External) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_External.Unmarshal(m, b)
}
func (m *External) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_External.Marshal(b, m, deterministic)
}
func (m *External) XXX_Merge(src proto.Message) {
	xxx_messageInfo_External.Merge(m, src)
}
func (m *External) XXX_Size() int {
	return xxx_messageInfo_External.Size(m)
}
func (m *External) XXX_DiscardUnknown() {
	xxx_messageInfo_External.DiscardUnknown(m)
}

var xxx_messageInfo_External proto.InternalMessageInfo

// Power distribution unit.  Network accessible power controller
type ExternalPdu struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExternalPdu) Reset()         { *m = ExternalPdu{} }
func (m *ExternalPdu) String() string { return proto.CompactTextString(m) }
func (*ExternalPdu) ProtoMessage()    {}
func (*ExternalPdu) Descriptor() ([]byte, []int) {
	return fileDescriptor_687eec8f588ec561, []int{0, 0}
}

func (m *ExternalPdu) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExternalPdu.Unmarshal(m, b)
}
func (m *ExternalPdu) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExternalPdu.Marshal(b, m, deterministic)
}
func (m *ExternalPdu) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExternalPdu.Merge(m, src)
}
func (m *ExternalPdu) XXX_Size() int {
	return xxx_messageInfo_ExternalPdu.Size(m)
}
func (m *ExternalPdu) XXX_DiscardUnknown() {
	xxx_messageInfo_ExternalPdu.DiscardUnknown(m)
}

var xxx_messageInfo_ExternalPdu proto.InternalMessageInfo

// Rack-level network switch.
type ExternalTor struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExternalTor) Reset()         { *m = ExternalTor{} }
func (m *ExternalTor) String() string { return proto.CompactTextString(m) }
func (*ExternalTor) ProtoMessage()    {}
func (*ExternalTor) Descriptor() ([]byte, []int) {
	return fileDescriptor_687eec8f588ec561, []int{0, 1}
}

func (m *ExternalTor) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExternalTor.Unmarshal(m, b)
}
func (m *ExternalTor) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExternalTor.Marshal(b, m, deterministic)
}
func (m *ExternalTor) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExternalTor.Merge(m, src)
}
func (m *ExternalTor) XXX_Size() int {
	return xxx_messageInfo_ExternalTor.Size(m)
}
func (m *ExternalTor) XXX_DiscardUnknown() {
	xxx_messageInfo_ExternalTor.DiscardUnknown(m)
}

var xxx_messageInfo_ExternalTor proto.InternalMessageInfo

type ExternalRack struct {
	Pdu *ExternalPdu `protobuf:"bytes,1,opt,name=pdu,proto3" json:"pdu,omitempty"`
	Tor *ExternalTor `protobuf:"bytes,2,opt,name=tor,proto3" json:"tor,omitempty"`
	// specify the blades in the rack.  Each blade is defined by an integer index within that rack, which is used
	// here as the key.
	Blades               map[int64]*common.BladeCapacity `protobuf:"bytes,3,rep,name=blades,proto3" json:"blades,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *ExternalRack) Reset()         { *m = ExternalRack{} }
func (m *ExternalRack) String() string { return proto.CompactTextString(m) }
func (*ExternalRack) ProtoMessage()    {}
func (*ExternalRack) Descriptor() ([]byte, []int) {
	return fileDescriptor_687eec8f588ec561, []int{0, 2}
}

func (m *ExternalRack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExternalRack.Unmarshal(m, b)
}
func (m *ExternalRack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExternalRack.Marshal(b, m, deterministic)
}
func (m *ExternalRack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExternalRack.Merge(m, src)
}
func (m *ExternalRack) XXX_Size() int {
	return xxx_messageInfo_ExternalRack.Size(m)
}
func (m *ExternalRack) XXX_DiscardUnknown() {
	xxx_messageInfo_ExternalRack.DiscardUnknown(m)
}

var xxx_messageInfo_ExternalRack proto.InternalMessageInfo

func (m *ExternalRack) GetPdu() *ExternalPdu {
	if m != nil {
		return m.Pdu
	}
	return nil
}

func (m *ExternalRack) GetTor() *ExternalTor {
	if m != nil {
		return m.Tor
	}
	return nil
}

func (m *ExternalRack) GetBlades() map[int64]*common.BladeCapacity {
	if m != nil {
		return m.Blades
	}
	return nil
}

// Finally, a zone is a collection of racks.  Each rack has a name, which is used as a key in the map below.
type ExternalZone struct {
	Racks                map[string]*ExternalRack `protobuf:"bytes,1,rep,name=racks,proto3" json:"racks,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ExternalZone) Reset()         { *m = ExternalZone{} }
func (m *ExternalZone) String() string { return proto.CompactTextString(m) }
func (*ExternalZone) ProtoMessage()    {}
func (*ExternalZone) Descriptor() ([]byte, []int) {
	return fileDescriptor_687eec8f588ec561, []int{0, 3}
}

func (m *ExternalZone) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExternalZone.Unmarshal(m, b)
}
func (m *ExternalZone) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExternalZone.Marshal(b, m, deterministic)
}
func (m *ExternalZone) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExternalZone.Merge(m, src)
}
func (m *ExternalZone) XXX_Size() int {
	return xxx_messageInfo_ExternalZone.Size(m)
}
func (m *ExternalZone) XXX_DiscardUnknown() {
	xxx_messageInfo_ExternalZone.DiscardUnknown(m)
}

var xxx_messageInfo_ExternalZone proto.InternalMessageInfo

func (m *ExternalZone) GetRacks() map[string]*ExternalRack {
	if m != nil {
		return m.Racks
	}
	return nil
}

func init() {
	proto.RegisterType((*External)(nil), "inventory.external")
	proto.RegisterType((*ExternalPdu)(nil), "inventory.external.pdu")
	proto.RegisterType((*ExternalTor)(nil), "inventory.external.tor")
	proto.RegisterType((*ExternalRack)(nil), "inventory.external.rack")
	proto.RegisterMapType((map[int64]*common.BladeCapacity)(nil), "inventory.external.rack.BladesEntry")
	proto.RegisterType((*ExternalZone)(nil), "inventory.external.zone")
	proto.RegisterMapType((map[string]*ExternalRack)(nil), "inventory.external.zone.RacksEntry")
}

func init() {
	proto.RegisterFile("github.com/Jim3Things/CloudChamber/pkg/protos/inventory/external.proto", fileDescriptor_687eec8f588ec561)
}

var fileDescriptor_687eec8f588ec561 = []byte{
	// 365 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xbd, 0x4e, 0xf3, 0x30,
	0x14, 0x86, 0xe5, 0xa6, 0xa9, 0x5a, 0x77, 0xf9, 0x94, 0xe1, 0x23, 0xca, 0x54, 0x01, 0x43, 0x91,
	0xa8, 0x2d, 0xb5, 0x03, 0x3f, 0x1b, 0x89, 0xe8, 0xc0, 0x46, 0xc4, 0xc4, 0x82, 0x9c, 0xc4, 0x4a,
	0xa3, 0x26, 0x76, 0xe4, 0x38, 0x51, 0xc3, 0xa5, 0xb0, 0x70, 0x13, 0xdc, 0x1b, 0x12, 0x13, 0x72,
	0xdc, 0x9f, 0x00, 0xed, 0x00, 0x53, 0x2c, 0x9f, 0xe7, 0xbc, 0x79, 0xce, 0x91, 0xe1, 0x3c, 0x4e,
	0xe4, 0xa2, 0x0c, 0x50, 0xc8, 0x33, 0x7c, 0x97, 0x64, 0xb3, 0x87, 0x45, 0xc2, 0xe2, 0x02, 0x7b,
	0x29, 0x2f, 0x23, 0x6f, 0x41, 0xb2, 0x80, 0x0a, 0x9c, 0x2f, 0x63, 0x9c, 0x0b, 0x2e, 0x79, 0x81,
	0x13, 0x56, 0x51, 0x26, 0xb9, 0xa8, 0x31, 0x5d, 0x49, 0x2a, 0x18, 0x49, 0x51, 0x53, 0xb1, 0x06,
	0xdb, 0x8a, 0x73, 0xd3, 0x8a, 0xa4, 0xac, 0xe2, 0x75, 0x2e, 0xf8, 0xaa, 0xd6, 0x09, 0xe1, 0x24,
	0xa6, 0x6c, 0x52, 0x91, 0x34, 0x89, 0x88, 0xa4, 0xf8, 0xc7, 0x41, 0xa7, 0x39, 0xde, 0xef, 0xac,
	0x42, 0x9e, 0x65, 0x9c, 0xe1, 0x90, 0xe4, 0x24, 0x4c, 0x64, 0xad, 0x43, 0x8e, 0xdf, 0x0c, 0xd8,
	0xdf, 0x58, 0x3a, 0x26, 0x34, 0xf2, 0xa8, 0x54, 0x1f, 0xc9, 0x85, 0xf3, 0x0e, 0x60, 0x57, 0x90,
	0x70, 0x69, 0x9d, 0x35, 0xd7, 0x36, 0x18, 0x81, 0xf1, 0x70, 0x7a, 0x84, 0xb6, 0x43, 0xa0, 0xdd,
	0x78, 0x51, 0xe9, 0x2b, 0x46, 0xa1, 0x92, 0x0b, 0xbb, 0x73, 0x18, 0x95, 0x5c, 0xf8, 0x8a, 0xb1,
	0xe6, 0xb0, 0x17, 0xa4, 0x24, 0xa2, 0x85, 0x6d, 0x8c, 0x8c, 0xf1, 0x70, 0x7a, 0xba, 0x8f, 0x56,
	0xff, 0x47, 0x6e, 0x83, 0xdd, 0x32, 0x29, 0x6a, 0xb7, 0xff, 0xe1, 0x9a, 0x2f, 0xa0, 0xd3, 0x07,
	0xfe, 0xba, 0xdb, 0xb9, 0x87, 0xc3, 0x16, 0x60, 0xfd, 0x83, 0xc6, 0x92, 0xd6, 0x8d, 0xac, 0xe1,
	0xab, 0xa3, 0x75, 0x0e, 0xcd, 0x8a, 0xa4, 0x25, 0x5d, 0x5b, 0xfd, 0x47, 0x7a, 0x13, 0xa8, 0xe9,
	0x7f, 0xda, 0xec, 0xc3, 0xd7, 0xd0, 0x75, 0xe7, 0x12, 0x38, 0xaf, 0x00, 0x76, 0x9f, 0x39, 0xa3,
	0x96, 0x07, 0x4d, 0x65, 0x50, 0xd8, 0xa0, 0x51, 0x3c, 0xd9, 0xa7, 0xa8, 0x40, 0xe4, 0x2b, 0xea,
	0xbb, 0xa1, 0xee, 0x75, 0x7c, 0x08, 0x77, 0xe5, 0xb6, 0xdf, 0x40, 0xfb, 0xa1, 0xaf, 0x7e, 0xf6,
	0xa1, 0x3d, 0xb4, 0x0c, 0xdd, 0xab, 0xc7, 0x8b, 0x3f, 0xbe, 0xc9, 0xa0, 0xd7, 0xdc, 0xcc, 0x3e,
	0x03, 0x00, 0x00, 0xff, 0xff, 0xaa, 0x76, 0xdf, 0xcd, 0xd5, 0x02, 0x00, 0x00,
}