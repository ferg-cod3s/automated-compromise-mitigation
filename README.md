# Automated Compromise Mitigation (ACM)

**Local-first credential breach response system** - Zero-knowledge, ToS-compliant automation for password managers.

[![Build Status](https://github.com/ferg-cod3s/automated-compromise-mitigation/actions/workflows/ci.yml/badge.svg)](https://github.com/ferg-cod3s/automated-compromise-mitigation/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ferg-cod3s/automated-compromise-mitigation)](https://goreportcard.com/report/github.com/ferg-cod3s/automated-compromise-mitigation)

## 🚀 Quick Start

See [Getting Started Guide](docs/GETTING_STARTED.md) for installation and setup.

## 📊 Project Status

- **Phase I (MVP)**: ✅ **COMPLETE** - Core gRPC service, CRS, audit logging, HIM, Bitwarden/1Password integration
- **Phase II (ACVS)**: ✅ **COMPLETE** - Automated Compliance Validation Service, evidence chains, ToS analysis
- **Phase III (Advanced)**: 🔄 **IN PROGRESS** - Production NLP, AWS IAM rotation, OpenTUI client, enhanced HIM

**Development Roadmap:** [ACM Development Roadmap](https://github.com/users/ferg-cod3s/projects/9)

## 📚 Documentation

- [00-INDEX.md](00-INDEX.md) - Complete documentation index
- [PHASE1_IMPLEMENTATION_SUMMARY.md](PHASE1_IMPLEMENTATION_SUMMARY.md) - Phase I details
- [PHASE2_IMPLEMENTATION_SUMMARY.md](PHASE2_IMPLEMENTATION_SUMMARY.md) - Phase II details
- [PHASE3_PLANNING.md](PHASE3_PLANNING.md) - Phase III roadmap
- [acm-tad.md](acm-tad.md) - Technical Architecture Document
- [acm-prd.md](acm-prd.md) - Product Requirements Document

## 🏗️ Architecture

- **Backend:** Go 1.21+ with gRPC/Protocol Buffers
- **Security:** mTLS authentication, zero-knowledge design
- **Storage:** SQLite (ACVS), in-memory (audit logs, Phase III upgrade)
- **Clients:** CLI, OpenTUI (Bubbletea), future Tauri GUI
- **Integrations:** Bitwarden, 1Password, GitHub PAT, AWS IAM (planned)

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

Licensed under the terms in [LICENSE](LICENSE).