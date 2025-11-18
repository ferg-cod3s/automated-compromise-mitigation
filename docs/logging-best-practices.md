# ACM Logging Best Practices

**Version:** 1.0
**Date:** November 2025
**Status:** Complete

---

## Table of Contents

1. [Overview](#overview)
2. [When to Log](#when-to-log)
3. [What to Log](#what-to-log)
4. [What NOT to Log](#what-not-to-log)
5. [Log Levels Guide](#log-levels-guide)
6. [Structured Logging](#structured-logging)
7. [Performance Considerations](#performance-considerations)
8. [Security Guidelines](#security-guidelines)
9. [Request Tracing](#request-tracing)
10. [Common Patterns](#common-patterns)
11. [Testing Logging](#testing-logging)

---

## Overview

This guide provides best practices for developers adding logging to ACM services. Following these guidelines ensures:

- **Operational visibility** - Logs provide insight into system behavior
- **Debuggability** - Issues can be diagnosed quickly
- **Security** - Sensitive data is never exposed
- **Performance** - Logging overhead is minimal
- **Consistency** - Log format is predictable across components

---

## When to Log

### DO Log

**Service Lifecycle Events:**
```go
logger.Info("ACM service starting", "version", version, "hostname", hostname)
logger.Info("ACM service ready", "address", listenAddr)
logger.Info("ACM service stopped", "uptime_seconds", uptime)
```

**Significant Operations:**
```go
logger.Info("GitHub rotation started", "credential_id_hash", hash)
logger.Info("GitHub rotation completed", "duration_ms", duration)
```

**Error Conditions:**
```go
logger.Error("failed to rotate credential",
    "error", err,
    "credential_id_hash", hash,
    "site", "github.com",
)
```

**Warning Conditions:**
```go
logger.Warn("operation slow", "duration_ms", 250, "threshold_ms", 100)
logger.Warn("retrying failed operation", "attempt", 2, "max_attempts", 3)
```

**State Changes:**
```go
logger.Info("ACVS enabled", "enabled_by", "user")
logger.Info("password manager detected", "manager", "Bitwarden")
```

### DON'T Log

**High-Frequency Events:**
```go
// BAD: Logs every iteration
for _, item := range items {
    logger.Debug("processing item", "item", item) // Avoid in loops
}

// GOOD: Log summary
logger.Debug("processing items", "count", len(items))
// ... process items ...
logger.Info("items processed", "count", len(items), "duration_ms", duration)
```

**Sensitive Data:**
```go
// NEVER log these directly:
logger.Info("token", "value", actualToken) // ❌ NEVER
logger.Info("password", "value", password) // ❌ NEVER
logger.Info("api_key", "value", apiKey)    // ❌ NEVER

// Use redaction:
logger.Info("token validated", "token_prefix", token[:8]) // ✓ OK
```

**Redundant Information:**
```go
// BAD: Same information logged multiple times
logger.Info("starting operation")
logger.Info("operation started") // Redundant

// GOOD: Log once with clear message
logger.Info("operation started", "operation", "github_rotation")
```

---

## What to Log

### Essential Context

Include these attributes when relevant:

**Request Context:**
```go
logger.WithContext(ctx).Info("operation completed",
    "request_id", "...",        // From context automatically
    "operation", "rotate",
    "duration_ms", duration,
)
```

**Resource Identifiers (Hashed):**
```go
logger.Info("credential rotated",
    "credential_id_hash", sha256.Sum256(credentialID),
    "site", "github.com",
    "username", username, // OK: Not sensitive
)
```

**Timing Information:**
```go
logger.Info("database query completed",
    "query", "select_credentials",
    "duration_ms", duration.Milliseconds(),
    "rows_returned", count,
)
```

**Error Details:**
```go
logger.Error("API call failed",
    "error", err,
    "api", "github",
    "endpoint", "/user/repos",
    "status_code", statusCode,
    "retry_attempt", retryCount,
)
```

---

## What NOT to Log

### Sensitive Data Categories

**Authentication Credentials:**
- Master passwords
- API tokens (full values)
- OAuth tokens
- Session tokens
- Private keys
- Encryption keys

**Personal Information:**
- Passwords
- Credit card numbers
- Social security numbers
- Unredacted email addresses (use domain-only)

**Security-Critical Data:**
- HMAC secrets
- Signing keys
- Certificate private keys
- Vault encryption keys

### Redaction Required

Use built-in redaction for:

```go
import "github.com/ferg-cod3s/automated-compromise-mitigation/internal/logging"

// Redact token (shows first 12 chars)
redactedToken := logging.RedactToken(token)
logger.Info("token created", "token", redactedToken) // "ghp_12345678..."

// Redact email (shows domain only)
redactedEmail := logging.RedactEmail(email, true)
logger.Info("user registered", "email", redactedEmail) // "***@example.com"

// Hash credential IDs
credHash := logging.HashCredentialID(credentialID)
logger.Info("credential accessed", "credential_id_hash", credHash)
```

---

## Log Levels Guide

### DEBUG

**Purpose:** Detailed diagnostic information
**Audience:** Developers debugging issues
**Production:** Usually disabled (enable for specific components)

```go
logger.Debug("validating token",
    "token_prefix", token[:8],
    "scopes_required", scopes,
)

logger.Debug("cache lookup",
    "key", key,
    "hit", hit,
    "ttl_seconds", ttl,
)
```

**When to use:**
- Internal state transitions
- Cache hits/misses
- Validation steps
- Algorithm decisions

### INFO

**Purpose:** Normal operational messages
**Audience:** Operators, SREs
**Production:** Always enabled

```go
logger.Info("rotation completed",
    "credential_id_hash", hash,
    "duration_ms", duration,
    "success", true,
)

logger.Info("service health check passed",
    "component", "database",
    "latency_ms", latency,
)
```

**When to use:**
- Successful operations
- State changes
- Configuration changes
- Health checks

### WARN

**Purpose:** Warning conditions (degraded but functional)
**Audience:** Operators, on-call engineers
**Production:** Always enabled

```go
logger.Warn("operation slow",
    "operation", "github_api_call",
    "duration_ms", 250,
    "threshold_ms", 100,
)

logger.Warn("retrying failed operation",
    "operation", "database_query",
    "attempt", 2,
    "max_attempts", 3,
    "error", err,
)

logger.Warn("fallback activated",
    "primary", "Bitwarden",
    "fallback", "1Password",
    "reason", err,
)
```

**When to use:**
- Slow operations (>100ms)
- Retry attempts
- Fallback scenarios
- Degraded functionality
- Resource limits approaching

### ERROR

**Purpose:** Error conditions requiring attention
**Audience:** On-call engineers, incident responders
**Production:** Always enabled, often triggers alerts

```go
logger.Error("failed to rotate credential",
    "error", err,
    "credential_id_hash", hash,
    "site", "github.com",
    "attempts", attempts,
)

logger.Error("database connection failed",
    "error", err,
    "database_path", dbPath,
    "retry_in_seconds", retryDelay,
)
```

**When to use:**
- Failed operations
- Unrecoverable errors
- External service failures
- Data corruption
- Security violations

---

## Structured Logging

### Key-Value Pairs

Always use key-value pairs, not string formatting:

```go
// BAD: String formatting
logger.Info(fmt.Sprintf("Rotated credential %s in %dms", id, duration))

// GOOD: Structured key-value pairs
logger.Info("credential rotated",
    "credential_id_hash", hash,
    "duration_ms", duration,
)
```

### Consistent Field Names

Use consistent naming across the codebase:

| Field | Type | Example | Notes |
|-------|------|---------|-------|
| `request_id` | string | `01HBAG5E7V9Q...` | From context |
| `duration_ms` | int64 | `156` | Milliseconds |
| `error` | error | `err` | Go error type |
| `component` | string | `rotation` | Service component |
| `operation` | string | `github_rotation` | Operation name |
| `credential_id_hash` | string | `sha256:a1b2...` | Hashed ID |
| `site` | string | `github.com` | Service site |
| `username` | string | `alice` | User (not sensitive) |
| `status_code` | int | `200` | HTTP status |
| `retry_attempt` | int | `2` | Retry count |

### Context-Aware Logging

Always use context when available:

```go
// Get context-aware logger
logger := logging.NewLogger("rotation")

func RotateCredential(ctx context.Context, credID string) error {
    // Extract request ID from context automatically
    log := logger.WithContext(ctx)

    log.Info("rotation started", "credential_id_hash", hash)

    // ... rotation logic ...

    log.Info("rotation completed", "duration_ms", duration)
    return nil
}
```

---

## Performance Considerations

### Timed Operations

Use the built-in `TimedOperation` helper:

```go
err := logger.TimedOperation(ctx, "github_api_call", func() error {
    return client.CreateToken(ctx, request)
})
// Automatically logs start, duration, and errors
```

### Avoid Expensive Operations

```go
// BAD: Expensive serialization in hot path
logger.Debug("request body", "body", string(largeBody)) // Allocates string

// GOOD: Only serialize when debug is enabled
if logger.Level() == slog.LevelDebug {
    logger.Debug("request body", "body", string(largeBody))
}
```

### Sampling (Future)

For high-frequency events, use sampling (future feature):

```go
// Sample 1% of events
if rand.Float64() < 0.01 {
    logger.Debug("cache access", "key", key)
}
```

---

## Security Guidelines

### Data Redaction

**Automatic Redaction:**

Certain field names are automatically redacted:
- `password`
- `token`
- `secret`
- `api_key`
- `private_key`

```go
// These are automatically redacted
logger.Info("config loaded",
    "password", "secret123",  // Logged as "[REDACTED]"
    "api_key", "ghp_abc123",  // Logged as "[REDACTED]"
)
```

**Manual Redaction:**

For other sensitive data:

```go
import "github.com/ferg-cod3s/automated-compromise-mitigation/internal/logging"

// Redact token
logger.Info("token created",
    "token", logging.RedactToken(token), // "ghp_12345678..."
)

// Redact email
logger.Info("user created",
    "email", logging.RedactEmail(email, true), // "***@example.com"
)

// Hash credential IDs
logger.Info("credential accessed",
    "credential_id_hash", logging.HashCredentialID(credID),
)
```

### Log Injection Prevention

Never concatenate user input into log messages:

```go
// BAD: Log injection vulnerability
userInput := "test\nlevel=ERROR msg=\"fake error\""
logger.Info("User input: " + userInput) // Can inject fake log entries

// GOOD: Use structured logging
logger.Info("user input received", "input", userInput) // Properly escaped
```

### Audit vs. Operational Logs

**Audit Logs:** Cryptographically signed, immutable, compliance-focused
```go
auditLogger.LogCredentialRotation(ctx, credentialID, outcome)
```

**Operational Logs:** Debugging, monitoring, operational visibility
```go
logger.Info("rotation completed", "credential_id_hash", hash)
```

**Use both:** Audit logs for compliance, operational logs for debugging

---

## Request Tracing

### Request ID Propagation

Request IDs are automatically propagated:

```go
// Server-side: Request ID added by middleware
func HandleRequest(ctx context.Context, req *Request) {
    logger := logging.NewLogger("server").WithContext(ctx)
    logger.Info("request received") // Includes request_id
}

// Client-side: Request ID propagated to outbound calls
func CallExternalAPI(ctx context.Context) {
    // Request ID automatically added to outbound metadata
    resp, err := client.GetUser(ctx, request)
}
```

### End-to-End Tracing

```
Client Request → gRPC Server → CRS → Password Manager → GitHub API
    [req-123]      [req-123]    [req-123]   [req-123]    [req-123]
```

All logs for a single request share the same `request_id`, enabling correlation.

### Searching by Request ID

```bash
# Find all logs for a specific request
grep "request_id.*01HBAG5E7V9Q" /var/log/acm/acm.log

# With jq (JSON logs)
cat /var/log/acm/acm.log | jq 'select(.request_id == "01HBAG5E7V9Q")'
```

---

## Common Patterns

### Service Startup

```go
func main() {
    logger := logging.NewLogger("main")

    logger.Info("ACM service starting",
        "version", version,
        "build_time", buildTime,
        "go_version", runtime.Version(),
        "hostname", hostname,
    )

    // ... initialization ...

    logger.Info("ACM service ready",
        "address", listenAddr,
        "mtls", true,
        "components", []string{"CRS", "ACVS", "rotation"},
    )
}
```

### Error Handling

```go
func RotateCredential(ctx context.Context, credID string) error {
    logger := logging.NewLogger("rotation").WithContext(ctx)

    result, err := githubClient.CreateToken(ctx, request)
    if err != nil {
        logger.Error("failed to create GitHub token",
            "error", err,
            "credential_id_hash", hash,
            "site", "github.com",
        )
        return fmt.Errorf("github token creation failed: %w", err)
    }

    logger.Info("GitHub token created",
        "credential_id_hash", hash,
        "expires_at", result.ExpiresAt,
    )

    return nil
}
```

### Retry Logic

```go
func CallWithRetry(ctx context.Context, fn func() error) error {
    logger := logging.NewLogger("retry").WithContext(ctx)

    for attempt := 1; attempt <= maxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        logger.Warn("operation failed, retrying",
            "error", err,
            "attempt", attempt,
            "max_attempts", maxAttempts,
            "retry_in_seconds", retryDelay,
        )

        time.Sleep(retryDelay)
    }

    logger.Error("operation failed after retries",
        "attempts", maxAttempts,
    )

    return fmt.Errorf("operation failed after %d attempts", maxAttempts)
}
```

### Database Operations

```go
func QueryCredentials(ctx context.Context, site string) ([]*Credential, error) {
    logger := logging.NewLogger("database").WithContext(ctx)

    err := logger.TimedOperation(ctx, "query_credentials", func() error {
        var err error
        credentials, err = db.Query(ctx, site)
        return err
    })

    if err != nil {
        return nil, err
    }

    logger.Info("credentials queried",
        "site", site,
        "count", len(credentials),
    )

    return credentials, nil
}
```

---

## Testing Logging

### Unit Tests

```go
func TestRotateCredential(t *testing.T) {
    // Initialize test logger
    config := logging.DefaultConfig()
    config.Level = "debug"
    logging.Initialize(config)

    // Your test code
    err := RotateCredential(ctx, credID)
    if err != nil {
        t.Errorf("RotateCredential failed: %v", err)
    }

    // Verify logs (if needed)
    // Note: Current implementation doesn't support log capture
    // Future: Add test helper for log verification
}
```

### Integration Tests

```go
func TestEndToEndRotation(t *testing.T) {
    // Setup test environment with logging
    config := logging.DevelopmentConfig()
    logging.Initialize(config)

    // Run integration test
    result := RunFullRotation(t, testCredential)

    // Logs will help debug failures
    if !result.Success {
        t.Error("Rotation failed - check logs for details")
    }
}
```

### Verify Redaction

```go
func TestSensitiveDataRedaction(t *testing.T) {
    token := "ghp_1234567890abcdefghijklmnopqrstuv"

    redacted := logging.RedactToken(token)

    if strings.Contains(redacted, "ghijklmnop") {
        t.Error("Token not properly redacted")
    }

    if !strings.Contains(redacted, "ghp_12345678") {
        t.Error("Token prefix should be visible")
    }
}
```

---

## Checklist

Before committing code with logging:

- [ ] No sensitive data logged directly
- [ ] Consistent field names used
- [ ] Appropriate log levels chosen
- [ ] Context-aware logging used (WithContext)
- [ ] Errors include sufficient context
- [ ] No logging in tight loops
- [ ] Request IDs propagated
- [ ] Timed operations use TimedOperation helper
- [ ] Redaction applied to tokens/emails
- [ ] Tests don't fail due to log output

---

## Related Documentation

- [Logging Architecture](./logging-architecture.md) - Design and implementation
- [Logging Configuration](./logging-configuration.md) - Configuration guide
- [Logging Troubleshooting](./logging-troubleshooting.md) - Common issues

---

**End of Logging Best Practices Guide**
