# Technical Architecture Document (TAD)
# Automated Compromise Mitigation (ACM)

**Version:** 1.0  
**Date:** November 2025  
**Status:** Draft  
**Architecture Type:** Local-First Microservice with Zero-Knowledge Security

---

## 1. Introduction

### 1.1 Purpose

This document defines the technical architecture and design principles for the Automated Compromise Mitigation (ACM) system. ACM is a local-first, zero-knowledge application composed of two modular services that operate entirely on the user's device to detect and remediate compromised credentials.

### 1.2 Scope

This architecture document covers:
- High-level system architecture and component interactions
- Service-client communication model with mTLS authentication
- Data flow and storage patterns
- Security architecture and threat model
- Deployment model and runtime requirements
- Technology stack and implementation details

### 1.3 Architectural Principles

| Principle | Description | Enforcement |
|-----------|-------------|-------------|
| **Zero-Knowledge** | Master password and vault encryption keys never accessible to ACM | No storage or transmission of password manager credentials; audit all memory access |
| **Local-First** | All sensitive processing occurs on user's device with no cloud dependencies | Network analysis; whitelist localhost-only communication |
| **Service-Client Separation** | Business logic centralized in service; clients are thin presentation layers | API contracts enforce separation; shared core logic prohibited |
| **Defense in Depth** | Multiple security layers: mTLS transport, certificate authentication, encrypted storage | Penetration testing validates layered approach |
| **Transparency** | Open-source, auditable code with documented security decisions | Public repository; reproducible builds; ADR documentation |
| **Fail-Safe Defaults** | Security-critical features (ACVS) disabled by default; explicit opt-in required | Configuration validation at startup |

---

## 2. High-Level Architecture

### 2.1 System Context Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         User's Device                           │
│                                                                 │
│  ┌──────────────┐                         ┌──────────────┐     │
│  │  OpenTUI     │                         │ Tauri GUI    │     │
│  │  (CLI)       │                         │  (Desktop)   │     │
│  └──────┬───────┘                         └──────┬───────┘     │
│         │                                        │             │
│         │        mTLS (localhost only)           │             │
│         └────────────┬──────────────────────────┘             │
│                      │                                         │
│            ┌─────────▼─────────┐                               │
│            │  ACM Service      │                               │
│            │  (Go Daemon)      │                               │
│            │                   │                               │
│            │  ┌─────────────┐  │                               │
│            │  │ API Gateway │  │                               │
│            │  │   (gRPC)    │  │                               │
│            │  └──────┬──────┘  │                               │
│            │         │         │                               │
│            │  ┌──────▼──────┐  │                               │
│            │  │     CRS     │  │  Credential Remediation       │
│            │  │   Service   │  │                               │
│            │  └──────┬──────┘  │                               │
│            │         │         │                               │
│            │  ┌──────▼──────┐  │                               │
│            │  │    ACVS     │  │  Compliance Validation        │
│            │  │   Service   │  │  (Opt-In)                     │
│            │  └──────┬──────┘  │                               │
│            │         │         │                               │
│            │  ┌──────▼──────┐  │                               │
│            │  │ HIM Manager │  │  Human-in-the-Middle          │
│            │  └─────────────┘  │                               │
│            └──────────┬─────────┘                               │
│                       │                                         │
│         ┌─────────────┼─────────────┐                           │
│         │             │             │                           │
│    ┌────▼────┐  ┌────▼────┐  ┌────▼────┐                       │
│    │ SQLite  │  │   OS    │  │Password │                       │
│    │ (Audit  │  │Keychain │  │Manager  │                       │
│    │  Logs)  │  │  /TPM   │  │   CLI   │                       │
│    └─────────┘  └─────────┘  └────┬────┘                       │
│                                    │                            │
│                              ┌─────▼─────┐                      │
│                              │ Encrypted │                      │
│                              │   Vault   │                      │
│                              └───────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Component Overview

| Component | Type | Technology | Responsibility |
|-----------|------|------------|----------------|
| **ACM Service** | Daemon/Service | Go | Core business logic; exposes gRPC API on localhost |
| **OpenTUI Client** | CLI/TUI | Go + Bubbletea | Terminal interface for developers and power users |
| **Tauri Client** | Desktop GUI | Rust + Tauri + Web UI | Visual interface for general users |
| **CRS (Credential Remediation Service)** | Core Module | Go | Detects compromised credentials, generates passwords, updates vault |
| **ACVS (Automated Compliance Validation Service)** | Core Module (Opt-In) | Go + Python (NLP) | Validates ToS compliance, manages automation rules |
| **HIM Manager** | Core Module | Go | Orchestrates Human-in-the-Middle workflows for MFA/CAPTCHA |
| **Legal NLP Engine** | Embedded Service | Python (spaCy/Transformers) | Parses ToS documents, generates Compliance Rule Sets |
| **Audit Logger** | Core Module | Go + SQLite | Cryptographically signed local audit trail |

### 2.3 Communication Model

**Service ↔ Client Communication:**
- Protocol: gRPC over mTLS
- Scope: `localhost:8443` (configurable port)
- Authentication: X.509 client certificates issued by local CA
- Encryption: TLS 1.3 with mutual authentication
- Session: Short-lived JWT tokens (15-30 min) issued after certificate validation

**Service ↔ Password Manager CLI:**
- Method: Subprocess execution (`os/exec`)
- Authentication: Inherits user's authenticated CLI session
- Data Exchange: Structured JSON output from CLI; parsed by CRS

**Service ↔ Local Storage:**
- Audit Logs: SQLite database with encrypted sensitive fields
- Certificates: OS Keychain (macOS), Windows Certificate Store, or Linux Secret Service
- Configuration: YAML files in `~/.acm/config/`

---

## 3. Detailed Component Architecture

### 3.1 ACM Service (Core Daemon)

