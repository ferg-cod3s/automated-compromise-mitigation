# ACM gRPC API Protocol Buffers

This directory contains the Protocol Buffer definitions for the ACM (Automated Compromise Mitigation) gRPC API. These `.proto` files define the service interfaces, messages, and data types used for communication between the ACM service and its clients (OpenTUI and Tauri GUI).

## Overview

The ACM API is organized into modular services, each handling a specific aspect of credential management:

| Service | Proto File | Description |
|---------|-----------|-------------|
| **CredentialService** | `credential.proto` | Core credential detection, rotation, and management |
| **AuditService** | `audit.proto` | Cryptographically-signed audit logging and reporting |
| **HIMService** | `him.proto` | Human-in-the-Middle workflow orchestration |
| **Common Types** | `common.proto` | Shared data types, enums, and error handling |

## Directory Structure

```
api/proto/
├── README.md                 # This file
├── buf.yaml                  # Buf configuration (linting, breaking changes)
├── buf.gen.yaml              # Code generation configuration
├── acm/v1/
│   ├── common.proto          # Common types and enums
│   ├── credential.proto      # CredentialService definition
│   ├── audit.proto           # AuditService definition
│   ├── him.proto             # HIMService definition
│   └── *.pb.go               # Generated Go code (created by scripts/generate-proto.sh)
```

## Prerequisites

### Option 1: Using Buf (Recommended)

