package main

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"

	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/auto"
	"github.com/monnoroch/go-inject/examples/weather/ai"
	"github.com/monnoroch/go-inject/examples/weather/blockchain"
	grpcinject "github.com/monnoroch/go-inject/examples/weather/grpc"
	proto "github.com/monnoroch/go-inject/examples/weather/proto"
)

/// The main server type for defining request handlers.
type Server struct {
	AiClient         ai.AiClient
	BlockchainClient blockchain.BlockchainClient
}

/// Handler for the WeatherPrediction.Predict RPCs.
func (self *Server) Predict(
	ctx context.Context,
	request *proto.SpaceTimeLocation,
) (*proto.Weather, error) {
	if !self.BlockchainClient.Pay(ctx, request.GetUserId()) {
		return &proto.Weather{}, errors.New("no money -- no weather!")
	}
	weather := self.AiClient.AskForWeather(
		ctx,
		request.GetLocation(),
		request.GetTimestamp(),
	)
	return &proto.Weather{Weather: weather}, nil
}

/// Annotation used by the weather prediction server module.
type WeatherPrediction struct{}

/// A module for providing a configured weather prediction server.
type weatherPredictionServerModule struct{}

/// Provider returning the AI service endpoint, to be used by the gRPC client module.
func (_ weatherPredictionServerModule) ProvideGrpcEndpoint() (string, ai.AiService) {
	return "ai-service:80", ai.AiService{}
}

/// Provider returning the blockchain service endpoint, to be used by the gRPC client module.
func (_ weatherPredictionServerModule) ProvideBlockchainGrpcEndpoint() (string, blockchain.BlockchainService) {
	return "blockchain-service:80", blockchain.BlockchainService{}
}

func (_ weatherPredictionServerModule) ProvideIsDevelInstance() (bool, blockchain.BlockchainService) {
	return true, blockchain.BlockchainService{} // TODO: make it into a flag
}

func (_ weatherPredictionServerModule) ProvideGrpcServer(
	grpcServer *grpc.Server, _ grpcinject.GrpcServer,
	weatherPredictionServer *Server, _ autoinject.Auto,
) (*grpc.Server, WeatherPrediction) {
	proto.RegisterWeatherPredictionServer(
		grpcServer,
		weatherPredictionServer,
	)
	return grpcServer, WeatherPrediction{}
}

func WeatherPredictionServerModule() inject.Module {
	return inject.CombineModules(
		weatherPredictionServerModule{},
		autoinject.AutoInjectModule(new(*Server)),
	)
}

func main() {
	injector, _ := inject.InjectorOf(
		grpcinject.GrpcServerModule{},
		grpcinject.GrpcClientModule(ai.AiService{}),
		grpcinject.GrpcClientModule(blockchain.BlockchainService{}),
		ai.AiServiceClientModule(),
		blockchain.BlockchainServiceClientModule(),
		WeatherPredictionServerModule(),
	)
	server := injector.MustGet(
		new(*grpc.Server), WeatherPrediction{},
	).(*grpc.Server)
	listener, _ := net.Listen("tcp", ":10080")
	server.Serve(listener)
}