**Purpose:** Centralized business logic handling all credential remediation, compliance validation, and audit operations.

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                      ACM Service (Go)                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              API Gateway (gRPC Server)               │   │
│  │                                                       │   │
│  │  • mTLS Enforcement                                   │   │
│  │  • Certificate Validation                             │   │
│  │  • JWT Session Management                             │   │
│  │  • Request Routing                                    │   │
│  └───────────────────────┬─────────────────────────────┘   │
│                          │                                  │
│            ┌─────────────┼─────────────┐                    │
│            │             │             │                    │
│  ┌─────────▼─────┐ ┌────▼────┐ ┌─────▼─────┐              │
│  │ CRS Module    │ │  ACVS   │ │    HIM    │              │
│  │               │ │ Module  │ │  Manager  │              │
│  │ • Breach      │ │         │ │           │              │
│  │   Detection   │ │ • Legal │ │ • Prompt  │              │
│  │ • Password    │ │   NLP   │ │   User    │              │
│  │   Generation  │ │ • CRC   │ │ • Resume  │              │
│  │ • Vault       │ │   Valid │ │   Flow    │              │
│  │   Update      │ │ • Evid  │ │           │              │
│  │               │ │   Chain │ │           │              │
│  └───────┬───────┘ └────┬────┘ └─────┬─────┘              │
│          │              │            │                     │
│  ┌───────▼──────────────▼────────────▼─────┐              │
│  │         Audit Logger (SQLite)            │              │
│  │                                           │              │
│  │  • Cryptographic Signing                 │              │
│  │  • Evidence Chain Generation              │              │
│  │  • Compliance Report Export               │              │
│  └───────────────────────────────────────────┘              │
│                                                             │
│  ┌───────────────────────────────────────────┐              │
│  │      Password Manager CLI Interface        │              │
│  │                                           │              │
│  │  • 1Password (`op`)                       │              │
│  │  • Bitwarden (`bw`)                       │              │
│  │  • LastPass (`lpass`)                     │              │
│  └───────────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

**Key Interfaces:**

```go
// Core service interfaces

type CredentialRemediationService interface {
    DetectCompromised(ctx context.Context) ([]CompromisedCredential, error)
    GeneratePassword(ctx context.Context, policy PasswordPolicy) (string, error)
    RotateCredential(ctx context.Context, cred CompromisedCredential, newPassword string) error
    VerifyRotation(ctx context.Context, credentialID string) (bool, error)
}

type ComplianceValidationService interface {
    AnalyzeToS(ctx context.Context, url string) (*ComplianceRuleSet, error)
    ValidateAction(ctx context.Context, action RotationAction, crc *ComplianceRuleSet) (*ValidationResult, error)
    GenerateEvidenceChain(ctx context.Context, action RotationAction, result *ValidationResult) (*EvidenceChain, error)
}

type HIMManager interface {
    RequiresHIM(ctx context.Context, action RotationAction) (bool, HIMType, error)
    PromptUser(ctx context.Context, prompt HIMPrompt) (*HIMResponse, error)
    ResumeAutomation(ctx context.Context, response *HIMResponse) error)
}

type AuditLogger interface {
    Log(ctx context.Context, event AuditEvent) error
    Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
    ExportReport(ctx context.Context, format ReportFormat, filter AuditFilter) (io.Reader, error)
    VerifyIntegrity(ctx context.Context, from, to time.Time) (bool, error)
}
```

**Deployment:**
- Binary: `acm-service` (single Go binary)
- Runtime: Daemon/service process (`systemd`, `launchd`, or Windows Service)
- Configuration: `~/.acm/config/service.yaml`
- Logs: `~/.acm/logs/service.log` (rotated, non-sensitive only)
- Data: `~/.acm/data/audit.db` (SQLite)

---

### 3.2 Credential Remediation Service (CRS)

**Purpose:** Core module responsible for detecting compromised credentials and performing safe, local vault updates.

**Workflow:**

```
┌─────────────────────────────────────────────────────────────┐
│              CRS Credential Rotation Workflow               │
└─────────────────────────────────────────────────────────────┘

1. Detection Phase
   ├─ Query Password Manager CLI for breach reports
   │  └─ `bw list items --exposed` OR `op item list --categories=Login --compromised`
   ├─ Parse JSON response to extract compromised item IDs
   └─ Return list of CompromisedCredential structs

2. Analysis Phase
   ├─ For each compromised credential:
   │  ├─ Extract metadata (site, username, last modified)
   │  ├─ Check rotation history (avoid duplicate rotation)
   │  └─ Determine rotation strategy (API, HIM, or skip)

3. Generation Phase
   ├─ Generate secure password using crypto/rand
   │  ├─ Default: 32 chars, mixed case, numbers, symbols
   │  └─ Policy-compliant (configurable per-site rules)
   └─ Store temporarily in secure memory (locked pages)

4. Update Phase
   ├─ Execute vault update via CLI
   │  └─ `bw edit item <id> --password "<new_password>"`
   ├─ Verify update success (query vault, compare timestamp)
   └─ Clear password from memory (explicit zeroing)

5. Audit Phase
   ├─ Log rotation event to audit database
   │  └─ {timestamp, credential_id_hash, action: "rotated", status: "success"}
   ├─ Generate evidence chain entry (if ACVS enabled)
   └─ Return RotationResult to caller
```

**Security Controls:**

| Control | Implementation | Rationale |
|---------|----------------|-----------|
| **Memory Protection** | `syscall.Mlock()` on password buffers | Prevent swap to disk |
| **Explicit Zeroing** | `memguard.Wipe()` after use | Clear sensitive data from RAM |
| **Subprocess Isolation** | Execute CLI in isolated context with minimal env vars | Reduce attack surface |
| **Atomic Transactions** | Verify vault state before and after update | Prevent partial rotations |
| **Rollback Capability** | Store old password temporarily (encrypted) until verified | Enable recovery from failed rotation |

**Error Handling:**

```go
type RotationError struct {
    Code    RotationErrorCode
    Message string
    Cause   error
    Retryable bool
}

const (
    ErrVaultLocked          RotationErrorCode = "VAULT_LOCKED"
    ErrCLINotFound         RotationErrorCode = "CLI_NOT_FOUND"
    ErrNetworkRequired     RotationErrorCode = "NETWORK_REQUIRED"  // Vault sync needed
    ErrHIMRequired         RotationErrorCode = "HIM_REQUIRED"
    ErrComplianceViolation RotationErrorCode = "COMPLIANCE_VIOLATION"
)
```

