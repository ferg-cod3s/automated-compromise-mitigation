# Security Planning & Threat Modeling
# Automated Compromise Mitigation (ACM)

**Version:** 1.0  
**Date:** November 2025  
**Status:** Planning Document  
**Document Type:** Security Architecture and Threat Analysis

---

## 1. Executive Summary

This document provides a comprehensive security plan for the ACM project, including:
- **Threat modeling** using STRIDE methodology
- **Attack surface analysis** for all components
- **Security implementation roadmap** by development phase
- **Security testing strategy** and validation criteria
- **Incident response playbooks** for common scenarios

### 1.1 Security Posture Goals

| Phase | Security Maturity Target | Key Milestones |
|-------|-------------------------|----------------|
| **Phase I (MVP)** | Basic security controls operational | Memory protection, mTLS, audit logging |
| **Phase II (ACVS)** | Enhanced security with compliance | Evidence chains, encrypted storage, security audit |
| **Phase III** | Hardened production security | HSM integration planning, advanced monitoring |
| **Phase IV** | Enterprise-grade security | Hardware security, formal verification |

---

## 2. Threat Modeling (STRIDE Analysis)

### 2.1 STRIDE Methodology Overview

**STRIDE** categorizes threats into six types:
- **S**poofing (impersonation)
- **T**ampering (data modification)
- **R**epudiation (denial of actions)
- **I**nformation Disclosure (data leaks)
- **D**enial of Service (availability attacks)
- **E**levation of Privilege (unauthorized access)

### 2.2 System Components for Threat Modeling

```
┌─────────────────────────────────────────────────────────┐
│                    Trust Boundaries                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────────────────────────────────┐      │
│  │  Trust Boundary: User's Device (Local)       │      │
│  │                                               │      │
│  │  ┌─────────────┐         ┌─────────────┐     │      │
│  │  │   Clients   │◄───────►│ ACM Service │     │      │
│  │  │ (TUI/GUI)   │  mTLS   │  (Go)       │     │      │
│  │  └─────────────┘         └──────┬──────┘     │      │
│  │                                  │            │      │
│  │                         ┌────────▼────────┐   │      │
│  │                         │  Password       │   │      │
│  │                         │  Manager CLI    │   │      │
│  │                         └────────┬────────┘   │      │
│  │                                  │            │      │
│  │                         ┌────────▼────────┐   │      │
│  │                         │  Encrypted      │   │      │
│  │                         │  Vault          │   │      │
│  │                         └─────────────────┘   │      │
│  └───────────────────────────────────────────────┘      │
│                                                         │
│  ┌──────────────────────────────────────────────┐      │
│  │  Trust Boundary: External (Internet)         │      │
│  │                                               │      │
│  │  ┌─────────────┐         ┌─────────────┐     │      │
│  │  │   ACVS      │───────►│  Third-Party │     │      │
│  │  │ (ToS Fetch) │  HTTPS │  Websites    │     │      │
│  │  └─────────────┘         └─────────────┘     │      │
│  └───────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────┘
```

---

### 2.3 Threat Analysis by Component

#### Component 1: Client-Service Communication (mTLS)

**Trust Boundary:** Localhost network stack

| Threat Type | Threat Description | Impact | Likelihood | Mitigation |
|-------------|-------------------|--------|------------|------------|
| **Spoofing** | Attacker impersonates legitimate client by stealing certificate | High | Low | • Short cert lifetimes (1 year)<br>• Hardware-backed key storage (TPM/Keychain)<br>• Certificate revocation |
| **Tampering** | MITM modifies gRPC messages between client and service | High | Very Low | • mTLS with certificate pinning<br>• TLS 1.3 with strong ciphers<br>• Message integrity via TLS |
| **Repudiation** | Client denies performing credential rotation | Medium | Low | • Audit logs with client cert fingerprint<br>• Cryptographic signatures on actions |
| **Info Disclosure** | Network sniffer captures credentials in transit | Critical | Very Low | • TLS encryption (all traffic encrypted)<br>• Localhost-only binding (127.0.0.1) |
| **Denial of Service** | Flood service with connection requests | Low | Medium | • Rate limiting per client cert<br>• Connection pooling<br>• Resource limits |
| **Elevation of Privilege** | Client with valid cert accesses admin functions | Medium | Low | • Role-based access control (RBAC) in future<br>• Audit all privileged operations |

**Security Controls to Implement:**

