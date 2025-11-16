#!/usr/bin/env bash
# Developer environment setup script for ACM
# Installs required development tools and dependencies

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
    else
        OS="unknown"
    fi

    log_info "Detected OS: $OS"
}

# Check Go installation
check_go() {
    log_info "Checking Go installation..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        log_info "Please install Go 1.21 or later from: https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_success "Go $GO_VERSION is installed"

    # Check minimum version (1.21)
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        log_error "Go $REQUIRED_VERSION or later is required (found $GO_VERSION)"
        exit 1
    fi

    # Ensure GOPATH is set
    if [ -z "${GOPATH:-}" ]; then
        GOPATH="$HOME/go"
        log_warn "GOPATH not set, using default: $GOPATH"
        export GOPATH
    fi

    log_info "GOPATH: $GOPATH"

    # Ensure GOPATH/bin is in PATH
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        log_warn "Add $GOPATH/bin to your PATH"
        echo "  export PATH=\"\$GOPATH/bin:\$PATH\""
    fi
}

# Install Go development tools
install_go_tools() {
    log_info "Installing Go development tools..."

    # golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        log_info "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOPATH/bin"
        log_success "Installed golangci-lint"
    else
        log_success "golangci-lint already installed ($(golangci-lint version | head -n1))"
    fi

    # gosec - Security linter
    if ! command -v gosec &> /dev/null; then
        log_info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
        log_success "Installed gosec"
    else
        log_success "gosec already installed"
    fi

    # govulncheck - Vulnerability checker
    if ! command -v govulncheck &> /dev/null; then
        log_info "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
        log_success "Installed govulncheck"
    else
        log_success "govulncheck already installed"
    fi

    # goimports - Import formatter
    if ! command -v goimports &> /dev/null; then
        log_info "Installing goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
        log_success "Installed goimports"
    else
        log_success "goimports already installed"
    fi

    # GoReleaser
    if ! command -v goreleaser &> /dev/null; then
        log_info "Installing goreleaser..."
        if [ "$OS" = "macos" ]; then
            if command -v brew &> /dev/null; then
                brew install goreleaser
            else
                go install github.com/goreleaser/goreleaser@latest
            fi
        else
            go install github.com/goreleaser/goreleaser@latest
        fi
        log_success "Installed goreleaser"
    else
        log_success "goreleaser already installed ($(goreleaser --version | head -n1))"
    fi
}

# Install Protocol Buffers tools
install_protobuf() {
    log_info "Checking Protocol Buffers installation..."

    # Check protoc
    if ! command -v protoc &> /dev/null; then
        log_warn "protoc not found"
        log_info "Please install Protocol Buffers compiler:"
        if [ "$OS" = "macos" ]; then
            echo "  brew install protobuf"
        elif [ "$OS" = "linux" ]; then
            echo "  # Ubuntu/Debian:"
            echo "  sudo apt-get install -y protobuf-compiler"
            echo "  # Fedora/RHEL:"
            echo "  sudo dnf install -y protobuf-compiler"
        fi
    else
        PROTOC_VERSION=$(protoc --version | awk '{print $2}')
        log_success "protoc $PROTOC_VERSION is installed"
    fi

    # Install Go protobuf plugins
    log_info "Installing Go protobuf plugins..."

    # protoc-gen-go
    if ! command -v protoc-gen-go &> /dev/null; then
        log_info "Installing protoc-gen-go..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        log_success "Installed protoc-gen-go"
    else
        log_success "protoc-gen-go already installed"
    fi

    # protoc-gen-go-grpc
    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        log_info "Installing protoc-gen-go-grpc..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
        log_success "Installed protoc-gen-go-grpc"
    else
        log_success "protoc-gen-go-grpc already installed"
    fi
}

# Install additional utilities
install_utilities() {
    log_info "Installing additional utilities..."

    # mockgen - Generate mocks for testing
    if ! command -v mockgen &> /dev/null; then
        log_info "Installing mockgen..."
        go install go.uber.org/mock/mockgen@latest
        log_success "Installed mockgen"
    else
        log_success "mockgen already installed"
    fi

    # gotestsum - Better test output
    if ! command -v gotestsum &> /dev/null; then
        log_info "Installing gotestsum..."
        go install gotest.tools/gotestsum@latest
        log_success "Installed gotestsum"
    else
        log_success "gotestsum already installed"
    fi
}

# Set up Git hooks
setup_git_hooks() {
    log_info "Setting up Git hooks..."

    cd "$ROOT_DIR"

    if [ ! -d ".git" ]; then
        log_warn "Not a git repository, skipping Git hooks setup"
        return
    fi

    # Create pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/usr/bin/env bash
# ACM pre-commit hook
# Runs formatting and linting checks before commit

set -e

echo "Running pre-commit checks..."

# Format check
echo "  - Checking formatting..."
if ! make fmt-check; then
    echo "ERROR: Code is not formatted. Run 'make fmt' to fix."
    exit 1
fi

# Lint check
echo "  - Running linters..."
if ! make lint; then
    echo "ERROR: Linting failed. Fix the issues above."
    exit 1
fi

echo "Pre-commit checks passed!"
EOF

    chmod +x .git/hooks/pre-commit
    log_success "Git pre-commit hook installed"
}

