# ACM Quick Start Guide

Get up and running with ACM development in 5 minutes.

## Prerequisites Check

```bash
# Check Go version (need 1.21+)
go version

# Check if make is installed
make --version

# Check if OpenSSL is installed
openssl version
```

If any are missing, see [BUILD.md](../BUILD.md#prerequisites) for installation instructions.

---

## Quick Setup (5 Steps)

### Step 1: Set Up Development Environment

This installs all required development tools (golangci-lint, gosec, protoc-gen-go, etc.):

```bash
make setup-dev
```

**Time:** ~2-3 minutes

### Step 2: Generate mTLS Certificates

Generate self-signed certificates for local development:

```bash
make cert-gen
```

**Time:** ~5 seconds

### Step 3: Build Binaries

Build both the service daemon and CLI client:

```bash
make build
```

**Time:** ~30 seconds

Binaries will be in `build/`:
- `build/acm-service` - Service daemon
- `build/acm-cli` - CLI client

### Step 4: Run Tests

Verify everything works:

```bash
make test
```

**Time:** Depends on number of tests

### Step 5: Start Development

Start the service in development mode:

```bash
make dev
```

In another terminal, start the CLI:

```bash
make dev-cli
```

---

## Common Commands

### Building

```bash
make build              # Build both service and CLI
make build-service      # Build service only
make build-cli          # Build CLI only
```

### Testing

```bash
make test               # Run tests
make test-coverage      # Run tests with coverage report
make lint               # Run linters
```

### Development

```bash
make dev                # Start service in dev mode
make dev-cli            # Start CLI in dev mode
make cert-gen           # Regenerate certificates
```

### Cleanup

```bash
make clean              # Remove build artifacts
```

---

## Project Structure

```
automated-compromise-mitigation/
├── cmd/
│   ├── acm-service/         # Service daemon entry point
│   └── acm-cli/             # CLI client entry point
│
├── internal/
│   ├── service/             # Core service logic
│   ├── crs/                 # Credential Remediation Service
│   ├── acvs/                # Compliance Validation Service
│   ├── auth/                # Authentication (mTLS)
│   └── config/              # Configuration management
│
├── api/proto/               # Protocol Buffer definitions
├── config/                  # Configuration files
├── certs/                   # mTLS certificates (generated)
├── build/                   # Build output (gitignored)
└── scripts/                 # Build and setup scripts
```

---

## Configuration

Example configuration files are in `config/`:

- `config/acm.example.yaml` - Production configuration template
- `config/dev.yaml` - Development configuration

Copy and customize:

```bash
cp config/acm.example.yaml config/acm.yaml
# Edit config/acm.yaml with your settings
```

---

## Troubleshooting

### Problem: `golangci-lint: command not found`

**Solution:**
```bash
make setup-dev
```

### Problem: `protoc: command not found`

**Solution:**

macOS:
```bash
brew install protobuf
```

Linux:
```bash
sudo apt-get install -y protobuf-compiler
```

### Problem: `Cannot connect to service`

**Solution:**

1. Check certificates exist:
   ```bash
   ls certs/
   ```

2. If missing, generate:
   ```bash
   make cert-gen
   ```

3. Verify service is running:
   ```bash
   make dev
   ```

### Problem: Tests failing

**Solution:**

1. Clean and rebuild:
   ```bash
   make clean
   make build
   make test
   ```

2. If still failing, check test output for specific errors

---

## Next Steps

1. **Read the Architecture**: [acm-tad.md](../acm-tad.md)
2. **Understand Security**: [acm-security-planning.md](../acm-security-planning.md)
3. **Review Threat Model**: [acm-threat-model.md](../acm-threat-model.md)
4. **Build System Details**: [BUILD.md](../BUILD.md)

---

## Getting Help

- **Build Issues**: See [BUILD.md](../BUILD.md#troubleshooting)
- **GitHub Issues**: [Report a bug](https://github.com/ferg-cod3s/automated-compromise-mitigation/issues)
- **Documentation**: [00-INDEX.md](../00-INDEX.md)

---

**Ready to build?** Run `make help` to see all available commands!
