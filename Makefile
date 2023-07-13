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
