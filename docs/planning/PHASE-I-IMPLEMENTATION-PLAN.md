# Phase I Implementation Plan - ACM Project
# Automated Compromise Mitigation (ACM)

**Version:** 1.0
**Date:** November 2025
**Status:** Active Development
**Duration:** 24 weeks (6 months)

---

## Executive Summary

This document provides the complete implementation roadmap for **Phase I (MVP)** of the ACM project. Phase I focuses on delivering the core Credential Remediation Service (CRS) with zero-knowledge security architecture, local-first operation, and comprehensive security controls.

### Phase I Goals

1. **Core Service Operational** - ACM service daemon with mTLS API
2. **Password Manager Integration** - 1Password and Bitwarden CLI support
3. **Security Foundation** - Memory protection, audit logging, certificate management
4. **OpenTUI Client** - Terminal interface for developers and power users
5. **Production-Ready Security** - All critical security controls implemented

---

## Implementation Timeline

### Week 1-2: Foundation & Development Environment
**Focus:** Project infrastructure, tooling, and development workflow

#### Atomic Tasks

1. **Project Initialization**
   - [ ] Initialize Go module (`go mod init github.com/acm-project/acm`)
   - [ ] Create `.gitignore` for Go project
   - [ ] Set up project directory structure (15+ directories)
   - [ ] Create `tools.go` for development dependencies
   - [ ] Initialize `go.sum` with dependency checksums

2. **Build Automation**
   - [ ] Create `Makefile` with targets:
     - `make build` - Build all binaries
     - `make test` - Run all tests
     - `make lint` - Run linters
     - `make security-scan` - Run security scanners
     - `make generate` - Generate protobuf code
     - `make clean` - Clean build artifacts
   - [ ] Create `scripts/setup-dev.sh` - Development environment setup
   - [ ] Create `scripts/run-tests.sh` - Comprehensive test runner

3. **Protocol Buffer Definitions**
   - [ ] Create `api/proto/acm/v1/service.proto` - Main service definition
   - [ ] Create `api/proto/acm/v1/crs.proto` - CRS service definition
   - [ ] Create `api/proto/acm/v1/audit.proto` - Audit service definition
   - [ ] Create `api/proto/acm/v1/him.proto` - HIM service definition
   - [ ] Create `api/proto/acm/v1/common.proto` - Common types
   - [ ] Set up protobuf generation in Makefile
   - [ ] Generate Go code from protobuf definitions

4. **Development Tooling**
   - [ ] Configure `golangci-lint` with security linters
   - [ ] Create `.golangci.yml` configuration
   - [ ] Set up `gosec` for security scanning
   - [ ] Configure `gofmt` and `goimports`
   - [ ] Create pre-commit hooks (optional)

5. **CI/CD Pipeline**
   - [ ] Create `.github/workflows/ci.yml` - Continuous integration
   - [ ] Create `.github/workflows/security.yml` - Security scanning
   - [ ] Create `.github/workflows/release.yml` - Release automation
   - [ ] Set up code coverage reporting (Codecov)
   - [ ] Configure dependency scanning (Dependabot)

6. **Certificate Infrastructure**
   - [ ] Create `scripts/generate-certs.sh` - Certificate generation
   - [ ] Create `configs/ca-config.json` - CA configuration for cfssl
   - [ ] Create `configs/tls/` directory structure
   - [ ] Document certificate generation process
   - [ ] Create certificate renewal workflow

7. **Database Schema**
   - [ ] Create `scripts/setup-sqlite.sh` - Database initialization
   - [ ] Create `internal/storage/audit/schema.sql` - Audit log schema
   - [ ] Create database migration framework
   - [ ] Document database schema

8. **Configuration Management**
   - [ ] Create `configs/service.yaml.example` - Service config template
   - [ ] Create `configs/client.yaml.example` - Client config template
   - [ ] Create configuration validation logic
   - [ ] Document all configuration options

9. **Documentation**
   - [ ] Create `docs/development/SETUP.md` - Development setup guide
   - [ ] Create `docs/development/ARCHITECTURE.md` - Architecture overview
   - [ ] Create `SECURITY.md` - Security policy
   - [ ] Update `README.md` with project overview
   - [ ] Create `CONTRIBUTING.md` - Contribution guidelines

**Deliverables:**
- ✅ Complete project scaffolding
- ✅ Working `make build` and `make test`
- ✅ CI/CD pipeline running on GitHub
- ✅ Certificate generation workflow
- ✅ Database schema initialized

