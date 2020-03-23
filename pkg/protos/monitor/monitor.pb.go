// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/Jim3Things/CloudChamber/pkg/protos/monitor/monitor.proto

package monitor

import (
	context "context"
	fmt "fmt"
	common "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ActualHealth int32

const (
	Actual_Invalid     ActualHealth = 0
	Actual_Unavailable ActualHealth = 1
	Actual_Draining    ActualHealth = 2
	Actual_Healthy     ActualHealth = 3
	Actual_Removing    ActualHealth = 4
)

var ActualHealth_name = map[int32]string{
	0: "Invalid",
	1: "Unavailable",
	2: "Draining",
	3: "Healthy",
	4: "Removing",
}

var ActualHealth_value = map[string]int32{
	"Invalid":     0,
	"Unavailable": 1,
	"Draining":    2,
	"Healthy":     3,
	"Removing":    4,
}

func (x ActualHealth) String() string {
	return proto.EnumName(ActualHealth_name, int32(x))
}

func (ActualHealth) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0}
}

type DesiredHealth int32

const (
	Desired_Invalid  DesiredHealth = 0
	Desired_Draining DesiredHealth = 1
	Desired_Stopped  DesiredHealth = 2
	Desired_Healthy  DesiredHealth = 3
	Desired_Removed  DesiredHealth = 4
)

var DesiredHealth_name = map[int32]string{
	0: "Invalid",
	1: "Draining",
	2: "Stopped",
	3: "Healthy",
	4: "Removed",
}

var DesiredHealth_value = map[string]int32{
	"Invalid":  0,
	"Draining": 1,
	"Stopped":  2,
	"Healthy":  3,
	"Removed":  4,
}

func (x DesiredHealth) String() string {
	return proto.EnumName(DesiredHealth_name, int32(x))
}

func (DesiredHealth) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0}
}

// NOTE: There is an aspect of this structure that I'm unhappy with - I'd like
//       the message structure to be sure that invalid combinations cannot be
//       created.  In other words, it is structurally impossible to create a
//       message that has an invalid combination of items.
//
//       This does not do that.  For example, capacity does not make sense if
//       the health is not 'Healthy'.  But it can be specified.
//
//       I considered extensive use of oneof to limit the options, but that
//       looked even worse.  Open to suggestions.
type Actual struct {
	Racks                []*ActualRack `protobuf:"bytes,1,rep,name=racks,proto3" json:"racks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Actual) Reset()         { *m = Actual{} }
func (m *Actual) String() string { return proto.CompactTextString(m) }
func (*Actual) ProtoMessage()    {}
func (*Actual) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0}
}

func (m *Actual) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Actual.Unmarshal(m, b)
}
func (m *Actual) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Actual.Marshal(b, m, deterministic)
}
func (m *Actual) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Actual.Merge(m, src)
}
func (m *Actual) XXX_Size() int {
	return xxx_messageInfo_Actual.Size(m)
}
func (m *Actual) XXX_DiscardUnknown() {
	xxx_messageInfo_Actual.DiscardUnknown(m)
}

var xxx_messageInfo_Actual proto.InternalMessageInfo

func (m *Actual) GetRacks() []*ActualRack {
	if m != nil {
		return m.Racks
	}
	return nil
}

type ActualRack struct {
	Name                 string                            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Pdu                  *ActualRackPduDetails             `protobuf:"bytes,2,opt,name=pdu,proto3" json:"pdu,omitempty"`
	Tor                  *ActualRackTorDetails             `protobuf:"bytes,3,opt,name=tor,proto3" json:"tor,omitempty"`
	Blades               map[int64]*ActualRackBladeDetails `protobuf:"bytes,4,rep,name=blades,proto3" json:"blades,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *ActualRack) Reset()         { *m = ActualRack{} }
func (m *ActualRack) String() string { return proto.CompactTextString(m) }
func (*ActualRack) ProtoMessage()    {}
func (*ActualRack) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0}
}

func (m *ActualRack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActualRack.Unmarshal(m, b)
}
func (m *ActualRack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActualRack.Marshal(b, m, deterministic)
}
func (m *ActualRack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActualRack.Merge(m, src)
}
func (m *ActualRack) XXX_Size() int {
	return xxx_messageInfo_ActualRack.Size(m)
}
func (m *ActualRack) XXX_DiscardUnknown() {
	xxx_messageInfo_ActualRack.DiscardUnknown(m)
}

var xxx_messageInfo_ActualRack proto.InternalMessageInfo

func (m *ActualRack) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ActualRack) GetPdu() *ActualRackPduDetails {
	if m != nil {
		return m.Pdu
	}
	return nil
}

