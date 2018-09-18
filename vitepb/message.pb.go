// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vitepb/message.proto

package vitepb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type StatusMsg struct {
	NetID                uint64   `protobuf:"varint,1,opt,name=NetID,proto3" json:"NetID,omitempty"`
	Version              uint64   `protobuf:"varint,2,opt,name=Version,proto3" json:"Version,omitempty"`
	Height               uint64   `protobuf:"varint,3,opt,name=Height,proto3" json:"Height,omitempty"`
	CurrentBlock         []byte   `protobuf:"bytes,4,opt,name=CurrentBlock,proto3" json:"CurrentBlock,omitempty"`
	GenesisBlock         []byte   `protobuf:"bytes,5,opt,name=GenesisBlock,proto3" json:"GenesisBlock,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatusMsg) Reset()         { *m = StatusMsg{} }
func (m *StatusMsg) String() string { return proto.CompactTextString(m) }
func (*StatusMsg) ProtoMessage()    {}
func (*StatusMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{0}
}
func (m *StatusMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatusMsg.Unmarshal(m, b)
}
func (m *StatusMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatusMsg.Marshal(b, m, deterministic)
}
func (dst *StatusMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusMsg.Merge(dst, src)
}
func (m *StatusMsg) XXX_Size() int {
	return xxx_messageInfo_StatusMsg.Size(m)
}
func (m *StatusMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusMsg.DiscardUnknown(m)
}

var xxx_messageInfo_StatusMsg proto.InternalMessageInfo

func (m *StatusMsg) GetNetID() uint64 {
	if m != nil {
		return m.NetID
	}
	return 0
}

func (m *StatusMsg) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *StatusMsg) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *StatusMsg) GetCurrentBlock() []byte {
	if m != nil {
		return m.CurrentBlock
	}
	return nil
}

func (m *StatusMsg) GetGenesisBlock() []byte {
	if m != nil {
		return m.GenesisBlock
	}
	return nil
}

type GetSnapshotBlocksMsg struct {
	Origin               []byte   `protobuf:"bytes,1,opt,name=Origin,proto3" json:"Origin,omitempty"`
	Count                uint64   `protobuf:"varint,2,opt,name=Count,proto3" json:"Count,omitempty"`
	Forward              bool     `protobuf:"varint,3,opt,name=Forward,proto3" json:"Forward,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetSnapshotBlocksMsg) Reset()         { *m = GetSnapshotBlocksMsg{} }
func (m *GetSnapshotBlocksMsg) String() string { return proto.CompactTextString(m) }
func (*GetSnapshotBlocksMsg) ProtoMessage()    {}
func (*GetSnapshotBlocksMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{1}
}
func (m *GetSnapshotBlocksMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSnapshotBlocksMsg.Unmarshal(m, b)
}
func (m *GetSnapshotBlocksMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSnapshotBlocksMsg.Marshal(b, m, deterministic)
}
func (dst *GetSnapshotBlocksMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSnapshotBlocksMsg.Merge(dst, src)
}
func (m *GetSnapshotBlocksMsg) XXX_Size() int {
	return xxx_messageInfo_GetSnapshotBlocksMsg.Size(m)
}
func (m *GetSnapshotBlocksMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSnapshotBlocksMsg.DiscardUnknown(m)
}

var xxx_messageInfo_GetSnapshotBlocksMsg proto.InternalMessageInfo

func (m *GetSnapshotBlocksMsg) GetOrigin() []byte {
	if m != nil {
		return m.Origin
	}
	return nil
}

func (m *GetSnapshotBlocksMsg) GetCount() uint64 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *GetSnapshotBlocksMsg) GetForward() bool {
	if m != nil {
		return m.Forward
	}
	return false
}

