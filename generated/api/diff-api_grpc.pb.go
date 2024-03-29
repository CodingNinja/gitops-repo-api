// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: api/diff-api.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DiffApiClient is the client API for DiffApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DiffApiClient interface {
	Diff(ctx context.Context, in *DiffRequest, opts ...grpc.CallOption) (*DiffResponse, error)
}

type diffApiClient struct {
	cc grpc.ClientConnInterface
}

func NewDiffApiClient(cc grpc.ClientConnInterface) DiffApiClient {
	return &diffApiClient{cc}
}

func (c *diffApiClient) Diff(ctx context.Context, in *DiffRequest, opts ...grpc.CallOption) (*DiffResponse, error) {
	out := new(DiffResponse)
	err := c.cc.Invoke(ctx, "/apipb.DiffApi/Diff", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DiffApiServer is the server API for DiffApi service.
// All implementations must embed UnimplementedDiffApiServer
// for forward compatibility
type DiffApiServer interface {
	Diff(context.Context, *DiffRequest) (*DiffResponse, error)
	mustEmbedUnimplementedDiffApiServer()
}

// UnimplementedDiffApiServer must be embedded to have forward compatible implementations.
type UnimplementedDiffApiServer struct {
}

func (UnimplementedDiffApiServer) Diff(context.Context, *DiffRequest) (*DiffResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Diff not implemented")
}
func (UnimplementedDiffApiServer) mustEmbedUnimplementedDiffApiServer() {}

// UnsafeDiffApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DiffApiServer will
// result in compilation errors.
type UnsafeDiffApiServer interface {
	mustEmbedUnimplementedDiffApiServer()
}

func RegisterDiffApiServer(s grpc.ServiceRegistrar, srv DiffApiServer) {
	s.RegisterService(&DiffApi_ServiceDesc, srv)
}

func _DiffApi_Diff_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiffRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiffApiServer).Diff(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/apipb.DiffApi/Diff",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiffApiServer).Diff(ctx, req.(*DiffRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DiffApi_ServiceDesc is the grpc.ServiceDesc for DiffApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DiffApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "apipb.DiffApi",
	HandlerType: (*DiffApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Diff",
			Handler:    _DiffApi_Diff_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/diff-api.proto",
}