---

### Week 3-4: Core Service - mTLS Authentication
**Focus:** Secure client-server communication with mutual TLS

#### Atomic Tasks

1. **TLS Server Implementation**
   - [ ] Create `internal/service/auth/mtls.go` - mTLS server
   - [ ] Implement TLS 1.3 configuration
   - [ ] Add localhost-only binding (`127.0.0.1:8443`)
   - [ ] Configure strong cipher suites (AES-256-GCM, ChaCha20-Poly1305)
   - [ ] Add certificate pinning logic
   - [ ] Write unit tests for TLS configuration

2. **Certificate Management**
   - [ ] Create `internal/crypto/cert/manager.go` - Certificate manager
   - [ ] Implement certificate loading (server cert, CA cert)
   - [ ] Add certificate validation logic
   - [ ] Implement certificate revocation list (CRL) support
   - [ ] Create certificate renewal workflow
   - [ ] Write tests for certificate operations

3. **Client Certificate Authentication**
   - [ ] Implement client certificate validation
   - [ ] Add certificate fingerprint verification
   - [ ] Create client certificate allowlist
   - [ ] Add certificate expiration checking
   - [ ] Implement revocation checking
   - [ ] Write integration tests

4. **JWT Session Management**
   - [ ] Create `internal/service/auth/jwt.go` - JWT token manager
   - [ ] Implement token generation (Ed25519 signing)
   - [ ] Add token validation with expiration (15-30 min)
   - [ ] Create token refresh endpoint
   - [ ] Implement in-memory revocation list
   - [ ] Write tests for JWT operations

5. **gRPC Server Setup**
   - [ ] Create `cmd/acm-service/main.go` - Service entry point
   - [ ] Implement gRPC server with mTLS credentials
   - [ ] Add authentication interceptor
   - [ ] Add rate limiting interceptor
   - [ ] Configure connection timeout
   - [ ] Add graceful shutdown

6. **gRPC Client Library**
   - [ ] Create `pkg/acmclient/client.go` - Client library
   - [ ] Implement mTLS client configuration
   - [ ] Add certificate loading from OS keychain
   - [ ] Implement connection pooling
   - [ ] Add retry logic with exponential backoff
   - [ ] Write client integration tests

7. **Integration Testing**
   - [ ] Create `test/integration/mtls_test.go`
   - [ ] Test: Valid client cert → connection succeeds
   - [ ] Test: Invalid client cert → connection rejected
   - [ ] Test: Expired cert → connection rejected
   - [ ] Test: Revoked cert → connection rejected
   - [ ] Test: Certificate pinning enforcement
   - [ ] Test: JWT token lifecycle

**Security Validation:**
- [ ] Verify localhost-only binding (netstat check)
- [ ] Verify TLS 1.3 enforcement (TLS handshake analysis)
- [ ] Verify strong ciphers only (cipher suite validation)
- [ ] Verify certificate pinning works (fingerprint mismatch test)
- [ ] Verify JWT expiration works (token timeout test)

**Deliverables:**
- ✅ Working mTLS server
- ✅ Client authentication with certificates
- ✅ JWT session management
- ✅ Integration tests passing

---

### Week 5-6: Password Manager CLI Integration
**Focus:** Secure integration with 1Password and Bitwarden CLIs

#### Atomic Tasks

1. **Password Manager Interface**
   - [ ] Create `internal/pwmanager/interface.go` - Common interface
   - [ ] Define `PasswordManager` interface methods
   - [ ] Create error types for CLI operations
   - [ ] Define credential data structures
   - [ ] Document interface contract

2. **Secure CLI Executor**
   - [ ] Create `internal/pwmanager/executor.go` - Secure CLI execution
   - [ ] Implement parameterized command execution (no shell)
   - [ ] Add command timeout (30s default)
   - [ ] Implement environment variable sanitization
   - [ ] Add output sanitization (remove credentials before logging)
   - [ ] Write executor tests

3. **1Password Integration**
   - [ ] Create `internal/pwmanager/onepassword/client.go`
   - [ ] Implement `DetectCompromised()` - Query for exposed items
   - [ ] Implement `GetItem()` - Retrieve item details
   - [ ] Implement `UpdateItem()` - Update password
   - [ ] Implement `GeneratePassword()` - Delegate to `op` CLI
   - [ ] Parse JSON output from `op` CLI
   - [ ] Write 1Password integration tests