```go
// Phase I: mTLS Server Configuration
func NewSecureGRPCServer() (*grpc.Server, error) {
    // Load server certificate
    cert, err := tls.LoadX509KeyPair(
        "/path/to/server.pem",
        "/path/to/server-key.pem",
    )
    if err != nil {
        return nil, err
    }
    
    // Load CA for client cert verification
    caCert, err := os.ReadFile("/path/to/ca.pem")
    if err != nil {
        return nil, err
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    // TLS configuration
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caCertPool,
        MinVersion:   tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
        // CRITICAL: Certificate pinning
        VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
            // Verify cert fingerprint matches expected
            if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
                return errors.New("no verified certificate chains")
            }
            clientCert := verifiedChains[0][0]
            fingerprint := sha256.Sum256(clientCert.Raw)
            
            // Check against allowlist (from config)
            if !isAllowedCertificate(fingerprint) {
                return fmt.Errorf("certificate not in allowlist: %x", fingerprint)
            }
            return nil
        },
    }
    
    // Bind to localhost only
    listener, err := net.Listen("tcp", "127.0.0.1:8443")
    if err != nil {
        return nil, err
    }
    
    // Create gRPC server with TLS
    creds := credentials.NewTLS(tlsConfig)
    server := grpc.NewServer(
        grpc.Creds(creds),
        grpc.MaxConcurrentStreams(100),  // DoS protection
        grpc.ConnectionTimeout(30*time.Second),
        grpc.UnaryInterceptor(rateLimitInterceptor),  // Rate limiting
    )
    
    return server, nil
}
```

---

#### Component 2: Password Manager CLI Integration

**Trust Boundary:** Subprocess execution environment

| Threat Type | Threat Description | Impact | Likelihood | Mitigation |
|-------------|-------------------|--------|------------|------------|
| **Spoofing** | Malicious binary masquerades as legitimate CLI | Critical | Low | • Verify CLI binary checksum on startup<br>• Check file signature (if available)<br>• Validate CLI version |
| **Tampering** | Attacker modifies CLI to leak credentials | Critical | Low | • Read-only CLI binary permissions<br>• File integrity monitoring<br>• Run CLI with minimal privileges |
| **Repudiation** | CLI denies credential modification | Medium | Very Low | • Parse CLI output for confirmation<br>• Verify vault changes independently |
| **Info Disclosure** | CLI output contains plaintext credentials | Critical | Medium | • Never log CLI output directly<br>• Parse and sanitize before logging<br>• Use `--format json` for structured output |
| **Denial of Service** | CLI hangs or crashes during operation | Medium | Medium | • Timeout for CLI operations (30s default)<br>• Graceful error handling<br>• Retry logic with exponential backoff |
| **Elevation of Privilege** | CLI exploited to gain system privileges | High | Low | • Run CLI with user's privileges (not root)<br>• Sandbox CLI execution (future) |

**Security Controls to Implement:**

```go
// Phase I: Secure CLI Execution
type SecureCLIExecutor struct {
    binaryPath    string
    expectedHash  string
    timeout       time.Duration
}

func (e *SecureCLIExecutor) Execute(args []string) ([]byte, error) {
    // 1. Verify CLI binary integrity
    if err := e.verifyBinaryHash(); err != nil {
        return nil, fmt.Errorf("CLI binary verification failed: %w", err)
    }
    
    // 2. Prepare command with timeout
    ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, e.binaryPath, args...)
    
    // 3. Minimal environment (no sensitive vars)
    cmd.Env = []string{
        "PATH=/usr/local/bin:/usr/bin:/bin",
        "HOME=" + os.Getenv("HOME"),
    }
    
    // 4. Capture output securely
    output, err := cmd.Output()
    if err != nil {
        // Log error but NEVER log output (may contain credentials)
        log.Error("CLI execution failed", "error", err)
        return nil, err
    }
    
    // 5. Sanitize output before any logging
    sanitized := sanitizeCredentials(output)
    log.Debug("CLI execution successful", "output_length", len(output))
    
    return output, nil
}

func (e *SecureCLIExecutor) verifyBinaryHash() error {
    f, err := os.Open(e.binaryPath)
    if err != nil {
        return err
    }
    defer f.Close()
    
    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return err
    }
    
    hash := hex.EncodeToString(h.Sum(nil))
    if hash != e.expectedHash {
        return fmt.Errorf("hash mismatch: expected %s, got %s", e.expectedHash, hash)
    }
    
    return nil
}
```

---

#### Component 3: Memory Handling (Credential Storage)

**Trust Boundary:** Process memory space

| Threat Type | Threat Description | Impact | Likelihood | Mitigation |
|-------------|-------------------|--------|------------|------------|
| **Spoofing** | N/A | - | - | - |
| **Tampering** | Memory corruption attack modifies credentials | High | Very Low | • Use safe Go memory patterns<br>• Enable ASLR/DEP |
| **Repudiation** | N/A | - | - | - |
| **Info Disclosure** | Memory dump reveals plaintext credentials | Critical | Low | • **mlock() memory pages**<br>• **Explicit zeroing of buffers**<br>• Disable core dumps |
| **Denial of Service** | Memory exhaustion from large credential sets | Medium | Low | • Memory limits per operation<br>• Streaming processing for large vaults |
| **Elevation of Privilege** | N/A | - | - | - |

