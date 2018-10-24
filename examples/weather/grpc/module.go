package grpc

import (
	"google.golang.org/grpc"
)

/// A module for providing gRPC client components.
type GrpcClientModule struct{}

func (_ GrpcClientModule) ProvideConnection(
	endpoint string,
) (*grpc.ClientConn, error) {
	connection, err := grpc.Dial(endpoint, grpc.WithInsecure())
	return connection, err
}