func (m *ActualRack) GetTor() *ActualRackTorDetails {
	if m != nil {
		return m.Tor
	}
	return nil
}

func (m *ActualRack) GetBlades() map[int64]*ActualRackBladeDetails {
	if m != nil {
		return m.Blades
	}
	return nil
}

type ActualRackBaseStatus struct {
	Health               ActualHealth      `protobuf:"varint,1,opt,name=health,proto3,enum=monitor.ActualHealth" json:"health,omitempty"`
	LastStart            *common.Timestamp `protobuf:"bytes,2,opt,name=last_start,json=lastStart,proto3" json:"last_start,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ActualRackBaseStatus) Reset()         { *m = ActualRackBaseStatus{} }
func (m *ActualRackBaseStatus) String() string { return proto.CompactTextString(m) }
func (*ActualRackBaseStatus) ProtoMessage()    {}
func (*ActualRackBaseStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0, 0}
}

func (m *ActualRackBaseStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActualRackBaseStatus.Unmarshal(m, b)
}
func (m *ActualRackBaseStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActualRackBaseStatus.Marshal(b, m, deterministic)
}
func (m *ActualRackBaseStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActualRackBaseStatus.Merge(m, src)
}
func (m *ActualRackBaseStatus) XXX_Size() int {
	return xxx_messageInfo_ActualRackBaseStatus.Size(m)
}
func (m *ActualRackBaseStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ActualRackBaseStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ActualRackBaseStatus proto.InternalMessageInfo

func (m *ActualRackBaseStatus) GetHealth() ActualHealth {
	if m != nil {
		return m.Health
	}
	return Actual_Invalid
}

func (m *ActualRackBaseStatus) GetLastStart() *common.Timestamp {
	if m != nil {
		return m.LastStart
	}
	return nil
}

type ActualRackBladeDetails struct {
	Status               *ActualRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Present              *common.BladeCapacity `protobuf:"bytes,2,opt,name=present,proto3" json:"present,omitempty"`
	Used                 *common.BladeCapacity `protobuf:"bytes,3,opt,name=used,proto3" json:"used,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ActualRackBladeDetails) Reset()         { *m = ActualRackBladeDetails{} }
func (m *ActualRackBladeDetails) String() string { return proto.CompactTextString(m) }
func (*ActualRackBladeDetails) ProtoMessage()    {}
func (*ActualRackBladeDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0, 1}
}

func (m *ActualRackBladeDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActualRackBladeDetails.Unmarshal(m, b)
}
func (m *ActualRackBladeDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActualRackBladeDetails.Marshal(b, m, deterministic)
}
func (m *ActualRackBladeDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActualRackBladeDetails.Merge(m, src)
}
func (m *ActualRackBladeDetails) XXX_Size() int {
	return xxx_messageInfo_ActualRackBladeDetails.Size(m)
}
func (m *ActualRackBladeDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ActualRackBladeDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ActualRackBladeDetails proto.InternalMessageInfo

func (m *ActualRackBladeDetails) GetStatus() *ActualRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ActualRackBladeDetails) GetPresent() *common.BladeCapacity {
	if m != nil {
		return m.Present
	}
	return nil
}

func (m *ActualRackBladeDetails) GetUsed() *common.BladeCapacity {
	if m != nil {
		return m.Used
	}
	return nil
}