**Security Controls to Implement:**

```go
// Phase I: Secure Memory Handling

import (
    "github.com/awnumar/memguard"
    "golang.org/x/sys/unix"
)

// SecureBuffer wraps memguard for automatic memory protection
type SecureBuffer struct {
    enclave *memguard.Enclave
}

func NewSecureBuffer(data []byte) (*SecureBuffer, error) {
    // Create memguard enclave (locked memory)
    enclave := memguard.NewEnclave(data)
    
    // Lock memory pages to prevent swapping
    if err := lockMemory(enclave.Seal().Buffer()); err != nil {
        return nil, err
    }
    
    return &SecureBuffer{enclave: enclave}, nil
}

func (sb *SecureBuffer) Read() []byte {
    lockedBuffer := sb.enclave.Open()
    defer lockedBuffer.Destroy()  // Automatic zeroing
    
    return lockedBuffer.Bytes()
}

func (sb *SecureBuffer) Destroy() {
    sb.enclave.Destroy()
}

func lockMemory(buf []byte) error {
    // Use mlock to prevent page swapping
    return unix.Mlock(buf)
}

// Password generation with secure memory
func GenerateSecurePassword(length int) (string, error) {
    // Allocate secure buffer
    buf := make([]byte, length)
    defer func() {
        // Explicit zeroing
        for i := range buf {
            buf[i] = 0
        }
    }()
    
    // Use crypto/rand for secure random
    if _, err := rand.Read(buf); err != nil {
        return "", err
    }
    
    // Convert to password charset
    password := encodeToPasswordCharset(buf)
    
    // Return password (caller responsible for zeroing)
    return password, nil
}
```

**System-Level Security:**

```bash
# Disable core dumps (prevents memory dumps on crash)
ulimit -c 0

# Or system-wide in /etc/security/limits.conf:
* hard core 0

# Enable ASLR (Address Space Layout Randomization)
echo 2 | sudo tee /proc/sys/kernel/randomize_va_space

# AppArmor/SELinux profile (future)
# Restrict ACM service to minimal file access
```

---

#### Component 4: Audit Logging System

**Trust Boundary:** Local filesystem

| Threat Type | Threat Description | Impact | Likelihood | Mitigation |
|-------------|-------------------|--------|------------|------------|
| **Spoofing** | Attacker creates fake audit entries | Medium | Low | • Cryptographic signatures on entries<br>• Sequence numbers with Merkle tree |
| **Tampering** | Attacker modifies or deletes audit logs | High | Medium | • **Cryptographic signatures (Ed25519)**<br>• Append-only log structure<br>• File permissions (read-only after write) |
| **Repudiation** | User denies performing action | Medium | Low | • Include client cert fingerprint<br>• Sign with user's session token |
| **Info Disclosure** | Audit logs leak credential information | Critical | Medium | • **Hash credential IDs (SHA-256)**<br>• Encrypt sensitive fields (AES-256-GCM)<br>• Never log passwords/tokens |
| **Denial of Service** | Log file fills disk, crashes service | Medium | Low | • Log rotation (max 100MB, 5 backups)<br>• Configurable retention<br>• Disk space monitoring |
| **Elevation of Privilege** | Attacker modifies logs to hide actions | High | Low | • Write-once log structure<br>• Periodic integrity verification<br>• Alert on verification failure |

**Security Controls to Implement:**