4. **Bitwarden Integration**
   - [ ] Create `internal/pwmanager/bitwarden/client.go`
   - [ ] Implement `DetectCompromised()` - Query exposed items
   - [ ] Implement `GetItem()` - Retrieve item details
   - [ ] Implement `UpdateItem()` - Update password
   - [ ] Implement session management (BW_SESSION)
   - [ ] Parse JSON output from `bw` CLI
   - [ ] Write Bitwarden integration tests

5. **CLI Binary Verification**
   - [ ] Create `internal/pwmanager/verify.go` - Binary verification
   - [ ] Implement checksum verification (SHA-256)
   - [ ] Add CLI version detection
   - [ ] Implement version compatibility checking
   - [ ] Add warning for unexpected CLI changes
   - [ ] Write verification tests

6. **Input Validation & Sanitization**
   - [ ] Create `internal/pwmanager/sanitize.go`
   - [ ] Implement credential ID validation (regex whitelist)
   - [ ] Add shell metacharacter detection
   - [ ] Implement command injection prevention
   - [ ] Add output sanitization (redact credentials)
   - [ ] Write fuzzing tests for sanitization

7. **Error Handling**
   - [ ] Define error types (VaultLocked, CLINotFound, etc.)
   - [ ] Implement retry logic with exponential backoff
   - [ ] Add timeout error handling
   - [ ] Create user-friendly error messages
   - [ ] Write error handling tests

**Security Validation:**
- [ ] Test CLI injection prevention (metacharacters sanitized)
- [ ] Verify parameterized execution (no shell invocation)
- [ ] Test binary verification (tampered binary detected)
- [ ] Verify credential sanitization (no leaks in logs)
- [ ] Test timeout handling (hung CLI processes killed)

**Deliverables:**
- ✅ Working 1Password CLI integration
- ✅ Working Bitwarden CLI integration
- ✅ Secure CLI execution framework
- ✅ Binary verification
- ✅ Integration tests passing

---

### Week 7-8: CRS - Credential Remediation Service
**Focus:** Core credential rotation with memory protection

#### Atomic Tasks

1. **CRS Service Core**
   - [ ] Create `internal/service/crs/service.go` - CRS service
   - [ ] Implement `DetectCompromised()` RPC handler
   - [ ] Implement `RotateCredential()` RPC handler
   - [ ] Implement `GetRotationStatus()` RPC handler
   - [ ] Add service initialization logic
   - [ ] Write CRS service tests

2. **Breach Detection**
   - [ ] Create `internal/service/crs/detector.go`
   - [ ] Implement password manager breach query
   - [ ] Parse breach report data
   - [ ] Filter and deduplicate compromised items
   - [ ] Add rotation history checking (avoid duplicates)
   - [ ] Write detection tests

3. **Secure Password Generation**
   - [ ] Create `internal/service/crs/generator.go`
   - [ ] Implement `crypto/rand` based generation
   - [ ] Add configurable password policies (length, complexity)
   - [ ] Implement policy validation
   - [ ] Add entropy measurement
   - [ ] Write generator tests

4. **Vault Update Logic**
   - [ ] Create `internal/service/crs/rotator.go`
   - [ ] Implement atomic vault update workflow
   - [ ] Add pre-rotation verification (confirm item exists)
   - [ ] Implement rotation via password manager CLI
   - [ ] Add post-rotation verification (confirm update)
   - [ ] Implement rollback on failure
   - [ ] Write rotation tests

5. **Secure Memory Handling**
   - [ ] Create `internal/crypto/memory/secure.go`
   - [ ] Implement memory locking (`syscall.Mlock()`)
   - [ ] Implement explicit zeroing (`memguard.Wipe()`)
   - [ ] Create `SecureBuffer` wrapper
   - [ ] Add locked memory allocation
   - [ ] Minimize credential lifetime (< 5 seconds)
   - [ ] Write memory security tests

6. **Transaction Management**
   - [ ] Create `internal/service/crs/transaction.go`
   - [ ] Implement transaction state tracking
   - [ ] Add atomic commit/rollback
   - [ ] Implement idempotency checking
   - [ ] Add transaction timeout handling
   - [ ] Write transaction tests

