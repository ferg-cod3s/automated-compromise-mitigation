#!/usr/bin/env bash
# GitHub Project Setup Script for ACM
# This script creates a GitHub Project board with automated workflows
# for tracking the ACM development roadmap

set -euo pipefail

# Configuration
PROJECT_NAME="ACM Development Roadmap"
REPO_OWNER="acm-project"
REPO_NAME="acm"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  ACM GitHub Project Setup Script                        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI (gh) not found. Please install it:${NC}"
    echo "  macOS:   brew install gh"
    echo "  Linux:   https://github.com/cli/cli/blob/trunk/docs/install_linux.md"
    echo "  Windows: choco install gh"
    exit 1
fi

# Check authentication
echo -e "${BLUE}[1/7] Checking GitHub authentication...${NC}"
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}Not authenticated. Please run: gh auth login${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Authenticated${NC}"

# Create repository (if it doesn't exist)
echo -e "${BLUE}[2/7] Creating repository...${NC}"
if gh repo view "$REPO_OWNER/$REPO_NAME" &> /dev/null; then
    echo -e "${GREEN}✓ Repository already exists${NC}"
else
    gh repo create "$REPO_OWNER/$REPO_NAME" \
        --public \
        --description "Automated Compromise Mitigation - Local-first credential breach response" \
        --homepage "https://acm-project.dev"
    echo -e "${GREEN}✓ Repository created${NC}"
fi

# Create GitHub Project (v2)
echo -e "${BLUE}[3/7] Creating GitHub Project...${NC}"
PROJECT_ID=$(gh project create \
    --owner "$REPO_OWNER" \
    --title "$PROJECT_NAME" \
    --format json | jq -r '.id')

if [ -z "$PROJECT_ID" ]; then
    echo -e "${YELLOW}Failed to create project. It may already exist.${NC}"
    # Attempt to get existing project
    PROJECT_ID=$(gh project list --owner "$REPO_OWNER" --format json | \
        jq -r ".[] | select(.title == \"$PROJECT_NAME\") | .id" | head -1)
    if [ -z "$PROJECT_ID" ]; then
        echo -e "${YELLOW}Error: Could not create or find project${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}✓ Project created/found: $PROJECT_ID${NC}"

# Add custom fields to project
echo -e "${BLUE}[4/7] Configuring project fields...${NC}"

# Add Phase field (single select)
gh project field-create "$PROJECT_ID" \
    --name "Phase" \
    --data-type "SINGLE_SELECT" \
    --single-select-options "Phase I: MVP" \
    --single-select-options "Phase II: ACVS" \
    --single-select-options "Phase III: Advanced" \
    --single-select-options "Phase IV: Enterprise" || true

# Add Priority field (single select)
gh project field-create "$PROJECT_ID" \
    --name "Priority" \
    --data-type "SINGLE_SELECT" \
    --single-select-options "🔴 Critical" \
    --single-select-options "🟠 High" \
    --single-select-options "🟡 Medium" \
    --single-select-options "🟢 Low" || true

# Add Effort field (single select)
gh project field-create "$PROJECT_ID" \
    --name "Effort" \
    --data-type "SINGLE_SELECT" \
    --single-select-options "XS (< 1 day)" \
    --single-select-options "S (1-3 days)" \
    --single-select-options "M (3-5 days)" \
    --single-select-options "L (1-2 weeks)" \
    --single-select-options "XL (2+ weeks)" || true

# Add Component field (single select)
gh project field-create "$PROJECT_ID" \
    --name "Component" \
    --data-type "SINGLE_SELECT" \
    --single-select-options "Core Service" \
    --single-select-options "CRS" \
    --single-select-options "ACVS" \
    --single-select-options "HIM Manager" \
    --single-select-options "OpenTUI" \
    --single-select-options "Tauri GUI" \
    --single-select-options "Security" \
    --single-select-options "Legal" \
    --single-select-options "Documentation" \
    --single-select-options "Infrastructure" || true

echo -e "${GREEN}✓ Project fields configured${NC}"

# Create project views
echo -e "${BLUE}[5/7] Creating project views...${NC}"

# View 1: Roadmap by Phase
gh project view-create "$PROJECT_ID" \
    --name "Roadmap by Phase" \
    --layout "BOARD" \
    --group-by "Phase" || true

# View 2: Current Sprint (filtered to current month)
gh project view-create "$PROJECT_ID" \
    --name "Current Sprint" \
    --layout "TABLE" || true

# View 3: By Component
gh project view-create "$PROJECT_ID" \
    --name "By Component" \
    --layout "BOARD" \
    --group-by "Component" || true

echo -e "${GREEN}✓ Project views created${NC}"

# Create initial issues from roadmap
echo -e "${BLUE}[6/7] Creating initial issues...${NC}"

# Phase I Issues
echo "Creating Phase I issues..."

# Core Service Issues
gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] Core ACM Service - Go Runtime" \
    --body "Implement the core ACM service in Go with gRPC API

