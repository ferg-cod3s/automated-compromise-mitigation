# CI/CD Setup Guide
# ACM Phase I Implementation

**Version:** 1.0
**Date:** November 2025
**Status:** Active
**Document Type:** Technical Guide

---

## Table of Contents

- [Overview](#overview)
- [Workflow Files](#workflow-files)
- [Security Features](#security-features)
- [Workflow Triggers](#workflow-triggers)
- [Performance & Timing](#performance--timing)
- [Required Configuration](#required-configuration)
- [Branch Protection](#branch-protection)
- [Secrets Management](#secrets-management)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The ACM project uses GitHub Actions for comprehensive CI/CD automation with a strong focus on security. Our pipeline includes:

- **Continuous Integration (CI)** - Automated testing, linting, and security scanning
- **Protocol Buffer Validation** - Ensuring API definitions are valid and up-to-date
- **Dependency Management** - Automated dependency updates and security reviews
- **Release Automation** - Secure release creation with SBOM generation
- **Security First** - Multiple security scans on every PR and release

### Design Principles

1. **Security by Default** - Every PR undergoes security scanning
2. **Fast Feedback** - Parallel execution and caching for speed
3. **Comprehensive Testing** - Multi-platform and multi-version testing
4. **Reproducible Builds** - Deterministic build process
5. **Supply Chain Security** - SBOM generation and artifact signing

---

## Workflow Files

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Purpose:** Main continuous integration pipeline for code quality and testing.

**Jobs:**

| Job | Purpose | Duration |
|-----|---------|----------|
| `lint` | Run golangci-lint for code quality | ~2 min |
| `test` | Run unit tests (matrix: Go 1.21/1.22, Linux/Mac/Windows) | ~3-5 min |
| `security-scan` | Run gosec and Trivy security scanners | ~2 min |
| `build` | Build binaries for all platforms | ~2 min |
| `integration-test` | Run integration tests | ~5 min |
| `verify-modules` | Ensure go.mod/go.sum are up-to-date | ~1 min |
| `license-check` | Verify license compliance | ~1 min |

**Key Features:**
- Matrix testing across Go versions (1.21, 1.22) and OS (Linux, macOS, Windows)
- Code coverage uploaded to Codecov
- Security results uploaded to GitHub Security tab
- Build artifacts retained for 7 days
- Comprehensive dependency caching

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

### 2. Proto Validation Workflow (`.github/workflows/proto.yml`)

**Purpose:** Validate Protocol Buffer definitions and ensure generated code is current.

**Jobs:**

| Job | Purpose | Duration |
|-----|---------|----------|
| `validate-proto` | Compile proto files to verify syntax | ~2 min |
| `lint-proto` | Lint proto files with buf | ~1 min |
| `proto-documentation` | Generate API documentation from protos | ~2 min |

**Key Features:**
- Validates all `.proto` files compile correctly
- Ensures generated Go code is up-to-date (fails if stale)
- Breaking change detection against `main` branch
- Auto-generated API documentation on PRs
- Uses `buf` for best-practice proto linting

**Triggers:**
- Changes to `api/proto/**/*.proto` or `api/proto/**/*.go`
- Push to `main` or pull requests

### 3. Dependency Review Workflow (`.github/workflows/dependency-review.yml`)

**Purpose:** Review and validate all dependencies for security and licensing.

**Jobs:**

| Job | Purpose | Duration |
|-----|---------|----------|
| `dependency-review` | Check for vulnerable dependencies | ~2 min |
| `govulncheck` | Go-specific vulnerability scanning | ~2 min |
| `dependency-analysis` | Analyze and report on dependencies | ~2 min |
| `license-compliance` | Verify acceptable licenses | ~2 min |

**Key Features:**
- Blocks PRs with critical vulnerabilities
- Denies GPL-3.0 and AGPL-3.0 licenses
- Reports available dependency updates
- Generates license compliance report
- Always posts summary to PR

**Triggers:**
- Pull requests to `main` branch

**Blocked Licenses:**
- GPL-3.0
- AGPL-3.0
- LGPL-3.0

### 4. Release Workflow (`.github/workflows/release.yml`)

**Purpose:** Automate secure release creation with comprehensive artifacts.

**Jobs:**

| Job | Purpose | Duration |
|-----|---------|----------|
| `security-check` | Pre-release security validation | ~3 min |
| `release` | Build and publish release with GoReleaser | ~10 min |
| `publish-checksums` | Generate additional checksums | ~2 min |
| `docker-build` | Build and push Docker images | ~8 min |
| `release-notes` | Generate comprehensive changelog | ~2 min |
| `security-audit` | Post-release security audit | ~3 min |

**Key Features:**
- Multi-platform binary builds (Linux, macOS, Windows, ARM64)
- SBOM (Software Bill of Materials) generation
- SHA-256 and SHA-512 checksums
- Docker multi-arch images (amd64, arm64)
- Automated changelog generation
- Security scanning of release artifacts
- GitHub Container Registry publishing

**Triggers:**
- Tag push matching `v*.*.*` (e.g., `v1.0.0`)

**Artifacts Generated:**
- Binaries for all platforms
- Checksums (SHA-256, SHA-512)
- SBOM in SPDX and CycloneDX formats
- Docker images
- Release notes with changelog

### 5. Dependabot Configuration (`.github/dependabot.yml`)

**Purpose:** Automated dependency updates.

**Configuration:**

| Ecosystem | Schedule | Open PRs Limit |
|-----------|----------|----------------|
| Go modules | Weekly (Monday 9 AM) | 10 |
| GitHub Actions | Weekly (Monday 9 AM) | 5 |
| Docker | Weekly (Monday 9 AM) | 5 |

**Features:**
- Groups minor and patch updates together
- Ignores major version updates for critical dependencies (gRPC, protobuf)
- Auto-labels PRs with `dependencies`, ecosystem, and `automated`
- Conventional commit messages (`deps:`, `ci:`)

---

## Security Features

### 1. Security Scanning Tools

| Tool | Purpose | Frequency |
|------|---------|-----------|
| **gosec** | Go security linter | Every PR + Release |
| **Trivy** | Vulnerability scanner | Every PR + Release |
| **govulncheck** | Go vulnerability database | Every PR + Release |
| **Dependency Review** | GitHub dependency analysis | Every PR |
| **CodeQL** | Semantic code analysis | Planned |

### 2. Security Checks by Phase

**Pull Request:**
- gosec security scanning
- Trivy vulnerability scanning
- govulncheck database check
- Dependency vulnerability review
- License compliance check

**Release:**
- Pre-release security validation
- SBOM generation (supply chain transparency)
- Post-release security audit
- Checksum generation and verification

### 3. Security Results

All security scan results are uploaded to **GitHub Security** tab for centralized review and tracking.

---

## Workflow Triggers

### CI Workflow

```yaml
Triggers:
  - Push to main
  - Pull request to main

Events:
  - Lint, test, security scan, build on every change
```

### Proto Validation

```yaml
Triggers:
  - Changes to api/proto/**/*.proto
  - Changes to api/proto/**/*.go
  - Push to main or PRs

Events:
  - Validate proto syntax
  - Check generated code freshness
  - Breaking change detection (PRs only)
```

### Dependency Review

```yaml
Triggers:
  - Pull requests to main

Events:
  - Dependency vulnerability check
  - License compliance verification
  - Update availability report
```

### Release Workflow

```yaml
Triggers:
  - Tag push matching v*.*.*

Events:
  - Security pre-check
  - Multi-platform builds
  - SBOM generation
  - Docker image creation
  - Changelog generation
```

### Dependabot

```yaml
Schedule:
  - Weekly on Monday at 09:00 UTC

Actions:
  - Check for dependency updates
  - Create PRs for updates
  - Group minor/patch updates
```

---

## Performance & Timing

### Estimated CI Run Times

| Workflow | Scenario | Estimated Time |
|----------|----------|----------------|
| **CI** | Single PR (cached) | 8-10 minutes |
| **CI** | Single PR (cold cache) | 12-15 minutes |
| **CI** | Push to main | 10-12 minutes |
| **Proto** | Proto changes | 4-5 minutes |
| **Dependency Review** | PR with dependencies | 6-8 minutes |
| **Release** | Tag push | 25-30 minutes |

### Optimization Features

1. **Caching:**
   - Go module cache
   - Go build cache
   - Docker layer cache

2. **Parallelization:**
   - Matrix builds run in parallel
   - Independent jobs run concurrently
   - Multi-platform builds parallelized

3. **Conditional Execution:**
   - Proto validation only on proto file changes
   - Integration tests only when needed
   - Docker builds only on releases

### Typical PR Timeline

```
0:00 - PR opened
0:01 - Lint starts (1-2 min)
0:01 - Tests start in parallel (3-5 min)
0:01 - Security scans start (2-3 min)
0:03 - Build starts (2-3 min)
0:06 - Integration tests start (5 min)
0:10 - All checks complete ✓
```

---

## Required Configuration

### 1. GitHub Secrets

No secrets are required for basic CI/CD. Optional secrets for enhanced features:

| Secret | Purpose | Required |
|--------|---------|----------|
| `CODECOV_TOKEN` | Upload coverage to Codecov | Optional |
| `GPG_PRIVATE_KEY` | Sign release artifacts | Future |
| `GPG_PASSPHRASE` | GPG key passphrase | Future |

### 2. GitHub Settings

**Repository Settings → Actions:**
- ✅ Allow all actions and reusable workflows
- ✅ Allow GitHub Actions to create and approve pull requests (for Dependabot)

**Repository Settings → Code security and analysis:**
- ✅ Dependency graph
- ✅ Dependabot alerts
- ✅ Dependabot security updates
- ✅ Secret scanning
- ✅ Push protection

### 3. Team Configuration

**Required Team:**
- Create team: `acm-core-team` for Dependabot PR reviews

### 4. GitHub Container Registry

For Docker image publishing:
- Container registry is automatically available
- Images published to `ghcr.io/<owner>/automated-compromise-mitigation`

---

## Branch Protection

### Recommended Branch Protection Rules

Apply these rules to the `main` branch:

**Repository Settings → Branches → Add rule:**

```yaml
Branch name pattern: main

Protect matching branches:
  ✅ Require a pull request before merging
    - Required approvals: 2
    - Dismiss stale reviews: true
    - Require review from Code Owners: true
    - Require approval of most recent push: true

  ✅ Require status checks to pass before merging
    - Require branches to be up to date: true
    - Status checks that are required:
      - lint
      - test (Go 1.22 / ubuntu-latest)
      - security-scan
      - build (ubuntu-latest)
      - dependency-review
      - govulncheck
      - license-compliance

  ✅ Require conversation resolution before merging

  ✅ Require signed commits

  ✅ Require linear history

  ✅ Include administrators

  ✅ Restrict who can push to matching branches
    - Teams: acm-core-team
    - Users: [Trusted maintainers]

  ✅ Allow force pushes
    - Specify who can force push: [None]

  ✅ Allow deletions: false
```

### Status Check Requirements

At minimum, require these checks to pass:

1. `lint` - Code quality
2. `test` (at least one matrix combination)
3. `security-scan` - Security validation
4. `build` (at least one platform)
5. `dependency-review` - Dependency security
6. `govulncheck` - Go vulnerabilities

---

## Secrets Management

### Current Secrets

No secrets are currently required for the base CI/CD pipeline.

### Future Secrets

| Secret | When Needed | Purpose |
|--------|-------------|---------|
| `CODECOV_TOKEN` | When coverage reporting is set up | Upload coverage reports |
| `GPG_PRIVATE_KEY` | When artifact signing is enabled | Sign release binaries |
| `DISCORD_WEBHOOK` | When Discord notifications are enabled | Release announcements |
| `AUR_KEY` | When Arch Linux packages are published | AUR package updates |

### How to Add Secrets

1. Go to **Repository Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Add name and value
4. Update workflow files to use: `${{ secrets.SECRET_NAME }}`

---

## Best Practices

### 1. Commit Message Conventions

Follow conventional commits for automated changelog generation:

```
feat: add new feature
fix: resolve bug
sec: security improvement
perf: performance optimization
docs: documentation update
test: add tests
deps: dependency update
ci: CI/CD changes
chore: maintenance tasks
```

### 2. Proto File Changes

When modifying proto files:

```bash
# 1. Edit .proto files
vim api/proto/service/v1/service.proto

# 2. Regenerate Go code
make proto-gen

# 3. Commit both .proto and generated .go files
git add api/proto/
git commit -m "feat(proto): add new RPC method"
```

### 3. Dependency Updates

Review Dependabot PRs promptly:

1. Check the changelog for breaking changes
2. Review security implications
3. Run tests locally if uncertain
4. Merge promptly to stay current

### 4. Release Process

```bash
# 1. Ensure main branch is clean and all PRs merged
git checkout main
git pull

# 2. Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GitHub Actions will:
#    - Run security checks
#    - Build binaries for all platforms
#    - Generate SBOM
#    - Create GitHub release
#    - Publish Docker images
#    - Generate changelog

# 4. Verify release
#    - Check GitHub Releases page
#    - Verify checksums
#    - Test binaries
```

### 5. Security Scan Failures

If security scans fail:

1. **Review findings** in GitHub Security tab
2. **Assess severity** - Critical/High must be fixed immediately
3. **Update dependencies** if vulnerability is in a dependency
4. **Fix code** if vulnerability is in ACM code
5. **Re-run checks** after fixes

---

## Troubleshooting

### Common Issues

#### 1. Go Module Cache Misses

**Symptom:** Slow build times, downloading modules every run

**Solution:**
```yaml
# Verify cache configuration in workflow
- uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache: true  # ← Should be enabled
```

#### 2. Proto Validation Fails

**Symptom:** "Generated proto code is out of date"

**Solution:**
```bash
# Regenerate proto files locally
make proto-gen

# Commit the changes
git add api/proto/
git commit -m "fix(proto): regenerate proto files"
```

#### 3. Dependency Review Blocks PR

**Symptom:** "Critical vulnerability detected"

**Solution:**
```bash
# Update the vulnerable dependency
go get -u <vulnerable-package>
go mod tidy

# Commit and push
git add go.mod go.sum
git commit -m "deps: update vulnerable dependency"
```

#### 4. License Check Fails

**Symptom:** "Found forbidden licenses!"

**Solution:**
- Review the dependency with the forbidden license
- Find an alternative dependency with acceptable license
- If no alternative exists, request an exemption from maintainers

#### 5. Build Fails on Specific Platform

**Symptom:** "Build fails on windows-latest but passes on linux"

**Solution:**
- Check for platform-specific code issues
- Test locally on the failing platform if possible
- Review build logs for specific errors
- Consider using build tags for platform-specific code

---

## Monitoring & Maintenance

### Weekly Tasks

- [ ] Review Dependabot PRs
- [ ] Check GitHub Security tab for new alerts
- [ ] Review workflow run failures

### Monthly Tasks

- [ ] Review CI/CD performance metrics
- [ ] Update pinned action versions if security updates available
- [ ] Review and update branch protection rules
- [ ] Audit security scan configurations

### Quarterly Tasks

- [ ] Review and update this documentation
- [ ] Evaluate new security scanning tools
- [ ] Review workflow efficiency and optimize
- [ ] Update Go versions in matrix testing

---

## Action Version Pinning

All GitHub Actions are pinned to specific SHA commits for security. To update:

```bash
# Find latest version
gh api repos/actions/checkout/releases/latest

# Get commit SHA for specific tag
gh api repos/actions/checkout/git/ref/tags/v4.1.1

# Update workflow file with SHA
uses: actions/checkout@<sha>  # v4.1.1
```

---

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [Dependabot Configuration](https://docs.github.com/en/code-security/dependabot)
- [Security Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [ACM Security Planning](../../acm-security-planning.md)
- [ACM Threat Model](../../acm-threat-model.md)

---

## Document History

| Date | Version | Changes |
|------|---------|---------|
| 2025-11-16 | 1.0 | Initial CI/CD setup documentation |

---

**End of CI/CD Setup Guide**