---

### 3.3 Automated Compliance Validation Service (ACVS)

**Purpose:** Opt-in module that validates automation actions against target site's Terms of Service and manages compliance evidence chains.

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│         ACVS (Automated Compliance Validation Service)      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────┐     │
│  │          Legal NLP Engine (Python)                 │     │
│  │                                                     │     │
│  │  • spaCy NLP Pipeline (en_core_web_trf)            │     │
│  │  • Custom ToS Classification Model                 │     │
│  │  • Named Entity Recognition (rate limits, etc.)    │     │
│  │                                                     │     │
│  │  Input:  ToS HTML/Text                             │     │
│  │  Output: Structured Compliance Rule Set (JSON)     │     │
│  └──────────────────────┬──────────────────────────────┘     │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │       Compliance Rule Set (CRC) Manager              │   │
│  │                                                       │   │
│  │  • Cache parsed ToS rules (TTL: 30 days)             │   │
│  │  • Version ToS documents (detect updates)            │   │
│  │  • Generate CRC signatures for audit trail           │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │         Compliance Validator                          │   │
│  │                                                       │   │
│  │  • Pre-Flight Check: Can automation proceed?         │   │
│  │  • Policy Enforcement: Block violations              │   │
│  │  • Evidence Generation: Log compliance proof         │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │         Evidence Chain Generator                      │   │
│  │                                                       │   │
│  │  • Cryptographic timestamps (RFC 3161)               │   │
│  │  • Linked evidence entries (Merkle tree)             │   │
│  │  • Export to PDF/JSON for compliance reporting       │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**Compliance Rule Set (CRC) Schema:**

```json
{
  "site": "example.com",
  "tos_url": "https://example.com/terms",
  "tos_version": "2025-01-15",
  "tos_hash": "sha256:abcd1234...",
  "parsed_at": "2025-11-13T10:00:00Z",
  "rules": [
    {
      "id": "CRC-001",
      "category": "automation",
      "severity": "high",
      "rule": "Prohibits automated login and scraping",
      "extracted_text": "Users may not use automated means to access the service...",
      "confidence": 0.95,
      "implications": {
        "allows_api_automation": false,
        "requires_human_interaction": true,
        "rate_limit": null
      }
    },
    {
      "id": "CRC-002",
      "category": "rate_limiting",
      "severity": "medium",
      "rule": "Maximum 60 requests per hour",
      "extracted_text": "API rate limit is 60 requests per hour per user",
      "confidence": 0.98,
      "implications": {
        "allows_api_automation": true,
        "rate_limit": {
          "requests": 60,
          "window": "1h"
        }
      }
    }
  ],
  "recommendation": "HIM_REQUIRED",
  "reasoning": "ToS explicitly prohibits automated login; recommend Human-in-the-Middle workflow"
}
```

**Validation Workflow:**

```
Pre-Rotation Validation:
1. Check if ACVS enabled (opt-in status)
   └─ If disabled: Skip validation, proceed to CRS
   
2. Check CRC cache for target site
   ├─ Cache hit (< 30 days old): Use cached CRC
   └─ Cache miss: Fetch ToS, parse with NLP, generate CRC

3. Evaluate rotation action against CRC
   ├─ Check CRC rules for automation prohibitions
   ├─ If API available and allowed: Proceed with API rotation
   ├─ If automation prohibited: Require HIM
   └─ If rate-limited: Check rotation frequency, enforce backoff

4. Generate pre-flight validation result
   └─ {allowed: bool, method: "API" | "HIM" | "BLOCKED", crc_rules_applied: []}

5. Log validation to evidence chain
   └─ Cryptographically sign: {timestamp, site, crc_version, decision, reasoning}
```

**NLP Model Training:**

The Legal NLP model is trained on a corpus of:
- 500+ website Terms of Service documents
- Manually annotated anti-automation clauses
- Rate limiting and API policy sections
- Legal definitions of "automated access"

Training pipeline:
1. Text preprocessing (HTML cleaning, section extraction)
2. Named Entity Recognition for: rate limits, prohibited actions, exceptions
3. Sentence classification (binary: automation-relevant vs. not)
4. Rule extraction using dependency parsing
5. Confidence scoring based on linguistic patterns

Model performance targets:
- Precision: > 90% (few false positives blocking valid automation)
- Recall: > 85% (catch most automation prohibitions)
- F1 Score: > 0.87

---

### 3.4 Human-in-the-Middle (HIM) Manager

**Purpose:** Orchestrate workflows where user intervention is required due to MFA, CAPTCHA, or policy restrictions.

**HIM Trigger Conditions:**

| Trigger | Detection Method | User Action Required |
|---------|------------------|----------------------|
| **MFA (TOTP)** | ACVS detects MFA during API call or UI analysis | User provides 6-digit TOTP code |
| **MFA (SMS)** | Rotation attempt returns "SMS code required" | User enters SMS code from phone |
| **MFA (Push)** | Push notification sent to user's device | User approves push notification |
| **CAPTCHA** | UI automation detects CAPTCHA challenge | User solves CAPTCHA in secure browser |
| **ToS Violation** | ACVS pre-flight check blocks automation | User reviews and manually rotates |
| **API Unavailable** | No documented API for target service | User manually navigates to site |

**HIM Workflow:**

```
┌─────────────────────────────────────────────────────────────┐
│              HIM Workflow State Machine                     │
└─────────────────────────────────────────────────────────────┘

State: AUTOMATION_IN_PROGRESS
  └─ Event: HIM_REQUIRED(type, context)
      └─ Transition to: AWAITING_USER_INPUT

State: AWAITING_USER_INPUT
  ├─ Send secure prompt to client via gRPC stream
  │  └─ Prompt includes: site, reason, expected input format
  ├─ Start timeout timer (5 minutes default)
  └─ Wait for user response or timeout
      ├─ Event: USER_RESPONDED(data)
      │   └─ Transition to: VALIDATING_RESPONSE
      └─ Event: TIMEOUT
          └─ Transition to: FAILED (reason: user timeout)

State: VALIDATING_RESPONSE
  ├─ Verify user input format (e.g., 6-digit TOTP)
  ├─ Submit to target service (if applicable)
  └─ Check result
      ├─ Success: Transition to AUTOMATION_RESUMING
      └─ Failure: Transition to AWAITING_USER_INPUT (retry)

State: AUTOMATION_RESUMING
  ├─ Continue rotation workflow from pause point
  └─ Transition to AUTOMATION_IN_PROGRESS
      └─ Eventually: COMPLETED or FAILED

State: COMPLETED
  └─ Log HIM event: {type, duration, attempts, success}

State: FAILED
  └─ Log HIM event: {type, reason, duration}
```

