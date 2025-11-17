# Phase III - Task 7: Structured Logging Infrastructure

**Date:** 2025-11-17
**Priority:** HIGH (foundational for operations and debugging)
**Estimated Time:** 3-4 days
**Dependencies:** None (can start immediately)

---

## Objective

Implement a comprehensive structured logging infrastructure for ACM using Go's `log/slog` standard library. This provides operational visibility, debugging capabilities, and audit compliance across all services.

## Success Criteria

- ✅ Structured JSON logging in production, pretty logging in development
- ✅ Log levels (DEBUG, INFO, WARN, ERROR) configurable per component
- ✅ Contextual logging with request IDs, user IDs, and service tags
- ✅ Performance instrumentation for gRPC, database, and external APIs
- ✅ Log rotation and retention policies configured
- ✅ Zero performance overhead in hot paths (<1% CPU impact)
- ✅ All existing services instrumented with structured logging

---

## Atomic Subtasks

### 7.1 Design Logging Architecture

**Objective:** Design comprehensive logging strategy for ACM ecosystem.

**Tasks:**
- [ ] Design log schema (fields, structure, standardization)
- [ ] Define log levels per component (service, CRS, ACVS, rotation, etc.)
- [ ] Plan log outputs (stdout, file, syslog, future: external aggregators)
- [ ] Design correlation ID strategy for request tracing
- [ ] Plan sensitive data redaction policy (tokens, passwords, PII)
- [ ] Define log rotation policy (size limits, retention periods)
- [ ] Design performance instrumentation strategy
- [ ] Plan log aggregation integration points (future: ELK, Loki, etc.)

**Files:** `docs/logging-architecture.md` (~400 lines)

**Success Criteria:** Architecture document complete and reviewed

---

### 7.2 Implement Core Logging Package

**Objective:** Create central logging package wrapping `log/slog`.

**Tasks:**
- [ ] Create `internal/logging/logger.go` with slog wrapper
- [ ] Implement log level configuration (env var `ACM_LOG_LEVEL`)
- [ ] Implement output configuration (stdout, file, both)
- [ ] Add JSON formatter for production
- [ ] Add pretty formatter for development (colorized)
- [ ] Implement context-aware logging (extract request ID from context)
- [ ] Add component-specific logger creation (with component tag)
- [ ] Implement log sampling for high-volume events
- [ ] Add panic recovery with stack trace logging

**Files:**
- `internal/logging/logger.go` (~300 lines)
- `internal/logging/config.go` (~150 lines)
- `internal/logging/formatter.go` (~200 lines)

**Code Example:**
```go
package logging

import (
    "context"
    "log/slog"
    "os"
)

// Logger wraps slog.Logger with ACM-specific functionality
type Logger struct {
    *slog.Logger
    component string
}

// NewLogger creates a component-specific logger
func NewLogger(component string) *Logger {
    level := getLogLevelFromEnv()
    handler := createHandler(level)

    return &Logger{
        Logger:    slog.New(handler),
        component: component,
    }
}

// WithContext extracts request ID from context and adds it to log
func (l *Logger) WithContext(ctx context.Context) *Logger {
    requestID := getRequestIDFromContext(ctx)
    if requestID != "" {
        return &Logger{
            Logger:    l.With("request_id", requestID),
            component: l.component,
        }
    }
    return l
}

// Redact creates a new logger that redacts sensitive fields
func (l *Logger) Redact(fields ...string) *Logger {
    // Implementation to redact sensitive data
}
```

**Success Criteria:** Core logging package compiles and basic tests pass

---

### 7.3 Add Request ID Middleware

**Objective:** Implement request correlation across service calls.

**Tasks:**
- [ ] Create `internal/logging/request_id.go`
- [ ] Implement gRPC unary interceptor for request ID injection
- [ ] Implement gRPC stream interceptor for request ID injection
- [ ] Add request ID to context on ingress
- [ ] Extract and propagate request ID in outbound calls
- [ ] Generate request IDs using UUID v7 (time-sortable)
- [ ] Add request ID to all log entries
- [ ] Document request ID header (`X-Request-ID`)

**Files:**
- `internal/logging/request_id.go` (~200 lines)
- `internal/logging/middleware.go` (~150 lines)