[Buf](https://buf.build) is a modern tool for working with Protocol Buffers that provides linting, breaking change detection, and code generation.

**Install Buf:**

```bash
# macOS/Linux
brew install bufbuild/buf/buf

# Or using Go
go install github.com/bufbuild/buf/cmd/buf@latest

# Verify installation
buf --version
```

### Option 2: Using protoc

If you prefer traditional `protoc`, you'll need:

**Install protoc:**

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt install protobuf-compiler

# Fedora
sudo dnf install protobuf-compiler

# Or download from: https://github.com/protocolbuffers/protobuf/releases
```

**Install Go plugins:**

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Generating Go Code

### Quick Start

```bash
# From project root
./scripts/generate-proto.sh
```

This script will:
1. Clean previously generated `.pb.go` files
2. Lint proto files (if using buf)
3. Generate Go code for all `.proto` files
4. Verify all expected files were created
5. Format the generated code

### Manual Generation with Buf

```bash
cd api/proto

# Lint proto files
buf lint

# Generate code
buf generate
```

### Manual Generation with protoc

```bash
cd api/proto

# Generate for each proto file
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  --go-grpc_opt=require_unimplemented_servers=true \
  --proto_path=. \
  acm/v1/*.proto
```

## Using the Generated Code

### Import in Go

```go
import (
    acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)
```

### Example: Creating a Client

```go
// Load mTLS credentials
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{clientCert},
    RootCAs:      caCertPool,
    MinVersion:   tls.VersionTLS13,
}

creds := credentials.NewTLS(tlsConfig)

// Connect to ACM service
conn, err := grpc.Dial(
    "localhost:8443",
    grpc.WithTransportCredentials(creds),
)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

// Create service clients
credClient := acmv1.NewCredentialServiceClient(conn)
auditClient := acmv1.NewAuditServiceClient(conn)
himClient := acmv1.NewHIMServiceClient(conn)

// Use the clients
ctx := context.Background()
resp, err := credClient.DetectCompromised(ctx, &acmv1.DetectRequest{
    Metadata: &acmv1.Metadata{
        RequestId:     "req-123",
        ClientVersion: "acm-tui/1.0.0",
    },
    PasswordManagerType: acmv1.PasswordManagerType_PASSWORD_MANAGER_TYPE_BITWARDEN,
})
if err != nil {
    log.Fatalf("DetectCompromised failed: %v", err)
}

fmt.Printf("Found %d compromised credentials\n", resp.TotalCount)
for _, cred := range resp.Credentials {
    fmt.Printf("  - %s (%s)\n", cred.Site, cred.Username)
}
```

### Example: Implementing a Server

```go
type credentialServer struct {
    acmv1.UnimplementedCredentialServiceServer
    // Your service implementation fields
}

func (s *credentialServer) DetectCompromised(
    ctx context.Context,
    req *acmv1.DetectRequest,
) (*acmv1.DetectResponse, error) {
    // Your implementation here
    return &acmv1.DetectResponse{
        Status: &acmv1.Status{
            Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
            Message: "Detection complete",
        },
        Credentials: []*acmv1.CompromisedCredential{
            // ... credentials
        },
        TotalCount: 5,
    }, nil
}

// Register server
grpcServer := grpc.NewServer(grpc.Creds(tlsCreds))
acmv1.RegisterCredentialServiceServer(grpcServer, &credentialServer{})
acmv1.RegisterAuditServiceServer(grpcServer, &auditServer{})
acmv1.RegisterHIMServiceServer(grpcServer, &himServer{})

// Start server
if err := grpcServer.Serve(listener); err != nil {
    log.Fatalf("Failed to serve: %v", err)
}
```

## API Design Principles

### 1. Zero-Knowledge Security

The API is designed to **never expose** master passwords or vault encryption keys:

- Credential IDs are **hashed** (SHA-256) before transmission
- Passwords are only returned after successful rotation (transmitted over mTLS)
- All sensitive fields are clearly documented

### 2. Comprehensive Error Handling

All responses include:

- `Status` with `StatusCode` enum for programmatic handling
- `Error` with detailed error codes and context
- Human-readable error messages

### 3. Audit Trail Integration

Every operation includes:

- `Metadata` with request ID for tracing
- Automatic audit log entries (handled by service)
- Cryptographic signatures for tamper-evidence

### 4. Extensibility for Future Phases

The API includes optional fields for Phase II+ features:

- `ComplianceValidation` in `RotateResponse` (for ACVS)
- `evidence_chain_id` in audit events
- Reserved fields for future extensions

## API Versioning

The API uses semantic versioning embedded in the package name:

- Current version: **v1** (`acm.v1`)
- Breaking changes will increment the version (e.g., `acm.v2`)
- Non-breaking additions can be made to v1

## Linting and Best Practices

The `buf.yaml` configuration enforces:

- **Naming conventions**: `snake_case` fields, `UPPER_SNAKE_CASE` enums
- **Enum zero values**: Must end with `_UNSPECIFIED`
- **Unique request/response types**: Each RPC has distinct message types
- **Standard naming**: Requests end with `Request`, responses with `Response`

Run linting:

```bash
cd api/proto
buf lint
```

## Breaking Change Detection

Buf can detect breaking changes between versions:

```bash
# Check against main branch
buf breaking --against '.git#branch=main'

# Check against a specific tag
buf breaking --against '.git#tag=v1.0.0'
```

## Service Documentation

### CredentialService

**Purpose:** Credential detection, rotation, and management

**RPCs:**
- `DetectCompromised`: Query password manager for breached credentials
- `RotateCredential`: Generate new password and update vault
- `GetRotationStatus`: Check status of rotation operation
- `ListCredentials`: Retrieve credential metadata (no passwords)
- `GeneratePassword`: Utility for secure password generation

### AuditService

**Purpose:** Cryptographically-signed audit logging

**RPCs:**
- `QueryLogs`: Search audit events with filters
- `VerifyIntegrity`: Verify cryptographic signatures on audit logs
- `ExportReport`: Generate compliance reports (PDF, JSON, CSV)
- `GetStatistics`: Aggregate statistics for dashboards

### HIMService

**Purpose:** Human-in-the-Middle workflow orchestration

**RPCs:**
- `PromptUser`: Bidirectional streaming for multi-step user interactions
- `GetHIMStatus`: Check status of active HIM workflows
- `CancelHIM`: Cancel a pending HIM request

## Common Types

The `common.proto` file defines shared types used across all services:

| Type | Purpose |
|------|---------|
| `Status` | Operation outcome with code and message |
| `Error` | Detailed error information with retry guidance |
| `Metadata` | Request tracing, client identification |
| `PasswordPolicy` | Password generation rules |
| `PaginationRequest/Response` | Pagination for list operations |

## Development Workflow

### Making Changes to Proto Files

1. **Edit** the `.proto` files in `acm/v1/`
2. **Lint** to catch issues: `buf lint`
3. **Check breaking changes**: `buf breaking --against '.git#branch=main'`
4. **Generate** new code: `./scripts/generate-proto.sh`
5. **Test** your changes in the service and client implementations
6. **Commit** both `.proto` files and generated `.pb.go` files

### Adding a New RPC

```protobuf
// In credential.proto
service CredentialService {
  // ... existing RPCs

  // New RPC: Describe what it does
  rpc YourNewRPC(YourRequest) returns (YourResponse);
}

message YourRequest {
  Metadata metadata = 1;
  // ... your fields
}

message YourResponse {
  Status status = 1;
  // ... your fields
  Error error = N;  // Always include error field
}
```

### Adding a New Service

1. Create `new_service.proto` in `acm/v1/`
2. Define the service interface and messages
3. Update `scripts/generate-proto.sh` to include new file in verification
4. Regenerate code
5. Implement the service in Go

## Troubleshooting

### "command not found: buf"

Install buf using instructions in Prerequisites section.

### "command not found: protoc-gen-go"

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure $GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### "cannot find package"

Ensure generated code is in the correct location and your Go module path is correct:

```bash
# Check go.mod module name
go mod edit -module=github.com/ferg-cod3s/automated-compromise-mitigation

# Regenerate code
./scripts/generate-proto.sh
```

### Linting Errors

Common linting issues:

```
ENUM_ZERO_VALUE_SUFFIX: Enum zero values must end with _UNSPECIFIED
→ Fix: Change `STATUS_UNKNOWN = 0` to `STATUS_UNSPECIFIED = 0`

FIELD_LOWER_SNAKE_CASE: Field names must be snake_case
→ Fix: Change `userId` to `user_id`

RPC_REQUEST_STANDARD_NAME: Request message must end with Request
→ Fix: Change `DetectParams` to `DetectRequest`
```

## References

- **Protocol Buffers Language Guide:** https://protobuf.dev/programming-guides/proto3/
- **gRPC Go Tutorial:** https://grpc.io/docs/languages/go/quickstart/
- **Buf Documentation:** https://buf.build/docs/
- **ACM Technical Architecture Document:** See `acm-tad.md` in project root

## License

Copyright 2025 ACM Project
SPDX-License-Identifier: Apache-2.0

---

**Questions or Issues?**

- File an issue on GitHub
- Consult the Technical Architecture Document (`acm-tad.md`)
- See the main project README for contribution guidelines
