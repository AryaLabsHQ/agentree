.PHONY: build test clean install run lint

# Binary name
BINARY_NAME=agentree
MAIN_PATH=./cmd/agentree

# Build the binary
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run integration tests
test-integration:
	go test -v -tags=integration ./cmd

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
install:
	go install $(MAIN_PATH)

# Run the application
run: build
	./$(BINARY_NAME)

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Update dependencies
deps:
	go mod tidy
	go mod download

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-integration- Run integration tests"
	@echo "  clean           - Remove build artifacts"
	@echo "  install         - Install to GOPATH/bin"
	@echo "  run             - Build and run the application"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  deps            - Update dependencies"
	@echo "  build-all       - Build for all platforms"
	@echo "  help            - Show this help message"