type ActualRackPduDetails struct {
	Status               *ActualRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	TorCable             bool                  `protobuf:"varint,2,opt,name=tor_cable,json=torCable,proto3" json:"tor_cable,omitempty"`
	Cables               map[int64]bool        `protobuf:"bytes,3,rep,name=cables,proto3" json:"cables,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ActualRackPduDetails) Reset()         { *m = ActualRackPduDetails{} }
func (m *ActualRackPduDetails) String() string { return proto.CompactTextString(m) }
func (*ActualRackPduDetails) ProtoMessage()    {}
func (*ActualRackPduDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0, 2}
}

func (m *ActualRackPduDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActualRackPduDetails.Unmarshal(m, b)
}
func (m *ActualRackPduDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActualRackPduDetails.Marshal(b, m, deterministic)
}
func (m *ActualRackPduDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActualRackPduDetails.Merge(m, src)
}
func (m *ActualRackPduDetails) XXX_Size() int {
	return xxx_messageInfo_ActualRackPduDetails.Size(m)
}
func (m *ActualRackPduDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ActualRackPduDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ActualRackPduDetails proto.InternalMessageInfo

func (m *ActualRackPduDetails) GetStatus() *ActualRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ActualRackPduDetails) GetTorCable() bool {
	if m != nil {
		return m.TorCable
	}
	return false
}

func (m *ActualRackPduDetails) GetCables() map[int64]bool {
	if m != nil {
		return m.Cables
	}
	return nil
}

type ActualRackTorDetails struct {
	Status               *ActualRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Cables               map[int64]bool        `protobuf:"bytes,2,rep,name=cables,proto3" json:"cables,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ActualRackTorDetails) Reset()         { *m = ActualRackTorDetails{} }
func (m *ActualRackTorDetails) String() string { return proto.CompactTextString(m) }
func (*ActualRackTorDetails) ProtoMessage()    {}
func (*ActualRackTorDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{0, 0, 3}
}

func (m *ActualRackTorDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActualRackTorDetails.Unmarshal(m, b)
}
func (m *ActualRackTorDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActualRackTorDetails.Marshal(b, m, deterministic)
}
func (m *ActualRackTorDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActualRackTorDetails.Merge(m, src)
}
func (m *ActualRackTorDetails) XXX_Size() int {
	return xxx_messageInfo_ActualRackTorDetails.Size(m)
}
func (m *ActualRackTorDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ActualRackTorDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ActualRackTorDetails proto.InternalMessageInfo

func (m *ActualRackTorDetails) GetStatus() *ActualRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ActualRackTorDetails) GetCables() map[int64]bool {
	if m != nil {
		return m.Cables
	}
	return nil
}

// This message describes a command from the monitor to the inventory.  These
// take the form of desired states for specific items.  Any item not mentioned
// has no actions to take.
//
// NOTE: This message has an even more obvious issue with legal-but-invalid
//       structures: teh last start time is not valid for several of the
//       health states.
type Desired struct {
	Racks                []*DesiredRack `protobuf:"bytes,1,rep,name=racks,proto3" json:"racks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Desired) Reset()         { *m = Desired{} }
func (m *Desired) String() string { return proto.CompactTextString(m) }
func (*Desired) ProtoMessage()    {}
func (*Desired) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1}
}

func (m *Desired) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Desired.Unmarshal(m, b)
}
func (m *Desired) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Desired.Marshal(b, m, deterministic)
}
func (m *Desired) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Desired.Merge(m, src)
}
func (m *Desired) XXX_Size() int {
	return xxx_messageInfo_Desired.Size(m)
}
func (m *Desired) XXX_DiscardUnknown() {
	xxx_messageInfo_Desired.DiscardUnknown(m)
}

var xxx_messageInfo_Desired proto.InternalMessageInfo

func (m *Desired) GetRacks() []*DesiredRack {
	if m != nil {
		return m.Racks
	}
	return nil
}

type DesiredRack struct {
	Name                 string                             `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Pdu                  *DesiredRackPduDetails             `protobuf:"bytes,2,opt,name=pdu,proto3" json:"pdu,omitempty"`
	Tor                  *DesiredRackTorDetails             `protobuf:"bytes,3,opt,name=tor,proto3" json:"tor,omitempty"`
	Blades               map[int64]*DesiredRackBladeDetails `protobuf:"bytes,4,rep,name=blades,proto3" json:"blades,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                           `json:"-"`
	XXX_unrecognized     []byte                             `json:"-"`
	XXX_sizecache        int32                              `json:"-"`
}

func (m *DesiredRack) Reset()         { *m = DesiredRack{} }
func (m *DesiredRack) String() string { return proto.CompactTextString(m) }
func (*DesiredRack) ProtoMessage()    {}
func (*DesiredRack) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0}
}

func (m *DesiredRack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DesiredRack.Unmarshal(m, b)
}
func (m *DesiredRack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DesiredRack.Marshal(b, m, deterministic)
}
func (m *DesiredRack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DesiredRack.Merge(m, src)
}
func (m *DesiredRack) XXX_Size() int {
	return xxx_messageInfo_DesiredRack.Size(m)
}
func (m *DesiredRack) XXX_DiscardUnknown() {
	xxx_messageInfo_DesiredRack.DiscardUnknown(m)
}

var xxx_messageInfo_DesiredRack proto.InternalMessageInfo

func (m *DesiredRack) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *DesiredRack) GetPdu() *DesiredRackPduDetails {
	if m != nil {
		return m.Pdu
	}
	return nil
}

func (m *DesiredRack) GetTor() *DesiredRackTorDetails {
	if m != nil {
		return m.Tor
	}
	return nil
}

func (m *DesiredRack) GetBlades() map[int64]*DesiredRackBladeDetails {
	if m != nil {
		return m.Blades
	}
	return nil
}

type DesiredRackBaseStatus struct {
	Health               DesiredHealth     `protobuf:"varint,1,opt,name=health,proto3,enum=monitor.DesiredHealth" json:"health,omitempty"`
	LastStart            *common.Timestamp `protobuf:"bytes,2,opt,name=last_start,json=lastStart,proto3" json:"last_start,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *DesiredRackBaseStatus) Reset()         { *m = DesiredRackBaseStatus{} }
func (m *DesiredRackBaseStatus) String() string { return proto.CompactTextString(m) }
func (*DesiredRackBaseStatus) ProtoMessage()    {}
func (*DesiredRackBaseStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0, 0}
}

func (m *DesiredRackBaseStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DesiredRackBaseStatus.Unmarshal(m, b)
}
func (m *DesiredRackBaseStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DesiredRackBaseStatus.Marshal(b, m, deterministic)
}
func (m *DesiredRackBaseStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DesiredRackBaseStatus.Merge(m, src)
}
func (m *DesiredRackBaseStatus) XXX_Size() int {
	return xxx_messageInfo_DesiredRackBaseStatus.Size(m)
}
func (m *DesiredRackBaseStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_DesiredRackBaseStatus.DiscardUnknown(m)
}

var xxx_messageInfo_DesiredRackBaseStatus proto.InternalMessageInfo

func (m *DesiredRackBaseStatus) GetHealth() DesiredHealth {
	if m != nil {
		return m.Health
	}
	return Desired_Invalid
}

func (m *DesiredRackBaseStatus) GetLastStart() *common.Timestamp {
	if m != nil {
		return m.LastStart
	}
	return nil
}

type DesiredRackBladeDetails struct {
	Status               *DesiredRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *DesiredRackBladeDetails) Reset()         { *m = DesiredRackBladeDetails{} }
func (m *DesiredRackBladeDetails) String() string { return proto.CompactTextString(m) }
func (*DesiredRackBladeDetails) ProtoMessage()    {}
func (*DesiredRackBladeDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0, 1}
}