7. **Error Recovery**
   - [ ] Create error types (RotationError with codes)
   - [ ] Implement rollback logic (restore old password)
   - [ ] Add partial rotation recovery
   - [ ] Implement error reporting to clients
   - [ ] Write error recovery tests

**Security Validation:**
- [ ] Memory dump test (verify no credentials in dump)
- [ ] Verify memory locking (pages not swapped)
- [ ] Test explicit zeroing (buffers contain zeros after use)
- [ ] Verify minimal credential lifetime (< 5s in memory)
- [ ] Test atomic transactions (no partial rotations)

**Deliverables:**
- ✅ Working credential detection
- ✅ Secure password generation
- ✅ Vault update with verification
- ✅ Memory protection implemented
- ✅ Transaction rollback capability

---

### Week 9-10: Audit Logging with Cryptographic Signing
**Focus:** Tamper-evident audit trail

#### Atomic Tasks

1. **SQLite Database Setup**
   - [ ] Create `internal/storage/audit/schema.sql`
   - [ ] Define `audit_events` table schema
   - [ ] Add indexes for performance
   - [ ] Create database migrations
   - [ ] Implement database connection pooling
   - [ ] Write database tests

2. **Audit Logger Core**
   - [ ] Create `internal/storage/audit/logger.go`
   - [ ] Implement `Log()` - Write audit entry
   - [ ] Implement `Query()` - Query audit logs
   - [ ] Implement `VerifyIntegrity()` - Verify chain
   - [ ] Add database initialization
   - [ ] Write logger tests

3. **Cryptographic Signing**
   - [ ] Create `internal/crypto/signing/ed25519.go`
   - [ ] Implement Ed25519 key generation
   - [ ] Implement audit entry signing
   - [ ] Add signature verification
   - [ ] Create key management (secure storage)
   - [ ] Write signing tests

4. **Merkle Tree Linking**
   - [ ] Implement entry hash computation
   - [ ] Link each entry to previous entry hash
   - [ ] Add chain initialization (genesis entry)
   - [ ] Implement chain verification algorithm
   - [ ] Write chain verification tests

5. **Credential ID Hashing**
   - [ ] Implement SHA-256 hashing for credential IDs
   - [ ] Add salt generation for hashing
   - [ ] Never log plaintext credential IDs
   - [ ] Write hashing tests

6. **Sensitive Field Encryption**
   - [ ] Create `internal/crypto/encrypt/aes.go`
   - [ ] Implement AES-256-GCM encryption
   - [ ] Encrypt sensitive audit log fields
   - [ ] Add key derivation (from service key)
   - [ ] Write encryption tests

7. **Log Rotation & Retention**
   - [ ] Implement log rotation (size-based)
   - [ ] Add retention policy enforcement
   - [ ] Create archive functionality
   - [ ] Implement old log cleanup
   - [ ] Write retention tests

8. **Audit CLI Commands**
   - [ ] Implement `acm audit query --since 7d`
   - [ ] Implement `acm audit verify`
   - [ ] Implement `acm audit export --format json`
   - [ ] Add filtering options (by date, action, status)
   - [ ] Write CLI tests

**Security Validation:**
- [ ] Test signature verification (tampered entry detected)
- [ ] Test chain verification (broken chain detected)
- [ ] Verify credential ID hashing (no plaintext IDs)
- [ ] Test encryption (sensitive fields encrypted)
- [ ] Verify append-only (no in-place updates)

**Deliverables:**
- ✅ Cryptographically signed audit logs
- ✅ Merkle tree chain linking
- ✅ Integrity verification command
- ✅ Credential ID hashing
- ✅ Sensitive field encryption

---

### Week 11-12: HIM (Human-in-the-Middle) Manager
**Focus:** User intervention for MFA/CAPTCHA

#### Atomic Tasks

1. **HIM State Machine**
   - [ ] Create `internal/service/him/statemachine.go`
   - [ ] Define state transitions (IDLE → AWAITING_INPUT → VALIDATING → RESUMING)
   - [ ] Implement state change logic
   - [ ] Add timeout handling (5 min default)
   - [ ] Write state machine tests

2. **HIM Manager Core**
   - [ ] Create `internal/service/him/manager.go`
   - [ ] Implement `RequiresHIM()` - Detect MFA/CAPTCHA need
   - [ ] Implement `PromptUser()` - Send prompt to client
   - [ ] Implement `ResumeAutomation()` - Resume after user input
   - [ ] Add concurrent prompt handling
   - [ ] Write manager tests

