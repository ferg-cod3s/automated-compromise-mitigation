# ACM Phase I - Atomic Task Breakdown
# Task Tracking & Progress

**Last Updated:** 2025-11-16
**Status:** Week 1-2 (Foundation)

---

## How to Use This Document

- Each task has a unique ID (e.g., `W1-T01`)
- Mark tasks as: `[ ]` Not started, `[>]` In progress, `[✓]` Complete
- Track blockers and dependencies
- Update weekly

---

## Week 1-2: Foundation & Development Environment

### Milestone M1: Project Infrastructure

**Goal:** Complete project scaffolding, tooling, and CI/CD pipeline

---

#### Project Initialization

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T01 | Initialize Go module | - | [>] | None | `go mod init` |
| W1-T02 | Create `.gitignore` | - | [ ] | None | Go + certs + config |
| W1-T03 | Create project directory structure | - | [ ] | W1-T01 | 15+ directories |
| W1-T04 | Create `tools.go` | - | [ ] | W1-T01 | Dev dependencies |
| W1-T05 | Add core dependencies to `go.mod` | - | [ ] | W1-T01 | gRPC, memguard, viper |
| W1-T06 | Run `go mod tidy` | - | [ ] | W1-T05 | Generate go.sum |

**Deliverable:** Working Go module with all dependencies

---

#### Build Automation

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T07 | Create `Makefile` skeleton | - | [ ] | W1-T01 | All targets |
| W1-T08 | Implement `make build` | - | [ ] | W1-T07, W1-T03 | Build all binaries |
| W1-T09 | Implement `make test` | - | [ ] | W1-T07 | Run all tests |
| W1-T10 | Implement `make lint` | - | [ ] | W1-T07 | golangci-lint |
| W1-T11 | Implement `make security-scan` | - | [ ] | W1-T07 | gosec + semgrep |
| W1-T12 | Implement `make generate` | - | [ ] | W1-T07 | Protobuf generation |
| W1-T13 | Implement `make clean` | - | [ ] | W1-T07 | Clean artifacts |
| W1-T14 | Create `scripts/setup-dev.sh` | - | [ ] | None | Dev env setup |
| W1-T15 | Create `scripts/run-tests.sh` | - | [ ] | None | Test runner |

**Deliverable:** Complete build automation

---

#### Protocol Buffer Definitions

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T16 | Create `api/proto/acm/v1/` directory | - | [ ] | W1-T03 | Proto directory |
| W1-T17 | Create `common.proto` | - | [ ] | W1-T16 | Common types |
| W1-T18 | Create `service.proto` | - | [ ] | W1-T17 | Main service |
| W1-T19 | Create `crs.proto` | - | [ ] | W1-T17 | CRS service |
| W1-T20 | Create `audit.proto` | - | [ ] | W1-T17 | Audit service |
| W1-T21 | Create `him.proto` | - | [ ] | W1-T17 | HIM service |
| W1-T22 | Add protobuf generation to Makefile | - | [ ] | W1-T12 | `make generate` |
| W1-T23 | Generate Go code from protos | - | [ ] | W1-T22 | Test generation |

**Deliverable:** Complete gRPC API definitions

---

#### Development Tooling

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T24 | Create `.golangci.yml` | - | [ ] | None | Linter config |
| W1-T25 | Configure security linters (gosec) | - | [ ] | W1-T24 | Add to config |
| W1-T26 | Configure code quality linters | - | [ ] | W1-T24 | gofmt, goimports |
| W1-T27 | Test golangci-lint on sample code | - | [ ] | W1-T26 | Verify works |
| W1-T28 | Configure gosec scanner | - | [ ] | None | Security scanning |
| W1-T29 | Test gosec on sample code | - | [ ] | W1-T28 | Verify works |

**Deliverable:** Working linter and security scanner

---

#### CI/CD Pipeline

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T30 | Create `.github/workflows/` directory | - | [ ] | None | Workflows dir |
| W1-T31 | Create `ci.yml` (CI workflow) | - | [ ] | W1-T30 | Build + test |
| W1-T32 | Add test coverage reporting | - | [ ] | W1-T31 | Codecov |
| W1-T33 | Create `security.yml` (Security workflow) | - | [ ] | W1-T30 | gosec + semgrep |
| W1-T34 | Create `release.yml` (Release workflow) | - | [ ] | W1-T30 | GoReleaser |
| W1-T35 | Configure Dependabot | - | [ ] | None | `.github/dependabot.yml` |
| W1-T36 | Test CI workflow with sample code | - | [ ] | W1-T31-W1-T34 | Push and verify |

**Deliverable:** Working CI/CD pipeline

---

#### Certificate Infrastructure

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T37 | Create `scripts/generate-certs.sh` | - | [ ] | None | cfssl wrapper |
| W1-T38 | Create `configs/ca-config.json` | - | [ ] | None | CA config |
| W1-T39 | Create `configs/tls/` directory | - | [ ] | W1-T03 | TLS configs |
| W1-T40 | Test certificate generation script | - | [ ] | W1-T37-W1-T38 | Generate certs |
| W1-T41 | Document certificate generation | - | [ ] | W1-T40 | Add to docs |
| W1-T42 | Create certificate renewal workflow | - | [ ] | W1-T37 | Renewal script |

