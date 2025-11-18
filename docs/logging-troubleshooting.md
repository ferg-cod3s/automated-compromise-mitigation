# ACM Logging Troubleshooting Guide

**Version:** 1.0
**Date:** November 2025
**Status:** Complete

---

## Table of Contents

1. [Overview](#overview)
2. [Common Issues](#common-issues)
3. [No Logs Appearing](#no-logs-appearing)
4. [Log Level Issues](#log-level-issues)
5. [Performance Problems](#performance-problems)
6. [File Rotation Issues](#file-rotation-issues)
7. [Request ID Issues](#request-id-issues)
8. [Debugging Tools](#debugging-tools)
9. [FAQ](#faq)

---

## Overview

This guide helps diagnose and resolve common logging issues in ACM. For each problem, we provide:

- **Symptoms** - What you observe
- **Diagnosis** - How to identify the root cause
- **Solution** - How to fix the issue

---

## Common Issues

### Quick Diagnostic Commands

```bash
# Check if service is logging
./acm-service 2>&1 | head -20

# Verify log file exists and is being written
ls -lh /var/log/acm/acm.log
tail -f /var/log/acm/acm.log

# Check environment variables
env | grep ACM_LOG

# Verify log file permissions
ls -la /var/log/acm/

# Check disk space
df -h /var/log
```

---

## No Logs Appearing

### Symptom

Service starts but no logs are visible (stdout or file).

### Diagnosis

**1. Check stdout:**
```bash
./acm-service 2>&1 | head
```

**2. Check environment variables:**
```bash
echo $ACM_LOG_LEVEL
echo $ACM_LOG_OUTPUT
echo $ACM_LOG_FORMAT
```

**3. Check file path:**
```bash
ls -la /var/log/acm/
```

### Solutions

**Solution 1: Log level too high**

```bash
# Set log level to info or debug
export ACM_LOG_LEVEL=info
./acm-service
```

**Solution 2: Output mode misconfigured**

```bash
# Ensure output is set correctly
export ACM_LOG_OUTPUT=stdout
./acm-service
```

**Solution 3: File permission issues**

```bash
# Create log directory with correct permissions
sudo mkdir -p /var/log/acm
sudo chown $USER:$USER /var/log/acm
chmod 755 /var/log/acm

# Try again
export ACM_LOG_OUTPUT=file
./acm-service
```

**Solution 4: Disk full**

```bash
# Check disk space
df -h /var/log

# Clean up old logs
rm /var/log/acm/acm.log.*.gz

# Reduce rotation settings
export ACM_LOG_MAX_SIZE=50
export ACM_LOG_MAX_BACKUPS=5
```

---

## Log Level Issues

### Symptom 1: Too Much Debug Output

**Problem:** Logs are too verbose, performance is degraded.

**Solution:**
```bash
# Set to info level in production
export ACM_LOG_LEVEL=info
./acm-service
```

**For specific component debugging:**
```go
// In code (future feature)
config := logging.DefaultConfig()
config.Level = "info"  // Global
config.SetComponentLevel("rotation", "debug")  // Component-specific
```

### Symptom 2: Missing Debug Logs

**Problem:** Debug logs not appearing when needed.

**Diagnosis:**
```bash
# Check current log level
env | grep ACM_LOG_LEVEL
```

**Solution:**
```bash
# Enable debug logging
export ACM_LOG_LEVEL=debug
./acm-service
```

### Symptom 3: Inconsistent Log Levels

**Problem:** Some components log at different levels.

**Diagnosis:**

Check if component-specific levels are set (code level).

**Solution:**

Ensure global log level is set correctly. Component-specific levels override global.

---

## Performance Problems

### Symptom: High CPU Usage

**Problem:** Logging is consuming excessive CPU.

**Diagnosis:**
```bash
# Check log volume
wc -l /var/log/acm/acm.log

# Check log rate
tail -f /var/log/acm/acm.log | pv -l > /dev/null
```

**Solutions:**

**1. Reduce log level:**
```bash
export ACM_LOG_LEVEL=info  # or warn
```

**2. Check for log loops:**
```bash
# Look for repeating messages
sort /var/log/acm/acm.log | uniq -c | sort -rn | head -20
```

**3. Disable debug logging in production:**
```bash
export ACM_ENV=production
export ACM_LOG_LEVEL=info
```

### Symptom: High Memory Usage

**Problem:** Logging is consuming excessive memory.

**Diagnosis:**
```bash
# Check process memory
ps aux | grep acm-service

# Check if logs are buffering
ls -lh /var/log/acm/
```

**Solutions:**

**1. Reduce rotation buffer:**
```bash
export ACM_LOG_MAX_SIZE=50  # Smaller file size
```

**2. Increase rotation frequency:**
```bash
export ACM_LOG_MAX_SIZE=10  # Rotate more frequently
```

### Symptom: Slow Application

**Problem:** Application is slower after enabling logging.

**Diagnosis:**

Check if expensive operations are in hot paths.

**Solutions:**

**1. Use async logging (future feature):**
```bash
export ACM_LOG_ASYNC=true
```

**2. Reduce logging in critical paths:**

Review code for unnecessary Debug logs in tight loops.

**3. Use sampling (future feature):**

Only log a percentage of high-frequency events.

---

## File Rotation Issues

### Symptom 1: Log Files Not Rotating

**Problem:** Main log file grows indefinitely.

**Diagnosis:**
```bash
# Check file size
ls -lh /var/log/acm/acm.log

# Check rotation settings
env | grep ACM_LOG_MAX
```

**Solutions:**

**1. Verify rotation is enabled:**
```bash
export ACM_LOG_OUTPUT=file  # or both
export ACM_LOG_FILE=/var/log/acm/acm.log
export ACM_LOG_MAX_SIZE=100
```

**2. Check file permissions:**
```bash
# Ensure write permissions
chmod 644 /var/log/acm/acm.log
```

**3. Manually trigger rotation (future feature):**
```bash
kill -HUP $(pidof acm-service)
```

### Symptom 2: Too Many Old Log Files

**Problem:** Disk space consumed by old rotated logs.

**Diagnosis:**
```bash
# Count rotated files
ls -1 /var/log/acm/acm.log.* | wc -l

# Check disk usage
du -sh /var/log/acm/
```

**Solutions:**

**1. Reduce backup count:**
```bash
export ACM_LOG_MAX_BACKUPS=5
```

**2. Reduce retention period:**
```bash
export ACM_LOG_MAX_AGE=7  # 7 days
```

**3. Enable compression:**
```bash
export ACM_LOG_COMPRESS=true
```

**4. Manual cleanup:**
```bash
# Remove old compressed logs
find /var/log/acm -name "*.gz" -mtime +30 -delete
```

### Symptom 3: Rotation Not Compressing

**Problem:** Rotated files are not compressed (large disk usage).

**Diagnosis:**
```bash
# Check for uncompressed rotated files
ls -lh /var/log/acm/acm.log.*
```

**Solutions:**

**1. Enable compression:**
```bash
export ACM_LOG_COMPRESS=true
./acm-service  # Restart
```

**2. Manually compress old files:**
```bash
gzip /var/log/acm/acm.log.2025-*
```

---

## Request ID Issues

### Symptom 1: Missing Request IDs

**Problem:** Logs don't include `request_id` field.

**Diagnosis:**
```bash
# Check for request_id in logs
grep "request_id" /var/log/acm/acm.log
```

**Solutions:**

**1. Use context-aware logging:**
```go
// Ensure you use WithContext
logger := logging.NewLogger("component").WithContext(ctx)
logger.Info("message")  // Will include request_id if present
```

**2. Verify middleware is enabled:**
```go
// gRPC server should have interceptors
grpc.ChainUnaryInterceptor(
    logging.UnaryServerInterceptor(logger),
)
```

### Symptom 2: Duplicate Request IDs

**Problem:** Multiple requests have the same request ID.

**Diagnosis:**
```bash
# Find duplicate request IDs
cat /var/log/acm/acm.log | jq -r '.request_id' | sort | uniq -d
```

**Solutions:**

**1. Verify UUID generation:**

Request IDs use UUID v7 (time-ordered). Duplicates should be extremely rare.

**2. Check for manual ID override:**

Ensure code isn't manually setting duplicate IDs.

### Symptom 3: Request ID Not Propagated

**Problem:** Request ID changes between service calls.

**Diagnosis:**
```bash
# Trace a request through logs
grep "request_id.*01HBAG" /var/log/acm/acm.log
```

**Solutions:**

**1. Use gRPC metadata propagation:**
```go
// Client calls should propagate context
resp, err := client.Method(ctx, request)  // ctx contains request ID
```

**2. Verify interceptors are configured:**

Both client and server interceptors must be enabled.

---

## Debugging Tools

### Log Analysis Commands

**View logs in real-time:**
```bash
tail -f /var/log/acm/acm.log
```

**Pretty-print JSON logs:**
```bash
tail -f /var/log/acm/acm.log | jq '.'
```

**Filter by level:**
```bash
cat /var/log/acm/acm.log | jq 'select(.level == "ERROR")'
```

**Filter by component:**
```bash
cat /var/log/acm/acm.log | jq 'select(.component == "rotation")'
```

**Filter by request ID:**
```bash
cat /var/log/acm/acm.log | jq 'select(.request_id == "01HBAG5E7V9Q")'
```

**Find slow operations:**
```bash
cat /var/log/acm/acm.log | jq 'select(.duration_ms > 100)'
```

**Count log entries by level:**
```bash
cat /var/log/acm/acm.log | jq -r '.level' | sort | uniq -c
```

**Extract errors with context:**
```bash
cat /var/log/acm/acm.log | jq 'select(.level == "ERROR") | {time, component, msg, error}'
```

### Performance Analysis

**Measure log throughput:**
```bash
# Count lines per second
tail -f /var/log/acm/acm.log | pv -l -i 1 > /dev/null
```

**Find high-frequency log sources:**
```bash
cat /var/log/acm/acm.log | jq -r '.component' | sort | uniq -c | sort -rn
```

**Analyze operation durations:**
```bash
cat /var/log/acm/acm.log | jq -r 'select(.duration_ms) | [.operation, .duration_ms] | @tsv' | \
    awk '{sum[$1]+=$2; count[$1]++} END {for (op in sum) print op, sum[op]/count[op]}'
```

### Configuration Verification

**Check current logging config:**
```bash
# Environment variables
env | grep ACM_LOG | sort

# Effective config (check startup logs)
./acm-service 2>&1 | grep -A 10 "logging config"
```

**Validate JSON format:**
```bash
# Verify all logs are valid JSON
cat /var/log/acm/acm.log | jq empty
# No output means valid JSON
```

---

## FAQ

### Q: Why aren't my debug logs appearing?

**A:** Check log level:
```bash
export ACM_LOG_LEVEL=debug
```

### Q: How do I log to both stdout and a file?

**A:** Set output mode to "both":
```bash
export ACM_LOG_OUTPUT=both
export ACM_LOG_FILE=/var/log/acm/acm.log
```

### Q: How do I rotate logs immediately?

**A:** Future feature - send SIGHUP:
```bash
kill -HUP $(pidof acm-service)
```

Currently, logs rotate automatically based on size/age.

### Q: How do I find all logs for a specific request?

**A:** Filter by request_id:
```bash
cat /var/log/acm/acm.log | jq 'select(.request_id == "YOUR_REQUEST_ID")'
```

### Q: Why are my logs not in JSON format?

**A:** Check format setting:
```bash
export ACM_LOG_FORMAT=json
```

Default in production should be JSON.

### Q: How do I reduce log verbosity in production?

**A:**
```bash
export ACM_LOG_LEVEL=info  # or warn
export ACM_LOG_FORMAT=json
```

### Q: Where are rotated logs stored?

**A:** Same directory as main log:
```bash
ls /var/log/acm/
# acm.log (current)
# acm.log.2025-11-17-143245 (rotated)
# acm.log.2025-11-16-102314.gz (rotated, compressed)
```

### Q: How do I enable logging for a specific component only?

**A:** Currently requires code-level configuration:
```go
config := logging.DefaultConfig()
config.Level = "warn"  // Global: minimal logging
config.SetComponentLevel("rotation", "debug")  // Component: verbose
```

### Q: Can I change log level without restarting?

**A:** Not currently supported. Requires service restart.

Future feature: Hot-reload via SIGHUP.

### Q: How do I know if sensitive data is being logged?

**A:** Automatic redaction is enabled for common sensitive fields. Review logs for:
- `password`, `token`, `secret`, `api_key` → Should show `[REDACTED]`
- Email addresses → Should show `***@domain.com`
- Tokens → Should show only prefix (e.g., `ghp_12345678...`)

**Test redaction:**
```bash
cat /var/log/acm/acm.log | grep -i "password\|token\|secret"
```

### Q: What's the performance impact of debug logging?

**A:** Debug logging can increase CPU by 5-10% and log volume significantly.

**Recommendation:** Use `info` in production, `debug` only for troubleshooting.

### Q: How do I troubleshoot missing logs in containers?

**A:**
1. Ensure logs go to stdout: `export ACM_LOG_OUTPUT=stdout`
2. Check container logs: `docker logs <container>`
3. Verify log format: `export ACM_LOG_FORMAT=json`

### Q: Why are timestamps in UTC?

**A:** All timestamps are in UTC for consistency across deployments.

If you need local time, convert:
```bash
cat /var/log/acm/acm.log | jq -r '.time' | xargs -I{} date -d {} +"%Y-%m-%d %H:%M:%S %Z"
```

---

## Getting Help

If you encounter issues not covered here:

1. **Check logs** - Often contain error details
2. **Check configuration** - Verify environment variables
3. **Enable debug logging** - Get detailed diagnostic info
4. **Review documentation:**
   - [Logging Architecture](./logging-architecture.md)
   - [Logging Configuration](./logging-configuration.md)
   - [Logging Best Practices](./logging-best-practices.md)

5. **File an issue** - Provide:
   - ACM version
   - Environment variables
   - Sample logs (redact sensitive data!)
   - Steps to reproduce

---

## Appendix: Log Schema

### Standard Fields

Every log entry includes:

```json
{
  "time": "2025-11-17T14:32:45.123456789Z",
  "level": "INFO",
  "msg": "operation completed",
  "service": "acm",
  "version": "0.3.0",
  "hostname": "acm-prod-01",
  "pid": 12345,
  "component": "rotation"
}
```

### Optional Fields

Depending on context:

```json
{
  "request_id": "01HBAG5E7V9QZXN8J4WTRGP0QX",
  "duration_ms": 156,
  "error": "connection timeout",
  "credential_id_hash": "sha256:a1b2c3d4...",
  "site": "github.com",
  "username": "alice",
  "operation": "github_rotation",
  "status_code": 200
}
```

---

**End of Logging Troubleshooting Guide**