type SnapshotBlocksMsg struct {
	Blocks               []*SnapshotBlockNet `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *SnapshotBlocksMsg) Reset()         { *m = SnapshotBlocksMsg{} }
func (m *SnapshotBlocksMsg) String() string { return proto.CompactTextString(m) }
func (*SnapshotBlocksMsg) ProtoMessage()    {}
func (*SnapshotBlocksMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{2}
}
func (m *SnapshotBlocksMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotBlocksMsg.Unmarshal(m, b)
}
func (m *SnapshotBlocksMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotBlocksMsg.Marshal(b, m, deterministic)
}
func (dst *SnapshotBlocksMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotBlocksMsg.Merge(dst, src)
}
func (m *SnapshotBlocksMsg) XXX_Size() int {
	return xxx_messageInfo_SnapshotBlocksMsg.Size(m)
}
func (m *SnapshotBlocksMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotBlocksMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotBlocksMsg proto.InternalMessageInfo

func (m *SnapshotBlocksMsg) GetBlocks() []*SnapshotBlockNet {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type GetAccountBlocksMsg struct {
	Origin               []byte   `protobuf:"bytes,1,opt,name=Origin,proto3" json:"Origin,omitempty"`
	Count                uint64   `protobuf:"varint,2,opt,name=Count,proto3" json:"Count,omitempty"`
	Forward              bool     `protobuf:"varint,3,opt,name=Forward,proto3" json:"Forward,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAccountBlocksMsg) Reset()         { *m = GetAccountBlocksMsg{} }
func (m *GetAccountBlocksMsg) String() string { return proto.CompactTextString(m) }
func (*GetAccountBlocksMsg) ProtoMessage()    {}
func (*GetAccountBlocksMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{3}
}
func (m *GetAccountBlocksMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountBlocksMsg.Unmarshal(m, b)
}
func (m *GetAccountBlocksMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountBlocksMsg.Marshal(b, m, deterministic)
}
func (dst *GetAccountBlocksMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountBlocksMsg.Merge(dst, src)
}
func (m *GetAccountBlocksMsg) XXX_Size() int {
	return xxx_messageInfo_GetAccountBlocksMsg.Size(m)
}
func (m *GetAccountBlocksMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountBlocksMsg.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountBlocksMsg proto.InternalMessageInfo

func (m *GetAccountBlocksMsg) GetOrigin() []byte {
	if m != nil {
		return m.Origin
	}
	return nil
}

func (m *GetAccountBlocksMsg) GetCount() uint64 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *GetAccountBlocksMsg) GetForward() bool {
	if m != nil {
		return m.Forward
	}
	return false
}

type AccountBlocksMsg struct {
	Blocks               []*AccountBlockNet `protobuf:"bytes,3,rep,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *AccountBlocksMsg) Reset()         { *m = AccountBlocksMsg{} }
func (m *AccountBlocksMsg) String() string { return proto.CompactTextString(m) }
func (*AccountBlocksMsg) ProtoMessage()    {}
func (*AccountBlocksMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{4}
}
func (m *AccountBlocksMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountBlocksMsg.Unmarshal(m, b)
}
func (m *AccountBlocksMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountBlocksMsg.Marshal(b, m, deterministic)
}
func (dst *AccountBlocksMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountBlocksMsg.Merge(dst, src)
}
func (m *AccountBlocksMsg) XXX_Size() int {
	return xxx_messageInfo_AccountBlocksMsg.Size(m)
}
func (m *AccountBlocksMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountBlocksMsg.DiscardUnknown(m)
}

var xxx_messageInfo_AccountBlocksMsg proto.InternalMessageInfo

func (m *AccountBlocksMsg) GetBlocks() []*AccountBlockNet {
	if m != nil {
		return m.Blocks
	}
	return nil
}

// version 2
type BlockID struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=Hash,proto3" json:"Hash,omitempty"`
	Height               uint64   `protobuf:"varint,2,opt,name=Height,proto3" json:"Height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockID) Reset()         { *m = BlockID{} }
func (m *BlockID) String() string { return proto.CompactTextString(m) }
func (*BlockID) ProtoMessage()    {}
func (*BlockID) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{5}
}
func (m *BlockID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockID.Unmarshal(m, b)
}
func (m *BlockID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockID.Marshal(b, m, deterministic)
}
func (dst *BlockID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockID.Merge(dst, src)
}
func (m *BlockID) XXX_Size() int {
	return xxx_messageInfo_BlockID.Size(m)
}
func (m *BlockID) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockID.DiscardUnknown(m)
}

var xxx_messageInfo_BlockID proto.InternalMessageInfo

func (m *BlockID) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *BlockID) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type Segment struct {
	From                 *BlockID `protobuf:"bytes,1,opt,name=From,proto3" json:"From,omitempty"`
	To                   *BlockID `protobuf:"bytes,2,opt,name=To,proto3" json:"To,omitempty"`
	Step                 uint64   `protobuf:"varint,3,opt,name=Step,proto3" json:"Step,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Segment) Reset()         { *m = Segment{} }
func (m *Segment) String() string { return proto.CompactTextString(m) }
func (*Segment) ProtoMessage()    {}
func (*Segment) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{6}
}
func (m *Segment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Segment.Unmarshal(m, b)
}
func (m *Segment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Segment.Marshal(b, m, deterministic)
}
func (dst *Segment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Segment.Merge(dst, src)
}
func (m *Segment) XXX_Size() int {
	return xxx_messageInfo_Segment.Size(m)
}
func (m *Segment) XXX_DiscardUnknown() {
	xxx_messageInfo_Segment.DiscardUnknown(m)
}

var xxx_messageInfo_Segment proto.InternalMessageInfo

func (m *Segment) GetFrom() *BlockID {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Segment) GetTo() *BlockID {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *Segment) GetStep() uint64 {
	if m != nil {
		return m.Step
	}
	return 0
}

type AccountSegment struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	Segment              *Segment `protobuf:"bytes,2,opt,name=Segment,proto3" json:"Segment,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountSegment) Reset()         { *m = AccountSegment{} }
func (m *AccountSegment) String() string { return proto.CompactTextString(m) }
func (*AccountSegment) ProtoMessage()    {}
func (*AccountSegment) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{7}
}
func (m *AccountSegment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountSegment.Unmarshal(m, b)
}
func (m *AccountSegment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountSegment.Marshal(b, m, deterministic)
}
func (dst *AccountSegment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountSegment.Merge(dst, src)
}
func (m *AccountSegment) XXX_Size() int {
	return xxx_messageInfo_AccountSegment.Size(m)
}
func (m *AccountSegment) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountSegment.DiscardUnknown(m)
}

var xxx_messageInfo_AccountSegment proto.InternalMessageInfo

func (m *AccountSegment) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *AccountSegment) GetSegment() *Segment {
	if m != nil {
		return m.Segment
	}
	return nil
}

type MultiAccountSegment struct {
	Segments             []*AccountSegment `protobuf:"bytes,1,rep,name=Segments,proto3" json:"Segments,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *MultiAccountSegment) Reset()         { *m = MultiAccountSegment{} }
func (m *MultiAccountSegment) String() string { return proto.CompactTextString(m) }
func (*MultiAccountSegment) ProtoMessage()    {}
func (*MultiAccountSegment) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{8}
}
func (m *MultiAccountSegment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MultiAccountSegment.Unmarshal(m, b)
}
func (m *MultiAccountSegment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MultiAccountSegment.Marshal(b, m, deterministic)
}
func (dst *MultiAccountSegment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MultiAccountSegment.Merge(dst, src)
}
func (m *MultiAccountSegment) XXX_Size() int {
	return xxx_messageInfo_MultiAccountSegment.Size(m)
}
func (m *MultiAccountSegment) XXX_DiscardUnknown() {
	xxx_messageInfo_MultiAccountSegment.DiscardUnknown(m)
}

