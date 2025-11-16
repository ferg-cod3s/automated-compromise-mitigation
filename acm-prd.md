# Product Requirements Document (PRD)
# Automated Compromise Mitigation (ACM)

**Version:** 1.0  
**Date:** November 2025  
**Status:** Draft  
**Project Type:** Open-Source Security Tool

---

## Executive Summary

### Overview

Automated Compromise Mitigation (ACM) is a local-first, zero-knowledge application that automatically detects and remediates compromised credentials following data breaches. The system operates entirely on the user's device, integrating with password manager CLIs to rotate credentials only when necessary—after confirmed compromise—while maintaining strict adherence to zero-knowledge security principles.

### Core Value Proposition

ACM resolves the fundamental conflict between security necessity (credential rotation post-breach) and user convenience (avoiding the behavioral burdens that lead to poor password hygiene). By automating rotation exclusively when required and operating under a local-first, zero-knowledge paradigm, ACM maximizes security while maintaining user trust.

### Strategic Architecture

ACM is composed of two separable, strictly local-first services:

1. **Credential Remediation Service (CRS)**: Interfaces with password manager CLIs, generates new credentials, and updates local vault entries. Operates purely locally with zero-knowledge security mandate.

2. **Automated Compliance Validation Service (ACVS)**: Executes automated rotation workflows and validates third-party Terms of Service (ToS) compliance. Operates as opt-in feature using personal API tokens and authenticated methods provided by end-users. All processing, including Legal NLP models for ToS analysis, runs locally on the user's device.

---

## 1. Product Vision and Goals

### 1.1 Vision Statement

To provide security-conscious users with a trustworthy, transparent, and automated solution for credential breach remediation that never compromises their master password, vault data, or privacy.

### 1.2 Strategic Goals

| Goal | Description | Success Metric |
|------|-------------|----------------|
| **Automate breach response** | Automatically detect and rotate compromised credentials via password manager CLIs | 100% success rate in rotating detected compromised credentials within local vault |
| **Maintain zero-knowledge security** | Ensure all sensitive data and processing remain local to user's device | Zero external data transmission; passes independent security audits |
| **Comply with third-party ToS** | Validate automation actions and avoid ToS violations through ACVS | Zero reports of ToS violations triggered by ACM automation |
| **Build user trust via transparency** | Operate as open-source, auditable project with clear security documentation | Active community contributions; positive security reviews |
| **Legal protection for contributors** | Enforce strong EULA with indemnification and limitation of liability | Legal framework approved by community legal review |
| **Community-driven governance** | Establish clear, consensus-based decision-making for technical and legal matters | Active contributor base; transparent RFC process |

### 1.3 Non-Goals (Explicit Scope Limitations)

- **Not a password manager replacement**: ACM augments existing password managers, never replaces them
- **Not universal unattended automation**: MFA, CAPTCHA, and anti-bot measures require Human-in-the-Middle (HIM) workflows
- **Not a cloud service**: ACM runs exclusively locally; no SaaS offering planned
- **Not a breach detection service**: ACM relies on existing password manager breach reports or user-provided breach data
- **Not suitable for non-technical users initially**: Phase I targets security professionals and developers

---

## 2. Target Users and Personas

### 2.1 Primary User Personas

#### Persona 1: Security-Conscious Developer
- **Profile**: Software engineer or security researcher who manages 100+ accounts
- **Pain Points**: Manual rotation is time-consuming; concerned about breach response time; distrusts cloud-based solutions
- **Needs**: CLI-first workflow, scriptable automation, transparent security model
- **Technical Skill**: High — comfortable with CLIs, understands mTLS and certificate-based auth

#### Persona 2: Privacy-Focused Professional
- **Profile**: Journalist, activist, or professional handling sensitive information
- **Pain Points**: Cannot trust centralized services; requires audit trail; needs immediate breach response
- **Needs**: Local-first operation, zero-knowledge guarantee, compliance evidence chain
- **Technical Skill**: Medium — can follow setup instructions, prefers GUI but will use TUI if necessary

#### Persona 3: Compliance Officer / IT Security Manager
- **Profile**: Manages credential policies for organization or personal high-security accounts
- **Pain Points**: Needs proof of timely credential rotation; must document compliance; requires automation that respects ToS
- **Needs**: Audit logs, compliance validation reports, ToS analysis, evidence chains
- **Technical Skill**: Medium-high — understands security concepts, needs clear documentation