**Deliverable:** Certificate generation infrastructure

---

#### Database Schema

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T43 | Create `scripts/setup-sqlite.sh` | - | [ ] | None | DB init script |
| W1-T44 | Create `internal/storage/audit/schema.sql` | - | [ ] | W1-T03 | Audit log schema |
| W1-T45 | Add indexes to schema | - | [ ] | W1-T44 | Performance |
| W1-T46 | Create migration framework | - | [ ] | W1-T44 | Schema migrations |
| W1-T47 | Test database initialization | - | [ ] | W1-T43-W1-T44 | Run script |
| W1-T48 | Document database schema | - | [ ] | W1-T47 | Add to docs |

**Deliverable:** SQLite database schema

---

#### Configuration Management

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T49 | Create `configs/service.yaml.example` | - | [ ] | W1-T03 | Service config |
| W1-T50 | Create `configs/client.yaml.example` | - | [ ] | W1-T03 | Client config |
| W1-T51 | Document all config options | - | [ ] | W1-T49-W1-T50 | Inline comments |
| W1-T52 | Create config validation logic | - | [ ] | W1-T49-W1-T50 | Validate on load |
| W1-T53 | Test config validation | - | [ ] | W1-T52 | Unit tests |

**Deliverable:** Configuration templates and validation

---

#### Documentation

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T54 | Create `docs/development/SETUP.md` | - | [ ] | None | Dev setup guide |
| W1-T55 | Create `docs/development/ARCHITECTURE.md` | - | [ ] | None | Architecture docs |
| W1-T56 | Create `SECURITY.md` | - | [ ] | None | Security policy |
| W1-T57 | Update `README.md` | - | [ ] | None | Project overview |
| W1-T58 | Create `CONTRIBUTING.md` | - | [ ] | None | Contribution guide |
| W1-T59 | Create `docs/` directory structure | - | [ ] | W1-T03 | Docs folders |

**Deliverable:** Foundation documentation

---

#### Additional Setup Tasks

| ID | Task | Owner | Status | Dependencies | Notes |
|----|------|-------|--------|--------------|-------|
| W1-T60 | Create `.editorconfig` | - | [ ] | None | Editor consistency |
| W1-T61 | Create `.vscode/settings.json` | - | [ ] | None | VSCode config |
| W1-T62 | Create `Dockerfile` for testing | - | [ ] | None | Container setup |
| W1-T63 | Create `docker-compose.yml` | - | [ ] | W1-T62 | Multi-container |
| W1-T64 | Test Docker build | - | [ ] | W1-T62-W1-T63 | Verify works |

**Deliverable:** Complete development environment

---

## Week 1-2 Checklist

**At the end of Week 2, you should have:**

- [ ] Complete Go module with all dependencies
- [ ] Full project directory structure (15+ directories)
- [ ] Working `Makefile` with all targets
- [ ] Protocol Buffer definitions for all services
- [ ] Generated Go code from protobuf
- [ ] CI/CD pipeline running on GitHub
- [ ] golangci-lint and gosec configured
- [ ] Certificate generation scripts working
- [ ] SQLite database schema created
- [ ] Configuration templates documented
- [ ] Foundation documentation complete
- [ ] Docker development environment working

**Success Criteria:**

```bash
# These commands should work:
make build        # ✓ Builds all binaries
make test         # ✓ Runs tests (even if none yet)
make lint         # ✓ Runs linters with no errors
make generate     # ✓ Generates protobuf code
./scripts/generate-certs.sh  # ✓ Creates certificates
./scripts/setup-sqlite.sh    # ✓ Creates database
```

---

## Week 3-4: mTLS Authentication (Preview)

Coming next: 25+ tasks for implementing secure mTLS authentication

---

## Progress Tracking

### Week 1 Progress (Days 1-5)

**Day 1:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 2:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 3:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 4:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 5:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

### Week 2 Progress (Days 6-10)

**Day 6:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 7:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 8:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 9:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

**Day 10:**
- [ ] Tasks completed: ___
- [ ] Blockers: ___

---

## Notes & Decisions

### Technical Decisions

- **Go Version:** 1.21+ (for latest security features)
- **gRPC Framework:** google.golang.org/grpc
- **TUI Framework:** Bubbletea (lightweight, performant)
- **Database:** SQLite (local-first, no server required)
- **Certificate Tool:** cfssl (production-grade)

### Deferred to Phase II

- ACVS (Automated Compliance Validation Service)
- Legal NLP engine
- Tauri GUI client
- Enterprise features

---

## Risk Log

| Date | Risk | Impact | Mitigation | Status |
|------|------|--------|------------|--------|
| 2025-11-16 | Legal review needed | High | Engage attorney ASAP | Open |
| - | - | - | - | - |

---

## Team Assignments

| Developer | Tasks Assigned | Status |
|-----------|----------------|--------|
| TBD | W1-T01 to W1-T20 | - |
| TBD | W1-T21 to W1-T40 | - |
| TBD | W1-T41 to W1-T64 | - |

---

**Document Status:** Active Task Tracker
**Update Frequency:** Daily during active development
**Next Review:** End of Week 2
