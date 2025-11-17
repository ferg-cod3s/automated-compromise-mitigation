# Getting Started with ACM

**Automated Compromise Mitigation (ACM)** is a local-first credential breach response tool that helps you detect and rotate compromised credentials using your password manager.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Architecture](#architecture)
- [Security](#security)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required

1. **Go 1.21 or later**
   ```bash
   go version  # Should show 1.21 or higher
   ```

2. **One of the following password managers:**
   - Bitwarden CLI (`bw`)
   - 1Password CLI v2 (`op`)

### Optional

- `make` for simplified build commands
- Protocol Buffer compiler (`protoc`) for development

---

## Installation

### From Source

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ferg-cod3s/automated-compromise-mitigation.git
   cd automated-compromise-mitigation
   ```

2. **Build the binaries:**
   ```bash
   make build
   ```

   Or manually:
   ```bash
   go build -o bin/acm-service ./cmd/acm-service
   go build -o bin/acm-cli ./cmd/acm-cli
   ```

3. **Verify installation:**
   ```bash
   ./bin/acm-service --version
   ./bin/acm-cli version
   ```

---

## Quick Start

### 1. Install and Configure Password Manager

**For Bitwarden:**
```bash
# Install Bitwarden CLI
npm install -g @bitwarden/cli

# Log in
bw login

# Unlock vault
bw unlock
# Copy the session key and export it:
export BW_SESSION="your-session-key-here"
```

**For 1Password:**
```bash
# Install 1Password CLI v2
# Download from: https://developer.1password.com/docs/cli/get-started/

# Sign in
op signin

# Verify
op whoami
```

### 2. Start ACM Service

```bash
./bin/acm-service
```

You should see:
```
╔═══════════════════════════════════════════════════════════╗
║  ACM Service - Automated Compromise Mitigation           ║
║  Local-First Credential Breach Response                  ║
╚═══════════════════════════════════════════════════════════╝

✓ mTLS certificates generated
✓ Using Bitwarden
✓ acm-service ready and listening on 127.0.0.1:8443 (mTLS enabled)
```

**Note:** On first run, ACM automatically generates mTLS certificates in `~/.acm/certs`.

### 3. Check Service Health

In a new terminal:
```bash
./bin/acm-cli health
```

Expected output:
```
ACM Service Health Check
==================================================
Status: HEALTH_STATUS_HEALTHY
Message: All systems operational

Components:
  ✓ credential_service
  ✓ audit_logger
  ✓ him_service
```

### 4. Detect Compromised Credentials

```bash
./bin/acm-cli detect
```

This scans your password vault for credentials exposed in known data breaches.

Example output:
```
Detection Results: Successfully detected 3 compromised credentials
======================================================================

Found 3 compromised credential(s):

1. Site: example.com
   Username: user@example.com
   Breach: Example Breach 2024
   Date: 2024-03-15
   Severity: high
   ID Hash: a1b2c3d4...

2. Site: testsite.com
   Username: test@test.com
   Breach: TestSite Leak 2023
   Date: 2023-11-20
   Severity: medium
   ID Hash: e5f6g7h8...

To rotate a credential, use: acm rotate <id-hash>
```

### 5. Rotate a Compromised Credential

```bash
./bin/acm-cli rotate a1b2c3d4...
```

This generates a strong password and updates it in your vault.

Example output:
```
Rotating credential: a1b2c3d4...
Using password policy:
  Length: 16
  Uppercase: true
  Lowercase: true
  Numbers: true
  Symbols: true

✓ Credential rotated successfully!
Status: Password successfully updated in vault

New password has been updated in your vault.
⚠ IMPORTANT: The password manager will sync this change.
```

---

## Usage

### CLI Commands

#### Health Check
```bash
acm-cli health
```
Verifies the ACM service is running and all components are operational.

#### Detect Compromised Credentials
```bash
acm-cli detect
```
Scans your password vault for credentials in known breaches.

**Requirements:**
- Password manager vault must be unlocked
- Password manager must support breach detection

#### Rotate Credential
```bash
acm-cli rotate <credential-id-hash>
```
Rotates a specific credential with a generated secure password.

**Password Policy (default):**
- Length: 16 characters
- Uppercase letters: Required
- Lowercase letters: Required
- Numbers: Required
- Symbols: Required

#### List Credentials
```bash
acm-cli list
```
Lists credentials in your vault (Phase I: limited implementation).

#### Version
```bash
acm-cli version
```
Shows ACM version information.

---

## Architecture

### Components

1. **ACM Service** (`acm-service`)
   - gRPC server running on localhost:8443
   - mTLS authentication
   - Business logic for credential rotation
   - Audit logging with Ed25519 signatures

2. **CLI Client** (`acm-cli`)
   - Command-line interface
   - mTLS client authentication
   - Human-readable output

3. **Password Manager Integration**
   - CLI subprocess invocation
   - Zero-knowledge architecture
   - No access to master passwords

### Data Flow

```
User → acm-cli → (mTLS) → acm-service → Password Manager CLI → Vault
                                ↓
                          Audit Logger (Ed25519)
```

### Security Architecture

- **Local-First:** All processing on your device
- **Zero-Knowledge:** No access to master passwords or vault keys
- **mTLS:** Mutual TLS 1.3 authentication
- **Audit Trail:** Cryptographically signed event logs

---

## Security

### Zero-Knowledge Principles

ACM **never** accesses:
- Your master password
- Vault encryption keys
- Decrypted passwords (except during rotation)

### Certificate Management

mTLS certificates are stored in `~/.acm/certs/`:
- `ca-cert.pem` - Certificate Authority
- `ca-key.pem` - CA private key
- `server-cert.pem` - Service certificate
- `server-key.pem` - Service private key
- `client-cert.pem` - CLI client certificate
- `client-key.pem` - Client private key

**Permissions:** `chmod 600` (owner read/write only)

### Audit Logging

All rotation events are logged with:
- Ed25519 cryptographic signatures
- Timestamp
- Credential ID (hashed for privacy)
- Operation status
- Error details (if failed)

Logs are currently in-memory (Phase I). SQLite persistence coming in Phase II.

---

## Troubleshooting

### Service Won't Start

**Error: "Failed to start gRPC server"**

Solution:
```bash
# Check if port 8443 is in use
lsof -i :8443

# Kill the process if needed
pkill acm-service

# Restart
./bin/acm-service
```

### Password Manager Not Detected

**Error: "No password manager configured"**

Solution:
```bash
# For Bitwarden
which bw
bw login
bw unlock

# For 1Password
which op
op signin
```

### Vault Locked Error

**Error: "Vault is locked. Please unlock..."**

Solution:
```bash
# Bitwarden
bw unlock
export BW_SESSION="your-session-key"

# 1Password
op signin
```

### Certificate Errors

**Error: "Failed to load client certificate"**

Solution:
```bash
# Regenerate certificates
rm -rf ~/.acm/certs
./bin/acm-service  # Will auto-generate on startup
```

### Connection Refused

**Error: "Failed to connect to service"**

Solution:
```bash
# Ensure service is running
ps aux | grep acm-service

# Check service logs
tail -f /tmp/acm-service.log

# Restart service
./bin/acm-service
```

---

## Advanced Usage

### Running Tests

```bash
# Unit tests
go test ./internal/crs -v
go test ./internal/audit -v

# Integration tests
go test ./test/integration -v

# All tests
make test
```

### Building from Scratch

```bash
# Clean previous builds
make clean

# Generate proto files
make proto

# Build binaries
make build

# Run tests
make test
```

### Development Mode

```bash
# Run service with verbose logging
./bin/acm-service --verbose

# Run CLI with debug output
./bin/acm-cli --debug detect
```

---

## Next Steps

- Read the [Architecture Documentation](../acm-tad.md)
- Review [Security Planning](../acm-security-planning.md)
- Check the [Threat Model](../acm-threat-model.md)
- Join community discussions (coming soon)

---

## Getting Help

- **Issues:** https://github.com/ferg-cod3s/automated-compromise-mitigation/issues
- **Documentation:** https://github.com/ferg-cod3s/automated-compromise-mitigation
- **Security:** See [SECURITY.md](../SECURITY.md) for vulnerability reporting

---

**Phase I Status:** Core functionality complete
**Ready for:** Alpha testing with Bitwarden or 1Password

**Note:** This is alpha software. Use in a test environment first!