3. **HIM Prompt Types**
   - [ ] Define MFA prompt structure (TOTP, SMS, Push)
   - [ ] Define CAPTCHA prompt structure
   - [ ] Define manual action prompt
   - [ ] Create prompt builder utilities
   - [ ] Write prompt tests

4. **gRPC Streaming**
   - [ ] Implement bidirectional gRPC stream for HIM
   - [ ] Add client stream handling
   - [ ] Implement server push for prompts
   - [ ] Add stream error handling
   - [ ] Write streaming tests

5. **Detection Logic**
   - [ ] Implement MFA detection (parse CLI output)
   - [ ] Implement CAPTCHA detection
   - [ ] Add ToS violation detection (future: ACVS)
   - [ ] Write detection tests

6. **Timeout Management**
   - [ ] Implement prompt timeout (5 min)
   - [ ] Add timeout warning (1 min remaining)
   - [ ] Handle timeout expiration (fail rotation)
   - [ ] Write timeout tests

7. **HIM Event Logging**
   - [ ] Log HIM events to audit trail
   - [ ] Record: type, duration, attempts, success/failure
   - [ ] Add context preservation across pause/resume
   - [ ] Write logging tests

**Deliverables:**
- ✅ HIM state machine
- ✅ MFA/CAPTCHA detection
- ✅ User prompt via gRPC stream
- ✅ Timeout handling
- ✅ HIM event logging

---

### Week 13-14: OpenTUI Client (Bubbletea)
**Focus:** Terminal user interface

#### Atomic Tasks

1. **TUI Framework Setup**
   - [ ] Create `clients/tui/main.go` - TUI entry point
   - [ ] Set up Bubbletea framework
   - [ ] Initialize Lipgloss styling
   - [ ] Configure Bubbles components
   - [ ] Create base model structure
   - [ ] Write TUI initialization tests

2. **Dashboard View**
   - [ ] Create `clients/tui/ui/dashboard.go`
   - [ ] Display service status (running/stopped)
   - [ ] Show compromised credential count
   - [ ] Display recent rotations
   - [ ] Add ACVS status indicator (disabled for Phase I)
   - [ ] Write dashboard tests

3. **Credential List View**
   - [ ] Create `clients/tui/ui/credentials.go`
   - [ ] List compromised credentials
   - [ ] Show metadata (site, username, breach date)
   - [ ] Add sorting and filtering
   - [ ] Implement pagination for large lists
   - [ ] Write credential list tests

4. **Rotation Workflow**
   - [ ] Create `clients/tui/ui/rotation.go`
   - [ ] Interactive credential selection
   - [ ] Rotation confirmation prompt
   - [ ] Progress indicator during rotation
   - [ ] Success/failure result display
   - [ ] Write rotation workflow tests

5. **HIM Prompt Handling**
   - [ ] Create `clients/tui/ui/him_prompt.go`
   - [ ] Display MFA/CAPTCHA prompts
   - [ ] Secure input field for codes
   - [ ] Timeout countdown display
   - [ ] Write HIM prompt tests

6. **Audit Log Viewer**
   - [ ] Create `clients/tui/ui/audit.go`
   - [ ] Display audit log entries
   - [ ] Add filtering by date/action
   - [ ] Implement search functionality
   - [ ] Add export functionality
   - [ ] Write audit viewer tests

7. **CLI Commands**
   - [ ] Create `clients/tui/commands/root.go` - Root command
   - [ ] Implement `acm status` - Service status
   - [ ] Implement `acm detect` - Detect compromised
   - [ ] Implement `acm rotate <id>` - Rotate credential
   - [ ] Implement `acm rotate --all` - Rotate all
   - [ ] Implement `acm audit` - View audit log
   - [ ] Implement `acm config` - Configuration management
   - [ ] Implement `acm cert renew` - Certificate renewal
   - [ ] Write CLI command tests

8. **Configuration Management**
   - [ ] Create `clients/tui/commands/config.go`
   - [ ] Implement `acm config show`
   - [ ] Implement `acm config set <key> <value>`
   - [ ] Implement `acm config validate`
   - [ ] Write config command tests

9. **gRPC Client Integration**
   - [ ] Create gRPC client connection
   - [ ] Implement certificate loading
   - [ ] Add retry logic
   - [ ] Handle connection errors gracefully
   - [ ] Write client integration tests

