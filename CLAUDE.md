# CLAUDE.md - AI Assistant Guide for ACM Repository

**Last Updated:** 2025-11-16
**Repository:** Automated Compromise Mitigation (ACM)
**Status:** Documentation/Planning Phase
**Project Type:** Open-Source Security Tool - Local-First Credential Breach Response

---

## 🎯 Quick Context

This repository contains **comprehensive planning and documentation** for the Automated Compromise Mitigation (ACM) project. It is currently in the **documentation/planning phase** with no production code yet. The repository consists of detailed technical specifications, legal frameworks, security assessments, and community building resources that will guide the implementation phase.

### Critical Security Context

**IMPORTANT:** ACM is a security-critical application dealing with credential management. When working on this repository:
- All changes must prioritize security over convenience
- Legal implications must be carefully considered (EULA, ToS compliance)
- Zero-knowledge and local-first principles are **non-negotiable**
- Any code suggestions must be reviewed against the threat model (acm-threat-model.md)

---

## 📂 Repository Structure

### Root Directory Organization

```
automated-compromise-mitigation/
├── 00-INDEX.md                      # Master index of all documentation
├── README.md                        # Project readme (minimal)
├── LICENSE                          # Open-source license
│
├── Core Planning Documents (58 KB)
│   ├── acm-prd.md                   # Product Requirements Document
│   └── acm-governance-roadmap.md    # 4-phase roadmap + governance model
│
├── Technical Architecture (115 KB)
│   ├── acm-tad.md                   # Technical Architecture Document
│   └── acm-threat-model.md          # STRIDE threat modeling analysis
│
├── Security & Risk (94 KB)
│   ├── acm-security-planning.md     # Security implementation roadmap
│   └── acm-risk-assessment.md       # 41 identified risks with mitigations
│
├── Legal & Compliance (47 KB)
│   └── acm-legal-framework.md       # EULA, indemnification, liability limits
│
├── Community Building (33 KB)
│   └── acm-community-building.md    # Community strategy and engagement
│
├── Executable Scripts (51 KB)
│   ├── setup-github-project.sh      # Creates GitHub Project board + issues
│   └── community-building-setup.sh  # Creates CONTRIBUTING.md, templates, etc.
│
├── Product Documentation Archive
│   └── product-docs/Archive.zip     # Additional archived documentation
│
└── Configuration
    ├── .gitignore                   # Go-focused gitignore
    └── CLAUDE.md                    # This file
```

---

## 📖 Key Documentation Files

### Must-Read Documents (in order)

1. **00-INDEX.md** - Start here for complete overview and navigation
2. **acm-prd.md** - Understand product vision, goals, user personas, features
3. **acm-tad.md** - Learn technical architecture, components, tech stack
4. **acm-security-planning.md** - Security roadmap and implementation requirements
5. **acm-threat-model.md** - Threat analysis using STRIDE methodology

### Document Purposes

| Document | Purpose | When to Consult |
|----------|---------|-----------------|
| **acm-prd.md** | Product requirements, user stories, success metrics | Understanding "what" and "why" |
| **acm-tad.md** | System architecture, component design, tech stack | Understanding "how" to build |
| **acm-threat-model.md** | Security threats, attack vectors, mitigations | Security reviews, threat analysis |
| **acm-security-planning.md** | Security roadmap, controls, implementation phases | Security feature planning |
| **acm-risk-assessment.md** | 41 identified risks with severity and mitigation | Risk analysis, project planning |
| **acm-legal-framework.md** | EULA, liability, indemnification, ToS compliance | Legal questions, license issues |
| **acm-governance-roadmap.md** | Project governance, decision-making, 4-phase roadmap | Process questions, timelines |
| **acm-community-building.md** | Community engagement, contributor onboarding | Community management tasks |

---

## 🔑 Core Project Principles

### Architectural Principles (Non-Negotiable)

1. **Zero-Knowledge Security**
   - Master passwords and vault encryption keys NEVER accessible to ACM
   - No storage or transmission of password manager credentials
   - Enforced through architecture, not just policy

2. **Local-First Operation**
   - All sensitive processing occurs on user's device
   - No cloud dependencies for core functionality
   - Network communication restricted to localhost only (mTLS)

3. **Service-Client Separation**
   - Business logic centralized in Go service daemon
   - Clients (OpenTUI/Tauri) are thin presentation layers
   - Communication via gRPC with mTLS authentication

4. **Defense in Depth**
   - Multiple security layers: mTLS, certificate auth, encrypted storage
   - Fail-safe defaults (ACVS disabled by default)
   - Security-critical features require explicit opt-in

5. **Transparency & Auditability**
   - Open-source, auditable codebase
   - Reproducible builds
   - Architecture Decision Records (ADRs) for major decisions

### Legal & Compliance Principles

- **Strong EULA Required** - Indemnification and liability limitation (requires legal review)
- **ToS Compliance Mandatory** - ACVS validates automation against third-party ToS
- **Human-in-the-Middle (HIM) Workflows** - User intervention required for MFA/CAPTCHA
- **Audit Trail Required** - All automated actions logged with cryptographic signatures