func (m *DesiredRackBladeDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DesiredRackBladeDetails.Unmarshal(m, b)
}
func (m *DesiredRackBladeDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DesiredRackBladeDetails.Marshal(b, m, deterministic)
}
func (m *DesiredRackBladeDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DesiredRackBladeDetails.Merge(m, src)
}
func (m *DesiredRackBladeDetails) XXX_Size() int {
	return xxx_messageInfo_DesiredRackBladeDetails.Size(m)
}
func (m *DesiredRackBladeDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_DesiredRackBladeDetails.DiscardUnknown(m)
}

var xxx_messageInfo_DesiredRackBladeDetails proto.InternalMessageInfo

func (m *DesiredRackBladeDetails) GetStatus() *DesiredRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

type DesiredRackPduDetails struct {
	Status               *DesiredRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	TorCable             bool                   `protobuf:"varint,2,opt,name=tor_cable,json=torCable,proto3" json:"tor_cable,omitempty"`
	Cables               map[int64]bool         `protobuf:"bytes,3,rep,name=cables,proto3" json:"cables,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *DesiredRackPduDetails) Reset()         { *m = DesiredRackPduDetails{} }
func (m *DesiredRackPduDetails) String() string { return proto.CompactTextString(m) }
func (*DesiredRackPduDetails) ProtoMessage()    {}
func (*DesiredRackPduDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0, 2}
}

func (m *DesiredRackPduDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DesiredRackPduDetails.Unmarshal(m, b)
}
func (m *DesiredRackPduDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DesiredRackPduDetails.Marshal(b, m, deterministic)
}
func (m *DesiredRackPduDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DesiredRackPduDetails.Merge(m, src)
}
func (m *DesiredRackPduDetails) XXX_Size() int {
	return xxx_messageInfo_DesiredRackPduDetails.Size(m)
}
func (m *DesiredRackPduDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_DesiredRackPduDetails.DiscardUnknown(m)
}

var xxx_messageInfo_DesiredRackPduDetails proto.InternalMessageInfo

func (m *DesiredRackPduDetails) GetStatus() *DesiredRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *DesiredRackPduDetails) GetTorCable() bool {
	if m != nil {
		return m.TorCable
	}
	return false
}

func (m *DesiredRackPduDetails) GetCables() map[int64]bool {
	if m != nil {
		return m.Cables
	}
	return nil
}

type DesiredRackTorDetails struct {
	Status               *DesiredRackBaseStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Cables               map[int64]bool         `protobuf:"bytes,2,rep,name=cables,proto3" json:"cables,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *DesiredRackTorDetails) Reset()         { *m = DesiredRackTorDetails{} }
func (m *DesiredRackTorDetails) String() string { return proto.CompactTextString(m) }
func (*DesiredRackTorDetails) ProtoMessage()    {}
func (*DesiredRackTorDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_318189739c1c4c20, []int{1, 0, 3}
}

func (m *DesiredRackTorDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DesiredRackTorDetails.Unmarshal(m, b)
}
func (m *DesiredRackTorDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DesiredRackTorDetails.Marshal(b, m, deterministic)
}
func (m *DesiredRackTorDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DesiredRackTorDetails.Merge(m, src)
}
func (m *DesiredRackTorDetails) XXX_Size() int {
	return xxx_messageInfo_DesiredRackTorDetails.Size(m)
}
func (m *DesiredRackTorDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_DesiredRackTorDetails.DiscardUnknown(m)
}

var xxx_messageInfo_DesiredRackTorDetails proto.InternalMessageInfo

func (m *DesiredRackTorDetails) GetStatus() *DesiredRackBaseStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *DesiredRackTorDetails) GetCables() map[int64]bool {
	if m != nil {
		return m.Cables
	}
	return nil
}

