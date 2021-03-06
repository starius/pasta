// Code generated by protoc-gen-go.
// source: record.proto
// DO NOT EDIT!

/*
Package database is a generated protocol buffer package.

It is generated from these files:
	record.proto

It has these top-level messages:
	Record
*/
package database

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

type Record struct {
	Content     []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Filename    string `protobuf:"bytes,2,opt,name=filename" json:"filename,omitempty"`
	SelfBurning bool   `protobuf:"varint,3,opt,name=self_burning,json=selfBurning" json:"self_burning,omitempty"`
	Redirect    bool   `protobuf:"varint,4,opt,name=redirect" json:"redirect,omitempty"`
	LongId      bool   `protobuf:"varint,5,opt,name=long_id,json=longId" json:"long_id,omitempty"`
	ContentType string `protobuf:"bytes,6,opt,name=content_type,json=contentType" json:"content_type,omitempty"`
}

func (m *Record) Reset()                    { *m = Record{} }
func (m *Record) String() string            { return proto.CompactTextString(m) }
func (*Record) ProtoMessage()               {}
func (*Record) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Record) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *Record) GetFilename() string {
	if m != nil {
		return m.Filename
	}
	return ""
}

func (m *Record) GetSelfBurning() bool {
	if m != nil {
		return m.SelfBurning
	}
	return false
}

func (m *Record) GetRedirect() bool {
	if m != nil {
		return m.Redirect
	}
	return false
}

func (m *Record) GetLongId() bool {
	if m != nil {
		return m.LongId
	}
	return false
}

func (m *Record) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func init() {
	proto.RegisterType((*Record)(nil), "database.Record")
}

func init() { proto.RegisterFile("record.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 180 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8f, 0xbd, 0x0e, 0xc2, 0x20,
	0x14, 0x46, 0x83, 0x3f, 0xb4, 0xd2, 0x4e, 0x2c, 0xde, 0x38, 0x55, 0xa7, 0x4e, 0x2e, 0xbe, 0x81,
	0x9b, 0x2b, 0x71, 0x6f, 0x68, 0xb9, 0x6d, 0x48, 0x2a, 0x10, 0x8a, 0x43, 0x1f, 0xcd, 0xb7, 0x33,
	0x60, 0xed, 0x78, 0xce, 0x09, 0x7c, 0xb9, 0xac, 0xf4, 0xd8, 0x59, 0xaf, 0xae, 0xce, 0xdb, 0x60,
	0x79, 0xae, 0x64, 0x90, 0xad, 0x9c, 0xf0, 0xf2, 0x21, 0x8c, 0x8a, 0x94, 0x38, 0xb0, 0xac, 0xb3,
	0x26, 0xa0, 0x09, 0x40, 0x2a, 0x52, 0x97, 0xe2, 0x8f, 0xfc, 0xc4, 0xf2, 0x5e, 0x8f, 0x68, 0xe4,
	0x0b, 0x61, 0x53, 0x91, 0xfa, 0x20, 0x56, 0xe6, 0x67, 0x56, 0x4e, 0x38, 0xf6, 0x4d, 0xfb, 0xf6,
	0x46, 0x9b, 0x01, 0xb6, 0x15, 0xa9, 0x73, 0x51, 0x44, 0x77, 0xff, 0xa9, 0xf8, 0xdc, 0xa3, 0xd2,
	0x1e, 0xbb, 0x00, 0xbb, 0x94, 0x57, 0xe6, 0x47, 0x96, 0x8d, 0xd6, 0x0c, 0x8d, 0x56, 0xb0, 0x4f,
	0x89, 0x46, 0x7c, 0xa8, 0xf8, 0xef, 0x32, 0xdf, 0x84, 0xd9, 0x21, 0xd0, 0xb4, 0x5b, 0x2c, 0xee,
	0x39, 0x3b, 0x6c, 0x69, 0x3a, 0xe6, 0xf6, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xde, 0x11, 0x00, 0x86,
	0xdc, 0x00, 0x00, 0x00,
}
