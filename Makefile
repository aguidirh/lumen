# Makefile for lumen

.PHONY: all build tidy test run clean help

.DEFAULT_GOAL := help

BINARY_NAME=lumen
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=cmd/lumen/main.go

all: tidy build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "$(BINARY_NAME) built in $(BINARY_PATH)"

# Tidy up dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run the application
run: tidy
	@echo "Running $(BINARY_NAME)..."
	@go run $(CMD_PATH)

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/

# Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Tidy and build the application (default)"
	@echo "  build        - Build the application binary"
	@echo "  tidy         - Tidy up Go module dependencies"
	@echo "  test         - Run all tests"
	@echo "  run          - Run the application"
	@echo "  clean        - Remove build artifacts"
	@echo "  help         - Show this help message" 