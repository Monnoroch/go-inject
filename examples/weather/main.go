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
	"github.com/monnoroch/go-inject/examples/weather/constant"
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
		constant.ConstantModule("ai-service:80", ai.AiService{}),
		constant.ConstantModule("blockchain-service:80", blockchain.BlockchainService{}),
		constant.ConstantModule(true, blockchain.BlockchainService{}),
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
	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
