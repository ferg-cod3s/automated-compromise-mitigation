#!/usr/bin/env bash
# Copyright 2025 ACM Project
# SPDX-License-Identifier: Apache-2.0

# Generate Go code from Protocol Buffer definitions
# This script uses buf (https://buf.build) for code generation.
# It can also fall back to protoc if buf is not available.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROTO_DIR="${PROJECT_ROOT}/api/proto"

echo -e "${GREEN}ACM Protocol Buffer Code Generation${NC}"
echo "======================================"
echo "Project root: ${PROJECT_ROOT}"
echo "Proto directory: ${PROTO_DIR}"
echo ""

# Check if buf is available
if command -v buf &> /dev/null; then
    echo -e "${GREEN}✓${NC} Found buf: $(buf --version)"
    USE_BUF=true
else
    echo -e "${YELLOW}⚠${NC}  buf not found. Falling back to protoc."
    USE_BUF=false
fi

# Function to check protoc installation
check_protoc() {
    if ! command -v protoc &> /dev/null; then
        echo -e "${RED}✗${NC} protoc not found!"
        echo ""
        echo "Please install protoc (Protocol Buffer Compiler):"
        echo "  macOS:    brew install protobuf"
        echo "  Linux:    sudo apt install protobuf-compiler"
        echo "  Windows:  choco install protoc"
        echo ""
        echo "Or download from: https://github.com/protocolbuffers/protobuf/releases"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Found protoc: $(protoc --version)"
}

# Function to check protoc-gen-go plugin
check_protoc_plugins() {
    if ! command -v protoc-gen-go &> /dev/null; then
        echo -e "${YELLOW}⚠${NC}  protoc-gen-go not found. Installing..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi

    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        echo -e "${YELLOW}⚠${NC}  protoc-gen-go-grpc not found. Installing..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi

    echo -e "${GREEN}✓${NC} Protocol Buffer Go plugins installed"
}

# Clean generated files
clean_generated() {
    echo ""
    echo "Cleaning previously generated files..."
    find "${PROTO_DIR}" -type f -name "*.pb.go" -delete
    find "${PROTO_DIR}" -type f -name "*_grpc.pb.go" -delete
    echo -e "${GREEN}✓${NC} Cleaned generated files"
}

# Generate using buf
generate_with_buf() {
    echo ""
    echo "Generating code with buf..."
    cd "${PROTO_DIR}"

    # Lint proto files
    echo "  → Linting proto files..."
    if buf lint; then
        echo -e "${GREEN}✓${NC} Lint passed"
    else
        echo -e "${YELLOW}⚠${NC}  Lint warnings (non-fatal)"
    fi

    # Generate code
    echo "  → Generating Go code..."
    if buf generate; then
        echo -e "${GREEN}✓${NC} Code generation complete"
    else
        echo -e "${RED}✗${NC} Code generation failed"
        exit 1
    fi
}

# Generate using protoc (fallback)
generate_with_protoc() {
    echo ""
    echo "Generating code with protoc..."

    check_protoc
    check_protoc_plugins

    cd "${PROTO_DIR}"

    # Generate Go code for each proto file
    for proto_file in acm/v1/*.proto; do
        echo "  → Generating from ${proto_file}..."
        protoc \
            --go_out=. \
            --go_opt=paths=source_relative \
            --go-grpc_out=. \
            --go-grpc_opt=paths=source_relative \
            --go-grpc_opt=require_unimplemented_servers=true \
            --proto_path=. \
            "${proto_file}"
    done

    echo -e "${GREEN}✓${NC} Code generation complete"
}

# Verify generated files
verify_generated() {
    echo ""
    echo "Verifying generated files..."

    local expected_files=(
        "acm/v1/common.pb.go"
        "acm/v1/credential.pb.go"
        "acm/v1/credential_grpc.pb.go"
        "acm/v1/audit.pb.go"
        "acm/v1/audit_grpc.pb.go"
        "acm/v1/him.pb.go"
        "acm/v1/him_grpc.pb.go"
    )

    local all_exist=true
    for file in "${expected_files[@]}"; do
        if [[ -f "${PROTO_DIR}/${file}" ]]; then
            echo -e "${GREEN}✓${NC} ${file}"
        else
            echo -e "${RED}✗${NC} ${file} (missing)"
            all_exist=false
        fi
    done

    if [[ "${all_exist}" == "true" ]]; then
        echo ""
        echo -e "${GREEN}✓${NC} All expected files generated successfully"
        return 0
    else
        echo ""
        echo -e "${RED}✗${NC} Some files are missing"
        return 1
    fi
}

# Format generated Go code
format_generated() {
    echo ""
    echo "Formatting generated Go code..."
    cd "${PROTO_DIR}"

    if command -v gofmt &> /dev/null; then
        gofmt -w -s acm/v1/*.go
        echo -e "${GREEN}✓${NC} Go code formatted"
    else
        echo -e "${YELLOW}⚠${NC}  gofmt not found, skipping formatting"
    fi
}

# Main execution
main() {
    # Clean old files
    if [[ "${1:-}" != "--no-clean" ]]; then
        clean_generated
    fi

    # Generate code
    if [[ "${USE_BUF}" == "true" ]]; then
        generate_with_buf
    else
        generate_with_protoc
    fi

    # Verify and format
    if verify_generated; then
        format_generated

        echo ""
        echo -e "${GREEN}════════════════════════════════════${NC}"
        echo -e "${GREEN}✓ Protocol Buffer generation complete!${NC}"
        echo -e "${GREEN}════════════════════════════════════${NC}"
        echo ""
        echo "Generated files are located in:"
        echo "  ${PROTO_DIR}/acm/v1/"
        echo ""
        echo "To use in your Go code:"
        echo '  import "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"'
        echo ""
    else
        echo ""
        echo -e "${RED}✗ Generation completed with errors${NC}"
        exit 1
    fi
}

# Help message
if [[ "${1:-}" == "--help" ]] || [[ "${1:-}" == "-h" ]]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Generate Go code from Protocol Buffer definitions."
    echo ""
    echo "Options:"
    echo "  --no-clean    Don't clean previously generated files"
    echo "  --help, -h    Show this help message"
    echo ""
    echo "Requirements:"
    echo "  - buf (recommended): https://buf.build/docs/installation"
    echo "  - OR protoc + protoc-gen-go + protoc-gen-go-grpc"
    echo ""
    echo "Install buf:"
    echo "  macOS/Linux:  brew install bufbuild/buf/buf"
    echo "  Go install:   go install github.com/bufbuild/buf/cmd/buf@latest"
    echo ""
    exit 0
fi

# Run main function
main "$@"