**Deliverables:**
- ✅ Working TUI with dashboard
- ✅ Credential detection and rotation UI
- ✅ HIM prompt handling
- ✅ Audit log viewer
- ✅ All CLI commands functional

---

### Week 15-16: Security Hardening
**Focus:** Comprehensive security testing and hardening

#### Atomic Tasks

1. **Memory Dump Testing**
   - [ ] Create `test/security/memory_test.go`
   - [ ] Generate memory dump during rotation
   - [ ] Search dump for credential patterns
   - [ ] Verify zero credentials found
   - [ ] Document memory dump testing procedure

2. **CLI Injection Testing**
   - [ ] Create `test/security/injection_test.go`
   - [ ] Test shell metacharacter injection
   - [ ] Test command separator injection (`;`, `&&`, `||`)
   - [ ] Test path traversal (`../../`)
   - [ ] Verify all injection attempts sanitized

3. **Fuzzing**
   - [ ] Set up go-fuzz for CLI parser
   - [ ] Fuzz gRPC message handlers
   - [ ] Fuzz password generation
   - [ ] Run 1M+ iterations
   - [ ] Document fuzzing results

4. **Certificate Validation Testing**
   - [ ] Test expired certificate rejection
   - [ ] Test wrong certificate rejection
   - [ ] Test revoked certificate rejection
   - [ ] Test certificate pinning enforcement
   - [ ] Document certificate testing

5. **Audit Log Tampering Detection**
   - [ ] Test signature verification (modified entry)
   - [ ] Test chain verification (deleted entry)
   - [ ] Test chain verification (inserted entry)
   - [ ] Verify all tampering detected

6. **SAST Scanning**
   - [ ] Run `gosec` security scanner
   - [ ] Run `semgrep` with security rules
   - [ ] Fix all critical/high findings
   - [ ] Document scan results
   - [ ] Add to CI/CD pipeline

7. **Dependency Scanning**
   - [ ] Run `go mod verify`
   - [ ] Scan dependencies with Snyk
   - [ ] Update vulnerable dependencies
   - [ ] Document dependency policy

8. **File Permission Validation**
   - [ ] Verify certificate files: 0600
   - [ ] Verify key files: 0600
   - [ ] Verify config directory: 0700
   - [ ] Verify audit database: 0600
   - [ ] Create permission validation script

9. **Penetration Testing**
   - [ ] Community security sprint planning
   - [ ] Internal penetration testing
   - [ ] Document findings
   - [ ] Fix identified vulnerabilities
   - [ ] Re-test to verify fixes

10. **Security Code Review**
    - [ ] Review all authentication code
    - [ ] Review all cryptographic code
    - [ ] Review all input validation
    - [ ] Review memory handling
    - [ ] Document review findings

**Security Checklist:**
- [ ] mTLS connections require valid client certificate
- [ ] Memory dumps contain zero plaintext credentials
- [ ] CLI injection attempts are sanitized
- [ ] Audit log tampering detected by verification
- [ ] No critical/high findings from SAST tools
- [ ] File permissions set correctly (0600 for certs/keys)
- [ ] All dependencies up-to-date, no CVEs
- [ ] Penetration testing complete with fixes

**Deliverables:**
- ✅ All security tests passing
- ✅ SAST scans clean
- ✅ Penetration testing complete
- ✅ Security hardening documented

---

### Week 17-18: Documentation & Testing
**Focus:** Comprehensive documentation and test coverage

#### Atomic Tasks

1. **API Documentation**
   - [ ] Generate protobuf API docs
   - [ ] Document all gRPC endpoints
   - [ ] Add request/response examples
   - [ ] Document error codes
   - [ ] Create `docs/api/README.md`

2. **Security Documentation**
   - [ ] Create `docs/security/ARCHITECTURE.md`
   - [ ] Document threat model
   - [ ] Document security controls
   - [ ] Create security testing guide
   - [ ] Update `SECURITY.md`

3. **User Documentation**
   - [ ] Create `docs/user/INSTALLATION.md`
   - [ ] Create `docs/user/QUICK-START.md`
   - [ ] Create `docs/user/CONFIGURATION.md`
   - [ ] Create `docs/user/TROUBLESHOOTING.md`
   - [ ] Create user FAQ