### 2.2 Secondary User Personas (Future Phases)

- **Enterprise Security Teams**: Deploy ACM across workstations with centralized policy management
- **Password Manager Power Users**: Early adopters who want cutting-edge security features
- **Open-Source Security Contributors**: Developers who extend ACM with new integrations or compliance models

---

## 3. Key Features and Functional Requirements

### 3.1 Core Components

| Component | Description | Key Functionality |
|-----------|-------------|-------------------|
| **Credential Remediation Service (CRS)** | Local-only remediation engine integrated with password manager CLIs | • Detect compromised credentials via built-in password manager reports<br>• Generate cryptographically secure new passwords<br>• Update vault entries via authenticated CLI<br>• Log actions to local audit chain |
| **Automated Compliance Validation Service (ACVS)** | Legal and automation compliance engine (opt-in) | • Validate automation actions via locally-run Legal NLP models<br>• Generate Compliance Rule Set (CRC) from ToS documents<br>• Pause for Human-in-the-Middle (HIM) workflows when MFA/CAPTCHA detected<br>• Record compliance validation results |
| **Human-in-the-Middle (HIM) Workflow** | User-assisted automation checkpoint | • Detect when automation cannot proceed (MFA, CAPTCHA, anti-bot)<br>• Prompt user securely via TUI or GUI<br>• Resume automation after user validation<br>• Log HIM events for audit trail |
| **Legal NLP Engine** | Locally executed ToS analysis engine | • Parse anti-bot and automation clauses from website ToS<br>• Extract rate limits, prohibited actions, authentication requirements<br>• Generate structured Compliance Rule Set (CRC) |
| **Audit & Evidence Chain System** | Compliance logging and verification mechanism | • Record all automated actions with timestamps<br>• Link actions to applicable ToS CRC rules<br>• Maintain cryptographically signed local audit logs<br>• Generate compliance reports |

### 3.2 Password Manager Integration Requirements

| Password Manager | CLI Tool | Required Capabilities | Implementation Priority |
|------------------|----------|----------------------|------------------------|
| **1Password** | `op` CLI | Item edit, password generation, secret reference syntax | Phase I (MVP) |
| **Bitwarden** | `bw` CLI | Local REST API via `bw serve`, item CRUD operations | Phase I (MVP) |
| **LastPass** | `lpass` CLI | Item edit, secure note access | Phase II |

**Integration Specifications:**
- Must support authenticated local access to vault
- Must allow programmatic item modification (read + write)
- Must support secure password generation
- Must maintain vault encryption during operations

### 3.3 Authentication and Security Architecture

| Component | Technology | Purpose |
|-----------|------------|---------|
| **ACM Service Core** | Go | High-performance local daemon exposing APIs |
| **Client Authentication** | mTLS with X.509 client certificates | Device-bound identity verification |
| **Session Management** | Short-lived JWT tokens (15-30 min) | Stateless authorization between service and clients |
| **Certificate Authority** | Local CA (cfssl or openssl) | Issue and revoke device certificates |
| **Secure Storage** | OS Keychain / TPM 2.0 | Store client private keys hardware-backed |

**Critical Security Requirements:**
- All client-service communication over localhost-only mTLS
- Zero external network dependencies for authentication
- Client certificates stored in platform secure storage (Keychain, TPM, Windows Certificate Store)
- Private keys never leave secure storage; all operations via secure enclave
- Session tokens stored in memory only (never persisted to disk)

### 3.4 User Interface Requirements

| Interface | Technology | Target Users | Key Features |
|-----------|------------|--------------|--------------|
| **OpenTUI (CLI)** | Go + Charm Bubbletea | Developers, security professionals, automation users | • Rich terminal UI with keyboard navigation<br>• Scriptable commands for automation<br>• Real-time status updates<br>• Audit log viewer |
| **Tauri Desktop GUI** | Rust + Tauri + Web UI | Privacy-focused professionals, compliance officers | • Visual workflow for credential management<br>• Interactive HIM prompts for MFA/CAPTCHA<br>• Compliance dashboard with ToS validation status<br>• Evidence chain export |