func init() {
	proto.RegisterEnum("monitor.ActualHealth", ActualHealth_name, ActualHealth_value)
	proto.RegisterEnum("monitor.DesiredHealth", DesiredHealth_name, DesiredHealth_value)
	proto.RegisterType((*Actual)(nil), "monitor.actual")
	proto.RegisterType((*ActualRack)(nil), "monitor.actual.rack")
	proto.RegisterMapType((map[int64]*ActualRackBladeDetails)(nil), "monitor.actual.rack.BladesEntry")
	proto.RegisterType((*ActualRackBaseStatus)(nil), "monitor.actual.rack.base_status")
	proto.RegisterType((*ActualRackBladeDetails)(nil), "monitor.actual.rack.blade_details")
	proto.RegisterType((*ActualRackPduDetails)(nil), "monitor.actual.rack.pdu_details")
	proto.RegisterMapType((map[int64]bool)(nil), "monitor.actual.rack.pdu_details.CablesEntry")
	proto.RegisterType((*ActualRackTorDetails)(nil), "monitor.actual.rack.tor_details")
	proto.RegisterMapType((map[int64]bool)(nil), "monitor.actual.rack.tor_details.CablesEntry")
	proto.RegisterType((*Desired)(nil), "monitor.desired")
	proto.RegisterType((*DesiredRack)(nil), "monitor.desired.rack")
	proto.RegisterMapType((map[int64]*DesiredRackBladeDetails)(nil), "monitor.desired.rack.BladesEntry")
	proto.RegisterType((*DesiredRackBaseStatus)(nil), "monitor.desired.rack.base_status")
	proto.RegisterType((*DesiredRackBladeDetails)(nil), "monitor.desired.rack.blade_details")
	proto.RegisterType((*DesiredRackPduDetails)(nil), "monitor.desired.rack.pdu_details")
	proto.RegisterMapType((map[int64]bool)(nil), "monitor.desired.rack.pdu_details.CablesEntry")
	proto.RegisterType((*DesiredRackTorDetails)(nil), "monitor.desired.rack.tor_details")
	proto.RegisterMapType((map[int64]bool)(nil), "monitor.desired.rack.tor_details.CablesEntry")
}

func init() {
	proto.RegisterFile("github.com/Jim3Things/CloudChamber/pkg/protos/monitor/monitor.proto", fileDescriptor_318189739c1c4c20)
}

