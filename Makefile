.PHONY: build test clean example

# Build the SDK
build:
	go build -o bin/paygent-sdk-go .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the example
example:
	go run example/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build example
build-example:
	go build -o bin/example example/main.go

# Run all checks
check: fmt lint test

# Install the SDK locally for testing
install:
	go install .