var xxx_messageInfo_MultiAccountSegment proto.InternalMessageInfo

func (m *MultiAccountSegment) GetSegments() []*AccountSegment {
	if m != nil {
		return m.Segments
	}
	return nil
}

type FileList struct {
	Files                []string `protobuf:"bytes,1,rep,name=Files,proto3" json:"Files,omitempty"`
	Start                uint64   `protobuf:"varint,2,opt,name=Start,proto3" json:"Start,omitempty"`
	End                  uint64   `protobuf:"varint,3,opt,name=End,proto3" json:"End,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileList) Reset()         { *m = FileList{} }
func (m *FileList) String() string { return proto.CompactTextString(m) }
func (*FileList) ProtoMessage()    {}
func (*FileList) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{9}
}
func (m *FileList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileList.Unmarshal(m, b)
}
func (m *FileList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileList.Marshal(b, m, deterministic)
}
func (dst *FileList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileList.Merge(dst, src)
}
func (m *FileList) XXX_Size() int {
	return xxx_messageInfo_FileList.Size(m)
}
func (m *FileList) XXX_DiscardUnknown() {
	xxx_messageInfo_FileList.DiscardUnknown(m)
}

var xxx_messageInfo_FileList proto.InternalMessageInfo

func (m *FileList) GetFiles() []string {
	if m != nil {
		return m.Files
	}
	return nil
}

func (m *FileList) GetStart() uint64 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *FileList) GetEnd() uint64 {
	if m != nil {
		return m.End
	}
	return 0
}

type SubLedger struct {
	SBlocks              []*SnapshotBlockNet `protobuf:"bytes,1,rep,name=SBlocks,proto3" json:"SBlocks,omitempty"`
	ABlocks              []*AccountBlockNet  `protobuf:"bytes,2,rep,name=ABlocks,proto3" json:"ABlocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *SubLedger) Reset()         { *m = SubLedger{} }
func (m *SubLedger) String() string { return proto.CompactTextString(m) }
func (*SubLedger) ProtoMessage()    {}
func (*SubLedger) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_8c333c64382bd7fd, []int{10}
}
func (m *SubLedger) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubLedger.Unmarshal(m, b)
}
func (m *SubLedger) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubLedger.Marshal(b, m, deterministic)
}
func (dst *SubLedger) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubLedger.Merge(dst, src)
}
func (m *SubLedger) XXX_Size() int {
	return xxx_messageInfo_SubLedger.Size(m)
}
func (m *SubLedger) XXX_DiscardUnknown() {
	xxx_messageInfo_SubLedger.DiscardUnknown(m)
}