# Create development directories
create_dev_directories() {
    log_info "Creating development directories..."

    cd "$ROOT_DIR"

    # Create directory structure
    mkdir -p cmd/acm-service
    mkdir -p cmd/acm-cli
    mkdir -p internal/service
    mkdir -p internal/client
    mkdir -p internal/crs
    mkdir -p internal/acvs
    mkdir -p internal/auth
    mkdir -p internal/storage
    mkdir -p internal/logging
    mkdir -p internal/config
    mkdir -p api/proto
    mkdir -p pkg/models
    mkdir -p pkg/utils
    mkdir -p tests/unit
    mkdir -p tests/integration
    mkdir -p tests/security
    mkdir -p config
    mkdir -p certs
    mkdir -p docs/development
    mkdir -p build
    mkdir -p dist

    log_success "Created development directory structure"
}

# Create example configuration files
create_config_files() {
    log_info "Creating example configuration files..."

    cd "$ROOT_DIR"

    # Create example service config
    if [ ! -f "config/acm.example.yaml" ]; then
        cat > config/acm.example.yaml << 'EOF'
# ACM Configuration Example
# Copy to config/acm.yaml and customize

service:
  # Service listening address
  address: "localhost:50051"

  # mTLS configuration
  tls:
    enabled: true
    cert_file: "certs/server-cert.pem"
    key_file: "certs/server-key.pem"
    ca_file: "certs/ca-cert.pem"
    client_auth: true

  # Logging configuration
  logging:
    level: "info"  # debug, info, warn, error
    format: "json"  # json, text
    output: "stdout"  # stdout, file path

# CRS (Credential Remediation Service) configuration
crs:
  enabled: true

  # Password manager configuration
  password_manager:
    type: "1password"  # 1password, bitwarden, pass
    cli_path: "/usr/local/bin/op"

  # Audit logging
  audit:
    enabled: true
    db_path: "data/audit.db"

# ACVS (Automated Compliance Validation Service) configuration
acvs:
  enabled: false  # Disabled by default - explicit opt-in required

  # ToS compliance configuration
  compliance:
    update_interval: "24h"
    rules_path: "data/compliance-rules"

# Security settings
security:
  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_minute: 60

  # Session management
  session:
    timeout: "15m"
    max_concurrent: 5
EOF
        log_success "Created config/acm.example.yaml"
    fi

    # Create development config
    if [ ! -f "config/dev.yaml" ]; then
        cat > config/dev.yaml << 'EOF'
# ACM Development Configuration

service:
  address: "localhost:50051"

  tls:
    enabled: true
    cert_file: "certs/server-cert.pem"
    key_file: "certs/server-key.pem"
    ca_file: "certs/ca-cert.pem"
    client_auth: true

  logging:
    level: "debug"
    format: "text"
    output: "stdout"

crs:
  enabled: true
  password_manager:
    type: "1password"
    cli_path: "/usr/local/bin/op"
  audit:
    enabled: true
    db_path: "data/dev-audit.db"

acvs:
  enabled: false

security:
  rate_limit:
    enabled: false
  session:
    timeout: "60m"
    max_concurrent: 10
EOF
        log_success "Created config/dev.yaml"
    fi
}

# Install system dependencies (optional)
install_system_deps() {
    log_info "Checking system dependencies..."

    # OpenSSL (for certificate generation)
    if ! command -v openssl &> /dev/null; then
        log_warn "OpenSSL not found (required for certificate generation)"
        if [ "$OS" = "macos" ]; then
            log_info "Install with: brew install openssl"
        elif [ "$OS" = "linux" ]; then
            log_info "Install with: sudo apt-get install openssl (Ubuntu/Debian)"
        fi
    else
        log_success "OpenSSL is installed"
    fi

    # SQLite (for audit logs)
    if ! command -v sqlite3 &> /dev/null; then
        log_warn "SQLite3 not found (required for audit logs)"
        if [ "$OS" = "macos" ]; then
            log_info "Install with: brew install sqlite"
        elif [ "$OS" = "linux" ]; then
            log_info "Install with: sudo apt-get install sqlite3 (Ubuntu/Debian)"
        fi
    else
        log_success "SQLite3 is installed"
    fi
}

# Print summary
print_summary() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║              ACM Development Environment                   ║"
    echo "╠════════════════════════════════════════════════════════════╣"
    echo "║  Status: READY                                             ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    log_success "Development environment setup complete!"
    echo ""
    echo "Next steps:"
    echo "  1. Review configuration: config/acm.example.yaml"
    echo "  2. Generate certificates: make cert-gen"
    echo "  3. Build binaries: make build"
    echo "  4. Run tests: make test"
    echo "  5. Start development: make dev"
    echo ""
    echo "For more targets: make help"
    echo ""
}

# Main execution
main() {
    log_info "ACM Development Environment Setup"
    echo ""

    detect_os
    check_go
    install_go_tools
    install_protobuf
    install_utilities
    create_dev_directories
    create_config_files
    setup_git_hooks
    install_system_deps
    print_summary
}

# Run main function
main "$@"
