# Makefile for lumen

.PHONY: all build tidy test run clean help generate-mocks

.DEFAULT_GOAL := help

BINARY_NAME=lumen
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=cmd/lumen/main.go

all: tidy generate-mocks build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_PATH) $(CMD_PATH)
	@echo "$(BINARY_NAME) built in $(BINARY_PATH)"

# Tidy up dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Generate mocks from interfaces
generate-mocks:
	@echo "Generating mocks..."
	@go generate ./pkg/catalog/...
	@go generate ./pkg/cli/...
	@go generate ./pkg/image/...
	@go generate ./pkg/list/...
	@go generate ./pkg/printer/...
	@echo "Mocks generated successfully"

# Run tests
test:
	@echo "Running tests..."
	@go test -v $$(go list ./... | grep -v '/mock$$')

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

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/ test/

# Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Tidy, generate mocks, and build the application (default)"
	@echo "  build          - Build the application binary"
	@echo "  tidy           - Tidy up Go module dependencies"
	@echo "  generate-mocks - Generate mocks from interfaces"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Generate a test coverage report"
	@echo "  view-coverage  - Open the HTML test coverage report in a browser"
	@echo "  run            - Run the application"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message" 