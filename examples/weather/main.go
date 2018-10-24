package main

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/monnoroch/go-inject"
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
type WeatherPredictionServerModule struct{}

/// Provider returning the AI service endpoint, to be used by the gRPC client module.
func (_ WeatherPredictionServerModule) ProvideGrpcEndpoint() (string, grpcinject.GrpcClient) {
	return "ai-service:80", grpcinject.GrpcClient{}
}

func (_ WeatherPredictionServerModule) ProvideServer(
	client ai.AiClient, _ ai.AiService,
) (*Server, WeatherPrediction) {
	return &Server{AiClient: client}, WeatherPrediction{}
}

func (_ WeatherPredictionServerModule) ProvideGrpcServer(
	grpcServer *grpc.Server, _ grpcinject.GrpcServer,
	weatherPredictionServer *Server, _ WeatherPrediction,
) (*grpc.Server, WeatherPrediction) {
	proto.RegisterWeatherPredictionServer(
		grpcServer,
		weatherPredictionServer,
	)
	return grpcServer, WeatherPrediction{}
}

func main() {
	injector, _ := inject.InjectorOf(
		grpcinject.GrpcServerModule{},
		grpcinject.GrpcClientModule{},
		ai.AiServiceClientModule{},
		WeatherPredictionServerModule{},
	)
	server := injector.MustGet(
		new(*grpc.Server), WeatherPrediction{},
	).(*grpc.Server)
	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