**Security Considerations:**

- **No Password Capture**: HIM prompts never ask for master password
- **Timeout Enforcement**: User has limited time to respond (prevent session hijacking)
- **Secure Input**: All user input transmitted over mTLS, never logged in plaintext
- **Context Preservation**: HIM manager maintains rotation state across pause/resume

---

### 3.5 Client Architecture

#### 3.5.1 OpenTUI (Terminal Interface)

**Technology Stack:**
- Language: Go
- TUI Framework: Charm Bubbletea
- Rendering: Lipgloss (styling), Bubbles (components)

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenTUI Application                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────┐     │
│  │           Bubbletea Runtime (Event Loop)           │     │
│  └────────────────┬──────────────────────────────────┘     │
│                   │                                         │
│     ┌─────────────┼─────────────┐                           │
│     │             │             │                           │
│  ┌──▼───┐   ┌─────▼────┐   ┌───▼────┐                      │
│  │ View │   │  Update  │   │  Init  │                      │
│  │      │   │  (State  │   │        │                      │
│  │Render│   │ Machine) │   │ Setup  │                      │
│  └──────┘   └──────────┘   └────────┘                      │
│                   │                                         │
│  ┌────────────────▼────────────────┐                        │
│  │      gRPC Client (mTLS)         │                        │
│  │                                 │                        │
│  │  • CredentialService.Detect()   │                        │
│  │  • CredentialService.Rotate()   │                        │
│  │  • HIMService.PromptUser()      │                        │
│  │  • AuditService.Query()         │                        │
│  └─────────────────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

**Key Commands:**

```bash
# Core commands
acm status                       # Show service status and configuration
acm detect                       # Detect compromised credentials
acm rotate <item-id>             # Rotate specific credential
acm rotate --all                 # Rotate all compromised credentials
acm audit --since 7d             # View audit log (last 7 days)
acm compliance check <site>      # Check ToS compliance for site

# ACVS commands (requires opt-in)
acm compliance enable            # Enable ACVS (accepts EULA)
acm compliance analyze <url>     # Analyze ToS and generate CRC
acm compliance status            # Show ACVS configuration

# Configuration commands
acm config show                  # Display current configuration
acm config set <key> <value>     # Update configuration
acm cert renew                   # Renew client certificate

# Scripting support
acm rotate --json                # Output results in JSON format
acm detect --format json | jq    # Pipe to jq for filtering
```

**User Experience:**

```
╭───────────────────────────────────────────────────────╮
│  ACM - Automated Compromise Mitigation               │
│  Status: Service Running • ACVS: Disabled             │
╰───────────────────────────────────────────────────────╯

Compromised Credentials Detected: 3

┌─────────────────────────────────────────────────────┐
│ [1] github.com                                       │
│     Username: user@example.com                       │
│     Breach: Collection #1 (2019-01-07)              │
│     Last Rotated: Never                              │
│     Action: [R]otate [S]kip [V]iew                   │
├─────────────────────────────────────────────────────┤
│ [2] linkedin.com                                     │
│     Username: user@example.com                       │
│     Breach: LinkedIn Data Breach (2021-06-01)       │
│     Last Rotated: 2024-11-01 (13 days ago)          │
│     Action: [R]otate [S]kip [V]iew                   │
├─────────────────────────────────────────────────────┤
│ [3] dropbox.com                                      │
│     Username: user@example.com                       │
│     Breach: Dropbox (2012-07-01)                    │
│     Last Rotated: Never                              │
│     ⚠️  ACVS: Requires HIM (MFA)                     │
│     Action: [R]otate [S]kip [V]iew                   │
└─────────────────────────────────────────────────────┘

Commands: [A]ll [Q]uit [C]onfig [?]Help
```

#### 3.5.2 Tauri Desktop GUI

**Technology Stack:**
- Framework: Tauri 1.5+ (Rust backend, web frontend)
- Frontend: React + TypeScript
- Styling: Tailwind CSS
- State Management: Zustand
- gRPC Client: grpc-web (proxied through Tauri backend)

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                Tauri Desktop Application                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────┐     │
│  │          Web Frontend (React)                      │     │
│  │                                                     │     │
│  │  Components:                                        │     │
│  │  • Dashboard (status overview)                      │     │
│  │  • CredentialList (detected compromises)            │     │
│  │  • RotationWorkflow (step-by-step wizard)           │     │
│  │  • ComplianceDashboard (ACVS status)                │     │
│  │  • AuditViewer (log explorer)                       │     │
│  │  • HIMPrompt (secure input for MFA/CAPTCHA)         │     │
│  └────────────────────┬──────────────────────────────┘     │
│                       │ Tauri IPC (invoke, emit)           │
│  ┌────────────────────▼──────────────────────────────┐     │
│  │         Rust Backend (Tauri Core)                  │     │
│  │                                                     │     │
│  │  • gRPC Client Manager (mTLS)                      │     │
│  │  • Certificate Store Interface                      │     │
│  │  • Secure Credential Display (masked)               │     │
│  │  • File System Access (config, logs)                │     │
│  └─────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

**Key Features:**

1. **Dashboard View**: Real-time status of service, detected compromises, recent rotations
2. **Workflow Wizard**: Step-by-step guided rotation with preview and confirmation
3. **HIM Integration**: Modal prompts for MFA/CAPTCHA with secure input fields
4. **Compliance Dashboard**: Visual ToS analysis results with rule explanations
5. **Audit Explorer**: Searchable, filterable view of all rotation events

**Security Considerations:**
- Frontend never has direct access to vault data (all via Tauri backend)
- Sensitive data (passwords, tokens) never stored in frontend state
- All API calls routed through Tauri backend with certificate validation
- Web content security policy (CSP) prevents external resource loading