---

## 🛠️ Development Workflows

### Current Phase: Documentation & Planning

The repository is currently in **Phase 0: Foundation & Documentation**. No production code exists yet.

#### Active Tasks

- Documentation maintenance and updates
- Community infrastructure setup
- Legal framework refinement (pending attorney review)
- Architecture decision documentation

#### Future Phases

1. **Phase I (Months 1-6)** - Core Service + CRS (1Password integration)
2. **Phase II (Months 7-12)** - ACVS + Legal NLP + Multi-password manager support
3. **Phase III (Months 13-18)** - Enhanced automation + community growth
4. **Phase IV (Months 19-24)** - Enterprise features + ecosystem expansion

See **acm-governance-roadmap.md** for complete phase breakdown.

### When Code Development Begins

Future code contributions should follow:

```bash
# Expected tech stack (per acm-tad.md)
Backend:    Go 1.21+
CLI/TUI:    Go + Bubbletea
GUI:        Tauri (Rust + TypeScript + React)
API:        gRPC with Protocol Buffers
Database:   SQLite (audit logs)
Security:   mTLS, x509 certificates, OS keychain integration
```

### Git Workflow

- Main branch: `main` (protected)
- Feature branches: `feature/<description>` or `claude/<session-id>`
- All changes require clear commit messages
- Security-critical changes require security review tag

---

## 📝 Documentation Conventions

### Markdown Style

- Use GitHub-flavored markdown (GFM)
- Include table of contents for documents > 200 lines
- Use tables for structured data (requirements, features, risks)
- Code blocks should specify language for syntax highlighting
- Use emoji sparingly and only where it adds clarity (status indicators)

### Document Headers

All major documentation files should include:

```markdown
# Document Title
# Project Name (if applicable)

**Version:** X.X
**Date:** Month YYYY
**Status:** Draft | Review | Approved
**Document Type:** PRD | TAD | Risk Assessment | etc.

---

## Executive Summary or Introduction
```

### Update Protocol

When updating documentation:

1. **Check Dependencies** - Update cross-references in related docs
2. **Update Metadata** - Increment version, update date
3. **Update 00-INDEX.md** - Reflect any structural changes
4. **Maintain Consistency** - Match terminology across documents
5. **Security Review** - Flag security-impacting changes

### Terminology Standards

| Term | Definition | Usage |
|------|------------|-------|
| **CRS** | Credential Remediation Service | Core service for vault operations |
| **ACVS** | Automated Compliance Validation Service | Opt-in ToS compliance engine |
| **HIM** | Human-in-the-Middle | User-assisted automation checkpoint |
| **CRC** | Compliance Rule Set | Structured ToS rules from Legal NLP |
| **Zero-Knowledge** | System cannot access master password or vault keys | Architecture principle |
| **Local-First** | All processing on user's device | Architecture principle |
| **mTLS** | Mutual TLS authentication | Service-client communication |

---

## 🔐 Security Considerations for AI Assistants

### Critical Security Requirements

1. **Never Suggest Cloud Solutions** for sensitive operations
   - ACM must remain local-first
   - No SaaS offerings or cloud dependencies for core features

2. **Master Password is Sacred**
   - ACM never has access to master passwords
   - Password manager CLIs are invoked by user with their credentials
   - Suggest architectures that maintain this boundary

3. **Audit Everything**
   - All automated actions must be logged
   - Logs must include timestamps, affected accounts, ToS compliance status
   - Suggest SQLite schemas that support tamper-evident logging

4. **Fail Securely**
   - When in doubt, require Human-in-the-Middle intervention
   - ACVS should be disabled by default
   - Error states should not leak sensitive information

### Threat Model Reference

When suggesting security features, consult **acm-threat-model.md** which includes:
- STRIDE threat analysis
- Attack surface mapping
- Mitigations for 41 identified risks
- Trust boundaries and data flow diagrams

### Legal Compliance

- **ToS Violations** - Any automation must respect third-party Terms of Service
- **EULA Required** - Contributors and users must accept strong liability protections
- **Export Controls** - Cryptography usage may have export implications
- **Legal Review Pending** - acm-legal-framework.md requires attorney review

---

## 🚀 Quick Start for AI Assistants

### First-Time Repository Analysis

1. Read **00-INDEX.md** (5 min) - Get complete overview
2. Read **acm-prd.md** sections 1-3 (10 min) - Understand product vision
3. Skim **acm-tad.md** sections 1-2 (5 min) - Grasp architecture
4. Check **acm-threat-model.md** section 2 (5 min) - Security context

### Answering User Questions

**Product Questions** → Reference **acm-prd.md**
**Architecture Questions** → Reference **acm-tad.md**
**Security Questions** → Reference **acm-threat-model.md** + **acm-security-planning.md**
**Risk Questions** → Reference **acm-risk-assessment.md**
**Legal Questions** → Reference **acm-legal-framework.md** (note: requires attorney review)
**Timeline Questions** → Reference **acm-governance-roadmap.md**
**Community Questions** → Reference **acm-community-building.md**

### Making Documentation Changes

