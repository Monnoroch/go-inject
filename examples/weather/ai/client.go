package ai

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	aiproto "github.com/monnoroch/go-inject/examples/weather/proto/ai"
)

/// A wrapper type around Ai client.
type AiClient struct {
	RawAiClient aiproto.AiClient
}

/// Ask AI service for weather at location and time specified in arguments.
func (self *AiClient) AskForWeather(
	ctx context.Context,
	location string,
	timestamp int64,
) string {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, requestMetadata)
	answer, _ := self.RawAiClient.Ask(ctx, &aiproto.Question{
		Question: fmt.Sprintf(
			"What's the weather at location '%s' at time '%d'",
			location,
			timestamp,
		),
	})
	return answer.GetAnswer()
}