---

## 4. Data Architecture

### 4.1 Data Flow Diagram

```
User Action (Detect Compromises)
         │
         ▼
  OpenTUI/Tauri Client
         │ gRPC Request (mTLS)
         ▼
    ACM Service API Gateway
         │ Authenticate (validate cert)
         ▼
      CRS Module
         │
         ├─ Execute: bw list items --exposed
         │           (subprocess)
         ▼
    Password Manager CLI
         │
         ▼
    Encrypted Vault (read-only)
         │ JSON Response
         ▼
      CRS Module (parse)
         │
         ▼
    Return: []CompromisedCredential
         │ gRPC Response
         ▼
  OpenTUI/Tauri Client (display)


User Action (Rotate Credential with ACVS)
         │
         ▼
  OpenTUI/Tauri Client
         │ gRPC Request (mTLS)
         ▼
    ACM Service API Gateway
         │
         ▼
      ACVS Module (if enabled)
         │
         ├─ Fetch ToS: https://target-site.com/terms
         │            (via Legal NLP)
         ▼
    Legal NLP Engine (Python subprocess)
         │ Parse, Extract Rules
         ▼
    Compliance Rule Set (CRC) [cached]
         │
         ▼
      ACVS Validator
         │ Check: Can automation proceed?
         ├─ Yes, API available ────────┐
         ├─ No, requires HIM ──────────┤
         └─ Blocked by ToS ────────────┤
                                       │
         ┌─────────────────────────────┘
         ▼
      CRS Module
         │ Generate Password
         ▼
    crypto/rand.Read() [secure random]
         │
         ▼
      CRS Module
         │ Update Vault
         ├─ Execute: bw edit item <id> --password "<new>"
         ▼
    Password Manager CLI
         │
         ▼
    Encrypted Vault (write)
         │ Success/Failure
         ▼
      CRS Module
         │ Verify Update
         ├─ Execute: bw get item <id>
         │           (verify new password set)
         ▼
    Password Manager CLI
         │ Confirmation
         ▼
      Audit Logger
         │ Log Event
         ├─ {timestamp, credential_hash, action, status}
         ▼
    SQLite Database (audit.db)
         │ Signed Entry
         ▼
    Evidence Chain (if ACVS enabled)
         │
         ▼
  OpenTUI/Tauri Client (display result)
```

### 4.2 Data Storage

| Data Type | Storage Location | Encryption | Retention |
|-----------|-----------------|------------|-----------|
| **Configuration** | `~/.acm/config/service.yaml` | Plaintext (no secrets) | Indefinite |
| **Client Certificates** | OS Keychain / Windows Cert Store / Linux Secret Service | Platform-managed | Certificate lifetime (1 year default) |
| **Audit Logs** | `~/.acm/data/audit.db` (SQLite) | Sensitive fields encrypted (AES-256-GCM) | User-configurable (default: 1 year) |
| **CRC Cache** | `~/.acm/data/crc_cache/` (JSON files) | Plaintext (no secrets) | 30 days (refreshed on access) |
| **Session Tokens (JWT)** | Memory only (never persisted) | N/A (in-memory only) | 15-30 minutes |
| **Legal NLP Models** | `~/.acm/models/` | Plaintext (public models) | Indefinite |

### 4.3 Audit Log Schema

```sql
CREATE TABLE audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,  -- ISO 8601 format
    event_type TEXT NOT NULL,  -- 'rotation', 'detection', 'compliance_check', 'him_prompt'
    credential_id_hash TEXT NOT NULL,  -- SHA-256 of credential ID (never store plaintext)
    action TEXT NOT NULL,  -- 'detected', 'rotated', 'skipped', 'failed'
    status TEXT NOT NULL,  -- 'success', 'failure', 'pending'
    details TEXT,  -- JSON blob with additional context (encrypted if sensitive)
    crc_rule_applied TEXT,  -- Reference to CRC rule ID (if ACVS enabled)
    evidence_chain_id TEXT,  -- Link to evidence chain entry
    signature TEXT NOT NULL,  -- Ed25519 signature of (id || timestamp || credential_id_hash || action)
    created_at INTEGER NOT NULL  -- Unix timestamp
);

CREATE INDEX idx_timestamp ON audit_events(timestamp);
CREATE INDEX idx_credential ON audit_events(credential_id_hash);
CREATE INDEX idx_event_type ON audit_events(event_type);
```

**Cryptographic Signing:**

```go
// Sign audit entry for tamper-evidence
func SignAuditEntry(entry AuditEvent, privateKey ed25519.PrivateKey) (signature string, err error) {
    message := fmt.Sprintf("%d|%s|%s|%s", 
        entry.ID, 
        entry.Timestamp, 
        entry.CredentialIDHash, 
        entry.Action)
    
    sig := ed25519.Sign(privateKey, []byte(message))
    return base64.StdEncoding.EncodeToString(sig), nil
}

// Verify audit log integrity
func VerifyAuditLog(db *sql.DB, publicKey ed25519.PublicKey) (valid bool, errors []string) {
    rows, err := db.Query("SELECT id, timestamp, credential_id_hash, action, signature FROM audit_events ORDER BY id")
    if err != nil {
        return false, []string{err.Error()}
    }
    defer rows.Close()
    
    var validationErrors []string
    for rows.Next() {
        var entry AuditEvent
        rows.Scan(&entry.ID, &entry.Timestamp, &entry.CredentialIDHash, &entry.Action, &entry.Signature)
        
        message := fmt.Sprintf("%d|%s|%s|%s", entry.ID, entry.Timestamp, entry.CredentialIDHash, entry.Action)
        sig, _ := base64.StdEncoding.DecodeString(entry.Signature)
        
        if !ed25519.Verify(publicKey, []byte(message), sig) {
            validationErrors = append(validationErrors, fmt.Sprintf("Entry %d signature invalid", entry.ID))
        }
    }
    
    return len(validationErrors) == 0, validationErrors
}
```

---

## 5. Security Architecture

### 5.1 Zero-Knowledge Architecture

**Principle:** ACM service never has access to the user's master password or the vault's encryption keys.

