# Changelog

All notable changes to the ACM (Automated Compromise Mitigation) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive build infrastructure with Makefile
- Cross-platform build scripts for Linux, macOS, and Windows
- GoReleaser configuration for multi-platform releases
- golangci-lint configuration with security-focused linters
- Development environment setup script
- mTLS certificate generation script
- Test runner with coverage reporting (targeting >80%)
- EditorConfig for consistent code style
- Development tool version pinning via tools.go

### Changed
- Updated Go toolchain from 1.21 to 1.24.2
- Updated gRPC from v1.60.1 to v1.76.0 for improved performance and security
- Updated Protocol Buffers from v1.32.0 to v1.36.10
- Updated Bubbletea TUI library from v0.25.0 to v1.3.10 with enhanced features
- Updated Lipgloss styling from v0.9.1 to v1.1.0
- Updated JWT library from v5.2.0 to v5.3.0
- Updated all transitive dependencies to latest secure versions
- Regenerated protocol buffer Go code with latest protoc versions

### Security
- mTLS mutual authentication for service-client communication
- Security scanning integration (gosec)
- Vulnerability checking (govulncheck)
- Certificate-based authentication infrastructure
- Zero vulnerabilities found in actively used code paths (govulncheck verified)

## [0.1.0] - TBD

### Added
- Initial project setup and documentation
- Product Requirements Document (PRD)
- Technical Architecture Document (TAD)
- Threat Model (STRIDE analysis)
- Security Planning and Risk Assessment
- Legal Framework
- Governance and Roadmap
- Community Building Strategy

### Documentation
- Comprehensive project documentation (>400KB)
- CLAUDE.md guide for AI assistants
- 00-INDEX.md master index
- Architecture Decision Records framework

---

## Release Types

- **Major Release (x.0.0)**: Breaking changes, major new features
- **Minor Release (0.x.0)**: New features, backwards compatible
- **Patch Release (0.0.x)**: Bug fixes, security patches

## Security Releases

Security vulnerabilities are disclosed according to our [Security Policy](SECURITY.md).

---

**Note**: Versions prior to 1.0.0 are considered pre-release and may include breaking changes in minor versions.