**Acceptance Criteria:**
- [ ] Go service starts and listens on localhost:8443
- [ ] gRPC server with mTLS authentication
- [ ] Health check endpoint
- [ ] Graceful shutdown handling
- [ ] Configuration file support

**Related Docs:** TAD Section 3.1" \
    --label "Phase I,Core Service,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] CRS - 1Password CLI Integration" \
    --body "Implement 1Password CLI integration for credential detection and rotation

**Acceptance Criteria:**
- [ ] Detect compromised credentials via \`op item list --compromised\`
- [ ] Generate secure passwords
- [ ] Update vault entries via \`op item edit\`
- [ ] Verify rotation success
- [ ] Unit tests with 80%+ coverage

**Related Docs:** TAD Section 3.2, Security Planning Component 2" \
    --label "Phase I,CRS,Priority: Critical" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] CRS - Bitwarden CLI Integration" \
    --body "Implement Bitwarden CLI integration for credential detection and rotation

**Acceptance Criteria:**
- [ ] Detect compromised credentials via \`bw list items --exposed\`
- [ ] Generate secure passwords  
- [ ] Update vault entries via \`bw edit item\`
- [ ] Verify rotation success
- [ ] Unit tests with 80%+ coverage

**Related Docs:** TAD Section 3.2, Security Planning Component 2" \
    --label "Phase I,CRS,Priority: Critical" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] mTLS Authentication System" \
    --body "Implement mTLS authentication with client certificates

**Acceptance Criteria:**
- [ ] Local CA generation (\`cfssl\` or \`openssl\`)
- [ ] Server certificate generation
- [ ] Client certificate generation and distribution
- [ ] TLS 1.3 enforcement
- [ ] Certificate pinning
- [ ] Certificate revocation support
- [ ] Integration tests

**Related Docs:** TAD Section 5.2, Security Planning Component 1" \
    --label "Phase I,Security,Priority: Critical" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] Secure Memory Handling" \
    --body "Implement memory protection for sensitive credentials

**Acceptance Criteria:**
- [ ] Memory locking (\`mlock\`) for credential buffers
- [ ] Explicit zeroing after use
- [ ] Integration with \`memguard\` library
- [ ] Memory dump test validates no credential leaks
- [ ] Documentation of memory safety patterns

**Related Docs:** Security Planning Component 3" \
    --label "Phase I,Security,Priority: Critical" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] Audit Logging System" \
    --body "Implement tamper-evident audit logging with SQLite

**Acceptance Criteria:**
- [ ] SQLite database with audit schema
- [ ] Cryptographic signatures (Ed25519)
- [ ] Merkle chain linking entries
- [ ] Credential ID hashing (SHA-256)
- [ ] Integrity verification command
- [ ] Unit tests for tampering detection

**Related Docs:** TAD Section 4.3, Security Planning Component 4" \
    --label "Phase I,Core Service,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] OpenTUI Client - Basic Functionality" \
    --body "Implement OpenTUI client with Bubbletea framework

**Acceptance Criteria:**
- [ ] Connection to ACM service via mTLS
- [ ] \`acm detect\` command (list compromised credentials)
- [ ] \`acm rotate <id>\` command
- [ ] \`acm status\` command
- [ ] \`acm audit\` command (view logs)
- [ ] Color-coded output with Lipgloss
- [ ] Keyboard navigation

**Related Docs:** TAD Section 3.5.1" \
    --label "Phase I,OpenTUI,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] EULA Implementation and Acceptance Flow" \
    --body "Implement EULA display and acceptance tracking

**Acceptance Criteria:**
- [ ] EULA displayed on first run
- [ ] User must scroll to end before accepting
- [ ] Require typing 'yes' or 'I accept'
- [ ] Log acceptance with timestamp and signature
- [ ] Store acceptance in \`~/.acm/legal/eula-acceptance.json\`
- [ ] Version EULA for future updates

**Related Docs:** Legal Framework Section 2.3" \
    --label "Phase I,Legal,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] Documentation - Installation Guide" \
    --body "Create comprehensive installation and setup guide

**Acceptance Criteria:**
- [ ] Installation instructions (macOS, Linux, Windows)
- [ ] \`acm setup\` wizard documentation
- [ ] Certificate generation steps
- [ ] Password manager CLI setup
- [ ] Troubleshooting section
- [ ] Screenshots/GIFs of setup process

**Related Docs:** PRD Section 10, Governance Section 5.4" \
    --label "Phase I,Documentation,Priority: Medium" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase I] CI/CD Pipeline Setup" \
    --body "Set up GitHub Actions for automated testing and builds

**Acceptance Criteria:**
- [ ] Run tests on every PR
- [ ] Run \`golangci-lint\` on every PR
- [ ] Run \`gosec\` security scan
- [ ] Dependency vulnerability scanning (Dependabot)
- [ ] Build artifacts for main branch merges
- [ ] Automated release process (GoReleaser)

**Related Docs:** TAD Section 6.3, Governance Section 5.3" \
    --label "Phase I,Infrastructure,Priority: High" || true

echo -e "${GREEN}✓ Phase I issues created${NC}"

# Phase II Issues (sample)
echo "Creating Phase II issues..."

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase II] ACVS - Legal NLP Engine" \
    --body "Implement Legal NLP engine for ToS parsing

**Acceptance Criteria:**
- [ ] spaCy or Transformers model integration
- [ ] ToS document fetching and caching
- [ ] Rule extraction (automation prohibitions, rate limits)
- [ ] CRC generation
- [ ] Accuracy > 85% F1 score on test set
- [ ] Unit tests with sample ToS documents

**Related Docs:** TAD Section 3.3, Security Planning Component 5" \
    --label "Phase II,ACVS,Priority: Critical" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase II] Evidence Chain System" \
    --body "Implement cryptographic evidence chain for compliance proof

**Acceptance Criteria:**
- [ ] HMAC signatures on evidence entries
- [ ] Timestamping (RFC 3161 compatible)
- [ ] Merkle tree linking
- [ ] Export to PDF and JSON
- [ ] Verification tool
- [ ] Integration tests

**Related Docs:** Legal Framework Section 3.3" \
    --label "Phase II,ACVS,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase II] Tauri Desktop GUI - Basic Functionality" \
    --body "Implement Tauri-based desktop GUI

**Acceptance Criteria:**
- [ ] React + TypeScript frontend
- [ ] Rust backend for gRPC communication
- [ ] Dashboard view (status, compromised list)
- [ ] Rotation workflow wizard
- [ ] HIM prompts for MFA/CAPTCHA
- [ ] Compliance dashboard (ACVS status)
- [ ] Cross-platform builds (macOS, Windows, Linux)

**Related Docs:** TAD Section 3.5.2" \
    --label "Phase II,Tauri GUI,Priority: High" || true

gh issue create --repo "$REPO_OWNER/$REPO_NAME" \
    --title "[Phase II] Third-Party Security Audit" \
    --body "Engage professional security firm for comprehensive audit

**Acceptance Criteria:**
- [ ] RFP to 3+ security firms
- [ ] Select firm and sign engagement
- [ ] Provide audit scope (Core Service, CRS, ACVS, mTLS)
- [ ] Receive audit report
- [ ] Create action plan for findings
- [ ] Remediate critical/high findings
- [ ] Publish redacted audit report

**Related Docs:** Security Planning Section 4.1, Risk Assessment RISK-SEC-002" \
    --label "Phase II,Security,Priority: Critical" || true

echo -e "${GREEN}✓ Phase II issues created${NC}"

# Apply project to repository
echo -e "${BLUE}[7/7] Linking project to repository...${NC}"
gh project link "$PROJECT_ID" --repo "$REPO_OWNER/$REPO_NAME" || true
echo -e "${GREEN}✓ Project linked to repository${NC}"

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  GitHub Project Setup Complete!                         ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "📊 Project URL: https://github.com/orgs/$REPO_OWNER/projects"
echo -e "📁 Repository: https://github.com/$REPO_OWNER/$REPO_NAME"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Review and prioritize issues in the project board"
echo "2. Assign owners to Phase I critical issues"
echo "3. Set up project automation (auto-add issues, auto-archive)"
echo "4. Create project milestones for Phase I/II/III"
echo "5. Begin development! 🚀"
echo ""