**Code Example:**
```go
// UnaryServerInterceptor injects request ID into context
func UnaryServerInterceptor(logger *Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
                handler grpc.UnaryHandler) (interface{}, error) {

        requestID := extractOrGenerateRequestID(ctx)
        ctx = context.WithValue(ctx, requestIDKey, requestID)

        logger.WithContext(ctx).Info("gRPC request started",
            "method", info.FullMethod,
            "request_id", requestID,
        )

        start := time.Now()
        resp, err := handler(ctx, req)
        duration := time.Since(start)

        if err != nil {
            logger.WithContext(ctx).Error("gRPC request failed",
                "method", info.FullMethod,
                "duration_ms", duration.Milliseconds(),
                "error", err,
            )
        } else {
            logger.WithContext(ctx).Info("gRPC request completed",
                "method", info.FullMethod,
                "duration_ms", duration.Milliseconds(),
            )
        }

        return resp, err
    }
}
```

**Success Criteria:** Request IDs appear in all logs, traceable end-to-end

---

### 7.4 Implement Sensitive Data Redaction

**Objective:** Ensure tokens, passwords, and PII never appear in logs.

**Tasks:**
- [ ] Create `internal/logging/redact.go`
- [ ] Implement token redaction (show only first 8 chars: `ghp_abc12345...`)
- [ ] Implement password redaction (replace with `[REDACTED]`)
- [ ] Implement email redaction (show only domain: `***@example.com`)
- [ ] Implement credential ID hashing for logs
- [ ] Add automatic redaction of fields named: `token`, `password`, `secret`, `key`
- [ ] Create whitelist of safe fields
- [ ] Add test suite with sensitive data detection

**Files:**
- `internal/logging/redact.go` (~250 lines)
- `internal/logging/redact_test.go` (~200 lines)

**Code Example:**
```go
// RedactSensitiveFields automatically redacts common sensitive fields
func RedactSensitiveFields(fields map[string]interface{}) map[string]interface{} {
    redacted := make(map[string]interface{})

    for key, value := range fields {
        switch {
        case isSensitiveFieldName(key):
            redacted[key] = redactValue(key, value)
        case isEmailField(key):
            redacted[key] = redactEmail(value)
        case isTokenField(key):
            redacted[key] = redactToken(value)
        default:
            redacted[key] = value
        }
    }

    return redacted
}

func redactToken(value interface{}) string {
    s := fmt.Sprint(value)
    if len(s) > 12 {
        return s[:12] + "..."
    }
    return "[REDACTED]"
}
```

**Success Criteria:** All test cases pass, no sensitive data in logs

---

### 7.5 Add Performance Instrumentation

**Objective:** Log timing metrics for performance monitoring.

**Tasks:**
- [ ] Create `internal/logging/metrics.go`
- [ ] Add timing helpers for operations
- [ ] Implement database query timing
- [ ] Implement GitHub API call timing
- [ ] Implement AWS API call timing
- [ ] Implement password manager CLI timing
- [ ] Add percentile tracking (p50, p95, p99)
- [ ] Implement slow query logging (threshold: 100ms)
- [ ] Add memory allocation tracking

**Files:**
- `internal/logging/metrics.go` (~300 lines)
- `internal/logging/timing.go` (~150 lines)

**Code Example:**
```go
// TimedOperation logs the duration of an operation
func (l *Logger) TimedOperation(ctx context.Context, operation string,
                                fn func() error) error {
    start := time.Now()

    l.WithContext(ctx).Debug("operation started",
        "operation", operation,
    )

    err := fn()
    duration := time.Since(start)

    if err != nil {
        l.WithContext(ctx).Error("operation failed",
            "operation", operation,
            "duration_ms", duration.Milliseconds(),
            "error", err,
        )
    } else {
        level := slog.LevelInfo
        if duration > 100*time.Millisecond {
            level = slog.LevelWarn // Slow operation
        }

        l.WithContext(ctx).Log(ctx, level, "operation completed",
            "operation", operation,
            "duration_ms", duration.Milliseconds(),
        )
    }

    return err
}
```

**Success Criteria:** Performance metrics logged for all major operations

---

### 7.6 Integrate Logging into Existing Services

**Objective:** Instrument all existing ACM services with structured logging.