**UI Shared Requirements:**
- Both interfaces communicate with same ACM Service via local gRPC API
- Consistent feature parity (TUI may have additional scriptable commands)
- Secure credential display (masked by default, reveal on explicit action)
- Clear indication when ACVS is enabled (opt-in status visible)

---

## 4. User Stories and Acceptance Criteria

### 4.1 Core User Stories

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| **US-01** | As a security-conscious user, I want the system to automatically detect compromised credentials from my password manager's breach reports | • CRS successfully queries password manager CLI for exposed credentials<br>• System displays list of compromised items<br>• Detection works offline using local breach databases |
| **US-02** | As a privacy-focused user, I want all credential processing to occur locally so my master password and vault data are never exposed | • No external API calls or network requests during credential rotation<br>• Independent security audit confirms zero-knowledge architecture<br>• Open-source code review shows no telemetry or data exfiltration |
| **US-03** | As a user, I want to automatically rotate a compromised password with a strong, unique replacement | • CRS generates cryptographically secure password (configurable length/complexity)<br>• CRS updates vault entry via password manager CLI<br>• Operation completes within 5 seconds<br>• Audit log records successful rotation |
| **US-04** | As a user, I want to be prompted to manually complete rotation when MFA or CAPTCHA is required | • ACVS detects MFA/CAPTCHA requirement during automated rotation<br>• System pauses automation and displays HIM prompt<br>• Prompt includes clear instructions and secure input method<br>• Automation resumes after user completes verification |
| **US-05** | As a compliance-focused user, I want ACM to validate that automation complies with target website's Terms of Service | • ACVS analyzes target site's ToS using local Legal NLP<br>• System generates Compliance Rule Set (CRC) with actionable rules<br>• Rotation attempt cross-references CRC before proceeding<br>• Non-compliant actions are blocked with explanation |
| **US-06** | As a legal contributor, I want to ensure the project has strong liability protection for developers | • EULA includes explicit indemnification clause<br>• Limitation of Liability caps project exposure<br>• ACVS opt-in requires user acceptance of EULA terms<br>• Legal framework reviewed by qualified counsel |
| **US-07** | As a developer, I want to verify ACM's security claims by auditing the source code | • Complete source code available on GitHub with permissive license (MIT/Apache 2.0)<br>• Security-critical code sections clearly documented<br>• Memory handling patterns documented and justified<br>• Build process is reproducible and verifiable |
| **US-08** | As a user, I want to view a complete audit trail of all credential rotations | • Audit log records: timestamp, credential ID (hashed), action taken, compliance status<br>• Logs cryptographically signed with local keypair<br>• Export functionality generates human-readable compliance reports<br>• Logs stored locally with configurable retention |
| **US-09** | As a CLI power user, I want to script credential rotation for automated workflows | • OpenTUI exposes scriptable commands (`acm rotate <item-id>`)<br>• Commands support JSON output for parsing<br>• Exit codes indicate success/failure/HIM-required<br>• Integration with shell scripts and cron jobs |
| **US-10** | As a GUI user, I want a visual workflow to review and approve automated rotations | • Tauri GUI displays pending rotations with context (site, last breach date)<br>• Batch approval with review step<br>• Visual HIM prompts for MFA with secure input<br>• Compliance status dashboard with ToS validation results |

### 4.2 ACVS-Specific User Stories (Opt-In Feature)

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| **US-11** | As a user opting into ACVS, I want explicit control over which credentials use automated rotation | • ACVS configuration UI shows list of credentials<br>• Per-credential toggle for automated rotation<br>• Bulk enable/disable with filtering<br>• Changes require re-confirmation of EULA acceptance |
| **US-12** | As a user, I want ACVS to pause automation if it detects potential ToS violation | • Legal NLP identifies anti-automation clauses in target site's ToS<br>• CRC flags rotation attempt as potentially violating<br>• ACVS displays violation details and recommended action<br>• User can override (accepting liability) or cancel |
| **US-13** | As a user, I want to provide my own API tokens for services that offer official APIs | • ACVS configuration accepts API tokens for supported services<br>• Tokens stored encrypted in OS Keychain<br>• API-based rotation preferred over UI scripting when available<br>• Token validation before rotation attempt |
| **US-14** | As a user, I want ACVS to generate evidence chains proving compliance attempts | • Evidence chain includes: ToS version analyzed, CRC rules applied, validation results<br>• Evidence cryptographically signed and timestamped<br>• Export generates PDF compliance report<br>• Evidence admissible for demonstrating good-faith compliance |