var xxx_messageInfo_SubLedger proto.InternalMessageInfo

func (m *SubLedger) GetSBlocks() []*SnapshotBlockNet {
	if m != nil {
		return m.SBlocks
	}
	return nil
}

func (m *SubLedger) GetABlocks() []*AccountBlockNet {
	if m != nil {
		return m.ABlocks
	}
	return nil
}

func init() {
	proto.RegisterType((*StatusMsg)(nil), "vitepb.StatusMsg")
	proto.RegisterType((*GetSnapshotBlocksMsg)(nil), "vitepb.GetSnapshotBlocksMsg")
	proto.RegisterType((*SnapshotBlocksMsg)(nil), "vitepb.SnapshotBlocksMsg")
	proto.RegisterType((*GetAccountBlocksMsg)(nil), "vitepb.GetAccountBlocksMsg")
	proto.RegisterType((*AccountBlocksMsg)(nil), "vitepb.AccountBlocksMsg")
	proto.RegisterType((*BlockID)(nil), "vitepb.BlockID")
	proto.RegisterType((*Segment)(nil), "vitepb.Segment")
	proto.RegisterType((*AccountSegment)(nil), "vitepb.AccountSegment")
	proto.RegisterType((*MultiAccountSegment)(nil), "vitepb.MultiAccountSegment")
	proto.RegisterType((*FileList)(nil), "vitepb.FileList")
	proto.RegisterType((*SubLedger)(nil), "vitepb.SubLedger")
}

func init() { proto.RegisterFile("vitepb/message.proto", fileDescriptor_message_8c333c64382bd7fd) }

