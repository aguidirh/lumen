# Makefile for lumen

.PHONY: all build tidy test run clean help generate-mocks build-mcp test-mcp

.DEFAULT_GOAL := help

BINARY_NAME=lumen
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=cmd/lumen/main.go

MCP_BINARY_NAME=mcp-server
MCP_BINARY_PATH=bin/$(MCP_BINARY_NAME)
MCP_SERVER_PATH=server

all: tidy generate-mocks build build-mcp

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "$(BINARY_NAME) built in $(BINARY_PATH)"

# Build the MCP server
build-mcp:
	@echo "Building $(MCP_BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(MCP_BINARY_PATH) ./$(MCP_SERVER_PATH)
	@echo "$(MCP_BINARY_NAME) built in $(MCP_BINARY_PATH)"

# Tidy up dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Generate mocks from interfaces
generate-mocks:
	@echo "Generating mocks..."
	@go generate ./internal/pkg/catalog/...
	@go generate ./internal/pkg/cli/...
	@go generate ./internal/pkg/image/...
	@go generate ./internal/pkg/list/...
	@go generate ./internal/pkg/printer/...
	@echo "Mocks generated successfully"

# Run tests
test:
	@echo "Running tests..."
	@go test -v $$(go list ./... | grep -v '/mock$$')

# Test the MCP server
test-mcp: build-mcp
	@echo "Testing MCP server..."
	@go test -v ./$(MCP_SERVER_PATH)

# Generate test coverage report
test-coverage:
	@echo "Generating test coverage report..."
	@mkdir -p test/coverage
	@go test -coverprofile=test/coverage/coverage.out $$(go list ./... | grep -v '/mock$$')
	@echo "Coverage report generated: test/coverage/coverage.out"
	@echo "To view the report, run: make view-coverage"

# View test coverage report in browser
view-coverage: test-coverage
	@echo "Opening coverage report in browser..."
	@go tool cover -html=test/coverage/coverage.out

# Run the application
run: tidy
	@echo "Running $(BINARY_NAME)..."
	@go run $(CMD_PATH)

# Run the MCP server
run-mcp: build-mcp
	@echo "Starting MCP server..."
	@echo "The server will read JSON requests from stdin and write responses to stdout."
	@echo "Press Ctrl+C to stop."
	@./$(MCP_BINARY_PATH)

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/ test/

# Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Tidy, generate mocks, build app and MCP server (default)"
	@echo "  build          - Build the application binary"
	@echo "  build-mcp      - Build the MCP server binary"
	@echo "  tidy           - Tidy up Go module dependencies"
	@echo "  generate-mocks - Generate mocks from interfaces"
	@echo "  test           - Run all tests"
	@echo "  test-mcp       - Test the MCP server functionality"
	@echo "  test-coverage  - Generate a test coverage report"
	@echo "  view-coverage  - Open the HTML test coverage report in a browser"
	@echo "  run            - Run the application"
	@echo "  run-mcp        - Start the MCP server"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message" 