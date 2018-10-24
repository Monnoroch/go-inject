package grpc

import (
	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/rewrite"
	"google.golang.org/grpc"
)

/// Annotation used by the gRPC client module.
type grpcClient struct{}

/// A module for providing gRPC client components.
type grpcClientModule struct{}

func (_ grpcClientModule) ProvideConnection(
	endpoint string, _ grpcClient,
) (*grpc.ClientConn, grpcClient, error) {
	connection, err := grpc.Dial(endpoint, grpc.WithInsecure())
	return connection, grpcClient{}, err
}

func GrpcClientModule(annotation inject.Annotation) inject.Module {
	return rewrite.RewriteAnnotations(
		grpcClientModule{},
		rewrite.AnnotationsMapping{
			grpcClient{}: annotation,
		},
	)
}

/// Annotation used by the gRPC server module.
type GrpcServer struct{}

/// A module for providing gRPC server components.
type GrpcServerModule struct{}

func (_ GrpcServerModule) ProvideServer() (*grpc.Server, GrpcServer) {
	return grpc.NewServer(), GrpcServer{}
}
