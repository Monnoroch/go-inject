package main

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/monnoroch/go-inject"
	"github.com/monnoroch/go-inject/auto"
	"github.com/monnoroch/go-inject/examples/weather/ai"
	grpcinject "github.com/monnoroch/go-inject/examples/weather/grpc"
	proto "github.com/monnoroch/go-inject/examples/weather/proto"
)

/// The main server type for defining request handlers.
type Server struct {
	AiClient ai.AiClient
}

/// Handler for the WeatherPrediction.Predict RPCs.
func (self *Server) Predict(
	ctx context.Context,
	request *proto.SpaceTimeLocation,
) (*proto.Weather, error) {
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
func (_ weatherPredictionServerModule) ProvideGrpcEndpoint() (string, grpcinject.GrpcClient) {
	return "ai-service:80", grpcinject.GrpcClient{}
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
		grpcinject.GrpcClientModule{},
		ai.AiServiceClientModule(),
		WeatherPredictionServerModule(),
	)
	server := injector.MustGet(
		new(*grpc.Server), WeatherPrediction{},
	).(*grpc.Server)
	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
