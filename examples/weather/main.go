package main

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/monnoroch/go-inject/examples/weather/ai"
	proto "github.com/monnoroch/go-inject/examples/weather/proto"
	aiproto "github.com/monnoroch/go-inject/examples/weather/proto/ai"
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

func main() {
	aiConnection, _ := grpc.Dial("ai-service:80", grpc.WithInsecure())
	weatherPredictionServer := &Server{
		AiClient: ai.AiClient{
			RawAiClient: aiproto.NewAiClient(aiConnection),
		},
	}
	server := grpc.NewServer()
	proto.RegisterWeatherPredictionServer(
		server,
		weatherPredictionServer,
	)

	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
