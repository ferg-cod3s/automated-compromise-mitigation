# ACM Logging Configuration Guide

**Version:** 1.0
**Date:** November 2025
**Status:** Complete

---

## Table of Contents

1. [Overview](#overview)
2. [Configuration Methods](#configuration-methods)
3. [Configuration Options](#configuration-options)
4. [Log Levels](#log-levels)
5. [Output Formats](#output-formats)
6. [Output Modes](#output-modes)
7. [Log Rotation](#log-rotation)
8. [Component-Specific Configuration](#component-specific-configuration)
9. [Environment Variables Reference](#environment-variables-reference)
10. [Examples](#examples)

---

## Overview

ACM's logging system is built on Go's `log/slog` standard library and provides flexible configuration for different deployment environments. Logging can be configured via:

- **Environment variables** (recommended for production)
- **Programmatic configuration** (for testing and development)
- **Sensible defaults** (zero-configuration startup)

---

## Configuration Methods

### Environment Variables

The primary configuration method for production deployments:

```bash
# Set log level
export ACM_LOG_LEVEL=info

# Set log format
export ACM_LOG_FORMAT=json

# Set output mode
export ACM_LOG_OUTPUT=both

# Set log file path
export ACM_LOG_FILE=/var/log/acm/acm.log

# Set rotation parameters
export ACM_LOG_MAX_SIZE=100
export ACM_LOG_MAX_AGE=30
export ACM_LOG_MAX_BACKUPS=10
export ACM_LOG_COMPRESS=true
```

### Programmatic Configuration

For development and testing:

```go
import "github.com/ferg-cod3s/automated-compromise-mitigation/internal/logging"

// Use pre-configured development settings
config := logging.DevelopmentConfig()

// Or customize
config := logging.Config{
    Level:      "debug",
    Format:     logging.FormatPretty,
    OutputMode: logging.OutputStdout,
}

// Initialize logging
if err := logging.Initialize(config); err != nil {
    log.Fatal(err)
}
```

### Default Configuration

If no configuration is provided, ACM uses sensible defaults:

- **Level:** `info`
- **Format:** `json`
- **Output:** `stdout`
- **Service Name:** `acm`
- **Version:** From build flags

---

## Configuration Options

### Core Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Level` | string | `info` | Minimum log level (debug, info, warn, error) |
| `Format` | Format | `json` | Output format (json, pretty, text) |
| `OutputMode` | OutputMode | `stdout` | Where to write logs (stdout, file, both) |
| `FilePath` | string | `/var/log/acm/acm.log` | Log file path (when using file output) |
| `ServiceName` | string | `acm` | Service name added to all logs |
| `Version` | string | From build | Service version added to all logs |
| `Hostname` | string | Auto-detected | Hostname added to all logs |
| `PID` | int | Auto-detected | Process ID added to all logs |

### Rotation Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `MaxSizeMB` | int | `100` | Maximum size in MB before rotation |
| `MaxAgeDays` | int | `30` | Maximum age in days before deletion |
| `MaxBackups` | int | `10` | Maximum number of old log files to keep |
| `Compress` | bool | `true` | Whether to compress rotated logs (gzip) |

---

## Log Levels

ACM supports four log levels, from most to least verbose:

### DEBUG

**Purpose:** Detailed diagnostic information for debugging
**When to use:** Development and troubleshooting
**Performance impact:** High (verbose output)

```bash
ACM_LOG_LEVEL=debug
```

**Example output:**
```json
{
  "time": "2025-11-17T14:32:45.123Z",
  "level": "DEBUG",
  "msg": "processing credential rotation",
  "component": "rotation",
  "credential_id_hash": "sha256:a1b2c3...",
  "request_id": "01HBAG5E7V9Q..."
}
```

### INFO (Default)

**Purpose:** Normal operational messages
**When to use:** Production
**Performance impact:** Low

```bash
ACM_LOG_LEVEL=info
```

**Example output:**
```json
{
  "time": "2025-11-17T14:32:45.123Z",
  "level": "INFO",
  "msg": "GitHub rotation completed",
  "component": "rotation",
  "duration_ms": 156
}
```

### WARN

**Purpose:** Warning conditions (degraded, retry, fallback)
**When to use:** Always enabled
**Performance impact:** Minimal

**Example scenarios:**
- Password manager CLI not found (fallback to another)
- Slow operations (>100ms)
- Retry attempts

### ERROR

**Purpose:** Error conditions requiring attention
**When to use:** Always enabled
**Performance impact:** Minimal

**Example scenarios:**
- Failed API calls
- Database errors
- Invalid credentials

---

## Output Formats

### JSON Format (Production)

**Recommended for:** Production, log aggregation, machine parsing

```bash
ACM_LOG_FORMAT=json
```

**Output:**
```json
{"time":"2025-11-17T14:32:45.123456789Z","level":"INFO","msg":"service started","service":"acm","version":"0.3.0","hostname":"acm-prod-01","pid":12345}
```

**Benefits:**
- Machine-readable
- Easy to parse
- Structured fields
- Compatible with log aggregators (Loki, ELK)

### Pretty Format (Development)

**Recommended for:** Development, local debugging

```bash
ACM_LOG_FORMAT=pretty
```

**Output:**
```
2025-11-17 14:32:45 INFO  [acm] service started version=0.3.0 hostname=acm-prod-01 pid=12345
```

**Benefits:**
- Human-readable
- Color-coded (in terminals)
- Easy to scan visually

### Text Format

**Recommended for:** Simple deployments, syslog

```bash
ACM_LOG_FORMAT=text
```

**Output:**
```
time=2025-11-17T14:32:45.123Z level=INFO msg="service started" service=acm version=0.3.0
```

---

## Output Modes

### Stdout (Default)

Write logs to standard output:

```bash
ACM_LOG_OUTPUT=stdout
```

**Use cases:**
- Container deployments (Docker, Kubernetes)
- Systemd services with journal logging
- Cloud platforms with log collection

### File

Write logs to a file with rotation:

```bash
ACM_LOG_OUTPUT=file
ACM_LOG_FILE=/var/log/acm/acm.log
```

**Use cases:**
- Traditional server deployments
- Local log retention required
- Compliance requirements

### Both

Write logs to both stdout and file:

```bash
ACM_LOG_OUTPUT=both
ACM_LOG_FILE=/var/log/acm/acm.log
```

**Use cases:**
- Development environments
- Hybrid deployments
- Debugging production issues

---

## Log Rotation

### Automatic Rotation

Logs are automatically rotated based on configuration:

```bash
# Rotate when file reaches 100MB
ACM_LOG_MAX_SIZE=100

# Delete logs older than 30 days
ACM_LOG_MAX_AGE=30

# Keep maximum 10 backup files
ACM_LOG_MAX_BACKUPS=10

# Compress old logs with gzip
ACM_LOG_COMPRESS=true
```

### Rotation Behavior

**When rotation occurs:**
1. Current log file renamed to `acm.log.YYYY-MM-DD-HHMMSS`
2. New empty `acm.log` created
3. Old rotated files compressed (if enabled)
4. Files older than `MaxAgeDays` deleted
5. Excess backups deleted (keeps only `MaxBackups`)

**Example rotated files:**
```
/var/log/acm/
  acm.log              (current)
  acm.log.2025-11-17-143245
  acm.log.2025-11-16-102314.gz
  acm.log.2025-11-15-084523.gz
```

### Manual Rotation

Trigger rotation via signal (future feature):

```bash
# Send SIGHUP to reload config and rotate logs
kill -HUP $(pidof acm-service)
```

---

## Component-Specific Configuration

Set different log levels for different components:

### Programmatic

```go
config := logging.DefaultConfig()
config.Level = "info"  // Global level

// Set component-specific levels
config.SetComponentLevel("rotation", "debug")
config.SetComponentLevel("acvs", "warn")
config.SetComponentLevel("crs", "info")

logging.Initialize(config)
```

### Component Names

| Component | Description |
|-----------|-------------|
| `main` | Service initialization and shutdown |
| `server` | gRPC server |
| `grpc` | gRPC middleware |
| `crs` | Credential Remediation Service |
| `acvs` | Compliance Validation Service |
| `rotation` | Credential rotation (all types) |
| `rotation.github` | GitHub PAT rotation |
| `rotation.aws` | AWS credential rotation |
| `pwmanager` | Password manager integration |
| `audit` | Audit logging |
| `evidence` | Evidence chain |
| `database` | Database operations |

---

## Environment Variables Reference

### Complete List

```bash
# Core Configuration
ACM_LOG_LEVEL=info              # Log level (debug|info|warn|error)
ACM_LOG_FORMAT=json             # Format (json|pretty|text)
ACM_LOG_OUTPUT=stdout           # Output (stdout|file|both)
ACM_LOG_FILE=/var/log/acm/acm.log  # File path

# Rotation Configuration
ACM_LOG_MAX_SIZE=100            # Max size in MB
ACM_LOG_MAX_AGE=30              # Max age in days
ACM_LOG_MAX_BACKUPS=10          # Number of backups
ACM_LOG_COMPRESS=true           # Compress (true|false|yes|no|1|0)

# Environment Detection
ACM_ENV=production              # Environment (production|development)
```

### Validation

Invalid values are replaced with defaults:
- Invalid log level → `info`
- Invalid format → `json`
- Invalid output mode → `stdout`
- Invalid boolean → `false`

---

## Examples

### Development Setup

```bash
# Enable debug logging with pretty format
export ACM_LOG_LEVEL=debug
export ACM_LOG_FORMAT=pretty
export ACM_LOG_OUTPUT=stdout

./acm-service
```

### Production Setup (Container)

```bash
# JSON logs to stdout for container orchestration
export ACM_ENV=production
export ACM_LOG_LEVEL=info
export ACM_LOG_FORMAT=json
export ACM_LOG_OUTPUT=stdout

./acm-service
```

### Production Setup (Server)

```bash
# JSON logs to file with rotation
export ACM_ENV=production
export ACM_LOG_LEVEL=info
export ACM_LOG_FORMAT=json
export ACM_LOG_OUTPUT=file
export ACM_LOG_FILE=/var/log/acm/acm.log
export ACM_LOG_MAX_SIZE=100
export ACM_LOG_MAX_AGE=30
export ACM_LOG_MAX_BACKUPS=10
export ACM_LOG_COMPRESS=true

./acm-service
```

### Debugging Specific Component

```bash
# Enable debug for rotation component only
# (Note: Requires code-level configuration currently)
export ACM_LOG_LEVEL=info

# In code:
# config.SetComponentLevel("rotation", "debug")
```

### Hybrid Setup (Development)

```bash
# Pretty logs to both stdout and file
export ACM_LOG_LEVEL=debug
export ACM_LOG_FORMAT=pretty
export ACM_LOG_OUTPUT=both
export ACM_LOG_FILE=/tmp/acm.log

./acm-service
```

---

## Best Practices

1. **Production:** Use `json` format with `stdout` output
2. **Development:** Use `pretty` format with `stdout` output
3. **Sensitive Data:** Never log passwords, tokens, or secrets (automatic redaction enabled)
4. **Log Levels:** Use `info` in production, `debug` only for troubleshooting
5. **Rotation:** Configure appropriate limits based on disk space and retention requirements
6. **Performance:** Monitor log volume in high-traffic environments

---

## Related Documentation

- [Logging Architecture](./logging-architecture.md) - Design and implementation
- [Logging Best Practices](./logging-best-practices.md) - Developer guide
- [Logging Troubleshooting](./logging-troubleshooting.md) - Common issues

---

**End of Logging Configuration Guide**
