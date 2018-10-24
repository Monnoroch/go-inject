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
