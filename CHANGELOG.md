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

### Security
- mTLS mutual authentication for service-client communication
- Security scanning integration (gosec)
- Vulnerability checking (govulncheck)
- Certificate-based authentication infrastructure

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
