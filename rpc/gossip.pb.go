// Code generated by protoc-gen-go. DO NOT EDIT.
// source: gossip.proto

/*
Package rpc is a generated protocol buffer package.

It is generated from these files:
	gossip.proto
	service.proto

It has these top-level messages:
	Version
	Entry
	PullRequest
	PullReply
	PushRequest
	PushReply
	GetRequest
	GetReply
	PutRequest
	PutReply
*/
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

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Version represents the latest conflict-free version number for an object.
type Version struct {
	Scalar uint64 `protobuf:"varint,1,opt,name=scalar" json:"scalar,omitempty"`
	Pid    uint64 `protobuf:"varint,2,opt,name=pid" json:"pid,omitempty"`
}

func (m *Version) Reset()                    { *m = Version{} }
func (m *Version) String() string            { return proto.CompactTextString(m) }
func (*Version) ProtoMessage()               {}
func (*Version) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Version) GetScalar() uint64 {
	if m != nil {
		return m.Scalar
	}
	return 0
}

func (m *Version) GetPid() uint64 {
	if m != nil {
		return m.Pid
	}
	return 0
}

// Entry represents a key/value entry that is being synchronized.
type Entry struct {
	Parent          *Version `protobuf:"bytes,1,opt,name=parent" json:"parent,omitempty"`
	Version         *Version `protobuf:"bytes,2,opt,name=version" json:"version,omitempty"`
	Value           []byte   `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	TrackVisibility bool     `protobuf:"varint,4,opt,name=trackVisibility" json:"trackVisibility,omitempty"`
}

func (m *Entry) Reset()                    { *m = Entry{} }
func (m *Entry) String() string            { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()               {}
func (*Entry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Entry) GetParent() *Version {
	if m != nil {
		return m.Parent
	}
	return nil
}

func (m *Entry) GetVersion() *Version {
	if m != nil {
		return m.Version
	}
	return nil
}

func (m *Entry) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Entry) GetTrackVisibility() bool {
	if m != nil {
		return m.TrackVisibility
	}
	return false
}

// PullRequest sends a vector of versions to a remote and expects any more
// recent versions of objects in reply.
type PullRequest struct {
	Versions map[string]*Version `protobuf:"bytes,1,rep,name=versions" json:"versions,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *PullRequest) Reset()                    { *m = PullRequest{} }
func (m *PullRequest) String() string            { return proto.CompactTextString(m) }
func (*PullRequest) ProtoMessage()               {}
func (*PullRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PullRequest) GetVersions() map[string]*Version {
	if m != nil {
		return m.Versions
	}
	return nil
}

// PullReply contains the entries for objects that have a later version. It
// may also contain an optional pull request to initiate a push in return.
// It returns successful acknowledgement if any synchronization takes place.
type PullReply struct {
	Success bool              `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Entries map[string]*Entry `protobuf:"bytes,2,rep,name=entries" json:"entries,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Pull    *PullRequest      `protobuf:"bytes,3,opt,name=pull" json:"pull,omitempty"`
}

func (m *PullReply) Reset()                    { *m = PullReply{} }
func (m *PullReply) String() string            { return proto.CompactTextString(m) }
func (*PullReply) ProtoMessage()               {}
func (*PullReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PullReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *PullReply) GetEntries() map[string]*Entry {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *PullReply) GetPull() *PullRequest {
	if m != nil {
		return m.Pull
	}
	return nil
}

// PushRequest sends a vector of entries to a remote expecting them to be
// synchronized at the remote namespace.
type PushRequest struct {
	Entries map[string]*Entry `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *PushRequest) Reset()                    { *m = PushRequest{} }
func (m *PushRequest) String() string            { return proto.CompactTextString(m) }
func (*PushRequest) ProtoMessage()               {}
func (*PushRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PushRequest) GetEntries() map[string]*Entry {
	if m != nil {
		return m.Entries
	}
	return nil
}

// PushReply returns successful acknowledgement if syncrhonization took place.
type PushReply struct {
	Success bool `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
}

func (m *PushReply) Reset()                    { *m = PushReply{} }
func (m *PushReply) String() string            { return proto.CompactTextString(m) }
func (*PushReply) ProtoMessage()               {}
func (*PushReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *PushReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*Version)(nil), "rpc.Version")
	proto.RegisterType((*Entry)(nil), "rpc.Entry")
	proto.RegisterType((*PullRequest)(nil), "rpc.PullRequest")
	proto.RegisterType((*PullReply)(nil), "rpc.PullReply")
	proto.RegisterType((*PushRequest)(nil), "rpc.PushRequest")
	proto.RegisterType((*PushReply)(nil), "rpc.PushReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Gossip service

type GossipClient interface {
	Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushReply, error)
	Pull(ctx context.Context, in *PullRequest, opts ...grpc.CallOption) (*PullReply, error)
}