**Tasks:**
- [ ] Update `cmd/acm-service/main.go` with logger initialization
- [ ] Add logging to CRS service (`internal/crs/service.go`)
- [ ] Add logging to ACVS service (`internal/acvs/service.go`)
- [ ] Add logging to GitHub rotator (`internal/rotation/github/rotator.go`)
- [ ] Add logging to password manager integrations
- [ ] Add logging to gRPC server handlers
- [ ] Add logging to database operations
- [ ] Replace all `fmt.Printf` with structured logging
- [ ] Add panic recovery middleware with logging

**Files Modified:**
- `cmd/acm-service/main.go`
- `internal/crs/service.go`
- `internal/acvs/service.go`
- `internal/rotation/github/rotator.go`
- `internal/server/*.go`
- `internal/pwmanager/*/adapter.go`

**Example Integration:**
```go
// In internal/rotation/github/rotator.go
func (r *Rotator) StartRotation(ctx context.Context, req RotationRequest) (*RotationResult, error) {
    log := r.logger.WithContext(ctx)

    log.Info("starting GitHub PAT rotation",
        "credential_id_hash", hashCredentialID(req.CredentialID),
        "site", req.Site,
    )

    return r.logger.TimedOperation(ctx, "github_rotation_start", func() error {
        // Existing rotation logic
        user, err := r.client.GetUser(ctx, req.CurrentToken)
        if err != nil {
            log.Error("failed to validate current token",
                "error", err,
                "site", req.Site,
            )
            return err
        }

        log.Debug("current token validated",
            "username", user.Login,
            "site", req.Site,
        )

        // ... rest of logic
    })
}
```

**Success Criteria:** All services emit structured logs with context

---

### 7.7 Add Log Rotation and Management

**Objective:** Implement log file rotation and retention policies.

**Tasks:**
- [ ] Add `gopkg.in/natefinch/lumberjack.v2` dependency
- [ ] Create `internal/logging/rotation.go`
- [ ] Implement file output with rotation
- [ ] Configure max file size (default: 100MB)
- [ ] Configure max age (default: 30 days)
- [ ] Configure max backups (default: 10 files)
- [ ] Add compression of rotated logs
- [ ] Implement log file cleanup on service start
- [ ] Add configuration via environment variables

**Files:**
- `internal/logging/rotation.go` (~200 lines)
- `internal/logging/output.go` (~150 lines)

**Configuration:**
```bash
# Environment variables
ACM_LOG_FILE=/var/log/acm/acm.log
ACM_LOG_MAX_SIZE=100      # MB
ACM_LOG_MAX_AGE=30        # days
ACM_LOG_MAX_BACKUPS=10    # number of old files
ACM_LOG_COMPRESS=true     # compress rotated logs
```

**Success Criteria:** Logs rotate automatically, old logs cleaned up

---

### 7.8 Add Logging Configuration

**Objective:** Make logging configurable via config file and env vars.

**Tasks:**
- [ ] Create `internal/logging/config.go` for configuration
- [ ] Support configuration via environment variables
- [ ] Support configuration via YAML/JSON config file
- [ ] Add per-component log level configuration
- [ ] Add output selection (stdout, file, both, syslog)
- [ ] Add format selection (json, pretty, logfmt)
- [ ] Implement hot-reload of log configuration (SIGHUP)
- [ ] Add validation of configuration
- [ ] Document all configuration options

**Files:**
- `internal/logging/config.go` (~250 lines)
- `docs/logging-configuration.md` (~300 lines)

**Example Config:**
```yaml
# config/logging.yaml
logging:
  level: info                    # Global level
  format: json                   # json, pretty, logfmt
  output: both                   # stdout, file, both, syslog

  file:
    path: /var/log/acm/acm.log
    max_size_mb: 100
    max_age_days: 30
    max_backups: 10
    compress: true

  components:
    crs: debug                   # Component-specific levels
    acvs: info
    rotation: debug
    server: info

  redaction:
    enabled: true
    fields:
      - password
      - token
      - secret
      - api_key
```

**Success Criteria:** Configuration fully functional, documented

---

### 7.9 Add Operational Logging

**Objective:** Add structured logging for operational events.

**Tasks:**
- [ ] Log service startup with version, build info, config
- [ ] Log service shutdown with graceful cleanup
- [ ] Log certificate loading and expiration warnings
- [ ] Log database connection establishment and health
- [ ] Log password manager detection and connectivity
- [ ] Log ACVS enable/disable events
- [ ] Log rotation start, progress, completion
- [ ] Log HIM workflow triggers
- [ ] Log evidence chain additions
- [ ] Log error conditions with stack traces

