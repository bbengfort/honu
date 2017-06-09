// Code generated by protoc-gen-go.
// source: service.proto
// DO NOT EDIT!

package rpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// GetRequest is sent from a client to the server to read a value for a key
type GetRequest struct {
	Key string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
}

func (m *GetRequest) Reset()                    { *m = GetRequest{} }
func (m *GetRequest) String() string            { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()               {}
func (*GetRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *GetRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

// GetReply is a response from the server to the client with the value
type GetReply struct {
	Success bool   `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Version string `protobuf:"bytes,2,opt,name=version" json:"version,omitempty"`
	Key     string `protobuf:"bytes,3,opt,name=key" json:"key,omitempty"`
	Value   []byte `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	Error   string `protobuf:"bytes,5,opt,name=error" json:"error,omitempty"`
}

func (m *GetReply) Reset()                    { *m = GetReply{} }
func (m *GetReply) String() string            { return proto.CompactTextString(m) }
func (*GetReply) ProtoMessage()               {}
func (*GetReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *GetReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *GetReply) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *GetReply) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *GetReply) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *GetReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// PutRequest is sent from a client to the server to put a value for a key
type PutRequest struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *PutRequest) Reset()                    { *m = PutRequest{} }
func (m *PutRequest) String() string            { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()               {}
func (*PutRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *PutRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *PutRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

// PutReply is a response from the leader to the client
type PutReply struct {
	Success bool   `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Key     string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	Version string `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
	Error   string `protobuf:"bytes,4,opt,name=error" json:"error,omitempty"`
}

func (m *PutReply) Reset()                    { *m = PutReply{} }
func (m *PutReply) String() string            { return proto.CompactTextString(m) }
func (*PutReply) ProtoMessage()               {}
func (*PutReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *PutReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *PutReply) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *PutReply) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *PutReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*GetRequest)(nil), "rpc.GetRequest")
	proto.RegisterType((*GetReply)(nil), "rpc.GetReply")
	proto.RegisterType((*PutRequest)(nil), "rpc.PutRequest")
	proto.RegisterType((*PutReply)(nil), "rpc.PutReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Storage service

type StorageClient interface {
	GetValue(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetReply, error)
	PutValue(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutReply, error)
}

type storageClient struct {
	cc *grpc.ClientConn
}

func NewStorageClient(cc *grpc.ClientConn) StorageClient {
	return &storageClient{cc}
}

func (c *storageClient) GetValue(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetReply, error) {
	out := new(GetReply)
	err := grpc.Invoke(ctx, "/rpc.Storage/GetValue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) PutValue(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutReply, error) {
	out := new(PutReply)
	err := grpc.Invoke(ctx, "/rpc.Storage/PutValue", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Storage service

type StorageServer interface {
	GetValue(context.Context, *GetRequest) (*GetReply, error)
	PutValue(context.Context, *PutRequest) (*PutReply, error)
}

func RegisterStorageServer(s *grpc.Server, srv StorageServer) {
	s.RegisterService(&_Storage_serviceDesc, srv)
}

func _Storage_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Storage/GetValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).GetValue(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_PutValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).PutValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Storage/PutValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).PutValue(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Storage_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Storage",
	HandlerType: (*StorageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetValue",
			Handler:    _Storage_GetValue_Handler,
		},
		{
			MethodName: "PutValue",
			Handler:    _Storage_PutValue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}

func init() { proto.RegisterFile("service.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 241 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0x31, 0x4f, 0xc3, 0x30,
	0x10, 0x85, 0x9b, 0xa4, 0xa5, 0xe1, 0x44, 0x05, 0xb2, 0x18, 0x2c, 0x06, 0x54, 0x79, 0xea, 0x80,
	0x32, 0x00, 0xff, 0x81, 0xb5, 0x0a, 0x12, 0x7b, 0xb1, 0x0e, 0x54, 0x11, 0x61, 0x73, 0xb6, 0x23,
	0x45, 0xe2, 0xc7, 0x23, 0x5f, 0xeb, 0x3a, 0x19, 0xe8, 0x96, 0xe7, 0x7b, 0xf7, 0xe9, 0xbd, 0x0b,
	0xac, 0x1c, 0x52, 0xbf, 0xd7, 0xd8, 0x58, 0x32, 0xde, 0x88, 0x8a, 0xac, 0x56, 0xf7, 0x00, 0x2f,
	0xe8, 0x5b, 0xfc, 0x09, 0xe8, 0xbc, 0xb8, 0x81, 0xea, 0x0b, 0x07, 0x59, 0xac, 0x8b, 0xcd, 0x65,
	0x1b, 0x3f, 0xd5, 0x2f, 0xd4, 0x3c, 0xb7, 0xdd, 0x20, 0x24, 0x2c, 0x5d, 0xd0, 0x1a, 0x9d, 0x63,
	0x47, 0xdd, 0x26, 0x19, 0x27, 0x3d, 0x92, 0xdb, 0x9b, 0x6f, 0x59, 0xf2, 0x6e, 0x92, 0x89, 0x58,
	0x9d, 0x88, 0xe2, 0x16, 0x16, 0xfd, 0xae, 0x0b, 0x28, 0xe7, 0xeb, 0x62, 0x73, 0xd5, 0x1e, 0x44,
	0x7c, 0x45, 0x22, 0x43, 0x72, 0xc1, 0xce, 0x83, 0x50, 0xcf, 0x00, 0xdb, 0xf0, 0x7f, 0xba, 0xcc,
	0x2a, 0x47, 0x2c, 0xf5, 0x01, 0x35, 0x6f, 0x9d, 0xcf, 0x7c, 0xa4, 0x95, 0x99, 0x36, 0x6a, 0x51,
	0x4d, 0x5b, 0x9c, 0xd2, 0xcd, 0x47, 0xe9, 0x1e, 0x11, 0x96, 0xaf, 0xde, 0xd0, 0xee, 0x13, 0xc5,
	0x03, 0x9f, 0xe9, 0x8d, 0xab, 0x5c, 0x37, 0x64, 0x75, 0x93, 0xaf, 0x7a, 0xb7, 0xca, 0x0f, 0xb6,
	0x1b, 0xd4, 0x2c, 0xba, 0xb7, 0x61, 0xe2, 0xce, 0x2d, 0x8f, 0xee, 0x54, 0x40, 0xcd, 0xde, 0x2f,
	0xf8, 0x77, 0x3d, 0xfd, 0x05, 0x00, 0x00, 0xff, 0xff, 0x99, 0xa4, 0x5d, 0x4c, 0xbf, 0x01, 0x00,
	0x00,
}
