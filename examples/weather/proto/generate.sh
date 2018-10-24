protoc -I . --go_out=plugins=grpc:. ./examples/weather/proto/ai/ai.proto
protoc -I . --go_out=plugins=grpc:. ./examples/weather/proto/weather.proto
protoc -I . --go_out=plugins=grpc:. ./examples/weather/proto/blockchain/blockchain.proto