---

## 5. Technical Requirements

### 5.1 Platform and Environment

| Requirement | Specification |
|-------------|---------------|
| **Supported Operating Systems** | macOS 11+, Windows 10+, Linux (Ubuntu 20.04+, Fedora 35+) |
| **Architecture** | x86_64, ARM64 (including Apple Silicon) |
| **Runtime Requirements** | Go 1.21+, Password manager CLI installed and configured |
| **Storage Requirements** | 100 MB for application, 50 MB for local NLP models, 10 MB for audit logs |
| **Memory Requirements** | 256 MB base, 512 MB during NLP processing |
| **Network Requirements** | None (localhost-only communication; external access only for password manager sync if enabled by user) |

### 5.2 Security Requirements

| Category | Requirement | Validation Method |
|----------|-------------|-------------------|
| **Zero-Knowledge Architecture** | Master password never leaves password manager; ACM never stores or transmits vault encryption keys | Security audit + code review |
| **Local-First Operation** | All sensitive processing occurs on-device; no cloud dependencies | Network traffic analysis during operation |
| **Data Encryption** | Audit logs encrypted at rest; certificates stored in platform secure storage | Penetration testing |
| **Authentication** | mTLS with client certificates for all service-client communication | TLS handshake analysis |
| **Secure Memory Handling** | Sensitive data cleared from memory after use; no swap for sensitive pages | Memory dump analysis |
| **Open-Source Transparency** | Complete source code auditable; reproducible builds | Community code review |

### 5.3 Performance Requirements

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| **Credential Detection** | < 2 seconds to query password manager for compromised items | Automated benchmark |
| **Password Generation** | < 100 ms to generate secure password | Unit test |
| **Vault Update** | < 3 seconds to update single credential via CLI | Integration test |
| **ToS Analysis (ACVS)** | < 10 seconds to parse and generate CRC from ToS document | NLP benchmark |
| **HIM Prompt Display** | < 500 ms from detection to user prompt | E2E test |
| **Audit Log Write** | < 50 ms to write and sign log entry | Performance test |

### 5.4 Compliance and Legal Requirements

| Requirement | Implementation | Enforcement Mechanism |
|-------------|----------------|----------------------|
| **EULA Acceptance** | User must accept EULA before first use | Application initialization gate |
| **ACVS Opt-In** | ACVS disabled by default; explicit opt-in with re-confirmation of EULA | Configuration flag + UI workflow |
| **Indemnification Clause** | EULA transfers liability for ToS violations to end-user | Legal document + acceptance log |
| **Limitation of Liability** | Caps project's aggregate financial exposure | Legal document |
| **Disclaimer of Warranties** | "As-is" and "as-available" software without guarantees | Legal document |
| **ToS Compliance Validation** | ACVS validates every rotation against target site's ToS CRC | Pre-flight check before automation |
| **Evidence Chain Generation** | All ACVS actions logged with compliance validation results | Automatic background process |

---

## 6. Non-Functional Requirements

### 6.1 Reliability

- **Vault Integrity**: Failed rotations must never corrupt vault data; atomic operations with rollback
- **Graceful Degradation**: If password manager CLI unavailable, fail safely with clear error message
- **Idempotency**: Repeated rotation attempts for same credential must be safe and predictable
- **Error Recovery**: Partial rotations must be resumable or cleanly reversed

### 6.2 Usability

- **Installation**: Single-command install via package managers (Homebrew, apt, Chocolatey)
- **Configuration**: Zero-config for basic use; advanced settings in `.acm/config.yaml`
- **Documentation**: Comprehensive user guide, security architecture documentation, API reference
- **Error Messages**: Clear, actionable error messages with suggested remediation steps

### 6.3 Maintainability

- **Code Quality**: Go code follows standard idioms; passes `golangci-lint` checks
- **Testing**: 80%+ unit test coverage; integration tests for all CLI interactions; E2E tests for HIM workflows
- **Documentation**: All public APIs documented with GoDoc; architecture decisions recorded in ADRs
- **Versioning**: Semantic versioning; clear upgrade paths with migration guides

### 6.4 Extensibility