```bash
# 1. Read the current document completely
# 2. Identify sections that need updates
# 3. Check for cross-references in other docs
# 4. Make changes maintaining consistent style
# 5. Update version number and date
# 6. Update 00-INDEX.md if structure changed
# 7. Commit with descriptive message
```

### Running Setup Scripts

**GitHub Project Setup:**
```bash
chmod +x setup-github-project.sh
./setup-github-project.sh
# Creates project board with 15+ pre-configured issues
# Requires: gh CLI, authenticated GitHub account
```

**Community Infrastructure Setup:**
```bash
chmod +x community-building-setup.sh
./community-building-setup.sh
# Creates: CONTRIBUTING.md, CODE_OF_CONDUCT.md, issue templates
# Requires: gh CLI, repository write access
```

---

## ⚠️ Common Pitfalls to Avoid

### Documentation Pitfalls

1. **Inconsistent Terminology** - Use the terminology table above
2. **Broken Cross-References** - Verify links work before committing
3. **Outdated Metadata** - Always update version and date
4. **Missing Context** - Documents should be readable standalone
5. **Security Handwaving** - Be specific about security mechanisms

### Architecture Pitfalls

1. **Suggesting Cloud Solutions** - ACM is strictly local-first
2. **Breaking Zero-Knowledge** - Never compromise master password boundary
3. **Skipping HIM Workflows** - Automation can't bypass MFA/CAPTCHA
4. **Ignoring Threat Model** - All features must address relevant threats
5. **Adding Dependencies Lightly** - Each dependency increases attack surface

### Legal Pitfalls

1. **ToS Violations** - ACVS must validate automation legality
2. **Missing Indemnification** - Contributors and users need legal protection
3. **Overpromising Compliance** - Compliance is best-effort, not guaranteed
4. **Ignoring Attorney Review** - Legal framework requires professional review

---

## 📊 Project Status & Metrics

### Documentation Completeness

| Category | Status | File Count | Size |
|----------|--------|------------|------|
| Planning | ✅ Complete | 2 files | 58 KB |
| Architecture | ✅ Complete | 2 files | 115 KB |
| Security | ✅ Complete | 2 files | 94 KB |
| Legal | ⚠️ Pending Review | 1 file | 47 KB |
| Governance | ✅ Complete | 2 files | 74 KB |
| Scripts | ✅ Complete | 2 files | 51 KB |

### Implementation Status

- **Phase 0 (Foundation):** 🟢 Complete (Documentation done)
- **Phase I (Core Service):** 🔴 Not started (Pending legal review)
- **Phase II (ACVS):** 🔴 Not started
- **Phase III (Enhancement):** 🔴 Not started
- **Phase IV (Enterprise):** 🔴 Not started

### Critical Blockers

1. **Legal Review Required** - acm-legal-framework.md must be reviewed by attorney
2. **No Code Repository** - Implementation repository not yet created
3. **Community Infrastructure** - Discord, GitHub Discussions not yet set up
4. **Governance Formation** - Core maintainers not yet recruited

---

## 🔗 External Resources

### Referenced Technologies

- **Password Managers:** 1Password CLI, Bitwarden CLI, pass (Unix Password Manager)
- **GUI Framework:** Tauri (https://tauri.app)
- **TUI Library:** Bubbletea (https://github.com/charmbracelet/bubbletea)
- **RPC Framework:** gRPC (https://grpc.io)
- **Security:** mTLS, x509 certificates, OS keychain/TPM integration

### Related Standards

- **STRIDE Threat Modeling** - Microsoft security framework
- **Zero-Knowledge Architecture** - Password manager security model
- **OWASP Top 10** - Web security risks
- **CWE/CVE** - Common vulnerabilities and exposures

---

## 🤝 Contributing to This Document

This CLAUDE.md file should be updated when:

- New major documentation is added
- Project phase changes (e.g., code development begins)
- Architecture principles are modified
- New conventions are established
- Security requirements change

To update CLAUDE.md:

1. Read the entire file to understand current state
2. Make targeted updates maintaining consistent structure
3. Update "Last Updated" date at top
4. Commit with message: "docs: update CLAUDE.md - <brief description>"
5. Consider if 00-INDEX.md also needs updating

---

## 📞 Getting Help

### For AI Assistants

- **Uncertain about security?** → Consult acm-threat-model.md and acm-security-planning.md
- **Need architecture context?** → Read acm-tad.md sections 1-4
- **Unsure about legal implications?** → Reference acm-legal-framework.md and note that attorney review is required
- **Question about roadmap?** → Check acm-governance-roadmap.md

### For Human Contributors

- **Questions:** Open a GitHub Discussion (once community infrastructure is set up)
- **Bugs:** File an issue using issue templates
- **Proposals:** Submit RFC via governance process (see acm-governance-roadmap.md)
- **Legal concerns:** Contact project legal review committee

---

## 📜 Document History

| Date | Version | Changes |
|------|---------|---------|
| 2025-11-16 | 1.0 | Initial CLAUDE.md creation - comprehensive repository guide |

---

**End of CLAUDE.md**

*This document is maintained for AI assistants working with the ACM repository. Keep it current as the project evolves.*
