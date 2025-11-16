# ACM Build System Documentation

This document describes the build infrastructure for the Automated Compromise Mitigation (ACM) project.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Build System Overview](#build-system-overview)
- [Makefile Targets](#makefile-targets)
- [Build Scripts](#build-scripts)
- [Development Workflow](#development-workflow)
- [Release Process](#release-process)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

```bash
# 1. Set up development environment (installs tools)
make setup-dev

# 2. Generate mTLS certificates for local development
make cert-gen

# 3. Build both service and CLI
make build

# 4. Run tests with coverage
make test-coverage

# 5. Start the service in development mode
make dev
```

---

## Prerequisites

### Required

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation
- **OpenSSL** - Certificate generation

### Optional

- **golangci-lint** - Installed by `make setup-dev`
- **Protocol Buffers compiler (protoc)** - For gRPC code generation
- **GoReleaser** - For release builds
- **Docker** - For containerized builds (future)

### Installation

#### macOS (Homebrew)
```bash
brew install go openssl protobuf
```

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y golang-go openssl protobuf-compiler
```

#### Fedora/RHEL
```bash
sudo dnf install -y golang openssl protobuf-compiler
```

---

## Build System Overview

The ACM build system consists of:

1. **Makefile** - Primary build automation (root directory)
2. **Build Scripts** - Shell scripts in `scripts/` directory
3. **GoReleaser** - Multi-platform release automation
4. **golangci-lint** - Comprehensive linting and security scanning

### Architecture

```
automated-compromise-mitigation/
├── Makefile                      # Main build file
├── .goreleaser.yml               # Release configuration
├── .golangci.yml                 # Linter configuration
├── VERSION                       # Version number
├── tools.go                      # Development tool dependencies
│
├── scripts/
│   ├── build.sh                  # Cross-platform build script
│   ├── test.sh                   # Test runner with coverage
│   ├── setup-dev.sh              # Dev environment setup
│   └── generate-certs.sh         # mTLS certificate generation
│
├── cmd/
│   ├── acm-service/              # Service daemon
│   └── acm-cli/                  # CLI client
│
├── build/                        # Local builds (gitignored)
└── dist/                         # Release builds (gitignored)
```

---

## Makefile Targets

### Help and Information

```bash
make help          # Show all available targets
make version       # Show version information
make info          # Show build environment info
```

### Building

```bash
make build                  # Build both service and CLI
make build-service          # Build service daemon only
make build-cli              # Build CLI client only
make build-all-platforms    # Cross-compile for all platforms
make install                # Install binaries to $GOPATH/bin
make uninstall              # Remove installed binaries
```

### Testing

```bash
make test                   # Run unit tests
make test-coverage          # Run tests with coverage report
make test-integration       # Run integration tests
make test-security          # Run security-focused tests
make bench                  # Run benchmarks
```

### Code Quality

```bash
make lint                   # Run golangci-lint
make lint-fix               # Run linter with auto-fix
make fmt                    # Format code with gofmt
make fmt-check              # Check code formatting
make vet                    # Run go vet
```

### Security

```bash
make security-scan          # Run gosec security scanner
make vuln-check             # Check for known vulnerabilities
```

### Development

```bash
make setup-dev              # Set up development environment
make cert-gen               # Generate mTLS certificates
make dev                    # Start service in dev mode
make dev-cli                # Start CLI in dev mode
make proto-gen              # Generate code from protobuf
```

### Dependencies

```bash
make mod-download           # Download dependencies
make mod-tidy               # Tidy dependencies
make mod-verify             # Verify dependencies
make mod-vendor             # Vendor dependencies
make deps-update            # Update dependencies
make deps-graph             # Generate dependency graph
```

### Release

```bash
make release                # Create release build
make release-snapshot       # Create snapshot (no tags)
make clean                  # Remove build artifacts
```

### CI/CD

```bash
make ci                     # Run all CI checks
make pre-commit             # Run pre-commit checks
```

---

## Build Scripts

### `scripts/build.sh`

Cross-platform build script that compiles binaries for all supported platforms.

**Platforms:**
- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64

**Usage:**
```bash
./scripts/build.sh
```

**Features:**
- Parallel builds
- Version injection from VERSION file
- Git commit hash embedding
- Build date stamping
- Archive creation (tar.gz for Unix, zip for Windows)
- Checksum generation

### `scripts/test.sh`

Test runner with comprehensive coverage reporting.

**Usage:**
```bash
./scripts/test.sh

# With integration tests
RUN_INTEGRATION=true ./scripts/test.sh

# With security tests
RUN_SECURITY=true ./scripts/test.sh

# Custom coverage threshold
COVERAGE_THRESHOLD=85 ./scripts/test.sh
```

**Features:**
- Unit test execution
- Race condition detection
- Coverage reporting (HTML and text)
- Coverage threshold enforcement (default: 80%)
- Package-level coverage breakdown
- Low-coverage file identification

### `scripts/setup-dev.sh`

Development environment setup and tool installation.

**Usage:**
```bash
./scripts/setup-dev.sh
```

**What it installs:**
- golangci-lint
- gosec (security linter)
- govulncheck (vulnerability scanner)
- goimports (import formatter)
- goreleaser (release tool)
- protoc-gen-go (protobuf Go plugin)
- protoc-gen-go-grpc (gRPC Go plugin)
- mockgen (mock generator)
- gotestsum (test output formatter)

**What it sets up:**
- Git pre-commit hooks
- Development directory structure
- Example configuration files
- Development config templates

### `scripts/generate-certs.sh`

mTLS certificate generation for local development.

**Usage:**
```bash
./scripts/generate-certs.sh
```

**Generates:**
- CA certificate and key (self-signed)
- Server certificate and key (for acm-service)
- Client certificate and key (for acm-cli)

**Certificate Validity:**
- CA: 10 years
- Server/Client: 1 year

**Security Notice:** Generated certificates are for development only. Do not use in production.

---

## Development Workflow

### Initial Setup

```bash
# 1. Clone repository
git clone https://github.com/ferg-cod3s/automated-compromise-mitigation.git
cd automated-compromise-mitigation

# 2. Set up development environment
make setup-dev

# 3. Generate certificates
make cert-gen

# 4. Verify setup
make info
```

### Daily Development

```bash
# 1. Update dependencies
make mod-tidy

# 2. Make code changes
# ... edit files ...

# 3. Format code
make fmt

# 4. Run tests
make test

# 5. Lint code
make lint

# 6. Build binaries
make build

# 7. Test locally
make dev
```

### Pre-Commit Checklist

```bash
# Automated by Git hook, or run manually:
make pre-commit
```

This runs:
1. Code formatting check
2. Linting
3. Unit tests

### Testing Changes

```bash
# Unit tests
make test

# With coverage
make test-coverage

# Integration tests
make test-integration

# Security tests
make test-security

# All tests with coverage
make ci
```

---

## Release Process

### Versioning

ACM follows [Semantic Versioning](https://semver.org/):

- **Major (x.0.0)**: Breaking changes
- **Minor (0.x.0)**: New features, backwards compatible
- **Patch (0.0.x)**: Bug fixes, security patches

### Creating a Release

```bash
# 1. Update VERSION file
echo "1.0.0" > VERSION

# 2. Update CHANGELOG.md
# ... document changes ...

# 3. Commit changes
git add VERSION CHANGELOG.md
git commit -m "chore: release v1.0.0"

# 4. Tag release
git tag -a v1.0.0 -m "Release v1.0.0"

# 5. Build release
make release

# 6. Push to GitHub
git push origin main --tags
```

### Automated Release (with GoReleaser)

When a tag is pushed to GitHub, GoReleaser automatically:

1. Builds for all platforms
2. Creates archives
3. Generates checksums
4. Creates SBOM (Software Bill of Materials)
5. Publishes GitHub release
6. Uploads artifacts

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Setup Dev Environment
        run: make setup-dev

      - name: Run CI Checks
        run: make ci

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### CI Target

The `make ci` target runs:

```bash
make ci
```

Which executes:
1. Linting (golangci-lint)
2. Unit tests with coverage
3. Integration tests
4. Security scanning (gosec)
5. Vulnerability checking (govulncheck)

---

## Troubleshooting

### Common Issues

#### Issue: `golangci-lint not found`

**Solution:**
```bash
make setup-dev
```

Or install manually:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

#### Issue: `protoc not found`

**Solution:**

macOS:
```bash
brew install protobuf
```

Linux:
```bash
sudo apt-get install -y protobuf-compiler
```

#### Issue: `Cannot connect to service`

**Solution:**

1. Check certificates exist:
   ```bash
   ls -la certs/
   ```

2. Regenerate if missing:
   ```bash
   make cert-gen
   ```

3. Verify service is running:
   ```bash
   make dev
   ```

#### Issue: Tests failing with race conditions

**Solution:**

1. Run tests with race detector:
   ```bash
   go test -race ./...
   ```

2. Fix identified races
3. Re-run tests

#### Issue: Coverage below threshold

**Solution:**

1. Identify low-coverage files:
   ```bash
   make test-coverage
   ```

2. Add tests for uncovered code
3. Aim for >80% coverage

### Build Cache Issues

```bash
# Clear Go build cache
go clean -cache -testcache -modcache

# Rebuild
make clean build
```

### Permission Issues

```bash
# Ensure scripts are executable
chmod +x scripts/*.sh

# Fix certificate permissions
make cert-gen
```

---

## Advanced Topics

### Custom Build Flags

```bash
# Build with custom LDFLAGS
LDFLAGS="-X main.CustomVar=value" make build

# Build with specific tags
TAGS="debug,verbose" make build
```

### Vendoring Dependencies

```bash
# Create vendor directory
make mod-vendor

# Build using vendor
go build -mod=vendor ./...
```

### Cross-Compilation

```bash
# Build for specific platform
GOOS=linux GOARCH=arm64 make build

# Build for all platforms
make build-all-platforms
```

---

## Resources

- [Go Documentation](https://golang.org/doc/)
- [GoReleaser Docs](https://goreleaser.com/)
- [golangci-lint Docs](https://golangci-lint.run/)
- [ACM Technical Architecture](acm-tad.md)
- [ACM Security Planning](acm-security-planning.md)

---

## Support

For build system issues:

1. Check this documentation
2. Review [Troubleshooting](#troubleshooting) section
3. Check [GitHub Issues](https://github.com/ferg-cod3s/automated-compromise-mitigation/issues)
4. Open a new issue with:
   - Go version (`go version`)
   - OS and architecture
   - Error messages
   - Steps to reproduce

---

**Last Updated:** 2025-11-16
**Version:** 1.0