type gossipClient struct {
	cc *grpc.ClientConn
}

func NewGossipClient(cc *grpc.ClientConn) GossipClient {
	return &gossipClient{cc}
}

func (c *gossipClient) Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushReply, error) {
	out := new(PushReply)
	err := grpc.Invoke(ctx, "/rpc.Gossip/Push", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gossipClient) Pull(ctx context.Context, in *PullRequest, opts ...grpc.CallOption) (*PullReply, error) {
	out := new(PullReply)
	err := grpc.Invoke(ctx, "/rpc.Gossip/Pull", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Gossip service

type GossipServer interface {
	Push(context.Context, *PushRequest) (*PushReply, error)
	Pull(context.Context, *PullRequest) (*PullReply, error)
}

func RegisterGossipServer(s *grpc.Server, srv GossipServer) {
	s.RegisterService(&_Gossip_serviceDesc, srv)
}

func _Gossip_Push_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).Push(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Gossip/Push",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).Push(ctx, req.(*PushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gossip_Pull_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PullRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).Pull(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Gossip/Pull",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).Pull(ctx, req.(*PullRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gossip_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Gossip",
	HandlerType: (*GossipServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Push",
			Handler:    _Gossip_Push_Handler,
		},
		{
			MethodName: "Pull",
			Handler:    _Gossip_Pull_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gossip.proto",
}

func init() { proto.RegisterFile("gossip.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 402 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xc1, 0x8e, 0x9b, 0x30,
	0x10, 0x8d, 0x03, 0x81, 0x64, 0xa0, 0x6d, 0x64, 0x55, 0x15, 0x4a, 0xd5, 0x0a, 0xa1, 0xb4, 0x42,
	0x3d, 0x70, 0x20, 0xaa, 0x5a, 0xe5, 0x9e, 0x56, 0xbd, 0x55, 0x3e, 0xe4, 0x5c, 0x42, 0xad, 0xd6,
	0x8a, 0x05, 0xae, 0x0d, 0x91, 0xf8, 0x89, 0x9e, 0x76, 0x7f, 0x6a, 0xbf, 0x6a, 0x85, 0x0d, 0x59,
	0xc2, 0x46, 0x7b, 0xda, 0x9b, 0x67, 0xde, 0x9b, 0x99, 0xf7, 0x5e, 0x02, 0xf8, 0x7f, 0x4a, 0xa5,
	0x98, 0x48, 0x84, 0x2c, 0xab, 0x12, 0x5b, 0x52, 0xe4, 0xd1, 0x06, 0xdc, 0x3d, 0x95, 0x8a, 0x95,
	0x05, 0x7e, 0x03, 0x8e, 0xca, 0x33, 0x9e, 0xc9, 0x00, 0x85, 0x28, 0xb6, 0x49, 0x57, 0xe1, 0x25,
	0x58, 0x82, 0xfd, 0x0e, 0xa6, 0xba, 0xd9, 0x3e, 0xa3, 0x5b, 0x04, 0xb3, 0x5d, 0x51, 0xc9, 0x06,
	0xaf, 0xc1, 0x11, 0x99, 0xa4, 0x45, 0xa5, 0x67, 0xbc, 0xd4, 0x4f, 0xa4, 0xc8, 0x93, 0x6e, 0x23,
	0xe9, 0x30, 0xfc, 0x11, 0xdc, 0x93, 0x69, 0xe9, 0x2d, 0x63, 0x5a, 0x0f, 0xe2, 0xd7, 0x30, 0x3b,
	0x65, 0xbc, 0xa6, 0x81, 0x15, 0xa2, 0xd8, 0x27, 0xa6, 0xc0, 0x31, 0xbc, 0xaa, 0x64, 0x96, 0x1f,
	0xf7, 0x4c, 0xb1, 0x03, 0xe3, 0xac, 0x6a, 0x02, 0x3b, 0x44, 0xf1, 0x9c, 0x8c, 0xdb, 0xd1, 0x0d,
	0x02, 0xef, 0x67, 0xcd, 0x39, 0xa1, 0xff, 0x6a, 0xaa, 0x2a, 0xbc, 0x85, 0x79, 0xb7, 0x5a, 0x05,
	0x28, 0xb4, 0x62, 0x2f, 0x7d, 0xaf, 0x0f, 0x0f, 0x38, 0xbd, 0x08, 0xa5, 0xfd, 0x90, 0x33, 0x7f,
	0xf5, 0x03, 0x5e, 0x5c, 0x40, 0x6d, 0x0c, 0x47, 0xda, 0x68, 0x9f, 0x0b, 0xd2, 0x3e, 0x71, 0xd4,
	0xcb, 0xbd, 0x66, 0xca, 0x40, 0xdb, 0xe9, 0x57, 0x14, 0xdd, 0x21, 0x58, 0x98, 0x93, 0x82, 0x37,
	0x38, 0x00, 0x57, 0xd5, 0x79, 0x4e, 0x95, 0xd2, 0xbb, 0xe6, 0xa4, 0x2f, 0xf1, 0x67, 0x70, 0x69,
	0x51, 0x49, 0x46, 0x55, 0x30, 0xd5, 0x6a, 0xdf, 0x0e, 0xd4, 0x0a, 0xde, 0x24, 0x3b, 0x83, 0x1a,
	0xa9, 0x3d, 0x17, 0xaf, 0xc1, 0x16, 0x35, 0xe7, 0x3a, 0x34, 0x2f, 0x5d, 0x8e, 0x1d, 0x12, 0x8d,
	0xae, 0xbe, 0x81, 0x3f, 0x1c, 0xbf, 0x62, 0x27, 0xbc, 0xb4, 0x03, 0x7a, 0x91, 0xb9, 0x35, 0x30,
	0xf3, 0x5f, 0x67, 0xac, 0xfe, 0xf6, 0x19, 0x7f, 0x79, 0x10, 0x6d, 0x22, 0x7e, 0xd7, 0x09, 0x38,
	0x53, 0xae, 0xcb, 0x7e, 0x36, 0x41, 0x1f, 0xda, 0x70, 0xdb, 0x63, 0x4f, 0x86, 0x9b, 0xfe, 0x02,
	0xe7, 0xbb, 0xfe, 0xf7, 0xe3, 0x4f, 0x60, 0xb7, 0x03, 0x78, 0x39, 0x16, 0xba, 0x7a, 0x39, 0xe8,
	0x08, 0xde, 0x44, 0x13, 0xc3, 0xe5, 0x1c, 0x3f, 0x4a, 0xf5, 0xcc, 0xed, 0x7e, 0x9b, 0x68, 0x72,
	0x70, 0xf4, 0x67, 0xb5, 0xb9, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x37, 0x1e, 0x4d, 0xf5, 0x66, 0x03,
	0x00, 0x00,
}
