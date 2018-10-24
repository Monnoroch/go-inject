package grpc

import (
	"google.golang.org/grpc"
)

/// Annotation used by the gRPC client module.
type GrpcClient struct{}

/// A module for providing gRPC client components.
type GrpcClientModule struct{}

func (_ GrpcClientModule) ProvideConnection(
	endpoint string, _ GrpcClient,
) (*grpc.ClientConn, GrpcClient, error) {
	connection, err := grpc.Dial(endpoint, grpc.WithInsecure())
	return connection, GrpcClient{}, err
}

/// Annotation used by the gRPC server module.
type GrpcServer struct{}

/// A module for providing gRPC server components.
type GrpcServerModule struct{}

func (_ GrpcServerModule) ProvideServer() (*grpc.Server, GrpcServer) {
	return grpc.NewServer(), GrpcServer{}
}