```go
// Phase I: Tamper-Evident Audit Logging

type AuditLogger struct {
    db           *sql.DB
    signingKey   ed25519.PrivateKey
    publicKey    ed25519.PublicKey
    lastEntryHash string
}

type AuditEntry struct {
    ID                int64
    Timestamp         time.Time
    EventType         string
    CredentialIDHash  string  // SHA-256 hash, never plaintext
    Action            string
    Status            string
    ClientCertFingerprint string
    Signature         string
    PreviousEntryHash string
}

func (al *AuditLogger) Log(event AuditEntry) error {
    // 1. Hash credential ID (never store plaintext)
    event.CredentialIDHash = hashCredentialID(event.CredentialIDHash)
    
    // 2. Link to previous entry (Merkle chain)
    event.PreviousEntryHash = al.lastEntryHash
    
    // 3. Generate entry hash
    entryHash := al.computeEntryHash(event)
    
    // 4. Sign entry with Ed25519
    message := fmt.Sprintf("%d|%s|%s|%s|%s",
        event.ID,
        event.Timestamp.Format(time.RFC3339),
        event.CredentialIDHash,
        event.Action,
        event.PreviousEntryHash,
    )
    signature := ed25519.Sign(al.signingKey, []byte(message))
    event.Signature = base64.StdEncoding.EncodeToString(signature)
    
    // 5. Insert into database (append-only)
    _, err := al.db.Exec(`
        INSERT INTO audit_events 
        (timestamp, event_type, credential_id_hash, action, status, 
         client_cert_fingerprint, signature, previous_entry_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `,
        event.Timestamp,
        event.EventType,
        event.CredentialIDHash,
        event.Action,
        event.Status,
        event.ClientCertFingerprint,
        event.Signature,
        event.PreviousEntryHash,
    )
    
    if err != nil {
        return err
    }
    
    // 6. Update last entry hash for next entry
    al.lastEntryHash = entryHash
    
    return nil
}

func (al *AuditLogger) VerifyIntegrity() error {
    rows, err := al.db.Query(`
        SELECT id, timestamp, credential_id_hash, action, 
               signature, previous_entry_hash
        FROM audit_events
        ORDER BY id ASC
    `)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    var prevHash string
    for rows.Next() {
        var entry AuditEntry
        err := rows.Scan(
            &entry.ID,
            &entry.Timestamp,
            &entry.CredentialIDHash,
            &entry.Action,
            &entry.Signature,
            &entry.PreviousEntryHash,
        )
        if err != nil {
            return err
        }
        
        // Verify previous hash linkage
        if entry.PreviousEntryHash != prevHash {
            return fmt.Errorf("chain broken at entry %d: expected prev %s, got %s",
                entry.ID, prevHash, entry.PreviousEntryHash)
        }
        
        // Verify signature
        message := fmt.Sprintf("%d|%s|%s|%s|%s",
            entry.ID,
            entry.Timestamp.Format(time.RFC3339),
            entry.CredentialIDHash,
            entry.Action,
            entry.PreviousEntryHash,
        )
        sig, _ := base64.StdEncoding.DecodeString(entry.Signature)
        if !ed25519.Verify(al.publicKey, []byte(message), sig) {
            return fmt.Errorf("signature verification failed for entry %d", entry.ID)
        }
        
        prevHash = al.computeEntryHash(entry)
    }
    
    return nil
}

func hashCredentialID(id string) string {
    h := sha256.Sum256([]byte(id))
    return hex.EncodeToString(h[:])
}
```

---

#### Component 5: ACVS Legal NLP Engine

**Trust Boundary:** Local Python subprocess

| Threat Type | Threat Description | Impact | Likelihood | Mitigation |
|-------------|-------------------|--------|------------|------------|
| **Spoofing** | Malicious ToS document impersonates legitimate site | Medium | Low | • Verify ToS URL matches target domain<br>• Check SSL certificate for HTTPS |
| **Tampering** | Attacker modifies cached CRC to bypass validation | High | Low | • Sign CRC with service key<br>• Validate CRC signature before use |
| **Repudiation** | Site denies ToS version used for CRC | Medium | Medium | • Store ToS hash and version<br>• Archive ToS snapshots (optional) |
| **Info Disclosure** | NLP model leaks training data | Low | Very Low | • Use public ToS corpus only<br>• No proprietary training data |
| **Denial of Service** | Large ToS document crashes NLP engine | Medium | Low | • Limit ToS size (1MB max)<br>• Timeout for NLP inference (30s) |
| **Elevation of Privilege** | Python code injection via malicious ToS | High | Very Low | • Sanitize ToS input<br>• Run NLP in isolated subprocess<br>• No eval() or exec() |

**Security Controls to Implement:**

```python
# Phase II: Secure NLP Processing

import hashlib
import hmac
from typing import Dict, List
import spacy
from datetime import datetime

class SecureLegalNLP:
    def __init__(self, signing_key: bytes):
        self.nlp = spacy.load("en_core_web_trf")
        self.signing_key = signing_key
        self.max_tos_size = 1024 * 1024  # 1MB limit
    
    def analyze_tos(self, tos_text: str, tos_url: str) -> Dict:
        # 1. Validate input size
        if len(tos_text) > self.max_tos_size:
            raise ValueError(f"ToS too large: {len(tos_text)} bytes")
        
        # 2. Sanitize input (remove potential code injection)
        sanitized_text = self.sanitize_tos(tos_text)
        
        # 3. Compute ToS hash for versioning
        tos_hash = hashlib.sha256(sanitized_text.encode()).hexdigest()
        
        # 4. Run NLP analysis (with timeout via subprocess wrapper)
        doc = self.nlp(sanitized_text)
        rules = self.extract_rules(doc)
        
        # 5. Generate CRC
        crc = {
            "site": self.extract_domain(tos_url),
            "tos_url": tos_url,
            "tos_hash": f"sha256:{tos_hash}",
            "analyzed_at": datetime.utcnow().isoformat(),
            "rules": rules,
        }
        
        # 6. Sign CRC for tamper protection
        crc["signature"] = self.sign_crc(crc)
        
        return crc
    
    def sanitize_tos(self, text: str) -> str:
        # Remove potential HTML/JS injection
        # (Use library like bleach for production)
        sanitized = text.replace("<script>", "").replace("</script>", "")
        return sanitized
    
    def sign_crc(self, crc: Dict) -> str:
        # Create HMAC signature of CRC
        message = f"{crc['site']}|{crc['tos_hash']}|{crc['analyzed_at']}"
        signature = hmac.new(
            self.signing_key,
            message.encode(),
            hashlib.sha256
        ).hexdigest()
        return signature
    
    def verify_crc_signature(self, crc: Dict) -> bool:
        stored_sig = crc.get("signature")
        crc_copy = crc.copy()
        del crc_copy["signature"]
        
        computed_sig = self.sign_crc(crc_copy)
        return hmac.compare_digest(stored_sig, computed_sig)
```

---

### 2.4 Attack Surface Analysis

| Surface | Exposure | Attack Vectors | Mitigation Priority |
|---------|----------|----------------|-------------------|
| **mTLS API (localhost:8443)** | Local only | • Certificate theft<br>• MITM (low risk) | High — Phase I |
| **Password Manager CLI** | Subprocess | • Binary replacement<br>• Command injection | High — Phase I |
| **File System (audit logs, config)** | Local disk | • Tampering<br>• Information disclosure | High — Phase I |
| **Memory (credentials in RAM)** | Process memory | • Memory dumps<br>• Debugger attachment | Critical — Phase I |
| **HTTPS (ToS fetching)** | Internet | • MITM<br>• Malicious ToS | Medium — Phase II |
| **OS Keychain/TPM** | Platform API | • Keychain vulnerabilities | Low — Rely on OS |

---

## 3. Security Implementation Roadmap

### 3.1 Phase I Security Deliverables (Months 1-4)

| Security Control | Implementation | Validation | Owner |
|------------------|----------------|------------|-------|
| **mTLS Authentication** | Go TLS 1.3, client certs, cert pinning | Penetration test | Security Lead |
| **Memory Protection** | mlock(), explicit zeroing, memguard | Memory dump analysis | Security Lead |
| **Secure CLI Execution** | Binary verification, timeout, sanitization | Fuzzing, code review | Technical Lead |
| **Audit Logging** | Ed25519 signatures, Merkle chain, encryption | Integrity verification test | Technical Lead |
| **Secrets Management** | OS Keychain integration for cert storage | Manual testing per platform | Security Lead |
| **Code Security** | gosec, Semgrep SAST, dependency scanning | CI/CD automated | All Developers |

**Security Acceptance Criteria:**
- [ ] All mTLS connections use TLS 1.3 with mutual cert authentication
- [ ] Zero credential leaks in logs (automated secret detection passes)
- [ ] Memory protection verified via controlled memory dump test
- [ ] Audit log integrity verification succeeds for 10,000 entry chain
- [ ] SAST scans show zero critical/high findings

---

### 3.2 Phase II Security Enhancements (Months 5-8)

| Security Control | Implementation | Validation | Owner |
|------------------|----------------|------------|-------|
| **Evidence Chain Cryptography** | HMAC signatures, timestamping | Legal review + crypto audit | Security Lead |
| **CRC Tamper Protection** | Sign CRC with service key | Tamper detection test | Legal NLP WG |
| **Enhanced Audit Encryption** | AES-256-GCM for sensitive fields | Decryption key protection test | Security Lead |
| **Third-Party Security Audit** | Engage professional security firm | Audit report with findings | Project Lead |
| **Certificate Rotation** | Automated cert renewal workflow | Expiration handling test | Technical Lead |
| **Secure ToS Fetching** | Certificate pinning for known sites | MITM simulation test | ACVS Lead |

**Security Acceptance Criteria:**
- [ ] Third-party security audit completed with action plan for findings
- [ ] Evidence chain signatures verifiable via public key
- [ ] Encrypted audit fields require service key to decrypt
- [ ] Certificate rotation tested 30 days before expiration

---

### 3.3 Phase III Security Hardening (Months 9-12)

| Security Control | Implementation | Validation | Owner |
|------------------|----------------|------------|-------|
| **Anomaly Detection** | Monitor for unusual certificate usage | Alert on suspicious patterns | Security Lead |
| **Advanced Rate Limiting** | Per-client, per-endpoint limits | Load testing | Technical Lead |
| **Secure Browser Integration** | Sandboxed browser for HIM CAPTCHA | Isolation verification | UX/Design WG |
| **Automated Security Testing** | DAST (OWASP ZAP) in CI/CD | Scan results | Security Lead |
| **Fuzzing** | AFL/libFuzzer for input validation | Crash-free fuzzing runs | Security Lead |

---

### 3.4 Phase IV Future Security (Year 2+)

| Security Control | Implementation | Owner |
|------------------|----------------|-------|
| **HSM/TPM Integration** | Hardware-backed certificate storage | Security Lead |
| **Formal Verification** | TLA+ specifications for critical paths | Security Researchers |
| **Zero-Knowledge Proofs** | Cryptographic proof of rotation without revealing password | Security Researchers |
| **Hardware Security Keys** | FIDO2/WebAuthn integration for HIM | Security Lead |

---

## 4. Security Testing Strategy

### 4.1 Testing Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E╲           ← Security-focused E2E tests
                 ╱ Sec. ╲          (Penetration testing scenarios)
                ╱────────╲
               ╱          ╲
              ╱ Integration╲       ← Integration security tests
             ╱   Security   ╲      (mTLS, CLI, audit chain)
            ╱────────────────╲
           ╱                  ╲
          ╱   Unit Security    ╲   ← Unit tests for security functions
         ╱     (80% coverage)   ╲  (Crypto, input validation, sanitization)
        ╱________________________╲
```

### 4.2 Unit Security Tests

**Example Test Cases:**

```go
// Test: Memory zeroing after use
func TestPasswordZeroingAfterUse(t *testing.T) {
    password := GenerateSecurePassword(32)
    buf := []byte(password)
    
    // Use password
    _ = buf
    
    // Zero memory
    ZeroMemory(buf)
    
    // Verify all bytes are zero
    for i, b := range buf {
        if b != 0 {
            t.Errorf("byte %d not zeroed: %d", i, b)
        }
    }
}

// Test: Signature verification
func TestAuditLogSignatureVerification(t *testing.T) {
    logger := NewAuditLogger()
    
    event := AuditEntry{
        Timestamp: time.Now(),
        Action: "rotate_credential",
        CredentialIDHash: "hash123",
    }
    
    // Log event
    err := logger.Log(event)
    require.NoError(t, err)
    
    // Verify integrity
    err = logger.VerifyIntegrity()
    require.NoError(t, err)
    
    // Tamper with log
    _, err = logger.db.Exec("UPDATE audit_events SET action = 'tampered' WHERE id = ?", event.ID)
    require.NoError(t, err)
    
    // Verification should fail
    err = logger.VerifyIntegrity()
    require.Error(t, err)
}

// Test: CLI output sanitization
func TestCLIOutputSanitization(t *testing.T) {
    maliciousOutput := `{
        "username": "user@example.com",
        "password": "SuperSecret123!",
        "token": "abc123def456"
    }`
    
    sanitized := SanitizeCLIOutput(maliciousOutput)
    
    // Should not contain plaintext credentials
    assert.NotContains(t, sanitized, "SuperSecret123!")
    assert.NotContains(t, sanitized, "abc123def456")
    
    // Should redact sensitive fields
    assert.Contains(t, sanitized, "password")
    assert.Contains(t, sanitized, "[REDACTED]")
}
```

---

### 4.3 Integration Security Tests

**Test Scenarios:**

1. **mTLS Authentication**
   - Valid client cert → connection succeeds
   - Invalid client cert → connection rejected
   - Expired client cert → connection rejected
   - Revoked client cert → connection rejected

2. **Password Manager CLI Integration**
   - CLI binary tampered → operation fails
   - CLI timeout → graceful error handling
   - CLI output malformed → parsing error with no crash

3. **Audit Log Chain**
   - 10,000 entries → verify entire chain in < 1 second
   - Tamper with middle entry → verification fails
   - Concurrent writes → no race conditions

---

### 4.4 Penetration Testing

**Annual Third-Party Penetration Test Scope:**

| Test Area | Methodology | Expected Findings |
|-----------|-------------|-------------------|
| **Authentication Bypass** | Attempt to connect without valid cert | Rejected with clear error |
| **Authorization Escalation** | Attempt privileged operations with user cert | Blocked (RBAC future) |
| **Data Exfiltration** | Monitor network traffic for credential leaks | Zero credentials transmitted |
| **Memory Analysis** | Dump process memory during operation | Credentials cleared after use |
| **Log Tampering** | Modify audit logs | Signature verification detects tampering |
| **DoS Attacks** | Flood service with requests | Rate limiting prevents crash |

**Internal Penetration Testing (Phase I):**

```bash
# Test 1: Attempt connection without client cert
curl --cacert ca.pem https://localhost:8443/health
# Expected: TLS handshake failure

# Test 2: Attempt connection with wrong client cert
curl --cert wrong-cert.pem --key wrong-key.pem \
     --cacert ca.pem https://localhost:8443/health
# Expected: Certificate verification failed

# Test 3: Memory dump analysis
gcore $(pgrep acm-service)
strings core.* | grep -i "password\|secret\|token"
# Expected: Zero credential strings found

# Test 4: Log tampering detection
sqlite3 ~/.acm/data/audit.db \
  "UPDATE audit_events SET action='tampered' WHERE id=1"
acm audit verify
# Expected: Verification failed — tampered entry detected
```

---

### 4.5 Fuzzing

**Fuzzing Targets:**

1. **gRPC API Handlers**
   ```bash
   # Use go-fuzz for gRPC message fuzzing
   go-fuzz -bin=./acm-fuzz.zip -workdir=./fuzz/grpc
   ```

2. **CLI Output Parser**
   ```bash
   # Fuzz JSON parser with invalid/malformed input
   AFL_I_DONT_CARE_ABOUT_MISSING_CRASHES=1 \
   afl-fuzz -i testcases/ -o findings/ -- ./acm-cli-parser @@
   ```

3. **ToS Parser (NLP)**
   ```bash
   # Fuzz HTML/text input to NLP engine
   python3 -m atheris fuzz_nlp.py
   ```

**Fuzzing Success Criteria:**
- 1,000,000+ iterations without crash
- No memory leaks detected (Valgrind)
- No secret disclosure in fuzzing output

---

## 5. Security Monitoring and Incident Response

### 5.1 Security Monitoring

**What to Monitor:**

| Metric | Threshold | Alert Action |
|--------|-----------|--------------|
| **Failed mTLS connections** | > 10/min from same IP | Log + investigate |
| **Certificate expiration** | < 30 days | Email warning to user |
| **Audit log verification failure** | Any failure | Critical alert + incident response |
| **Memory usage spike** | > 1GB for service | Warning (potential leak) |
| **CLI execution timeout** | > 5 timeouts/hour | Warning (CLI issue) |
| **SAST/Dependency CVE** | Any critical/high | Block merge + immediate review |

---

### 5.2 Security Incident Response Playbooks

#### Playbook 1: Credential Exposure in Logs

**Trigger:** Automated secret detection finds plaintext credential in logs

**Response:**
1. **Immediate (< 1 hour):**
   - Identify affected log files
   - Quarantine logs (move to secure location, restrict access)
   - Determine scope: which credentials exposed, for how long

2. **Short-term (< 24 hours):**
   - Rotate all potentially exposed credentials
   - Patch code to prevent future logging
   - Deploy emergency release with fix

3. **Long-term (< 1 week):**
   - Conduct root cause analysis (RCA)
   - Implement additional secret detection in CI/CD
   - Update security testing to catch similar issues
   - Publish post-mortem (if public project)

---

#### Playbook 2: Certificate Theft Detected

**Trigger:** Anomaly detection flags certificate used from unexpected location/time

**Response:**
1. **Immediate (< 1 hour):**
   - Revoke compromised certificate
   - Force user to re-authenticate with new certificate
   - Audit logs for actions performed with stolen cert

2. **Short-term (< 24 hours):**
   - Investigate how certificate was stolen (keychain vulnerability?)
   - Notify user of compromise
   - Provide guidance on securing certificates (enable OS keychain encryption)

3. **Long-term (< 1 week):**
   - Implement hardware-backed key storage (TPM/Secure Enclave)
   - Shorten certificate lifetime (6 months instead of 1 year)
   - Add user notification on certificate usage (optional)

---

#### Playbook 3: Vulnerability Disclosed

**Trigger:** Security researcher reports vulnerability via security@acm.dev

**Response:**
1. **Immediate (< 48 hours):**
   - Acknowledge receipt of report
   - Triage severity (P0/P1/P2)
   - Assign Security Lead + relevant developer

2. **Short-term (varies by severity):**
   - **P0 (Critical):** 24-48 hour fix, emergency release
   - **P1 (High):** 1 week fix, patch release
   - **P2 (Medium):** 2-4 week fix, next minor release

3. **Long-term (< 2 weeks after fix):**
   - Coordinated disclosure with researcher
   - Publish CVE and GitHub Security Advisory
   - Blog post with technical details (after users have time to patch)
   - Credit researcher in security hall of fame

---

## 6. Security Documentation Requirements

### 6.1 Required Security Docs

| Document | Purpose | Owner | Update Frequency |
|----------|---------|-------|-----------------|
| **SECURITY.md** | Responsible disclosure policy | Security Lead | Annually |
| **Threat Model** (this doc) | Comprehensive threat analysis | Security Lead | Quarterly |
| **Security Audit Reports** | Third-party audit findings | Project Lead | Per audit |
| **Incident Post-Mortems** | Lessons learned from incidents | Security Lead | Per incident |
| **Cryptography Spec** | Document all crypto usage | Security Lead | On changes |
| **Vulnerability Disclosure Log** | Public CVE history | Security Lead | On disclosure |

---

### 6.2 SECURITY.md Template

```markdown
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues to: **security@acm.dev**

You can also use our PGP key for encrypted communication:
- Fingerprint: `XXXX XXXX XXXX XXXX`
- Key: https://acm-project.dev/pgp

### What to Include

- Description of the vulnerability
- Steps to reproduce (proof-of-concept)
- Affected versions
- Potential impact
- Suggested fix (if available)

### What to Expect

- **Acknowledgment:** Within 48 hours
- **Initial Assessment:** Within 1 week
- **Fix Timeline:**
  - Critical: 24-48 hours
  - High: 1 week
  - Medium: 2-4 weeks
- **Coordinated Disclosure:** We'll work with you on timing

### Bug Bounty

We currently do not have a bug bounty program, but we deeply appreciate
security research and will publicly acknowledge your contribution (if desired).

## Security Best Practices

Users should follow these security practices:

1. **Keep ACM Updated:** Always use the latest version
2. **Secure Your Device:** Use full disk encryption, strong passwords
3. **Verify Downloads:** Check GPG signatures on releases
4. **Review Audit Logs:** Regularly check `acm audit` for suspicious activity
5. **Protect Certificates:** Enable OS keychain encryption

## Security Features

- **Zero-Knowledge Architecture:** Master password never leaves your device
- **mTLS Authentication:** All service-client communication encrypted
- **Tamper-Evident Audit Logs:** Cryptographic signatures detect tampering
- **Memory Protection:** Sensitive data locked in memory, zeroed after use

## Past Security Advisories

See [https://github.com/acm-project/acm/security/advisories](advisories)
```

---

## 7. Security Compliance Checklist

### 7.1 Pre-Release Security Checklist

**Before Public Release (Phase I):**

- [ ] **Authentication & Authorization**
  - [ ] mTLS implemented with TLS 1.3
  - [ ] Client certificate validation working
  - [ ] Certificate revocation mechanism in place
  - [ ] Localhost-only binding (127.0.0.1)

- [ ] **Cryptography**
  - [ ] All crypto uses standard libraries (Go crypto/*)
  - [ ] No custom crypto implementations
  - [ ] Secure random for password generation (crypto/rand)
  - [ ] Audit log signatures (Ed25519)

- [ ] **Memory Safety**
  - [ ] Memory locking (mlock) implemented
  - [ ] Explicit zeroing of sensitive buffers
  - [ ] No credential strings in logs verified

- [ ] **Input Validation**
  - [ ] All CLI output sanitized before use
  - [ ] JSON parsing with strict schema validation
  - [ ] No command injection vulnerabilities

- [ ] **Audit & Logging**
  - [ ] Credential IDs hashed (SHA-256)
  - [ ] Logs cryptographically signed
  - [ ] Log integrity verification working
  - [ ] No sensitive data in logs verified

- [ ] **Code Security**
  - [ ] gosec scan passes (zero critical/high)
  - [ ] Semgrep scan passes
  - [ ] Dependency vulnerability scan clean
  - [ ] Unit test coverage > 80%

- [ ] **Documentation**
  - [ ] SECURITY.md published
  - [ ] Threat model documented
  - [ ] Security architecture in TAD

- [ ] **Testing**
  - [ ] Security unit tests passing
  - [ ] Integration security tests passing
  - [ ] Manual penetration testing completed
  - [ ] Memory dump analysis clean

---

## 8. Conclusion

This security planning document provides a comprehensive roadmap for building and maintaining a secure ACM project. Key takeaways:

1. **Defense in Depth:** Multiple layers of security controls protect against various threat vectors
2. **Security by Design:** Security integrated from Phase I, not bolted on later
3. **Continuous Monitoring:** Ongoing security testing, audits, and incident response
4. **Transparency:** Open-source nature allows community security review

**Next Steps:**
1. Review this document with Security Lead and Core Team
2. Assign security implementation owners for Phase I
3. Set up security testing infrastructure (SAST, dependency scanning)
4. Begin threat modeling workshops with contributors
5. Engage external security auditor for Phase II

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-13 | Claude (AI Assistant) | Complete security planning document with threat modeling and implementation roadmap |

---

**Document Status:** Planning Document — Ready for Implementation  
**Next Review Date:** Phase I Completion  
**Distribution:** Core Team, Security Lead, Contributors
