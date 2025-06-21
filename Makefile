.PHONY: all build clean run test lint help setup

# Go parameters
BINARY_NAME=zikrr
MAIN_PACKAGE=./cmd/zikrr
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Colors for terminal output
GREEN=\033[0;32m
NC=\033[0m # No Color

all: clean build

help:
	@echo "Available commands:"
	@echo "  ${GREEN}make${NC}          - Clean and build the project"
	@echo "  ${GREEN}make build${NC}    - Build the binary"
	@echo "  ${GREEN}make run${NC}      - Run the application"
	@echo "  ${GREEN}make clean${NC}    - Remove build artifacts"
	@echo "  ${GREEN}make test${NC}     - Run tests"
	@echo "  ${GREEN}make lint${NC}     - Run linters"
	@echo "  ${GREEN}make setup${NC}    - Install development dependencies"
	@echo "  ${GREEN}make help${NC}     - Show this help message"

build:
	@echo "Building ${BINARY_NAME}..."
	@go build ${LDFLAGS} -o bin/${BINARY_NAME} ${MAIN_PACKAGE}
	@echo "${GREEN}Build successful!${NC}"

run: build
	@echo "Running ${BINARY_NAME}..."
	@./bin/${BINARY_NAME}

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "${GREEN}Cleanup complete!${NC}"

test:
	@echo "Running tests..."
	@go test -v -race ./...

lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please run 'make setup' first."; \
		exit 1; \
	fi

setup:
	@echo "Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go mod download
	@go mod tidy
	@echo "${GREEN}Setup complete!${NC}"

# Create necessary directories
bin:
	@mkdir -p bin 