# ACM Phase 1 Code Structure

**Version:** 1.0
**Date:** November 2025
**Status:** Initial Setup
**Phase:** Phase I - Core Service + CRS

---

## Table of Contents

1. [Overview](#overview)
2. [Directory Structure](#directory-structure)
3. [Package Organization](#package-organization)
4. [Key Interfaces](#key-interfaces)
5. [Build and Development](#build-and-development)
6. [Next Steps](#next-steps)

---

## Overview

This document describes the Go project structure for the ACM (Automated Compromise Mitigation) Phase 1 implementation. The structure follows Go best practices and the architecture defined in the Technical Architecture Document (TAD).

### Project Goals

Phase 1 focuses on:
- Core ACM service daemon with gRPC API
- Credential Remediation Service (CRS)
- 1Password CLI integration (primary)
- Bitwarden CLI integration (secondary)
- Audit logging with cryptographic signatures
- OpenTUI client for terminal interaction

### Design Principles

- **Zero-Knowledge Security**: Never access master passwords or vault keys
- **Local-First**: All operations on user's device
- **Service-Client Separation**: Business logic in service, thin clients
- **Defense in Depth**: Multiple security layers (mTLS, JWT, encryption)
- **Clear Interfaces**: Well-defined contracts between components

---

## Directory Structure

```
automated-compromise-mitigation/
├── cmd/                          # Executable entry points
│   ├── acm-service/              # ACM service daemon
│   │   └── main.go               # Service entry point
│   └── acm-cli/                  # OpenTUI client
│       └── main.go               # CLI entry point
│
├── internal/                     # Internal packages (not importable)
│   ├── crs/                      # Credential Remediation Service
│   │   ├── doc.go                # Package documentation
│   │   └── interface.go          # CRS interface definitions
│   │
│   ├── him/                      # Human-in-the-Middle Manager
│   │   ├── doc.go                # Package documentation
│   │   └── interface.go          # HIM interface definitions
│   │
│   ├── audit/                    # Audit Logger
│   │   ├── doc.go                # Package documentation
│   │   └── interface.go          # Audit interface definitions
│   │
│   ├── auth/                     # Authentication (mTLS, JWT)
│   │   └── doc.go                # Package documentation
│   │
│   └── pwmanager/                # Password Manager CLI integrations
│       ├── doc.go                # Package documentation
│       ├── interface.go          # PasswordManager interface
│       ├── onepassword/          # 1Password CLI implementation
│       │   └── doc.go            # Implementation documentation
│       └── bitwarden/            # Bitwarden CLI implementation
│           └── doc.go            # Implementation documentation
│
├── api/                          # API definitions
│   └── proto/                    # gRPC protobuf definitions
│       └── acm/
│           └── v1/               # Version 1 API definitions
│
├── pkg/                          # Public libraries (if needed)
│
├── configs/                      # Configuration file examples
│
├── scripts/                      # Build and setup scripts
│
├── test/                         # Test files
│   ├── integration/              # Integration tests
│   └── testdata/                 # Test fixtures and data
│
├── docs/                         # Documentation
│   └── phase1/                   # Phase 1 specific docs
│       └── CODE_STRUCTURE.md     # This file
│
├── go.mod                        # Go module definition
└── go.sum                        # Go module checksums (generated)
```

---

## Package Organization

### cmd/

Contains executable entry points. Each subdirectory under `cmd/` produces a binary.

#### cmd/acm-service/

The main ACM service daemon. This is the core business logic server that:
- Exposes gRPC API over mTLS on localhost:8443
- Orchestrates CRS, ACVS, HIM Manager, and Audit Logger
- Manages password manager CLI interactions
- Provides health check endpoints

**TODO in Phase 1:**
- Configuration loading from `~/.acm/config/service.yaml`
- mTLS server setup with certificate validation
- gRPC service registration
- Graceful shutdown handling
- Logging infrastructure

#### cmd/acm-cli/

The OpenTUI client application. This provides:
- Command-line interface for ACM operations
- Beautiful terminal UI using Bubbletea
- gRPC client with mTLS authentication
- Interactive and scriptable modes

**TODO in Phase 1:**
- Command parsing and routing
- gRPC client initialization
- TUI implementation with Bubbletea
- Configuration loading
- Certificate management

### internal/

Contains internal packages that are not importable by external projects. This enforces proper API boundaries.

#### internal/crs/

**Credential Remediation Service (CRS)** - Core module for credential rotation.

**Key Responsibilities:**
- Detect compromised credentials via password manager CLI
- Generate secure passwords using crypto/rand
- Update vault entries atomically
- Verify rotation success
- Maintain rotation history

**Key Interface:**
```go
type CredentialRemediationService interface {
    DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error)
    GeneratePassword(ctx context.Context, policy pwmanager.PasswordPolicy) (string, error)
    RotateCredential(ctx context.Context, cred pwmanager.CompromisedCredential, newPassword string) (*RotationResult, error)
    VerifyRotation(ctx context.Context, credentialID string) (bool, error)
    GetRotationHistory(ctx context.Context, credentialID string) ([]RotationEvent, error)
}
```

**TODO in Phase 1:**
- Implement CRS service struct
- Password generation with policy enforcement
- CLI subprocess execution and parsing
- Rotation workflow orchestration
- Integration with audit logger

#### internal/him/

**Human-in-the-Middle (HIM) Manager** - Orchestrates user intervention workflows.

**Key Responsibilities:**
- Determine when HIM is required (MFA, CAPTCHA, ToS)
- Manage HIM session state machine
- Prompt users via gRPC streaming
- Handle timeouts and retries
- Resume automation after user input

**Key Interface:**
```go
type HIMManager interface {
    RequiresHIM(ctx context.Context, action RotationAction) (bool, HIMType, error)
    PromptUser(ctx context.Context, prompt HIMPrompt) (*HIMResponse, error)
    ResumeAutomation(ctx context.Context, sessionID string, response *HIMResponse) error
    GetSessionState(ctx context.Context, sessionID string) (*HIMSessionState, error)
    CancelSession(ctx context.Context, sessionID string) error
    ListActiveSessions(ctx context.Context) ([]*HIMSessionState, error)
}
```

**TODO in Phase 1:**
- Implement HIM state machine
- Session storage (in-memory or SQLite)
- gRPC streaming for prompts/responses
- Timeout handling
- Basic MFA (TOTP) support

#### internal/audit/

**Audit Logger** - Cryptographically signed audit trail.

**Key Responsibilities:**
- Log all rotation events with Ed25519 signatures
- Store events in SQLite database
- Query and filter audit log
- Verify integrity of log entries
- Export compliance reports (JSON, CSV, PDF)

**Key Interface:**
```go
type AuditLogger interface {
    Log(ctx context.Context, event AuditEvent) (eventID string, signature string, error error)
    Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
    ExportReport(ctx context.Context, format ReportFormat, filter AuditFilter) (io.Reader, error)
    VerifyIntegrity(ctx context.Context, from, to time.Time) (valid bool, errors []string, err error)
    GetStatistics(ctx context.Context, filter AuditFilter) (*AuditStatistics, error)
    Cleanup(ctx context.Context, retentionDays int) (int, error)
}
```

**TODO in Phase 1:**
- SQLite database schema creation
- Ed25519 signing implementation
- Event logging and querying
- JSON/CSV export
- Integrity verification

#### internal/auth/

**Authentication and Authorization** - mTLS and JWT management.

**Key Responsibilities:**
- Generate and manage X.509 certificates
- Configure mTLS server and client
- Issue and validate JWT tokens
- Store certificates in OS keychain
- Handle certificate renewal

**TODO in Phase 1:**
- Self-signed CA generation
- Client certificate generation
- mTLS server configuration
- mTLS client configuration
- JWT token issuance and validation
- OS keychain integration (macOS, Linux)

#### internal/pwmanager/

**Password Manager Integrations** - Zero-knowledge CLI interactions.

**Key Responsibilities:**
- Define PasswordManager interface
- Implement 1Password CLI integration
- Implement Bitwarden CLI integration
- Auto-detect available password managers
- Handle CLI subprocess execution
- Parse structured output (JSON)

**Key Interface:**
```go
type PasswordManager interface {
    DetectCompromised(ctx context.Context) ([]CompromisedCredential, error)
    GetCredential(ctx context.Context, id string) (*Credential, error)
    UpdatePassword(ctx context.Context, id string, newPassword string) error
    VerifyUpdate(ctx context.Context, id string, expectedModifiedAfter time.Time) (bool, error)
    IsAvailable(ctx context.Context) (bool, error)
    IsVaultLocked(ctx context.Context) (bool, error)
    Type() string
}
```

**TODO in Phase 1:**
- Implement 1Password CLI integration
- Implement Bitwarden CLI integration
- CLI detection and validation
- Subprocess execution with timeouts
- JSON parsing and error handling
- Memory protection for sensitive data

### api/proto/

Contains Protocol Buffer definitions for gRPC API.

**TODO in Phase 1:**
- Define `CredentialService` (detect, rotate, status)
- Define `AuditService` (query logs, verify integrity)
- Define `HIMService` (streaming prompts/responses)
- Define message types (DetectRequest, RotateResponse, etc.)
- Generate Go code with `protoc`

### configs/

Example configuration files for service and client.

**TODO in Phase 1:**
- `service.yaml.example` - Service configuration template
- `client.yaml.example` - Client configuration template
- Documentation on configuration options

### scripts/

Build and setup automation scripts.

**TODO in Phase 1:**
- `setup.sh` - Initialize ACM (generate certs, create dirs, etc.)
- `build.sh` - Build all binaries
- `test.sh` - Run tests
- `install.sh` - Install binaries and configs

### test/

Test files and fixtures.

**TODO in Phase 1:**
- Unit tests for each package
- Integration tests for CRS workflow
- Test fixtures (mock CLI output, certs, etc.)
- E2E tests for service + client

---

## Key Interfaces

### CredentialRemediationService (CRS)

Located in `internal/crs/interface.go`. Core service for credential rotation.

**Methods:**
- `DetectCompromised()` - Query password manager for breached credentials
- `GeneratePassword()` - Create secure password with policy enforcement
- `RotateCredential()` - Perform complete rotation workflow
- `VerifyRotation()` - Confirm successful password update
- `GetRotationHistory()` - Retrieve rotation audit trail

**Key Types:**
- `RotationResult` - Outcome of rotation operation
- `RotationStatus` - Success, failure, HIM required, etc.
- `RotationError` - Structured error with retry information
- `RotationEvent` - Historical rotation record

### HIMManager

Located in `internal/him/interface.go`. Manages user intervention workflows.

**Methods:**
- `RequiresHIM()` - Determine if action needs user input
- `PromptUser()` - Send prompt and wait for response (blocking)
- `ResumeAutomation()` - Continue after user input
- `GetSessionState()` - Query session status
- `CancelSession()` - Cancel in-progress session
- `ListActiveSessions()` - Get all active HIM sessions

**Key Types:**
- `HIMPrompt` - Prompt sent to user
- `HIMResponse` - User's response
- `HIMSessionState` - Current session state
- `SessionState` - State machine states (created, awaiting_input, etc.)
- `HIMType` - Type of intervention (MFA, CAPTCHA, manual, etc.)

### AuditLogger

Located in `internal/audit/interface.go`. Tamper-evident audit trail.

**Methods:**
- `Log()` - Record signed audit event
- `Query()` - Retrieve events matching filter
- `ExportReport()` - Generate compliance report
- `VerifyIntegrity()` - Validate signatures
- `GetStatistics()` - Aggregate statistics
- `Cleanup()` - Remove old entries

**Key Types:**
- `AuditEvent` - Signed log entry
- `AuditFilter` - Query criteria (time range, event types, etc.)
- `ReportFormat` - JSON, PDF, CSV, HTML
- `AuditStatistics` - Aggregate metrics

### PasswordManager

Located in `internal/pwmanager/interface.go`. Abstract password manager CLI.

**Methods:**
- `DetectCompromised()` - Find breached credentials
- `GetCredential()` - Retrieve credential metadata
- `UpdatePassword()` - Change password in vault
- `VerifyUpdate()` - Confirm update success
- `IsAvailable()` - Check if CLI is installed
- `IsVaultLocked()` - Check vault lock status
- `Type()` - Return manager type ("1password", "bitwarden")

**Key Types:**
- `CompromisedCredential` - Breached credential info
- `Credential` - Credential metadata (no password)
- `PasswordPolicy` - Password generation constraints
- `PasswordManagerError` - Structured error

---

## Build and Development

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins
- 1Password CLI (`op`) or Bitwarden CLI (`bw`)

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/ferg-cod3s/automated-compromise-mitigation.git
cd automated-compromise-mitigation

# Install Go dependencies
go mod download

# Install protobuf compiler plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installation
go version
protoc --version
```

### Build Binaries

```bash
# Build all binaries
go build -o bin/acm-service ./cmd/acm-service
go build -o bin/acm ./cmd/acm-cli

# Or use build script (TODO: create script)
./scripts/build.sh
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/crs/...

# Run integration tests
go test ./test/integration/...
```

### Generate Protobuf Code

```bash
# Generate Go code from .proto files
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/acm/v1/*.proto
```

### Code Style

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run security checks
gosec ./...
```

---

## Next Steps

### Phase 1 Implementation Priority

1. **Week 1-2: Core Infrastructure**
   - Complete `internal/auth` implementation (mTLS, JWT)
   - Implement basic `internal/audit` (SQLite, signing)
   - Create protobuf API definitions

2. **Week 3-4: Password Manager Integration**
   - Implement `internal/pwmanager/onepassword`
   - Implement `internal/pwmanager/bitwarden`
   - Add CLI detection and subprocess execution

3. **Week 5-6: CRS Implementation**
   - Implement `internal/crs` core logic
   - Password generation with crypto/rand
   - Rotation workflow orchestration
   - Integration with audit logger

4. **Week 7-8: HIM Manager**
   - Implement `internal/him` state machine
   - gRPC streaming for prompts
   - Basic MFA (TOTP) support

5. **Week 9-10: Service Daemon**
   - Complete `cmd/acm-service/main.go`
   - gRPC server setup with mTLS
   - Service registration and routing
   - Configuration loading

6. **Week 11-12: CLI Client**
   - Complete `cmd/acm-cli/main.go`
   - Bubbletea TUI implementation
   - gRPC client with mTLS
   - Interactive rotation workflow

7. **Week 13-14: Testing and Documentation**
   - Unit tests for all packages
   - Integration tests
   - E2E tests
   - User documentation

8. **Week 15-16: Polish and Security Audit**
   - Code review and refactoring
   - Security testing (SAST, DAST)
   - Performance optimization
   - Prepare for Phase I release

### Future Phases

- **Phase II (Months 7-12):** ACVS implementation, Legal NLP, multi-manager support
- **Phase III (Months 13-18):** Enhanced automation, Tauri GUI
- **Phase IV (Months 19-24):** Enterprise features, mobile apps

---

## Dependencies

### Direct Dependencies

Listed in `go.mod`:

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `modernc.org/sqlite` - Pure Go SQLite driver
- `github.com/golang-jwt/jwt/v5` - JWT tokens
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling

### Development Dependencies

- `golangci-lint` - Comprehensive Go linter
- `gosec` - Security scanner
- `protoc` - Protocol Buffers compiler
- `protoc-gen-go` - Go code generator for protobuf
- `protoc-gen-go-grpc` - gRPC code generator

---

## Architecture Decision Records (ADRs)

Future ADRs will be documented in `docs/architecture/decisions/`:

- ADR-001: Why pure Go SQLite instead of CGo-based?
- ADR-002: Ed25519 vs RSA for audit signatures
- ADR-003: gRPC streaming vs polling for HIM prompts
- ADR-004: In-memory vs SQLite for HIM session state
- ADR-005: Password generation algorithm selection

---

## References

- [ACM Technical Architecture Document](../../acm-tad.md)
- [ACM Product Requirements Document](../../acm-prd.md)
- [ACM Security Planning](../../acm-security-planning.md)
- [ACM Governance Roadmap](../../acm-governance-roadmap.md)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-16 | Initial code structure documentation for Phase 1 |

---

**Status:** Initial Setup - Ready for Phase 1 Implementation
**Next Review:** After Week 2 of Phase 1 Development