**Implementation:**

1. **Password Manager Integration**: ACM interacts with password manager CLI, which handles all vault encryption/decryption
2. **Subprocess Isolation**: CLI executed in separate process with minimal environment variables
3. **Transient Decryption**: Vault entries decrypted by CLI only for duration of read/write operation
4. **No Key Storage**: ACM never stores or logs the master password or derived encryption keys

**Validation:**
- Code audit verifies no password/key storage
- Memory dump analysis during rotation confirms no keys in ACM process memory
- Network traffic analysis confirms zero external transmission of credentials

### 5.2 Authentication and Authorization

#### mTLS Client Certificate Authentication

**Certificate Hierarchy:**

```
ACM Local CA (Self-Signed Root)
├── ACM Service Certificate (server cert)
│   └── CN: localhost
│       Validity: 1 year
│       Key Usage: Digital Signature, Key Encipherment
│       Extended Key Usage: TLS Web Server Authentication
│
├── OpenTUI Client Certificate (client cert #1)
│   └── CN: acm-tui-<device-id>
│       Validity: 1 year
│       Key Usage: Digital Signature
│       Extended Key Usage: TLS Web Client Authentication
│
└── Tauri GUI Client Certificate (client cert #2)
    └── CN: acm-gui-<device-id>
        Validity: 1 year
        Key Usage: Digital Signature
        Extended Key Usage: TLS Web Client Authentication
```

**Certificate Generation Workflow:**

```bash
# Initial setup (automated by acm-service setup command)

1. Generate Local CA
   cfssl gencert -initca ca-csr.json | cfssljson -bare ca

2. Generate Server Certificate
   cfssl gencert -ca=ca.pem -ca-key=ca-key.pem \
     -config=ca-config.json -profile=server \
     server-csr.json | cfssljson -bare server

3. Generate Client Certificates (per client)
   cfssl gencert -ca=ca.pem -ca-key=ca-key.pem \
     -config=ca-config.json -profile=client \
     client-tui-csr.json | cfssljson -bare client-tui
   
   # Store private key in OS Keychain
   security add-generic-password -s "acm-client-tui" \
     -a "$USER" -w "$(cat client-tui-key.pem)" \
     -T /usr/local/bin/acm

4. Configure Service
   # service.yaml
   tls:
     cert_file: ~/.acm/certs/server.pem
     key_file: ~/.acm/certs/server-key.pem
     ca_file: ~/.acm/certs/ca.pem
     client_auth: required
```

**Go TLS Configuration:**

```go
func NewMTLSServer(certFile, keyFile, caFile string, addr string) (*grpc.Server, error) {
    // Load server certificate
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }
    
    // Load CA certificate for client validation
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, err
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    // TLS configuration with mutual authentication
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caCertPool,
        MinVersion:   tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
    }
    
    // Create gRPC server with TLS credentials
    creds := credentials.NewTLS(tlsConfig)
    server := grpc.NewServer(grpc.Creds(creds))
    
    return server, nil
}

func NewMTLSClient(certFile, keyFile, caFile string, addr string) (*grpc.ClientConn, error) {
    // Load client certificate
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }
    
    // Load CA certificate for server validation
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, err
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    // TLS configuration
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
        ServerName:   "localhost",
    }
    
    // Create gRPC client with TLS credentials
    creds := credentials.NewTLS(tlsConfig)
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
    if err != nil {
        return nil, err
    }
    
    return conn, nil
}
```

#### Session Management with JWT

**JWT Structure:**

```json
{
  "header": {
    "alg": "EdDSA",  // Ed25519 signature
    "typ": "JWT"
  },
  "payload": {
    "iss": "acm-service",
    "sub": "acm-tui-<device-id>",
    "aud": "acm-api",
    "exp": 1700000000,  // Unix timestamp (30 min from issue)
    "iat": 1699998200,
    "jti": "unique-token-id",
    "cert_fingerprint": "sha256:abcd1234..."  // Client cert fingerprint for validation
  },
  "signature": "..."
}
```

**Token Lifecycle:**

1. **Issuance**: After successful mTLS handshake, service issues JWT token
2. **Storage**: Client stores token in memory only (never persisted to disk)
3. **Usage**: Client includes token in gRPC metadata for each request
4. **Validation**: Service validates token signature and expiration on each request
5. **Renewal**: Client requests new token before expiration (proactive refresh)
6. **Revocation**: Service maintains in-memory revocation list (for emergency revocation)

### 5.3 Threat Model

| Threat | Attack Vector | Impact | Mitigation |
|--------|---------------|--------|------------|
| **Local System Compromise** | Attacker gains root/admin access to user's machine | Critical — full vault compromise | • Memory locking for sensitive data<br>• Secure enclave integration (future)<br>• User education on endpoint security |
| **Man-in-the-Middle (Network)** | Attacker intercepts localhost traffic | High — potential credential theft | • mTLS with certificate pinning<br>• Localhost-only binding<br>• Network monitoring for unexpected listeners |
| **Certificate Theft** | Attacker steals client certificate | High — can impersonate client | • Private keys in OS Keychain/TPM<br>• Short certificate lifetimes (1 year)<br>• Certificate revocation capability |
| **Password Manager CLI Exploit** | Vulnerability in CLI allows unauthorized vault access | Critical — vault compromise | • Pin specific CLI versions<br>• Monitor CVE disclosures<br>• Community security testing |
| **ToS Violation (Legal)** | User's automation violates site ToS, account terminated | Medium — account loss, potential legal action | • ACVS mandatory for automation<br>• Strong indemnification clause in EULA<br>• Evidence chain for good-faith compliance |
| **Supply Chain Attack** | Malicious dependency introduced | Critical — full system compromise | • Dependency pinning and verification<br>• Reproducible builds<br>• SBOM generation and scanning |
| **Memory Dump/Swap Attack** | Attacker reads process memory or swap file | High — sensitive data exposure | • Memory locking (`mlock`)<br>• Explicit zeroing of buffers<br>• Disable swap for sensitive pages |
| **UI Injection** | Malicious input in HIM prompt | Medium — phishing attack | • Input sanitization and validation<br>• Clear visual indicators of legitimate prompts<br>• No clickable links in prompts |