- **Password Manager Plugins**: Modular architecture allows adding new password manager integrations
- **Legal NLP Models**: Plugin system for custom ToS analysis models
- **Audit Log Exporters**: Extensible format system (JSON, CSV, PDF)
- **API Stability**: Public gRPC API versioned; breaking changes follow deprecation policy

---

## 7. Dependencies and Integration Points

### 7.1 External Dependencies

| Dependency | Purpose | Version | License |
|------------|---------|---------|---------|
| **Password Manager CLI** | Vault access and modification | 1Password CLI 2.x, Bitwarden CLI 2023.x | Proprietary / GPL-3.0 |
| **Go Cryptography** | TLS, certificate handling, secure random | stdlib crypto/* | BSD-3-Clause |
| **Legal NLP Models** | ToS parsing and CRC generation | spaCy 3.x or HuggingFace Transformers | MIT / Apache 2.0 |
| **gRPC** | Service-client communication | grpc-go 1.60+ | Apache 2.0 |
| **SQLite** | Local audit log storage | 3.40+ | Public Domain |

### 7.2 Integration Requirements

**Password Manager CLI Integration:**
- Must be installed and authenticated before ACM can operate
- ACM service user must have permission to execute CLI commands
- CLI must support machine-readable output (JSON preferred)

**Operating System Integration:**
- Keychain/Credential Manager access for certificate storage
- File system permissions for `.acm/` configuration directory
- Network loopback interface for localhost-only communication

**Optional Integrations:**
- Breach notification APIs (e.g., HaveIBeenPwned) for supplemental breach detection
- Organization CA for enterprise certificate issuance

---

## 8. Risk Assessment Summary

### 8.1 Critical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Automation blocked by MFA/CAPTCHA** | High — prevents unattended rotation | High — widespread deployment of anti-bot measures | **HIM Workflow**: Mandatory pause for user intervention; clear UX for verification |
| **ToS violation liability** | High — user account termination, potential legal action | Medium — varies by site and automation method | **ACVS**: Mandatory compliance validation with CRC; opt-in with EULA indemnification |
| **Local compromise exposes decrypted credentials** | Critical — full vault compromise | Low — requires local system access | **Security hardening**: Memory locking, secure storage integration, process isolation |
| **Legal liability for project contributors** | High — lawsuits from users or third parties | Low-Medium — depends on EULA enforcement | **Legal Framework**: Strong indemnification + limitation of liability + open-source license |
| **Password manager CLI API changes** | Medium — breaks integration | Medium — CLIs evolve over time | **Version detection**: Pin to specific CLI versions; graceful degradation; community-driven updates |

### 8.2 Secondary Risks

- **NLP model inaccuracy**: Legal NLP misinterprets ToS → false positives/negatives in CRC
- **Certificate management complexity**: Users struggle with mTLS setup → abandonment
- **Performance on low-end hardware**: NLP processing slow on older machines → poor UX
- **Open-source governance disputes**: Community disagreements stall development → fork risk

*(Full risk assessment in separate Risk Assessment & Mitigation Plan document)*

---

## 9. Success Metrics and KPIs

### 9.1 Adoption Metrics

| Metric | Target (6 months) | Target (1 year) | Measurement |
|--------|-------------------|-----------------|-------------|
| **GitHub Stars** | 500+ | 2,000+ | GitHub API |
| **Downloads** | 5,000+ | 25,000+ | Release analytics |
| **Active Users** | 1,000+ | 10,000+ | Opt-in telemetry (anonymous) |
| **Community Contributors** | 10+ | 50+ | GitHub contributor graph |

### 9.2 Security Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Security Audits Passed** | 2+ independent audits with no critical findings | Audit reports |
| **Reported Vulnerabilities** | < 5 CVEs with CVSS > 7.0 | Security disclosure program |
| **Time to Patch Critical Vulnerabilities** | < 48 hours | Incident response logs |

### 9.3 Quality Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Code Coverage** | > 80% | CI pipeline (codecov) |
| **User-Reported Bugs** | < 20 open bugs with "critical" severity | GitHub Issues |
| **Documentation Completeness** | 100% of public APIs documented | Automated doc checks |

### 9.4 Compliance Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **ToS Violation Reports** | 0 confirmed violations caused by ACM automation | User feedback + incident reports |
| **ACVS Adoption Rate** | > 30% of users enable ACVS | Configuration telemetry (opt-in) |
| **Evidence Chain Generation Success** | 100% of ACVS rotations generate valid evidence | Audit log analysis |

---

## 10. Development Phases and Roadmap

### Phase I: MVP — Credential Remediation Service (CRS) Focus
**Timeline:** Months 1-4  
**Status:** Not Started

**Objectives:**
- Establish zero-knowledge foundation using password manager CLIs
- Implement local breach detection via built-in password manager reports
- Enable manual rotation workflow with HIM

**Deliverables:**
- CRS service with 1Password and Bitwarden CLI integration
- OpenTUI client with credential detection and rotation commands
- Basic HIM workflow (user manually visits site to complete rotation)
- Local audit logging (non-compliance-focused)
- Security documentation and threat model
- Initial EULA and ToS (without ACVS provisions)

**Success Criteria:**
- Successfully detects compromised credentials from Bitwarden/1Password breach reports
- Generates secure passwords and updates vault entries via CLI
- Completes full rotation workflow (detect → generate → update → verify) in < 10 seconds
- Passes initial security review by community security experts

---

### Phase II: ACVS — Legal NLP & ToS Compliance
**Timeline:** Months 5-8  
**Status:** Not Started

**Objectives:**
- Introduce non-fragile automation using user-provided API tokens
- Implement compliance validation with local Legal NLP
- Establish evidence chain and audit system

**Deliverables:**
- ACVS service with Legal NLP engine (spaCy or Transformers)
- Compliance Rule Set (CRC) generator from ToS documents
- Opt-in ACVS configuration UI in both OpenTUI and Tauri clients
- API-based rotation for services with documented APIs
- Updated EULA with ACVS provisions and indemnification
- Evidence chain export (JSON, PDF formats)

**Success Criteria:**
- Legal NLP accurately extracts anti-automation clauses from 10 popular websites' ToS
- CRC validation prevents prohibited automation attempts
- API-based rotation works for 5+ major services (e.g., GitHub, AWS, Google, Microsoft)
- Evidence chain logs are cryptographically verifiable

---

### Phase III: Advanced Automation & HIM Workflows
**Timeline:** Months 9-12  
**Status:** Not Started

**Objectives:**
- Expand automation capabilities with controlled UI scripting
- Enhance HIM workflows for complex MFA scenarios
- Improve NLP model accuracy and coverage

**Objectives:**
- Limited UI scripting for low-risk, non-financial sites (with ACVS validation)
- Advanced HIM manager with support for TOTP, SMS, push notifications
- Tauri GUI with full feature parity to OpenTUI
- Enhanced audit/evidence system with compliance reporting dashboard
- Performance optimizations for NLP processing
- Multi-language Legal NLP models (support non-English ToS)

**Deliverables:**
- UI scripting framework with browser automation (Playwright or Selenium)
- HIM workflows for hardware security keys (FIDO2/WebAuthn)
- Compliance dashboard in Tauri GUI showing ToS validation history
- Batch rotation workflows (rotate all compromised credentials with one command)
- Integration with additional password managers (LastPass, KeePassXC)

**Success Criteria:**
- UI scripting successfully rotates credentials for 20+ websites with ACVS validation
- HIM workflows handle 95% of common MFA methods (TOTP, SMS, push, hardware key)
- NLP model achieves 90%+ accuracy on ToS clause extraction (validated against legal expert annotations)
- User satisfaction score > 4.0/5.0 (from opt-in surveys)

---

### Phase IV: Enterprise and Ecosystem (Future)
**Timeline:** Year 2+  
**Status:** Future Consideration

**Potential Features:**
- Enterprise deployment with centralized policy management
- Integration with SSO providers and identity platforms
- Federated Legal NLP model sharing (community-maintained ToS database)
- Hardware security module (HSM) support for certificate storage
- Peer-to-peer encrypted synchronization of audit logs
- Browser extension for real-time breach alerts and rotation triggers

---

## 11. Open Questions and Decisions Needed

### 11.1 Technical Decisions

| Question | Options | Decision Owner | Deadline |
|----------|---------|----------------|----------|
| **Legal NLP Framework** | spaCy vs. HuggingFace Transformers vs. Custom | Tech Lead | Before Phase II |
| **gRPC vs. REST API** | gRPC for performance vs. REST for simplicity | Architecture Team | Before Phase I |
| **Certificate Management UX** | Auto-generate certs vs. User-managed CA | Security + UX Team | Before Phase I |
| **Tauri vs. Electron for GUI** | Already decided (Tauri) but confirm | Architecture Team | Confirmed ✓ |

### 11.2 Legal and Governance Decisions

| Question | Options | Decision Owner | Deadline |
|----------|---------|----------------|----------|
| **Open-Source License** | MIT vs. Apache 2.0 vs. GPL-3.0 | Governance Committee | Before public release |
| **EULA Review** | Internal draft vs. External counsel | Legal Advisor | Before Phase I release |
| **Indemnification Enforceability** | Jurisdiction-specific variations needed? | Legal Advisor | Before ACVS release |
| **Community Governance Model** | BDFL vs. Steering Committee vs. Foundation | Founder + Community | Before Phase II |

### 11.3 Product Strategy Decisions

| Question | Options | Decision Owner | Deadline |
|----------|---------|----------------|----------|
| **ACVS Default State** | Opt-in (current plan) vs. Opt-out with warnings | Product + Legal | Confirmed: Opt-in ✓ |
| **Supported Password Managers Priority** | 1Password + Bitwarden only vs. Broader coverage | Product Team | Before Phase I |
| **Telemetry and Analytics** | Strictly opt-in vs. Privacy-preserving analytics | Privacy + Product Team | Before Phase I |

---

## 12. Stakeholder Matrix

| Stakeholder | Role | Interest | Influence | Engagement Strategy |
|-------------|------|----------|-----------|---------------------|
| **End Users** | Primary users of ACM | High — security and convenience | High — adoption and feedback | Regular community calls, GitHub Discussions, feature voting |
| **Open-Source Contributors** | Developers extending ACM | High — code quality and architecture | High — code contributions and reviews | Clear contribution guidelines, ADR documentation, mentorship |
| **Security Researchers** | Auditors and vulnerability reporters | High — security correctness | Medium — vulnerability disclosure | Responsible disclosure program, bug bounties (future), acknowledgments |
| **Legal Experts** | Review EULA and ToS compliance | Medium — legal framework correctness | Medium — legal guidance | Community legal review committee, external counsel (future) |
| **Password Manager Vendors** | Providers of CLI tools ACM integrates with | Low-Medium — API stability | Low — CLI breaking changes | Monitor release notes, maintain compatibility layer |

---

## 13. Appendices

### Appendix A: Terminology and Definitions

- **ACM**: Automated Compromise Mitigation
- **CRS**: Credential Remediation Service — local service handling password rotation
- **ACVS**: Automated Compliance Validation Service — opt-in service validating ToS compliance
- **HIM**: Human-in-the-Middle — workflow where user manually completes automation step
- **CRC**: Compliance Rule Set — structured rules derived from Legal NLP analysis of ToS
- **mTLS**: Mutual Transport Layer Security — authentication using client and server certificates
- **Zero-Knowledge**: Architecture where service never has access to user's master password or vault encryption keys
- **Local-First**: Application operates entirely on user's device with no cloud dependencies

### Appendix B: References

1. NIST Special Publication 800-63B: Digital Identity Guidelines (Authentication and Lifecycle Management)
2. OWASP Password Storage Cheat Sheet
3. Password Manager CLI Documentation:
   - 1Password CLI Reference: https://developer.1password.com/docs/cli
   - Bitwarden CLI Reference: https://bitwarden.com/help/cli/
4. Legal NLP Resources:
   - spaCy Industrial-Strength NLP: https://spacy.io/
   - HuggingFace Transformers: https://huggingface.co/transformers/
5. Secure Local-First Authentication:
   - @localfirst/auth: https://github.com/local-first-web/auth

### Appendix C: Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2025-11-13 | Initial Draft | Created from research document |
| 1.0 | 2025-11-13 | Claude (AI Assistant) | Comprehensive PRD based on ACM research and architectural decisions |

---

## 14. Approval and Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| **Product Owner** | [TBD] | _______________ | ________ |
| **Technical Lead** | [TBD] | _______________ | ________ |
| **Security Lead** | [TBD] | _______________ | ________ |
| **Legal Advisor** | [TBD] | _______________ | ________ |

---

**Document Status:** Draft — Pending Stakeholder Review  
**Next Review Date:** [TBD]  
**Distribution:** Public (Open-Source Project Documentation)