4. **Developer Documentation**
   - [ ] Create `docs/development/CONTRIBUTING.md`
   - [ ] Create `docs/development/ARCHITECTURE.md`
   - [ ] Create `docs/development/TESTING.md`
   - [ ] Create `docs/development/RELEASE.md`
   - [ ] Document code conventions

5. **Unit Test Coverage**
   - [ ] Achieve 80%+ coverage for `internal/service/`
   - [ ] Achieve 80%+ coverage for `internal/crypto/`
   - [ ] Achieve 80%+ coverage for `internal/pwmanager/`
   - [ ] Achieve 80%+ coverage for `internal/storage/`
   - [ ] Generate coverage reports

6. **Integration Tests**
   - [ ] End-to-end rotation workflow
   - [ ] mTLS authentication scenarios
   - [ ] Password manager integration
   - [ ] Audit log verification
   - [ ] HIM workflow

7. **E2E Tests**
   - [ ] Full user workflow (detect → rotate → verify)
   - [ ] Multi-client scenarios
   - [ ] Error recovery scenarios
   - [ ] Write E2E test documentation

**Deliverables:**
- ✅ Complete API documentation
- ✅ Security architecture documentation
- ✅ User guides (installation, quick start, config)
- ✅ Developer guides (contributing, architecture)
- ✅ 80%+ unit test coverage
- ✅ Comprehensive integration tests

---

### Week 19-20: Packaging & Release Preparation
**Focus:** Cross-platform packaging and release automation

#### Atomic Tasks

1. **GoReleaser Configuration**
   - [ ] Create `.goreleaser.yml`
   - [ ] Configure cross-platform builds (linux, macOS, Windows)
   - [ ] Configure architecture support (amd64, arm64)
   - [ ] Add checksum generation
   - [ ] Add code signing (future)

2. **systemd Service Files**
   - [ ] Create `build/systemd/acm-service.service`
   - [ ] Configure service restart policy
   - [ ] Add service dependencies
   - [ ] Test on Ubuntu/Fedora
   - [ ] Document systemd setup

3. **launchd Plists**
   - [ ] Create `build/launchd/com.acm-project.acm.plist`
   - [ ] Configure auto-start
   - [ ] Add service dependencies
   - [ ] Test on macOS
   - [ ] Document launchd setup

4. **Package Managers**
   - [ ] Create Homebrew formula
   - [ ] Create Debian package (.deb)
   - [ ] Create RPM package (.rpm)
   - [ ] Test package installation
   - [ ] Document package installation

5. **Release Automation**
   - [ ] Create GitHub release workflow
   - [ ] Add release notes generation
   - [ ] Configure artifact uploads
   - [ ] Add changelog generation
   - [ ] Test release workflow

6. **Reproducible Builds**
   - [ ] Configure build reproducibility
   - [ ] Document build process
   - [ ] Verify builds are reproducible
   - [ ] Add build verification instructions

**Deliverables:**
- ✅ Cross-platform binaries (Linux, macOS, Windows)
- ✅ systemd service files
- ✅ launchd plists
- ✅ Package manager support (Homebrew, apt, dnf)
- ✅ Automated release workflow
- ✅ Reproducible builds

---

### Week 21-24: Alpha Testing & Iteration
**Focus:** Community testing and iterative improvements

#### Atomic Tasks

1. **Internal Alpha Testing**
   - [ ] Core team testing (1 week)
   - [ ] Bug tracking and triage
   - [ ] Performance profiling
   - [ ] Fix critical bugs
   - [ ] Document known issues

2. **Community Alpha Program**
   - [ ] Recruit alpha testers (security-focused users)
   - [ ] Create alpha testing guide
   - [ ] Set up feedback channels (Discord, GitHub Discussions)
   - [ ] Distribute alpha binaries
   - [ ] Collect and triage feedback

3. **Bug Fixes**
   - [ ] Prioritize bugs (P0/P1/P2)
   - [ ] Fix critical bugs (P0)
   - [ ] Fix high-priority bugs (P1)
   - [ ] Create bug fix releases
   - [ ] Update changelog

4. **Performance Optimization**
   - [ ] Profile CPU usage
   - [ ] Profile memory usage
   - [ ] Optimize hot paths
   - [ ] Reduce startup time
   - [ ] Document performance benchmarks

5. **Security Hardening (Based on Feedback)**
   - [ ] Address security concerns from alpha testers
   - [ ] Additional penetration testing
   - [ ] Fix identified vulnerabilities
   - [ ] Update security documentation