### 5.4 Secure Development Practices

- **Code Review**: All changes reviewed by 2+ maintainers before merge
- **Security Testing**: Automated SAST (Semgrep, gosec), DAST (OWASP ZAP), dependency scanning (Dependabot)
- **Vulnerability Disclosure**: Responsible disclosure program with security@acm.dev contact
- **Incident Response**: Documented playbook for security incidents with <48h critical patch SLA
- **Security Audits**: Annual third-party penetration testing and code audit

---

## 6. Technology Stack

### 6.1 Core Technologies

| Layer | Technology | Version | Purpose |
|-------|------------|---------|---------|
| **Service Runtime** | Go | 1.21+ | Core ACM service, CRS, ACVS, HIM |
| **API Protocol** | gRPC | 1.60+ | Service-client communication |
| **TLS Library** | crypto/tls (Go stdlib) | stdlib | mTLS implementation |
| **Database** | SQLite | 3.40+ | Audit log storage |
| **NLP Framework** | spaCy | 3.7+ | Legal NLP for ToS analysis |
| **Certificate Management** | cfssl | 1.6+ | Local CA and cert generation |

### 6.2 Client Technologies

| Client | Technology Stack | Rationale |
|--------|------------------|-----------|
| **OpenTUI** | Go + Bubbletea + Lipgloss | • Native Go for seamless gRPC integration<br>• Rich TUI with minimal dependencies<br>• Scriptable CLI commands |
| **Tauri GUI** | Rust + Tauri + React + TypeScript | • Lightweight desktop app (< 5MB binary)<br>• Secure IPC between frontend and backend<br>• Native OS integration (notifications, file pickers)<br>• Low memory footprint |

### 6.3 Build and Deployment

| Tool | Purpose |
|------|---------|
| **GoReleaser** | Cross-platform Go binary builds for ACM service and OpenTUI |
| **Tauri Bundler** | Desktop app packaging (DMG, MSI, AppImage, deb, rpm) |
| **GitHub Actions** | CI/CD pipeline (test, build, release) |
| **Dependabot** | Automated dependency updates |
| **Codecov** | Test coverage tracking |

### 6.4 Development Tools

- **Linting**: golangci-lint (Go), ESLint (TypeScript), Clippy (Rust)
- **Formatting**: gofmt (Go), Prettier (TypeScript), rustfmt (Rust)
- **Testing**: Go test (unit/integration), Playwright (E2E for Tauri)
- **Protobuf**: protoc + protoc-gen-go for gRPC definitions

---

## 7. Deployment Architecture

### 7.1 Deployment Model

**Single-User Deployment (Primary):**

```
User's Workstation
├── ACM Service (daemon)
│   └── Listens: localhost:8443 (gRPC over mTLS)
│
├── OpenTUI Client
│   └── Connects: localhost:8443
│
├── Tauri GUI Client
│   └── Connects: localhost:8443
│
├── Password Manager
│   └── 1Password, Bitwarden, etc. (user-installed)
│
└── OS Services
    ├── Keychain / Certificate Store
    └── systemd / launchd / Windows Service
```

### 7.2 Installation

**macOS:**
```bash
# Install via Homebrew
brew tap acm-project/acm
brew install acm

# Start service (launchd)
brew services start acm

# Install GUI
brew install --cask acm-gui

# Setup (generates certificates)
acm setup
```

**Linux:**
```bash
# Install via package manager
sudo apt install acm          # Debian/Ubuntu
sudo dnf install acm          # Fedora

# Start service (systemd)
sudo systemctl enable acm
sudo systemctl start acm

# Setup
acm setup
```

**Windows:**
```powershell
# Install via Chocolatey
choco install acm

# Start service (Windows Service)
Start-Service ACM

# Setup
acm.exe setup
```

### 7.3 Configuration

**Service Configuration** (`~/.acm/config/service.yaml`):

```yaml
service:
  listen_addr: "127.0.0.1:8443"
  tls:
    cert_file: ~/.acm/certs/server.pem
    key_file: ~/.acm/certs/server-key.pem
    ca_file: ~/.acm/certs/ca.pem
    client_auth_required: true
  
  jwt:
    signing_key_file: ~/.acm/keys/jwt-signing-key.pem
    token_lifetime: 30m
    refresh_threshold: 5m

password_manager:
  type: bitwarden  # or: 1password, lastpass
  cli_path: /usr/local/bin/bw
  session_timeout: 15m

crs:
  password_policy:
    default_length: 32
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: true
  
  rotation_strategy: auto  # auto, manual, prompt

acvs:
  enabled: false  # Explicit opt-in required
  legal_nlp:
    model_path: ~/.acm/models/legal-tos-v1
    cache_dir: ~/.acm/data/crc_cache
    cache_ttl: 720h  # 30 days
  
  evidence_chain:
    signing_key_file: ~/.acm/keys/evidence-signing-key.pem
    export_format: pdf  # pdf, json, both

audit:
  database_path: ~/.acm/data/audit.db
  retention_days: 365
  signing_key_file: ~/.acm/keys/audit-signing-key.pem

logging:
  level: info  # debug, info, warn, error
  file: ~/.acm/logs/service.log
  max_size_mb: 100
  max_backups: 5
```

---

## 8. Monitoring and Observability

### 8.1 Logging

**Log Levels:**
- **DEBUG**: Detailed diagnostic information (CRS/ACVS internal state)
- **INFO**: Normal operational events (rotation started, completed)
- **WARN**: Unexpected but non-critical events (HIM timeout, CRC cache miss)
- **ERROR**: Failure events requiring attention (vault update failed, CLI not found)

**Sensitive Data Protection:**
- Passwords, tokens, master passwords: Never logged
- Credential IDs: Always hashed (SHA-256) before logging
- User input: Sanitized and masked in logs

### 8.2 Metrics (Optional, Opt-In)

If user enables telemetry (opt-in, privacy-preserving):

```go
// Example metrics (exported to Prometheus format on localhost:9090)

acm_rotation_total{status="success|failure|him_required"} counter
acm_rotation_duration_seconds histogram
acm_compliance_check_total{result="allowed|blocked|him_required"} counter
acm_him_prompt_duration_seconds histogram
acm_active_sessions gauge
```