var fileDescriptor_318189739c1c4c20 = []byte{
	// 852 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x96, 0xcf, 0x6e, 0xeb, 0x44,
	0x14, 0xc6, 0x71, 0x92, 0xe6, 0xcf, 0xf1, 0x85, 0x6b, 0x46, 0x97, 0x12, 0xf9, 0x6e, 0x42, 0xd8,
	0x54, 0x17, 0x62, 0x5f, 0x25, 0x42, 0xb7, 0x81, 0x15, 0xc9, 0xad, 0x44, 0x2b, 0x41, 0x25, 0xa7,
	0x6c, 0x90, 0xa0, 0x9a, 0xd8, 0x53, 0xc7, 0xaa, 0xed, 0xb1, 0xc6, 0xe3, 0x08, 0xb3, 0xe4, 0x0d,
	0x58, 0xc2, 0x0b, 0xf0, 0x12, 0x6c, 0x78, 0x05, 0xc4, 0x7b, 0x20, 0xb6, 0xac, 0xd0, 0x8c, 0x6d,
	0xea, 0x38, 0x71, 0x22, 0xda, 0xdb, 0x55, 0x66, 0x32, 0xdf, 0xf9, 0xce, 0x39, 0x9e, 0x99, 0x9f,
	0x06, 0xe6, 0xae, 0xc7, 0x57, 0xc9, 0xd2, 0xb0, 0x69, 0x60, 0x5e, 0x78, 0xc1, 0xe4, 0x6a, 0xe5,
	0x85, 0x6e, 0x6c, 0xce, 0x7d, 0x9a, 0x38, 0xf3, 0x15, 0x0e, 0x96, 0x84, 0x99, 0xd1, 0xad, 0x6b,
	0x46, 0x8c, 0x72, 0x1a, 0x9b, 0x01, 0x0d, 0x3d, 0x4e, 0x59, 0xf1, 0x6b, 0xc8, 0xbf, 0x51, 0x27,
	0x9f, 0xea, 0x9f, 0x97, 0xdc, 0x48, 0xb8, 0xa6, 0x69, 0xc4, 0xe8, 0xf7, 0x69, 0x16, 0x6c, 0x8f,
	0x5c, 0x12, 0x8e, 0xd6, 0xd8, 0xf7, 0x1c, 0xcc, 0x89, 0xb9, 0x35, 0xc8, 0xbc, 0xf4, 0xff, 0x59,
	0x90, 0x4d, 0x83, 0x80, 0x86, 0xa6, 0x8d, 0x23, 0x6c, 0x7b, 0x3c, 0xcd, 0x4d, 0x5e, 0xdf, 0xcb,
	0x84, 0x7b, 0x01, 0x89, 0x39, 0x0e, 0xa2, 0xdc, 0xe5, 0xb9, 0x4b, 0xa9, 0xeb, 0x93, 0x4c, 0xb5,
	0x4c, 0x6e, 0x4c, 0x12, 0x44, 0x45, 0x8a, 0xe1, 0xcf, 0x3d, 0x68, 0x63, 0x9b, 0x27, 0xd8, 0x47,
	0x2f, 0xe0, 0x88, 0x61, 0xfb, 0x36, 0xee, 0x2b, 0x83, 0xe6, 0x89, 0x3a, 0x7e, 0x66, 0x14, 0x5f,
	0x27, 0x5b, 0x37, 0xc4, 0xa2, 0x95, 0x49, 0xf4, 0x9f, 0xba, 0xd0, 0x12, 0x23, 0x84, 0xa0, 0x15,
	0xe2, 0x80, 0xf4, 0x95, 0x81, 0x72, 0xd2, 0xb3, 0xe4, 0x18, 0x8d, 0xa1, 0x19, 0x39, 0x49, 0xbf,
	0x31, 0x50, 0x4e, 0xd4, 0xf1, 0x60, 0x97, 0x8d, 0x11, 0x39, 0xc9, 0xb5, 0x43, 0x38, 0xf6, 0xfc,
	0xd8, 0x12, 0x62, 0x11, 0xc3, 0x29, 0xeb, 0x37, 0xf7, 0xc4, 0x70, 0xca, 0xee, 0x62, 0x38, 0x65,
	0x68, 0x06, 0xed, 0xa5, 0x8f, 0x1d, 0x12, 0xf7, 0x5b, 0xb2, 0xe2, 0xdd, 0x61, 0x33, 0x29, 0x39,
	0x0b, 0x39, 0x4b, 0x67, 0xdd, 0x7f, 0x66, 0x47, 0xbf, 0x28, 0x8d, 0xae, 0x62, 0xe5, 0x91, 0x7a,
	0x0a, 0xea, 0x12, 0xc7, 0xe4, 0x3a, 0xe6, 0x98, 0x27, 0x31, 0x3a, 0x85, 0xf6, 0x8a, 0x60, 0x9f,
	0xaf, 0x64, 0x43, 0xef, 0x8c, 0x8f, 0xab, 0x96, 0xd9, 0xaa, 0x34, 0xfa, 0x51, 0x69, 0x68, 0x8a,
	0x95, 0xeb, 0xd1, 0x4b, 0x00, 0x1f, 0xc7, 0x5c, 0x18, 0x31, 0x9e, 0xf7, 0xfe, 0xae, 0x91, 0x6d,
	0x89, 0x71, 0x55, 0x6c, 0x89, 0xd5, 0x13, 0xa2, 0x85, 0xd0, 0xe8, 0xbf, 0x2a, 0xf0, 0xb6, 0xac,
	0xa2, 0xe8, 0x4a, 0x64, 0xcf, 0xea, 0x90, 0xd9, 0xeb, 0x1a, 0x2a, 0xd5, 0x6b, 0xe5, 0x7a, 0xf4,
	0x12, 0x3a, 0x11, 0x23, 0x31, 0x09, 0x8b, 0xd4, 0xc7, 0x45, 0xea, 0x2c, 0x43, 0x71, 0xb0, 0xac,
	0x42, 0x86, 0x5e, 0x40, 0x2b, 0x89, 0x89, 0x93, 0x7f, 0xf1, 0x3a, 0xb9, 0xd4, 0xe8, 0x7f, 0x29,
	0xa0, 0x96, 0x76, 0xec, 0x01, 0x75, 0x3e, 0x87, 0x9e, 0xd8, 0x46, 0x1b, 0x2f, 0x7d, 0x22, 0x2b,
	0xed, 0x5a, 0x5d, 0x4e, 0xd9, 0x5c, 0xcc, 0xd1, 0x57, 0xd0, 0x96, 0x0b, 0x71, 0xbf, 0x29, 0xf7,
	0xf3, 0xe3, 0x43, 0x47, 0xc7, 0x90, 0x71, 0xdb, 0x7b, 0x9b, 0xb9, 0xe8, 0x53, 0x50, 0x4b, 0x02,
	0xa4, 0x41, 0xf3, 0x96, 0xa4, 0xb2, 0xe4, 0xa6, 0x25, 0x86, 0xe8, 0x19, 0x1c, 0xad, 0xb1, 0x9f,
	0x14, 0x95, 0x64, 0x93, 0x4f, 0x1b, 0xa7, 0x8a, 0xfe, 0x87, 0x02, 0x6a, 0xe9, 0xbc, 0x3d, 0xa0,
	0xe3, 0xbb, 0xa6, 0x1a, 0x7b, 0x9a, 0x2a, 0xe5, 0x7a, 0xbc, 0xa6, 0xbe, 0x05, 0xb5, 0x74, 0x19,
	0x76, 0x84, 0x9e, 0x96, 0x43, 0xd5, 0xf1, 0x70, 0x77, 0x93, 0xe5, 0x23, 0x5b, 0xb2, 0x1f, 0x5e,
	0x16, 0x77, 0x07, 0xa9, 0xd0, 0x39, 0x0f, 0x25, 0x10, 0xb5, 0xb7, 0xd0, 0x53, 0x50, 0xbf, 0x0e,
	0xf1, 0x1a, 0x7b, 0xbe, 0x28, 0x5b, 0x53, 0xd0, 0x13, 0xe8, 0xbe, 0x66, 0xd8, 0x0b, 0xbd, 0xd0,
	0xd5, 0x1a, 0x42, 0xfb, 0x85, 0x8c, 0x4a, 0xb5, 0xa6, 0x58, 0xb2, 0x48, 0x40, 0xd7, 0x62, 0xa9,
	0x35, 0xfc, 0xbd, 0x0b, 0x1d, 0x87, 0xc4, 0x1e, 0x23, 0x0e, 0xfa, 0x68, 0x13, 0x4e, 0xef, 0xfd,
	0x57, 0x5a, 0x2e, 0xd8, 0xa0, 0xd3, 0x6f, 0x9d, 0x3d, 0x74, 0x9a, 0x94, 0xe9, 0xf4, 0xc1, 0x4e,
	0x9f, 0x6d, 0x3c, 0x4d, 0xca, 0x78, 0xaa, 0x09, 0xda, 0xe2, 0xd3, 0xbc, 0xc2, 0xa7, 0x9a, 0xb8,
	0xfd, 0x80, 0xfa, 0x61, 0x13, 0x50, 0xd3, 0x0a, 0xa0, 0xde, 0xdf, 0xf2, 0x7c, 0x83, 0x84, 0xba,
	0xa8, 0x02, 0x6a, 0x5a, 0xb9, 0x06, 0x35, 0x1d, 0xed, 0xb8, 0x07, 0xfa, 0xdf, 0x15, 0x86, 0xdc,
	0xdf, 0x6a, 0x3f, 0x44, 0x2e, 0x2b, 0x10, 0x19, 0x1d, 0xdc, 0xe1, 0xc7, 0xbb, 0x70, 0x7f, 0x56,
	0x28, 0xf2, 0x80, 0x9e, 0x2f, 0x2b, 0x18, 0x19, 0x1d, 0x3c, 0x83, 0x8f, 0xd7, 0xd6, 0x77, 0x87,
	0x38, 0x32, 0xdd, 0xe4, 0xc8, 0x87, 0x35, 0x6d, 0xd6, 0x81, 0xe4, 0x62, 0x37, 0x48, 0xca, 0xdc,
	0x50, 0xc4, 0xd2, 0x82, 0xd3, 0x28, 0x22, 0x4e, 0x15, 0x22, 0x2a, 0x74, 0x24, 0x44, 0x88, 0xa3,
	0xb5, 0xc6, 0x33, 0xe8, 0x7c, 0x99, 0x25, 0x47, 0xaf, 0xe0, 0x89, 0x45, 0x22, 0xca, 0x78, 0x26,
	0x45, 0x4f, 0x2b, 0x78, 0xd3, 0x8f, 0x8d, 0xec, 0xa5, 0x64, 0x14, 0x2f, 0x25, 0xe3, 0x4c, 0xbc,
	0x94, 0xc6, 0x29, 0xf4, 0xce, 0xc3, 0x35, 0x09, 0x39, 0x65, 0x29, 0xfa, 0x0c, 0x7a, 0x0b, 0xc2,
	0xe7, 0x34, 0xbc, 0xf1, 0x5c, 0x54, 0x13, 0x51, 0xe7, 0x84, 0x4c, 0x80, 0x05, 0xe1, 0x57, 0x98,
	0xb9, 0x84, 0xc7, 0x48, 0xab, 0x7e, 0x17, 0xbd, 0x5a, 0xd2, 0xec, 0xd5, 0x37, 0x9f, 0xdc, 0xeb,
	0x65, 0xbb, 0x6c, 0xcb, 0xf9, 0xe4, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x38, 0x64, 0x9d, 0x82,
	0x19, 0x0b, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MonitorClient is the client API for Monitor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MonitorClient interface {
	// Report the health of a set of inventory items to the monitor
	ReportHealth(ctx context.Context, in *Actual, opts ...grpc.CallOption) (*empty.Empty, error)
}

type monitorClient struct {
	cc grpc.ClientConnInterface
}

func NewMonitorClient(cc grpc.ClientConnInterface) MonitorClient {
	return &monitorClient{cc}
}

func (c *monitorClient) ReportHealth(ctx context.Context, in *Actual, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/monitor.Monitor/ReportHealth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorServer is the server API for Monitor service.
type MonitorServer interface {
	// Report the health of a set of inventory items to the monitor
	ReportHealth(context.Context, *Actual) (*empty.Empty, error)
}

// UnimplementedMonitorServer can be embedded to have forward compatible implementations.
type UnimplementedMonitorServer struct {
}

func (*UnimplementedMonitorServer) ReportHealth(ctx context.Context, req *Actual) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportHealth not implemented")
}

func RegisterMonitorServer(s *grpc.Server, srv MonitorServer) {
	s.RegisterService(&_Monitor_serviceDesc, srv)
}

func _Monitor_ReportHealth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Actual)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).ReportHealth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/ReportHealth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).ReportHealth(ctx, req.(*Actual))
	}
	return interceptor(ctx, in, info, handler)
}

