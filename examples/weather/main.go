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

/// A module for providing a configured weather prediction server.
type WeatherPredictionServerModule struct{}

/// Provider returning the AI service endpoint, to be used by the gRPC client module.
func (_ WeatherPredictionServerModule) ProvideGrpcEndpoint() string {
	return "ai-service:80"
}

func (_ WeatherPredictionServerModule) ProvideServer(
	client ai.AiClient,
) *Server {
	return &Server{AiClient: client}
}

func main() {
	injector, _ := inject.InjectorOf(
		grpcinject.GrpcClientModule{},
		ai.AiServiceClientModule{},
		WeatherPredictionServerModule{},
	)
	weatherPredictionServer := injector.MustGet(new(*Server)).(*Server)

	server := grpc.NewServer()
	proto.RegisterWeatherPredictionServer(
		server,
		weatherPredictionServer,
	)
	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
