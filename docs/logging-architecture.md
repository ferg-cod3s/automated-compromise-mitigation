# ACM Logging Architecture

**Version:** 1.0
**Date:** November 2025
**Status:** Phase III Implementation
**Authors:** ACM Development Team

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Design Goals](#design-goals)
3. [Log Schema](#log-schema)
4. [Log Levels](#log-levels)
5. [Component Taxonomy](#component-taxonomy)
6. [Correlation & Tracing](#correlation--tracing)
7. [Sensitive Data Policy](#sensitive-data-policy)
8. [Output Strategies](#output-strategies)
9. [Performance Considerations](#performance-considerations)
10. [Integration Architecture](#integration-architecture)

---

## Executive Summary

This document defines the logging architecture for the Automated Compromise Mitigation (ACM) system. ACM implements **structured logging** using Go's standard library `log/slog`, providing operational visibility, debugging capabilities, and compliance audit trails across all services.

### Key Principles

1. **Structured over Unstructured** - All logs are JSON-structured in production
2. **Context Propagation** - Request IDs trace operations end-to-end
3. **Security First** - Sensitive data automatically redacted
4. **Performance Conscious** - <1% CPU overhead requirement
5. **Operational Focus** - Logs enable debugging and monitoring

### Architecture Highlights

- **Library:** Go `log/slog` (standard library, Go 1.21+)
- **Formats:** JSON (production), Pretty (development)
- **Outputs:** stdout, file, or both
- **Rotation:** 100MB max size, 30 day retention, compression
- **Correlation:** UUID v7 request IDs for tracing
- **Redaction:** Automatic token/password/PII redaction

---

## Design Goals

### Primary Goals

1. **Operational Visibility**
   - Trace credential rotations from start to completion
   - Monitor ACVS compliance validations
   - Track password manager interactions
   - Debug gRPC service calls

2. **Security Compliance**
   - Never log master passwords or vault keys
   - Redact authentication tokens
   - Hash credential IDs before logging
   - Prevent log injection attacks

3. **Performance**
   - <100μs log call overhead
   - <1% total CPU impact
   - Asynchronous file writes
   - Minimal memory allocations

4. **Developer Experience**
   - Easy to add logging to new code
   - Contextual logging with request IDs
   - Pretty output for local development
   - Clear log message conventions

### Non-Goals

- **Not a metrics system** - Use Prometheus for metrics
- **Not distributed tracing** - Use OpenTelemetry for tracing (future)
- **Not log aggregation** - Use Loki/ELK for centralization (future)
- **Not real-time streaming** - Logs are file/stdout based

---

## Log Schema

### Standard Log Entry Structure

Every log entry follows this schema:

```json
{
  "timestamp": "2025-11-17T14:32:45.123456Z",
  "level": "info",
  "component": "rotation.github",
  "request_id": "01HBAG5E7V9QZXN8J4WTRGP0QX",
  "message": "GitHub rotation started",

  "operation": "github_rotation_start",
  "duration_ms": 156,

  "credential_id_hash": "sha256:a1b2c3d4e5f6...",
  "site": "github.com",
  "username": "testuser",

  "error": "token validation failed: invalid credentials",
  "error_code": "AUTH_FAILED",
  "stack_trace": "goroutine 42 [running]:\n...",

  "service": "acm",
  "version": "0.3.0",
  "hostname": "acm-prod-01",
  "pid": 12345
}
```

### Required Fields

Every log entry MUST include:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | RFC3339Nano | UTC timestamp | `2025-11-17T14:32:45.123456Z` |
| `level` | string | Log level | `debug`, `info`, `warn`, `error` |
| `component` | string | Component tag | `rotation.github` |
| `message` | string | Human-readable message | `GitHub rotation started` |
| `service` | string | Service name | `acm` |
| `version` | string | Service version | `0.3.0` |

### Optional Fields

Context-dependent fields:

| Field | Type | Description | When to Include |
|-------|------|-------------|-----------------|
| `request_id` | string | Request correlation ID | All gRPC requests |
| `operation` | string | Operation name | Timed operations |
| `duration_ms` | int64 | Operation duration | After operation completes |
| `error` | string | Error message | When operation fails |
| `error_code` | string | Machine-readable error code | When operation fails |
| `stack_trace` | string | Stack trace | Panics and critical errors |
| `credential_id_hash` | string | SHA-256 of credential ID | Credential operations |
| `site` | string | Target site domain | Rotation/ACVS operations |
| `username` | string | Username (not sensitive) | User-specific operations |
| `hostname` | string | Server hostname | Multi-instance deployments |
| `pid` | int | Process ID | Debugging |

### Field Naming Conventions

- **Snake case** for all field names: `request_id`, `credential_id_hash`
- **Suffix units** for durations: `duration_ms`, `timeout_secs`
- **Prefix hashes** with algorithm: `sha256:...`, `md5:...`
- **Use standard names**: `error`, `message`, `timestamp`

---

## Log Levels

### Level Definitions

ACM uses four log levels:

#### DEBUG
- **Purpose:** Detailed debugging information
- **Production:** Disabled by default
- **Examples:**
  - Function entry/exit
  - Variable values
  - Internal state transitions
  - Fine-grained flow control

```json
{
  "level": "debug",
  "message": "validating current GitHub token",
  "token_prefix": "ghp_abc12345...",
  "token_length": 40
}
```

#### INFO
- **Purpose:** Normal operational messages
- **Production:** Default level
- **Examples:**
  - Service startup/shutdown
  - Rotation started/completed
  - ACVS validation results
  - Configuration changes

```json
{
  "level": "info",
  "message": "GitHub rotation started",
  "credential_id_hash": "sha256:a1b2c3...",
  "site": "github.com"
}
```

#### WARN
- **Purpose:** Warning conditions (degraded but functional)
- **Production:** Always logged
- **Examples:**
  - Retry attempts
  - Fallback to alternative method
  - Slow operations (>100ms)
  - Certificate expiration warnings

```json
{
  "level": "warn",
  "message": "GitHub API rate limit approaching",
  "remaining": 50,
  "limit": 5000,
  "reset_at": "2025-11-17T15:00:00Z"
}
```

#### ERROR
- **Purpose:** Error conditions requiring attention
- **Production:** Always logged
- **Examples:**
  - Operation failures
  - External API errors
  - Database connection failures
  - Authentication errors

```json
{
  "level": "error",
  "message": "GitHub rotation failed",
  "error": "token validation failed: 401 Unauthorized",
  "error_code": "AUTH_FAILED",
  "credential_id_hash": "sha256:a1b2c3..."
}
```

### Level Selection Guidelines

**Use DEBUG when:**
- Tracing execution flow for debugging
- Logging variable values
- Entry/exit of internal functions
- Not needed in production

**Use INFO when:**
- Recording important business events
- User-initiated actions
- Service lifecycle events
- Normal operational milestones

**Use WARN when:**
- Degraded performance detected
- Approaching resource limits
- Retrying after transient failure
- Using fallback mechanisms

**Use ERROR when:**
- Operation failed and cannot continue
- External dependency unavailable
- Data validation failed
- User action blocked

---

## Component Taxonomy

### Component Hierarchy

Components are hierarchical with dot notation:

```
acm                          Root service
├── server                   gRPC server
│   ├── server.credential    Credential service handlers
│   ├── server.rotation      Rotation service handlers
│   ├── server.acvs          ACVS service handlers
│   └── server.audit         Audit service handlers
├── crs                      Credential Remediation Service
├── acvs                     Compliance Validation Service
│   ├── acvs.crc             CRC Manager
│   ├── acvs.validator       Validator
│   ├── acvs.evidence        Evidence Chain
│   └── acvs.nlp             NLP Engine
├── rotation                 Rotation framework
│   ├── rotation.github      GitHub rotation
│   └── rotation.aws         AWS rotation
├── pwmanager                Password manager integration
│   ├── pwmanager.bitwarden  Bitwarden adapter
│   └── pwmanager.onepassword 1Password adapter
├── audit                    Audit logging
├── database                 Database operations
└── auth                     Authentication (mTLS)
```

### Component Configuration

Each component has independent log level configuration:

```yaml
components:
  acm: info                    # Global default
  server: info                 # gRPC server
  crs: info                    # CRS
  acvs: warn                   # ACVS (reduce noise)
  rotation: debug              # Rotation (verbose for debugging)
  rotation.github: debug       # GitHub rotation
  pwmanager: warn              # Password managers (reduce noise)
  database: info               # Database
```

---

## Correlation & Tracing

### Request ID Strategy

Every gRPC request gets a unique **Request ID** for end-to-end tracing.

#### Request ID Format

- **Type:** UUID v7 (time-sortable)
- **Example:** `01HBAG5E7V9QZXN8J4WTRGP0QX`
- **Benefits:**
  - Time-ordered (first 48 bits are timestamp)
  - Globally unique
  - Sortable in logs

#### Request ID Lifecycle

```
1. gRPC Request Arrives
   ↓
2. Interceptor generates/extracts Request ID
   ↓
3. Request ID added to context.Context
   ↓
4. All logs use WithContext(ctx) to include Request ID
   ↓
5. Outbound calls propagate Request ID via gRPC metadata
   ↓
6. Request ID in final response
```

#### Code Example

```go
// Client sends request
ctx := metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)

// Server interceptor extracts or generates
func UnaryServerInterceptor(logger *Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {

        // Extract from metadata or generate new
        requestID := extractOrGenerateRequestID(ctx)
        ctx = context.WithValue(ctx, requestIDKey, requestID)

        // Log with request ID
        logger.WithContext(ctx).Info("request started",
            "method", info.FullMethod,
            "request_id", requestID,
        )

        resp, err := handler(ctx, req)

        logger.WithContext(ctx).Info("request completed",
            "method", info.FullMethod,
            "request_id", requestID,
        )

        return resp, err
    }
}

// Application code logs with context
func (r *Rotator) StartRotation(ctx context.Context, req RotationRequest) (*RotationResult, error) {
    r.logger.WithContext(ctx).Info("starting rotation",
        "site", req.Site,
    )
    // Request ID automatically included
}
```

### Tracing Example

Following a GitHub rotation through logs:

```json
{"timestamp":"2025-11-17T14:32:45.000Z","level":"info","component":"server.rotation","request_id":"01HBAG5E...","message":"gRPC request started","method":"/acm.v1.RotationService/StartGitHubRotation"}

{"timestamp":"2025-11-17T14:32:45.010Z","level":"info","component":"rotation.github","request_id":"01HBAG5E...","message":"starting GitHub rotation","site":"github.com"}

{"timestamp":"2025-11-17T14:32:45.050Z","level":"debug","component":"rotation.github","request_id":"01HBAG5E...","message":"validating current token"}

{"timestamp":"2025-11-17T14:32:45.120Z","level":"info","component":"acvs","request_id":"01HBAG5E...","message":"validating action","site":"github.com","action":"PAT_ROTATION"}

{"timestamp":"2025-11-17T14:32:45.150Z","level":"info","component":"rotation.github","request_id":"01HBAG5E...","message":"ACVS validation passed","result":"ALLOWED"}

{"timestamp":"2025-11-17T14:32:45.200Z","level":"info","component":"server.rotation","request_id":"01HBAG5E...","message":"gRPC request completed","duration_ms":200}
```

All entries share the same `request_id`, enabling complete trace reconstruction.

---

## Sensitive Data Policy

### Never Log

The following data MUST NEVER appear in logs:

1. **Master Passwords**
   - Password manager master passwords
   - Vault encryption keys
   - Any user-entered passwords

2. **Full Authentication Tokens**
   - GitHub PATs (except first 12 chars)
   - AWS access keys (except prefix)
   - OAuth tokens
   - Session tokens

3. **Vault Contents**
   - Password values from vault
   - Secret notes
   - Secure fields

4. **Personally Identifiable Information (PII)**
   - Email addresses (show domain only)
   - Phone numbers
   - Social security numbers
   - Credit card numbers

### Redaction Rules

#### Tokens

Show only first 12 characters:

```
Original: ghp_1234567890abcdefghijklmnopqrstuvwxyz
Logged:   ghp_1234567890...
```

#### Passwords

Replace with `[REDACTED]`:

```
Original: MySecurePassword123!
Logged:   [REDACTED]
```

#### Email Addresses

Show domain only:

```
Original: alice@example.com
Logged:   ***@example.com
```

#### Credential IDs

Use SHA-256 hash:

```
Original: credential-abc-123
Logged:   sha256:a1b2c3d4e5f6789...
```

### Automatic Redaction

Fields with these names are automatically redacted:

- `password`
- `token`
- `secret`
- `api_key`
- `access_key`
- `secret_key`
- `private_key`
- `master_password`
- `passphrase`

### Redaction Implementation

```go
// Redact token to show only prefix
func redactToken(token string) string {
    if len(token) > 12 {
        return token[:12] + "..."
    }
    return "[REDACTED]"
}

// Hash credential ID
func hashCredentialID(id string) string {
    hash := sha256.Sum256([]byte(id))
    return "sha256:" + hex.EncodeToString(hash[:])
}

// Redact email
func redactEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) == 2 {
        return "***@" + parts[1]
    }
    return "[REDACTED]"
}
```

---

## Output Strategies

### Output Modes

ACM supports three output modes:

#### 1. Stdout (Default for Containers)

- **Use Case:** Docker, Kubernetes, cloud deployments
- **Benefits:**
  - 12-factor app compliant
  - Works with container log drivers
  - Easy integration with log aggregators

```bash
ACM_LOG_OUTPUT=stdout
```

#### 2. File (Traditional Deployments)

- **Use Case:** VM deployments, bare metal
- **Benefits:**
  - Persistent logs on disk
  - Log rotation built-in
  - Independent of process lifecycle

```bash
ACM_LOG_OUTPUT=file
ACM_LOG_FILE=/var/log/acm/acm.log
```

#### 3. Both (Debugging)

- **Use Case:** Development, troubleshooting
- **Benefits:**
  - Console visibility
  - Persistent logs

```bash
ACM_LOG_OUTPUT=both
```

### Format Selection

#### JSON (Production)

Machine-parsable, structured:

```json
{"timestamp":"2025-11-17T14:32:45.123Z","level":"info","component":"rotation.github","message":"rotation started"}
```

#### Pretty (Development)

Human-readable, colorized:

```
2025-11-17 14:32:45.123 INFO  [rotation.github] rotation started site=github.com
```

### Log Rotation

Automatic rotation prevents disk exhaustion:

```yaml
rotation:
  max_size_mb: 100        # Rotate at 100MB
  max_age_days: 30        # Delete logs older than 30 days
  max_backups: 10         # Keep 10 old files
  compress: true          # Compress rotated logs (.gz)
```

**Rotation Example:**
```
/var/log/acm/
├── acm.log              (current, 95MB)
├── acm-2025-11-16.log.gz (compressed, 80MB)
├── acm-2025-11-15.log.gz (compressed, 85MB)
└── ...
```

---

## Performance Considerations

### Latency Budget

| Operation | Target | Acceptable | Critical |
|-----------|--------|------------|----------|
| Log call | <100μs | <500μs | <1ms |
| JSON serialize | <500μs | <1ms | <5ms |
| File write (buffered) | <1ms | <10ms | <50ms |
| Request ID generation | <1μs | <10μs | <100μs |

### CPU & Memory Overhead

- **CPU:** <1% of total service CPU
- **Memory:** <50MB for log buffers
- **Disk I/O:** Asynchronous, buffered writes

### Optimization Strategies

1. **Lazy Evaluation**
   - Only format messages if level is enabled
   - Skip expensive operations for filtered logs

2. **Buffered Writes**
   - Buffer log entries before flushing to disk
   - Flush on: full buffer, timeout (1s), or service shutdown

3. **Sampling**
   - High-frequency logs can be sampled
   - Example: Log 1 in 100 password manager checks

4. **Sync vs Async**
   - ERROR/WARN: Synchronous (immediate flush)
   - INFO: Asynchronous (buffered)
   - DEBUG: Asynchronous (buffered)

### Benchmarks

Target performance:

```
BenchmarkLoggerInfo-8             1000000    1054 ns/op    320 B/op    5 allocs/op
BenchmarkLoggerWithContext-8       500000    2134 ns/op    640 B/op   10 allocs/op
BenchmarkRedactToken-8            5000000     345 ns/op    128 B/op    3 allocs/op
BenchmarkHashCredentialID-8       1000000    1234 ns/op    256 B/op    4 allocs/op
```

---

## Integration Architecture

### Service Integration Points

```
┌─────────────────────────────────────────────────────────────┐
│                      ACM Service                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Logger (slog wrapper)                     │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐       │
│  │ Request ID  │  │  Redaction  │  │ Performance  │       │
│  │ Middleware  │  │   Engine    │  │Instrumentation│       │
│  └─────────────┘  └─────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
          ┌─────────────────┐  ┌─────────────────┐
          │  JSON Handler   │  │  Pretty Handler │
          │  (Production)   │  │  (Development)  │
          └─────────────────┘  └─────────────────┘
                    │                   │
                    └─────────┬─────────┘
                              ▼
                    ┌─────────────────┐
                    │  Output Router  │
                    └─────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    ┌─────────┐         ┌─────────┐        ┌─────────┐
    │ Stdout  │         │  File   │        │  Both   │
    └─────────┘         │(Rotated)│        └─────────┘
                        └─────────┘
```

### Component Integration

Each ACM component gets its own logger instance:

```go
// main.go
func main() {
    // Initialize logging system
    logging.Initialize(config.Logging)

    // Create component loggers
    serverLogger := logging.NewLogger("server")
    crsLogger := logging.NewLogger("crs")
    acvsLogger := logging.NewLogger("acvs")
    rotationLogger := logging.NewLogger("rotation.github")

    // Pass to components
    server := server.NewServer(serverLogger, ...)
    crs := crs.NewService(crsLogger, ...)
    acvs := acvs.NewService(acvsLogger, ...)
    rotator := rotation.NewGitHubRotator(rotationLogger, ...)
}
```

### gRPC Integration

Logging integrates with gRPC via interceptors:

```go
// Server setup
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(logging.UnaryServerInterceptor(serverLogger)),
    grpc.StreamInterceptor(logging.StreamServerInterceptor(serverLogger)),
)
```

**Interceptor logs:**
- Request start (method, request ID)
- Request completion (duration, status)
- Errors with details

---

## Configuration Schema

### Environment Variables

```bash
# Log level (debug, info, warn, error)
ACM_LOG_LEVEL=info

# Log format (json, pretty)
ACM_LOG_FORMAT=json

# Log output (stdout, file, both)
ACM_LOG_OUTPUT=stdout

# Log file path (if output=file or both)
ACM_LOG_FILE=/var/log/acm/acm.log

# Rotation settings
ACM_LOG_MAX_SIZE=100      # MB
ACM_LOG_MAX_AGE=30        # days
ACM_LOG_MAX_BACKUPS=10    # number of old files
ACM_LOG_COMPRESS=true     # compress rotated logs

# Component-specific levels (comma-separated)
ACM_LOG_COMPONENTS="rotation.github=debug,pwmanager=warn"
```

### YAML Configuration

```yaml
# config/logging.yaml
logging:
  level: info
  format: json
  output: stdout

  file:
    path: /var/log/acm/acm.log
    max_size_mb: 100
    max_age_days: 30
    max_backups: 10
    compress: true

  components:
    server: info
    crs: info
    acvs: warn
    rotation: debug
    rotation.github: debug
    pwmanager: warn
    database: info

  redaction:
    enabled: true
    auto_detect: true
    fields:
      - password
      - token
      - secret
      - api_key

  performance:
    buffer_size: 1000
    flush_interval_ms: 1000
    sample_rate: 1.0  # 1.0 = log everything, 0.1 = log 10%
```

---

## Migration Strategy

### Phase 1: Infrastructure (Days 1-2)

1. Implement core logging package
2. Add request ID middleware
3. Implement redaction

### Phase 2: Integration (Days 2-3)

4. Integrate into existing services
5. Replace all `fmt.Printf` with structured logs
6. Add operational logging

### Phase 3: Production (Days 3-4)

7. Add log rotation
8. Add configuration
9. Performance tuning
10. Documentation

---

## Future Enhancements

### Phase IV Considerations

1. **Log Aggregation**
   - Loki integration for centralized logging
   - Elasticsearch/Kibana for search
   - CloudWatch Logs for AWS deployments

2. **Distributed Tracing**
   - OpenTelemetry integration
   - Jaeger/Zipkin compatibility
   - Trace IDs in addition to request IDs

3. **Metrics Integration**
   - Prometheus metrics from logs
   - Log-based alerting
   - Anomaly detection

4. **Advanced Features**
   - Log sampling for high-volume endpoints
   - Log-based profiling
   - Automated log analysis

---

## Appendix A: Log Message Conventions

### Message Formatting

- **Present tense** for actions: "starting rotation", "validating token"
- **Past tense** for completed: "rotation started", "token validated"
- **Lowercase** first word (unless proper noun)
- **No trailing punctuation**
- **Concise** but descriptive

### Good Examples

```
"starting GitHub rotation"
"token validation failed"
"ACVS validation passed"
"database connection established"
```

### Bad Examples

```
"Starting GitHub Rotation."  // Capitalized, punctuation
"Token failed"               // Too vague
"The token validation process has encountered an error" // Too verbose
```

---

## Appendix B: Component Reference

Complete list of component tags:

| Component | Tag | Description |
|-----------|-----|-------------|
| Main | `acm` | Service entry point |
| gRPC Server | `server` | gRPC server infrastructure |
| Credential Service | `server.credential` | Credential gRPC handlers |
| Rotation Service | `server.rotation` | Rotation gRPC handlers |
| ACVS Service | `server.acvs` | ACVS gRPC handlers |
| CRS | `crs` | Credential Remediation Service |
| ACVS | `acvs` | Compliance Validation Service |
| CRC Manager | `acvs.crc` | CRC cache management |
| Validator | `acvs.validator` | Compliance validation |
| Evidence Chain | `acvs.evidence` | Evidence chain generation |
| NLP Engine | `acvs.nlp` | Legal NLP analysis |
| Rotation | `rotation` | Rotation framework |
| GitHub Rotation | `rotation.github` | GitHub PAT rotation |
| AWS Rotation | `rotation.aws` | AWS IAM rotation |
| Password Manager | `pwmanager` | Password manager base |
| Bitwarden | `pwmanager.bitwarden` | Bitwarden integration |
| 1Password | `pwmanager.onepassword` | 1Password integration |
| Audit | `audit` | Audit logging |
| Database | `database` | Database operations |
| Auth | `auth` | mTLS authentication |

---

**Document Status:** Architecture Complete
**Next Step:** Implement Task 7.2 (Core Logging Package)
**Review Status:** Approved for Implementation

---

**End of Logging Architecture Document**
