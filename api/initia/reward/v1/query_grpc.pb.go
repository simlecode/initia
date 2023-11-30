// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: initia/reward/v1/query.proto

package rewardv1

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

const (
	Query_Params_FullMethodName                = "/initia.reward.v1.Query/Params"
	Query_AnnualProvisions_FullMethodName      = "/initia.reward.v1.Query/AnnualProvisions"
	Query_LastDilutionTimestamp_FullMethodName = "/initia.reward.v1.Query/LastDilutionTimestamp"
)

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueryClient interface {
	// Params returns the total set of minting parameters.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// AnnualProvisions returns current minting annual provisions value.
	AnnualProvisions(ctx context.Context, in *QueryAnnualProvisionsRequest, opts ...grpc.CallOption) (*QueryAnnualProvisionsResponse, error)
	// LastDilutionTimestamp returns the time when the last release rate dilution occurred.
	LastDilutionTimestamp(ctx context.Context, in *QueryLastDilutionTimestampRequest, opts ...grpc.CallOption) (*QueryLastDilutionTimestampResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, Query_Params_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) AnnualProvisions(ctx context.Context, in *QueryAnnualProvisionsRequest, opts ...grpc.CallOption) (*QueryAnnualProvisionsResponse, error) {
	out := new(QueryAnnualProvisionsResponse)
	err := c.cc.Invoke(ctx, Query_AnnualProvisions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) LastDilutionTimestamp(ctx context.Context, in *QueryLastDilutionTimestampRequest, opts ...grpc.CallOption) (*QueryLastDilutionTimestampResponse, error) {
	out := new(QueryLastDilutionTimestampResponse)
	err := c.cc.Invoke(ctx, Query_LastDilutionTimestamp_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
// All implementations must embed UnimplementedQueryServer
// for forward compatibility
type QueryServer interface {
	// Params returns the total set of minting parameters.
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// AnnualProvisions returns current minting annual provisions value.
	AnnualProvisions(context.Context, *QueryAnnualProvisionsRequest) (*QueryAnnualProvisionsResponse, error)
	// LastDilutionTimestamp returns the time when the last release rate dilution occurred.
	LastDilutionTimestamp(context.Context, *QueryLastDilutionTimestampRequest) (*QueryLastDilutionTimestampResponse, error)
	mustEmbedUnimplementedQueryServer()
}

// UnimplementedQueryServer must be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) AnnualProvisions(context.Context, *QueryAnnualProvisionsRequest) (*QueryAnnualProvisionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AnnualProvisions not implemented")
}
func (UnimplementedQueryServer) LastDilutionTimestamp(context.Context, *QueryLastDilutionTimestampRequest) (*QueryLastDilutionTimestampResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LastDilutionTimestamp not implemented")
}
func (UnimplementedQueryServer) mustEmbedUnimplementedQueryServer() {}

// UnsafeQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServer will
// result in compilation errors.
type UnsafeQueryServer interface {
	mustEmbedUnimplementedQueryServer()
}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&Query_ServiceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_AnnualProvisions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAnnualProvisionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).AnnualProvisions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_AnnualProvisions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).AnnualProvisions(ctx, req.(*QueryAnnualProvisionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_LastDilutionTimestamp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryLastDilutionTimestampRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).LastDilutionTimestamp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_LastDilutionTimestamp_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).LastDilutionTimestamp(ctx, req.(*QueryLastDilutionTimestampRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Query_ServiceDesc is the grpc.ServiceDesc for Query service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Query_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "initia.reward.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "AnnualProvisions",
			Handler:    _Query_AnnualProvisions_Handler,
		},
		{
			MethodName: "LastDilutionTimestamp",
			Handler:    _Query_LastDilutionTimestamp_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "initia/reward/v1/query.proto",
}