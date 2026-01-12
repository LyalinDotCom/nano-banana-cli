# Nanobanana CLI Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-s -w \
	-X github.com/lyalindotcom/nano-banana-cli/internal/cli.Version=$(VERSION) \
	-X github.com/lyalindotcom/nano-banana-cli/internal/cli.Commit=$(COMMIT) \
	-X github.com/lyalindotcom/nano-banana-cli/internal/cli.BuildDate=$(BUILD_DATE)"

.PHONY: build clean test lint install release-local help

## Build the CLI for current platform
build:
	go build $(LDFLAGS) -o nanobanana ./cmd/nanobanana

## Build for all platforms
build-all: build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/nanobanana-darwin-arm64 ./cmd/nanobanana

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/nanobanana-darwin-amd64 ./cmd/nanobanana

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/nanobanana-linux-amd64 ./cmd/nanobanana

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/nanobanana-linux-arm64 ./cmd/nanobanana

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/nanobanana-windows-amd64.exe ./cmd/nanobanana

## Clean build artifacts
clean:
	rm -f nanobanana
	rm -rf dist/

## Run tests
test:
	go test -v ./...

## Run linter
lint:
	golangci-lint run

## Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/nanobanana

## Create a local release (requires goreleaser)
release-local:
	goreleaser release --snapshot --clean

## Update dependencies
deps:
	go mod tidy
	go mod download

## Show help
help:
	@echo "Nanobanana CLI - Available targets:"
	@echo ""
	@echo "  build          Build for current platform"
	@echo "  build-all      Build for all supported platforms"
	@echo "  clean          Clean build artifacts"
	@echo "  test           Run tests"
	@echo "  lint           Run linter"
	@echo "  install        Install to GOPATH/bin"
	@echo "  release-local  Create local release snapshot"
	@echo "  deps           Update dependencies"
	@echo "  help           Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