**Privacy Guarantee:** Metrics contain no PII, credential data, or site-specific information.

### 8.3 Health Checks

```bash
# Service health check (returns 200 OK if healthy)
curl --cert client.pem --key client-key.pem \
     --cacert ca.pem \
     https://localhost:8443/health

# Response
{
  "status": "healthy",
  "version": "1.0.0",
  "components": {
    "crs": "healthy",
    "acvs": "disabled",
    "him_manager": "healthy",
    "audit_logger": "healthy"
  },
  "password_manager": {
    "type": "bitwarden",
    "cli_available": true,
    "vault_locked": false
  }
}
```

---

## 9. Scalability and Performance

### 9.1 Performance Targets

| Operation | Target Latency | Max Latency | Measurement |
|-----------|----------------|-------------|-------------|
| Detect compromised credentials | < 2s | 5s | CLI query time |
| Generate secure password | < 100ms | 500ms | crypto/rand overhead |
| Update vault entry | < 3s | 10s | CLI write + verify |
| ToS NLP analysis | < 10s | 30s | spaCy inference time |
| HIM prompt display | < 500ms | 2s | gRPC stream latency |

### 9.2 Scalability Considerations

**Concurrent Operations:**
- ACM service handles concurrent requests from multiple clients (TUI + GUI)
- Go goroutines for parallel credential rotation (configurable max concurrency)
- SQLite Write-Ahead Logging (WAL) mode for concurrent reads

**Resource Limits:**
- Memory: 256MB base, 512MB during NLP processing
- Disk: 100MB for application, 50MB for models, growing audit logs
- CPU: Single-core sufficient for normal use; multi-core beneficial for batch rotations

**Bottlenecks:**
- Password Manager CLI: Sequential execution required (no parallel writes to vault)
- Legal NLP: CPU-bound; consider GPU acceleration for large-scale ToS analysis (future)

---

## 10. Future Enhancements

### 10.1 Planned Features (Phase IV+)

| Feature | Description | Timeline |
|---------|-------------|----------|
| **HSM Integration** | Hardware security module support for certificate and key storage | Year 2 |
| **Enterprise Deployment** | Centralized policy management for organization-wide credential rotation | Year 2 |
| **Federated Legal NLP** | Community-maintained ToS database with shared CRC rules | Year 2 |
| **Browser Extension** | Real-time breach alerts and rotation triggers from browser | Year 2+ |
| **P2P Audit Sync** | Encrypted peer-to-peer synchronization of audit logs across devices | Year 2+ |
| **Mobile App** | iOS/Android app using same ACM service API | Year 3 |

### 10.2 Research Areas

- **Zero-Knowledge Proof of Rotation**: Cryptographic proof of credential change without revealing password
- **Decentralized Identity**: Integration with DIDs (Decentralized Identifiers) for portable identity
- **Automated ToS Monitoring**: Continuous monitoring of ToS changes with diff analysis
- **AI-Assisted Risk Assessment**: ML model predicting breach likelihood based on credential metadata

---

## 11. Appendices

### Appendix A: gRPC API Definition

```protobuf
syntax = "proto3";

package acm.v1;

service CredentialService {
  rpc DetectCompromised(DetectRequest) returns (DetectResponse);
  rpc RotateCredential(RotateRequest) returns (RotateResponse);
  rpc GetRotationStatus(StatusRequest) returns (StatusResponse);
}

service ComplianceService {
  rpc AnalyzeToS(AnalyzeRequest) returns (AnalyzeResponse);
  rpc ValidateAction(ValidateRequest) returns (ValidateResponse);
  rpc ExportEvidenceChain(ExportRequest) returns (ExportResponse);
}

service HIMService {
  rpc PromptUser(stream HIMPrompt) returns (stream HIMResponse);
}

service AuditService {
  rpc QueryLogs(QueryRequest) returns (QueryResponse);
  rpc VerifyIntegrity(VerifyRequest) returns (VerifyResponse);
}

message DetectRequest {
  string password_manager_type = 1;  // "bitwarden", "1password", "lastpass"
}

message DetectResponse {
  repeated CompromisedCredential credentials = 1;
}

message CompromisedCredential {
  string id = 1;  // Hashed credential ID
  string site = 2;
  string username = 3;
  string breach_name = 4;
  int64 breach_date = 5;  // Unix timestamp
  int64 last_rotated = 6;  // Unix timestamp (0 if never)
  bool requires_him = 7;
}

message RotateRequest {
  string credential_id = 1;
  PasswordPolicy policy = 2;
  bool acvs_enabled = 3;
}

message RotateResponse {
  string credential_id = 1;
  string status = 2;  // "success", "failure", "him_required"
  string new_password = 3;  // Only if success
  string error_message = 4;  // Only if failure
  ComplianceValidation compliance = 5;  // Only if acvs_enabled
}

// ... additional message definitions
```

### Appendix B: Certificate Authority Setup Script

*(Complete bash script for automated CA and certificate generation available in project repository)*

### Appendix C: Security Audit Checklist

| Check | Description | Status |
|-------|-------------|--------|
| **Authentication** | mTLS with client certificates verified | ✓ |
| **Authorization** | JWT tokens validated on each request | ✓ |
| **Encryption** | TLS 1.3 enforced, strong ciphers only | ✓ |
| **Memory Safety** | Sensitive data locked and zeroed | ✓ |
| **Input Validation** | All user input sanitized and validated | ✓ |
| **Output Encoding** | Log output sanitized (no credential data) | ✓ |
| **Dependency Security** | Regular vulnerability scanning (Dependabot) | ✓ |
| **Code Review** | All PRs reviewed by 2+ maintainers | ✓ |
| **Penetration Testing** | Annual third-party security assessment | Planned |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2025-11-13 | Initial Draft | Created from ACM research and PRD |
| 1.0 | 2025-11-13 | Claude (AI Assistant) | Complete TAD with service-client architecture, mTLS, and local-first design |

---

**Document Status:** Draft — Pending Technical Review  
**Next Review Date:** [TBD]  
**Distribution:** Public (Open-Source Project Documentation)