6. **Third-Party Security Audit (Optional)**
   - [ ] Engage security audit firm
   - [ ] Provide codebase access
   - [ ] Review audit findings
   - [ ] Fix audit findings
   - [ ] Publish audit report

7. **Phase II Planning**
   - [ ] Review Phase I outcomes
   - [ ] Plan ACVS implementation
   - [ ] Legal NLP research
   - [ ] Community feedback integration
   - [ ] Create Phase II roadmap

**Deliverables:**
- ✅ Alpha testing complete
- ✅ Critical bugs fixed
- ✅ Performance optimizations applied
- ✅ Security audit (if conducted)
- ✅ Phase II roadmap

---

## Success Criteria

### Functional Requirements

- [ ] ACM service runs as daemon (systemd/launchd)
- [ ] OpenTUI client successfully connects via mTLS
- [ ] Detects compromised credentials from 1Password/Bitwarden
- [ ] Generates secure passwords (32 chars, crypto/rand)
- [ ] Updates vault entries via password manager CLI
- [ ] Verifies rotation success
- [ ] Logs all actions to cryptographically signed audit trail
- [ ] HIM workflow prompts user for MFA/CAPTCHA

### Security Requirements

- [ ] All mTLS connections use TLS 1.3
- [ ] Zero credential leaks in logs (automated secret detection passes)
- [ ] Memory protection verified (controlled memory dump test)
- [ ] Audit log integrity verification succeeds for 10,000+ entries
- [ ] SAST scans show zero critical/high findings
- [ ] Integration tests achieve 80%+ code coverage
- [ ] Security tests pass (memory, CLI injection, certificate validation)

### Performance Requirements

- [ ] Detect compromised credentials: < 2 seconds
- [ ] Generate secure password: < 100ms
- [ ] Update vault entry: < 3 seconds
- [ ] HIM prompt display: < 500ms
- [ ] Audit log write: < 50ms
- [ ] Service startup time: < 2 seconds
- [ ] Memory footprint: < 256MB (base), < 512MB (peak)

---

## Dependencies & Prerequisites

### Required Tools

- Go 1.21+
- Protocol Buffers (protoc) 3.21+
- golangci-lint v1.55+
- gosec v2.18+
- SQLite 3.40+
- cfssl 1.6+
- Docker 20.10+
- Make 4.0+

### External Dependencies

- 1Password CLI (`op`) 2.x
- Bitwarden CLI (`bw`) 2023.x
- Git 2.x

### Development Environment

- Linux, macOS, or Windows with WSL2
- 8GB+ RAM
- 10GB+ disk space
- Internet connection (for dependencies)

---

## Risk Management

### Critical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Legal review not completed | Cannot release | Engage attorney ASAP |
| Password manager CLI changes | Integration breaks | Pin CLI versions, version detection |
| Security vulnerability found | Reputation damage | Rapid patching, disclosure program |
| Low community adoption | Project stalls | Early outreach, security marketing |

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance issues | Poor UX | Profiling, optimization |
| Platform compatibility | Limited adoption | Cross-platform testing |
| Dependency vulnerabilities | Security risk | Continuous scanning |
| Complex installation | Adoption barrier | Package managers, docs |

---

## Team & Responsibilities

| Role | Responsibilities | Skills Required |
|------|------------------|-----------------|
| **Project Lead** | Timeline, coordination, legal | Project management |
| **Technical Lead** | Architecture, code review | Go, gRPC, security |
| **Security Lead** | Threat modeling, audits | Cryptography, pentesting |
| **Backend Dev** | Service, CRS, audit | Go, SQLite, crypto |
| **CLI Dev** | TUI, password manager | Go, Bubbletea, CLIs |
| **QA Engineer** | Testing, automation | Testing frameworks |
| **DevOps** | CI/CD, packaging | GitHub Actions, Docker |

---

## Next Steps

1. **Review and approve this plan** with core team
2. **Assign task owners** for Week 1-2 tasks
3. **Set up development environment** (all developers)
4. **Begin Week 1 tasks** (project initialization)
5. **Schedule weekly sync meetings** (progress tracking)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-16 | Claude (AI Assistant) | Initial Phase I implementation plan |

---

**Document Status:** Active Development Plan
**Next Review Date:** End of Week 2 (Milestone M1)
**Distribution:** Core Team, Contributors