var _Monitor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monitor.Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReportHealth",
			Handler:    _Monitor_ReportHealth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/Jim3Things/CloudChamber/pkg/protos/monitor/monitor.proto",
}

// InventoryClient is the client API for Inventory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type InventoryClient interface {
	// Set the configuration options for reporting health
	SetConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	// Set the desired states for various inventory items.  Provide the
	// current actual state as informative return values.
	SetTargets(ctx context.Context, in *Desired, opts ...grpc.CallOption) (*Actual, error)
}

type inventoryClient struct {
	cc grpc.ClientConnInterface
}

func NewInventoryClient(cc grpc.ClientConnInterface) InventoryClient {
	return &inventoryClient{cc}
}

func (c *inventoryClient) SetConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/monitor.Inventory/SetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryClient) SetTargets(ctx context.Context, in *Desired, opts ...grpc.CallOption) (*Actual, error) {
	out := new(Actual)
	err := c.cc.Invoke(ctx, "/monitor.Inventory/SetTargets", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InventoryServer is the server API for Inventory service.
type InventoryServer interface {
	// Set the configuration options for reporting health
	SetConfig(context.Context, *empty.Empty) (*empty.Empty, error)
	// Set the desired states for various inventory items.  Provide the
	// current actual state as informative return values.
	SetTargets(context.Context, *Desired) (*Actual, error)
}

// UnimplementedInventoryServer can be embedded to have forward compatible implementations.
type UnimplementedInventoryServer struct {
}

func (*UnimplementedInventoryServer) SetConfig(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConfig not implemented")
}
func (*UnimplementedInventoryServer) SetTargets(ctx context.Context, req *Desired) (*Actual, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetTargets not implemented")
}

func RegisterInventoryServer(s *grpc.Server, srv InventoryServer) {
	s.RegisterService(&_Inventory_serviceDesc, srv)
}

func _Inventory_SetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServer).SetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Inventory/SetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServer).SetConfig(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Inventory_SetTargets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Desired)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServer).SetTargets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Inventory/SetTargets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServer).SetTargets(ctx, req.(*Desired))
	}
	return interceptor(ctx, in, info, handler)
}

var _Inventory_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monitor.Inventory",
	HandlerType: (*InventoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetConfig",
			Handler:    _Inventory_SetConfig_Handler,
		},
		{
			MethodName: "SetTargets",
			Handler:    _Inventory_SetTargets_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/Jim3Things/CloudChamber/pkg/protos/monitor/monitor.proto",
}