var fileDescriptor_message_8c333c64382bd7fd = []byte{
	// 489 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0x4d, 0x6f, 0xda, 0x4c,
	0x10, 0xc7, 0x65, 0x43, 0x6c, 0x98, 0xa0, 0xe7, 0x49, 0x37, 0x88, 0x5a, 0x39, 0x34, 0x68, 0x7b,
	0xa1, 0x17, 0xd2, 0x52, 0xf5, 0x03, 0x10, 0x12, 0x5e, 0xa4, 0x24, 0x95, 0xec, 0xb4, 0xb7, 0x36,
	0x32, 0x78, 0x64, 0xac, 0x82, 0x17, 0xed, 0xae, 0xdb, 0x0f, 0xd3, 0x2f, 0x5b, 0xed, 0x1b, 0x72,
	0xa0, 0xaa, 0x7a, 0xe8, 0x6d, 0xfe, 0x33, 0xff, 0x9d, 0xf9, 0x31, 0x63, 0xa0, 0xfb, 0xbd, 0x90,
	0xb8, 0x5b, 0x5e, 0x6d, 0x51, 0x88, 0x34, 0xc7, 0xe1, 0x8e, 0x33, 0xc9, 0x48, 0x60, 0xb2, 0x17,
	0xaf, 0x6c, 0x35, 0x5d, 0xad, 0x58, 0x55, 0xca, 0xa7, 0xe5, 0x86, 0xad, 0xbe, 0x3d, 0x95, 0x28,
	0x8d, 0xef, 0xe2, 0xd2, 0xd6, 0x45, 0x99, 0xee, 0xc4, 0x9a, 0x1d, 0x19, 0xe8, 0x4f, 0x0f, 0xda,
	0x89, 0x4c, 0x65, 0x25, 0xee, 0x45, 0x4e, 0xba, 0x70, 0xf2, 0x80, 0x72, 0x71, 0x13, 0x79, 0x7d,
	0x6f, 0xd0, 0x8c, 0x8d, 0x20, 0x11, 0x84, 0x9f, 0x91, 0x8b, 0x82, 0x95, 0x91, 0xaf, 0xf3, 0x4e,
	0x92, 0x1e, 0x04, 0x73, 0x2c, 0xf2, 0xb5, 0x8c, 0x1a, 0xba, 0x60, 0x15, 0xa1, 0xd0, 0x99, 0x54,
	0x9c, 0x63, 0x29, 0xaf, 0xd5, 0xbc, 0xa8, 0xd9, 0xf7, 0x06, 0x9d, 0xf8, 0x59, 0x4e, 0x79, 0x66,
	0x58, 0xa2, 0x28, 0x84, 0xf1, 0x9c, 0x18, 0x4f, 0x3d, 0x47, 0xbf, 0x42, 0x77, 0x86, 0x32, 0xb1,
	0xf0, 0x3a, 0xa7, 0x39, 0x7b, 0x10, 0x7c, 0xe4, 0x45, 0x5e, 0x94, 0x1a, 0xb4, 0x13, 0x5b, 0xa5,
	0xf8, 0x27, 0x6a, 0x0f, 0x96, 0xd3, 0x08, 0xc5, 0x3f, 0x65, 0xfc, 0x47, 0xca, 0x33, 0x8d, 0xd9,
	0x8a, 0x9d, 0xa4, 0xb7, 0xf0, 0xe2, 0xb8, 0xf9, 0x5b, 0x08, 0xf4, 0x96, 0x44, 0xe4, 0xf5, 0x1b,
	0x83, 0xd3, 0x51, 0x34, 0x34, 0x4b, 0x1c, 0x3e, 0xb3, 0x3e, 0xa0, 0x8c, 0xad, 0x8f, 0x7e, 0x81,
	0xf3, 0x19, 0xca, 0xb1, 0xb9, 0xc1, 0xbf, 0xa7, 0x9c, 0xc0, 0xd9, 0x51, 0xef, 0xab, 0x3d, 0x64,
	0x43, 0x43, 0xbe, 0x74, 0x90, 0x75, 0x67, 0x9d, 0xf1, 0x03, 0x84, 0x3a, 0xb7, 0xb8, 0x21, 0x04,
	0x9a, 0xf3, 0x54, 0xac, 0x2d, 0x95, 0x8e, 0x6b, 0x97, 0xf4, 0xeb, 0x97, 0xa4, 0x2b, 0x08, 0x13,
	0xcc, 0xb7, 0x58, 0x4a, 0xf2, 0x1a, 0x9a, 0x53, 0xce, 0xb6, 0xfa, 0xd9, 0xe9, 0xe8, 0x7f, 0x37,
	0xd0, 0x76, 0x8d, 0x75, 0x91, 0x5c, 0x82, 0xff, 0xc8, 0x74, 0x8f, 0xdf, 0x58, 0xfc, 0x47, 0xa6,
	0x86, 0x27, 0x12, 0x77, 0xf6, 0x83, 0xd1, 0x31, 0xfd, 0x04, 0xff, 0x59, 0x6c, 0x37, 0x2b, 0x82,
	0x70, 0x9c, 0x65, 0x1c, 0x85, 0xb0, 0x94, 0x4e, 0x92, 0x37, 0x7b, 0xa0, 0xc3, 0x29, 0x36, 0x1d,
	0xbb, 0x3a, 0x5d, 0xc0, 0xf9, 0x7d, 0xb5, 0x91, 0xc5, 0x41, 0xef, 0x11, 0xb4, 0x6c, 0xe8, 0x2e,
	0xdc, 0x3b, 0x58, 0x9e, 0xeb, 0xb4, 0xf7, 0xd1, 0x39, 0xb4, 0xa6, 0xc5, 0x06, 0xef, 0x0a, 0x21,
	0xd5, 0xf9, 0x54, 0x6c, 0x1e, 0xb7, 0x63, 0x23, 0x54, 0x36, 0x91, 0x29, 0xdf, 0x1f, 0x55, 0x0b,
	0x72, 0x06, 0x8d, 0xdb, 0x32, 0xb3, 0x3f, 0x56, 0x85, 0x94, 0x43, 0x3b, 0xa9, 0x96, 0x77, 0x98,
	0xe5, 0xc8, 0xc9, 0x08, 0xc2, 0xe4, 0xfa, 0xef, 0xbe, 0x35, 0x67, 0x24, 0xef, 0x20, 0x1c, 0xdb,
	0x37, 0xfe, 0x9f, 0x4f, 0xef, 0x7c, 0xcb, 0x40, 0xff, 0xd7, 0xdf, 0xff, 0x0a, 0x00, 0x00, 0xff,
	0xff, 0x5a, 0x9c, 0x8f, 0x57, 0x4c, 0x04, 0x00, 0x00,
}
