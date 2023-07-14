.PHONY: build run lint test docker

GO_FILES=$(shell find . -name '*.go')

build:
	@go build -o ./cmd/generator/main ./cmd/generator

run:
	@go run ./cmd/generator/main.go

lint:
	@golangci-lint run

test:
	@go test -v -race ./...

proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	./protobuf/telematics_data.proto