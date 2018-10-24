package blockchain

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/auto"
	grpcinject "github.com/monnoroch/go-inject/examples/weather/grpc"
	blockchainproto "github.com/monnoroch/go-inject/examples/weather/proto/blockchain"
)

/// A wrapper type around blockchain client.
type BlockchainClient struct {
	RawBlockchainClient blockchainproto.BlockchainClient
}

/// Make payment using the blockchain service.
func (self *BlockchainClient) Pay(
	ctx context.Context,
	userId int64,
) bool {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, requestMetadata)
	_, err := self.RawBlockchainClient.Pay(ctx, &blockchainproto.PayRequest{
		From:           userId,
		To:             12345, // our app's user id, not enough funding to make it a flag
		AmountMicroEth: 5,
	})
	return err == nil
}

func (self BlockchainClient) ProvideAutoInjectAnnotations() interface{} {
	return struct {
		RawBlockchainClient private
	}{}
}

/// Annotation used by the AI service client module.
type BlockchainService struct{}

/// Annotation for private providers.
type private struct{}

/// A module for providing AI service client components.
type blockchainServiceClientModule struct{}

func (_ blockchainServiceClientModule) ProvideCachedGrpcClient(
	connection *grpc.ClientConn, _ grpcinject.GrpcClient,
) (blockchainproto.BlockchainClient, private) {
	return blockchainproto.NewBlockchainClient(connection), private{}
}

func BlockchainServiceClientModule() inject.Module {
	return inject.CombineModules(
		blockchainServiceClientModule{},
		autoinject.AutoInjectModule(new(BlockchainClient)),
	)
}
