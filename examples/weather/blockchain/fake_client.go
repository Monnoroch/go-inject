package blockchain

import (
	"context"

	"google.golang.org/grpc"

	blockchainproto "github.com/monnoroch/go-inject/examples/weather/proto/blockchain"
)

type develBlockchainClient struct{}

/// Make all payments succeed.
func (self develBlockchainClient) Pay(
	_ context.Context,
	_ *blockchainproto.PayRequest,
	_ ...grpc.CallOption,
) (*blockchainproto.PayResponse, error) {
	return &blockchainproto.PayResponse{}, nil
}
