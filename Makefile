# Makefile for ACM Project
# Automated Compromise Mitigation

.PHONY: help proto-gen proto-lint test test-integration lint build clean install-tools

# Default target
help:
	@echo "ACM Project Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  help              - Show this help message"
	@echo "  install-tools     - Install required development tools"
	@echo "  proto-gen         - Generate Go code from proto files"
	@echo "  proto-lint        - Lint proto files with buf"
	@echo "  lint              - Run golangci-lint"
	@echo "  test              - Run unit tests"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-all          - Run all tests"
	@echo "  build             - Build all binaries"
	@echo "  clean             - Clean build artifacts"
	@echo ""

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✓ Tools installed"

# Generate proto files
proto-gen:
	@echo "Generating proto files..."
	@if [ ! -d "api/proto" ]; then \
		echo "No api/proto directory found - skipping"; \
		exit 0; \
	fi
	@find api/proto -name "*.proto" -print0 | while IFS= read -r -d '' proto_file; do \
		echo "  Generating $$proto_file"; \
		protoc --proto_path=api/proto \
		       --go_out=paths=source_relative:api/proto \
		       --go-grpc_out=paths=source_relative:api/proto \
		       "$$proto_file"; \
	done
	@echo "✓ Proto generation complete"

# Lint proto files
proto-lint:
	@echo "Linting proto files..."
	@if [ ! -f "buf.yaml" ]; then \
		echo "No buf.yaml found - skipping buf lint"; \
		exit 0; \
	fi
	buf lint api/proto
	@echo "✓ Proto lint complete"

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --config .golangci.yml
	@echo "✓ Lint complete"

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✓ Tests complete"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration -timeout=15m ./test/integration/...
	@echo "✓ Integration tests complete"

# Run all tests
test-all: test test-integration

# Build binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	@if [ -d "cmd/acm-service" ]; then \
		echo "  Building acm-service..."; \
		go build -o bin/acm-service ./cmd/acm-service; \
	fi
	@if [ -d "cmd/acm-cli" ]; then \
		echo "  Building acm-cli..."; \
		go build -o bin/acm-cli ./cmd/acm-cli; \
	fi
	@echo "✓ Build complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
	@echo "✓ Clean complete"
