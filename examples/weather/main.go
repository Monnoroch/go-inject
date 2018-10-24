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

func AiServiceEndpoint() string {
	return "ai-service:80"
}

func NewAiServiceGrpcConnection(aiServiceEndpoint string) *grpc.ClientConn {
	connection, _ := grpc.Dial(aiServiceEndpoint, grpc.WithInsecure())
	return connection
}

func NewGrpcAiClient(connection *grpc.ClientConn) aiproto.AiClient {
	return aiproto.NewAiClient(connection)
}

func NewAiClient(aiClient aiproto.AiClient) ai.AiClient {
	return ai.AiClient{RawAiClient: aiClient}
}

func NewServer(client ai.AiClient) *Server {
	return &Server{AiClient: client}
}

func main() {
	endpoint := AiServiceEndpoint()
	connection := NewAiServiceGrpcConnection(endpoint)
	aiClient := NewAiClient(NewGrpcAiClient(connection))
	weatherPredictionServer := NewServer(aiClient)
	server := grpc.NewServer()
	proto.RegisterWeatherPredictionServer(
		server,
		weatherPredictionServer,
	)

	listener, _ := net.Listen("tcp", ":80")
	server.Serve(listener)
}
