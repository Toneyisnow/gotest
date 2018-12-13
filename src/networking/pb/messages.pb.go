// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messages.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type NetMessageType int32

const (
	NetMessageType_SendEvents NetMessageType = 0
	NetMessageType_AddNode    NetMessageType = 1
	NetMessageType_SyncEvents NetMessageType = 2
)

var NetMessageType_name = map[int32]string{
	0: "SendEvents",
	1: "AddNode",
	2: "SyncEvents",
}

var NetMessageType_value = map[string]int32{
	"SendEvents": 0,
	"AddNode":    1,
	"SyncEvents": 2,
}

func (x NetMessageType) String() string {
	return proto.EnumName(NetMessageType_name, int32(x))
}

func (NetMessageType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{0}
}

type NetMessage struct {
	OwnerId     string               `protobuf:"bytes,1,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`
	Hash        string               `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	TimeStamp   *timestamp.Timestamp `protobuf:"bytes,3,opt,name=time_stamp,json=timeStamp,proto3" json:"time_stamp,omitempty"`
	MessageType NetMessageType       `protobuf:"varint,4,opt,name=messageType,proto3,enum=pb.NetMessageType" json:"messageType,omitempty"`
	// Here are the different types of messages
	SendEventMessage     *SendEventMessage `protobuf:"bytes,6,opt,name=send_event_message,json=sendEventMessage,proto3" json:"send_event_message,omitempty"`
	CommandMessage       *CommandMessage   `protobuf:"bytes,7,opt,name=command_message,json=commandMessage,proto3" json:"command_message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *NetMessage) Reset()         { *m = NetMessage{} }
func (m *NetMessage) String() string { return proto.CompactTextString(m) }
func (*NetMessage) ProtoMessage()    {}
func (*NetMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{0}
}

func (m *NetMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NetMessage.Unmarshal(m, b)
}
func (m *NetMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NetMessage.Marshal(b, m, deterministic)
}
func (m *NetMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetMessage.Merge(m, src)
}
func (m *NetMessage) XXX_Size() int {
	return xxx_messageInfo_NetMessage.Size(m)
}
func (m *NetMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NetMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NetMessage proto.InternalMessageInfo

func (m *NetMessage) GetOwnerId() string {
	if m != nil {
		return m.OwnerId
	}
	return ""
}

func (m *NetMessage) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *NetMessage) GetTimeStamp() *timestamp.Timestamp {
	if m != nil {
		return m.TimeStamp
	}
	return nil
}

func (m *NetMessage) GetMessageType() NetMessageType {
	if m != nil {
		return m.MessageType
	}
	return NetMessageType_SendEvents
}

func (m *NetMessage) GetSendEventMessage() *SendEventMessage {
	if m != nil {
		return m.SendEventMessage
	}
	return nil
}

func (m *NetMessage) GetCommandMessage() *CommandMessage {
	if m != nil {
		return m.CommandMessage
	}
	return nil
}

type NetEvent struct {
	EventId              string               `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	OwnerId              string               `protobuf:"bytes,2,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`
	Hash                 string               `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	TimeStamp            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=time_stamp,json=timeStamp,proto3" json:"time_stamp,omitempty"`
	Transactions         []*NetTransaction    `protobuf:"bytes,5,rep,name=transactions,proto3" json:"transactions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *NetEvent) Reset()         { *m = NetEvent{} }
func (m *NetEvent) String() string { return proto.CompactTextString(m) }
func (*NetEvent) ProtoMessage()    {}
func (*NetEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{1}
}

func (m *NetEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NetEvent.Unmarshal(m, b)
}
func (m *NetEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NetEvent.Marshal(b, m, deterministic)
}
func (m *NetEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetEvent.Merge(m, src)
}
func (m *NetEvent) XXX_Size() int {
	return xxx_messageInfo_NetEvent.Size(m)
}
func (m *NetEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_NetEvent.DiscardUnknown(m)
}

var xxx_messageInfo_NetEvent proto.InternalMessageInfo

func (m *NetEvent) GetEventId() string {
	if m != nil {
		return m.EventId
	}
	return ""
}

func (m *NetEvent) GetOwnerId() string {
	if m != nil {
		return m.OwnerId
	}
	return ""
}

func (m *NetEvent) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *NetEvent) GetTimeStamp() *timestamp.Timestamp {
	if m != nil {
		return m.TimeStamp
	}
	return nil
}

func (m *NetEvent) GetTransactions() []*NetTransaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

type NetTransaction struct {
	OwnerId              string   `protobuf:"bytes,1,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`
	Hash                 string   `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	InAccount            string   `protobuf:"bytes,3,opt,name=in_account,json=inAccount,proto3" json:"in_account,omitempty"`
	OutAccount           string   `protobuf:"bytes,4,opt,name=out_account,json=outAccount,proto3" json:"out_account,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NetTransaction) Reset()         { *m = NetTransaction{} }
func (m *NetTransaction) String() string { return proto.CompactTextString(m) }
func (*NetTransaction) ProtoMessage()    {}
func (*NetTransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{2}
}

func (m *NetTransaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NetTransaction.Unmarshal(m, b)
}
func (m *NetTransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NetTransaction.Marshal(b, m, deterministic)
}
func (m *NetTransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetTransaction.Merge(m, src)
}
func (m *NetTransaction) XXX_Size() int {
	return xxx_messageInfo_NetTransaction.Size(m)
}
func (m *NetTransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_NetTransaction.DiscardUnknown(m)
}

var xxx_messageInfo_NetTransaction proto.InternalMessageInfo

func (m *NetTransaction) GetOwnerId() string {
	if m != nil {
		return m.OwnerId
	}
	return ""
}

func (m *NetTransaction) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *NetTransaction) GetInAccount() string {
	if m != nil {
		return m.InAccount
	}
	return ""
}

func (m *NetTransaction) GetOutAccount() string {
	if m != nil {
		return m.OutAccount
	}
	return ""
}

type SendEventMessage struct {
	Events               []*NetEvent `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SendEventMessage) Reset()         { *m = SendEventMessage{} }
func (m *SendEventMessage) String() string { return proto.CompactTextString(m) }
func (*SendEventMessage) ProtoMessage()    {}
func (*SendEventMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{3}
}

func (m *SendEventMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendEventMessage.Unmarshal(m, b)
}
func (m *SendEventMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendEventMessage.Marshal(b, m, deterministic)
}
func (m *SendEventMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendEventMessage.Merge(m, src)
}
func (m *SendEventMessage) XXX_Size() int {
	return xxx_messageInfo_SendEventMessage.Size(m)
}
func (m *SendEventMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_SendEventMessage.DiscardUnknown(m)
}

var xxx_messageInfo_SendEventMessage proto.InternalMessageInfo

func (m *SendEventMessage) GetEvents() []*NetEvent {
	if m != nil {
		return m.Events
	}
	return nil
}

type CommandMessage struct {
	CommandId            string   `protobuf:"bytes,1,opt,name=command_id,json=commandId,proto3" json:"command_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommandMessage) Reset()         { *m = CommandMessage{} }
func (m *CommandMessage) String() string { return proto.CompactTextString(m) }
func (*CommandMessage) ProtoMessage()    {}
func (*CommandMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{4}
}

func (m *CommandMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommandMessage.Unmarshal(m, b)
}
func (m *CommandMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommandMessage.Marshal(b, m, deterministic)
}
func (m *CommandMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommandMessage.Merge(m, src)
}
func (m *CommandMessage) XXX_Size() int {
	return xxx_messageInfo_CommandMessage.Size(m)
}
func (m *CommandMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_CommandMessage.DiscardUnknown(m)
}

var xxx_messageInfo_CommandMessage proto.InternalMessageInfo

func (m *CommandMessage) GetCommandId() string {
	if m != nil {
		return m.CommandId
	}
	return ""
}

func init() {
	proto.RegisterEnum("pb.NetMessageType", NetMessageType_name, NetMessageType_value)
	proto.RegisterType((*NetMessage)(nil), "pb.NetMessage")
	proto.RegisterType((*NetEvent)(nil), "pb.NetEvent")
	proto.RegisterType((*NetTransaction)(nil), "pb.NetTransaction")
	proto.RegisterType((*SendEventMessage)(nil), "pb.SendEventMessage")
	proto.RegisterType((*CommandMessage)(nil), "pb.CommandMessage")
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor_4dc296cbfe5ffcd5) }

var fileDescriptor_4dc296cbfe5ffcd5 = []byte{
	// 419 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x41, 0x8f, 0x93, 0x40,
	0x14, 0x16, 0x8a, 0xed, 0xf2, 0xd8, 0x20, 0x99, 0x78, 0xc0, 0x4d, 0xcc, 0x12, 0xe2, 0xa1, 0xf1,
	0x40, 0x93, 0xd5, 0x18, 0x8d, 0xf1, 0xb0, 0x1a, 0x0f, 0x3d, 0xc8, 0x81, 0xed, 0x9d, 0x00, 0x33,
	0xb6, 0x24, 0x32, 0x43, 0x3a, 0x83, 0xa6, 0x27, 0x7f, 0x91, 0xbf, 0xc3, 0xbf, 0x65, 0xe6, 0x31,
	0xd0, 0xd2, 0xb8, 0x87, 0xde, 0x98, 0xf7, 0x7d, 0xef, 0xfb, 0xde, 0xfb, 0x1e, 0xe0, 0x37, 0x4c,
	0xca, 0x62, 0xcb, 0x64, 0xd2, 0xee, 0x85, 0x12, 0xc4, 0x6e, 0xcb, 0x9b, 0xdb, 0xad, 0x10, 0xdb,
	0x1f, 0x6c, 0x85, 0x95, 0xb2, 0xfb, 0xbe, 0x52, 0x75, 0xc3, 0xa4, 0x2a, 0x9a, 0xb6, 0x27, 0xc5,
	0x7f, 0x6c, 0x80, 0x94, 0xa9, 0x6f, 0x7d, 0x2b, 0x79, 0x01, 0x57, 0xe2, 0x17, 0x67, 0xfb, 0xbc,
	0xa6, 0xa1, 0x15, 0x59, 0x4b, 0x37, 0x5b, 0xe0, 0x7b, 0x4d, 0x09, 0x01, 0x67, 0x57, 0xc8, 0x5d,
	0x68, 0x63, 0x19, 0xbf, 0xc9, 0x07, 0x00, 0x2d, 0x98, 0xa3, 0x62, 0x38, 0x8b, 0xac, 0xa5, 0x77,
	0x77, 0x93, 0xf4, 0x9e, 0xc9, 0xe0, 0x99, 0x6c, 0x06, 0xcf, 0xcc, 0xd5, 0xec, 0x07, 0xfd, 0x49,
	0xde, 0x82, 0x67, 0xe6, 0xdd, 0x1c, 0x5a, 0x16, 0x3a, 0x91, 0xb5, 0xf4, 0xef, 0x48, 0xd2, 0x96,
	0xc9, 0x71, 0x1c, 0x8d, 0x64, 0xa7, 0x34, 0xf2, 0x19, 0x88, 0x64, 0x9c, 0xe6, 0xec, 0x27, 0xe3,
	0x2a, 0x37, 0x48, 0x38, 0x47, 0xe3, 0xe7, 0xba, 0xf9, 0x81, 0x71, 0xfa, 0x55, 0x83, 0x46, 0x22,
	0x0b, 0xe4, 0x59, 0x85, 0x7c, 0x84, 0x67, 0x95, 0x68, 0x9a, 0x82, 0xd3, 0x51, 0x60, 0x81, 0x02,
	0xe8, 0xfe, 0xa5, 0x87, 0x86, 0x76, 0xbf, 0x9a, 0xbc, 0xe3, 0xbf, 0x16, 0x5c, 0xa5, 0x4c, 0xa1,
	0xa0, 0x4e, 0xab, 0x1f, 0xe4, 0x98, 0x16, 0xbe, 0xd7, 0x74, 0x12, 0xa4, 0xfd, 0xff, 0x20, 0x67,
	0x8f, 0x06, 0xe9, 0x5c, 0x12, 0xe4, 0x3b, 0xb8, 0x56, 0xfb, 0x82, 0xcb, 0xa2, 0x52, 0xb5, 0xe0,
	0x32, 0x7c, 0x1a, 0xcd, 0x86, 0x5d, 0x52, 0xa6, 0x36, 0x47, 0x28, 0x9b, 0xf0, 0xe2, 0xdf, 0xe0,
	0x4f, 0xf1, 0x4b, 0x8f, 0xff, 0x12, 0xa0, 0xe6, 0x79, 0x51, 0x55, 0xa2, 0xe3, 0xca, 0x6c, 0xe3,
	0xd6, 0xfc, 0xbe, 0x2f, 0x90, 0x5b, 0xf0, 0x44, 0xa7, 0x46, 0xdc, 0x41, 0x1c, 0x44, 0xa7, 0x0c,
	0x21, 0x7e, 0x0f, 0xc1, 0xf9, 0xb5, 0xc8, 0x2b, 0x98, 0x63, 0x82, 0x32, 0xb4, 0x70, 0x8d, 0x6b,
	0xb3, 0x06, 0x92, 0x32, 0x83, 0xc5, 0x2b, 0xf0, 0xa7, 0x67, 0xd2, 0xb3, 0x0c, 0x37, 0x1d, 0x87,
	0x77, 0x4d, 0x65, 0x4d, 0x5f, 0x7f, 0xc2, 0x5d, 0x4f, 0xfe, 0x2a, 0xe2, 0x03, 0x8c, 0xe6, 0x32,
	0x78, 0x42, 0x3c, 0x58, 0xdc, 0x53, 0x9a, 0x0a, 0xca, 0x02, 0x0b, 0xc1, 0x03, 0xaf, 0x0c, 0x68,
	0x97, 0x73, 0xbc, 0xc0, 0x9b, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x85, 0x16, 0x02, 0xab, 0x62,
	0x03, 0x00, 0x00,
}