**Files Modified:**
- `cmd/acm-service/main.go`
- All service initialization code

**Example:**
```go
func main() {
    logger := logging.NewLogger("main")

    logger.Info("ACM service starting",
        "version", version.Version,
        "build_time", version.BuildTime,
        "go_version", runtime.Version(),
        "hostname", hostname,
    )

    // Service initialization
    db, err := initDatabase(ctx, config)
    if err != nil {
        logger.Error("database initialization failed",
            "error", err,
            "config", redact.DatabaseConfig(config.Database),
        )
        os.Exit(1)
    }

    logger.Info("database initialized",
        "path", config.Database.Path,
        "version", dbVersion,
    )

    // ... rest of initialization

    logger.Info("ACM service ready",
        "address", config.Server.Address,
        "tls_enabled", config.Server.TLS.Enabled,
    )
}
```

**Success Criteria:** All operational events logged with context

---

### 7.10 Testing

**Objective:** Comprehensive testing of logging infrastructure.

**Tasks:**
- [ ] Create unit tests for logger creation
- [ ] Test log level filtering
- [ ] Test request ID propagation
- [ ] Test sensitive data redaction
- [ ] Test performance impact (benchmark)
- [ ] Test log rotation functionality
- [ ] Test configuration loading and validation
- [ ] Test JSON and pretty formatting
- [ ] Test context propagation across services
- [ ] Test panic recovery and logging
- [ ] Create integration test with full service
- [ ] Verify no log injection vulnerabilities

**Files:**
- `internal/logging/logger_test.go` (~300 lines)
- `internal/logging/redact_test.go` (~200 lines)
- `internal/logging/config_test.go` (~150 lines)
- `internal/logging/benchmark_test.go` (~200 lines)

**Test Coverage:**
- Unit tests: >90% coverage
- Benchmarks: <1% CPU overhead
- Integration: End-to-end request tracing

**Success Criteria:** All tests pass, performance acceptable

---

### 7.11 Documentation

**Objective:** Complete documentation for logging infrastructure.

**Tasks:**
- [ ] Create `docs/logging-architecture.md` (architecture overview)
- [ ] Create `docs/logging-configuration.md` (config reference)
- [ ] Create `docs/logging-best-practices.md` (developer guide)
- [ ] Document log schema and fields
- [ ] Document request ID tracing
- [ ] Document sensitive data redaction policy
- [ ] Create troubleshooting guide
- [ ] Add logging examples for common scenarios
- [ ] Document log aggregation integration (future)
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md

**Files:**
- `docs/logging-architecture.md` (~400 lines)
- `docs/logging-configuration.md` (~300 lines)
- `docs/logging-best-practices.md` (~500 lines)
- `docs/logging-troubleshooting.md` (~300 lines)

**Documentation Sections:**
1. Architecture Overview
2. Configuration Reference
3. Developer Guide (how to add logging)
4. Log Schema Reference
5. Request Tracing Guide
6. Performance Considerations
7. Security & Redaction
8. Troubleshooting

**Success Criteria:** Complete logging documentation published

---

## Log Schema Standard

### Standard Log Fields

All log entries should include:

```json
{
  "timestamp": "2025-11-17T14:32:45.123Z",
  "level": "info",
  "component": "rotation",
  "request_id": "01HBAG5E7V9QZXN8J4WTRGP0QX",
  "message": "GitHub rotation started",

  "operation": "github_rotation_start",
  "duration_ms": 156,

  "credential_id_hash": "sha256:a1b2c3...",
  "site": "github.com",
  "username": "testuser",

  "error": "token validation failed",
  "stack_trace": "...",

  "service": "acm",
  "version": "0.3.0",
  "hostname": "acm-prod-01"
}
```

### Log Levels

- **DEBUG**: Detailed debugging information (disabled in production)
- **INFO**: Normal operational messages (default level)
- **WARN**: Warning conditions (degraded, retry, fallback)
- **ERROR**: Error conditions requiring attention

### Component Tags

- `main` - Service initialization
- `server` - gRPC server
- `crs` - Credential Remediation Service
- `acvs` - Compliance Validation Service
- `rotation` - Credential rotation
- `rotation.github` - GitHub rotation
- `rotation.aws` - AWS rotation
- `pwmanager` - Password manager integration
- `audit` - Audit logging
- `evidence` - Evidence chain
- `database` - Database operations

