# Makefile for lhcli

BINARY_NAME=lhcli
MODULE_NAME=github.com/longhorn/lhcli
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}"

.PHONY: all build clean test coverage lint fmt help

all: build

build: ## Build the binary
       go build ${LDFLAGS} -o ${BINARY_NAME}

install: ## Install the binary
	go install ${LDFLAGS}

clean: ## Remove build artifacts
	go clean
	rm -f ${BINARY_NAME}
	rm -rf dist/

test: ## Run tests
	go test -v ./...

coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	golangci-lint run

fmt: ## Format code
	go fmt ./...

run: ## Run the CLI
	go run main.go

deps: ## Download dependencies
	go mod download
	go mod tidy

cross-build: ## Build for multiple platforms
       GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64
       GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64
       GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64
       GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64
       #GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