---

## Performance Requirements

### Latency Targets

- Log call overhead: <100μs (microseconds)
- JSON serialization: <500μs
- File write (buffered): <1ms
- No blocking in hot paths

### Resource Limits

- CPU overhead: <1% of total
- Memory: <50MB for log buffers
- Disk I/O: Asynchronous, buffered

### Benchmarks

```go
BenchmarkLoggerInfo-8         1000000    1054 ns/op    320 B/op    5 allocs/op
BenchmarkLoggerWithContext-8   500000    2134 ns/op    640 B/op   10 allocs/op
BenchmarkRedaction-8           300000    4567 ns/op   1280 B/op   15 allocs/op
```

---

## Security Considerations

### Sensitive Data Protection

1. **Never log**:
   - Master passwords
   - Vault encryption keys
   - Full authentication tokens
   - Credentials in plaintext
   - PII without hashing

2. **Always redact**:
   - Tokens (show first 12 chars only)
   - Passwords (replace with `[REDACTED]`)
   - Email addresses (show domain only)
   - Credential IDs (use SHA-256 hash)

3. **Log injection prevention**:
   - Validate all log input
   - Escape newlines and special characters
   - Use structured logging (no string concatenation)

### Audit Trail

Logging complements but doesn't replace:
- Audit logging (Ed25519 signed events)
- Evidence chain (cryptographic proof)
- Database audit tables

---

## Integration with Existing Systems

### Phase I Integration

- **Audit Logger**: Structured logs for audit events
- **CRS**: Log credential operations with context
- **Password Managers**: Log CLI invocations and results

### Phase II Integration

- **ACVS**: Log ToS analysis, validation results
- **Evidence Chain**: Log chain additions with signatures
- **NLP Engine**: Log analysis requests and responses

### Phase III Integration

- **Rotation**: Log all rotation workflow steps
- **SQLite**: Log database operations and performance
- **OpenTUI**: Log user interactions (future)

---

## Future Enhancements (Phase IV)

### Log Aggregation

- Integration with Loki/Grafana
- Integration with Elasticsearch/Kibana
- Integration with CloudWatch Logs
- Integration with Splunk

### Advanced Features

- Distributed tracing (OpenTelemetry)
- Metrics export (Prometheus)
- Log-based alerting
- Anomaly detection

---

## Dependencies

### Go Libraries

- `log/slog` - Standard library (Go 1.21+)
- `gopkg.in/natefinch/lumberjack.v2` - Log rotation
- `github.com/google/uuid` - Request ID generation (already used)

### No External Services Required

- All logging is local-first
- Optional future integration with external systems
- Fully functional without network dependencies

---

## Rollout Plan

### Phase 1: Core Infrastructure (Day 1-2)

1. Implement core logging package (7.2)
2. Add request ID middleware (7.3)
3. Implement redaction (7.4)

### Phase 2: Integration (Day 2-3)

4. Integrate into existing services (7.6)
5. Add operational logging (7.9)
6. Add performance instrumentation (7.5)

### Phase 3: Production Readiness (Day 3-4)

7. Add log rotation (7.7)
8. Add configuration (7.8)
9. Testing (7.10)
10. Documentation (7.11)

---

## Definition of Done

Task 7 is complete when:

- [ ] All 11 subtasks completed with checkboxes marked
- [ ] Core logging package implemented and tested
- [ ] All services instrumented with structured logging
- [ ] Sensitive data redaction working and tested
- [ ] Request ID tracing end-to-end functional
- [ ] Log rotation and retention configured
- [ ] Performance benchmarks meet targets (<1% overhead)
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] No sensitive data in test logs verified
- [ ] Complete documentation published
- [ ] Code review completed
- [ ] All commits pushed to branch

---

**Task Status:** Ready to Start
**Estimated Completion:** 3-4 days
**Priority:** HIGH (foundational for Phase III tasks 4-6)

---

**Next Steps:**
1. Review and approve this task plan
2. Create feature branch: `feature/phase3-logging`
3. Begin with subtask 7.1 (Design)
4. Implement subtasks 7.2-7.11 sequentially
5. Commit and push regularly
6. Update PHASE3_IMPLEMENTATION_SUMMARY.md upon completion
