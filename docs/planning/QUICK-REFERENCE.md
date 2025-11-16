# ACM Phase I - Quick Reference Guide

**Last Updated:** 2025-11-16

---

## Quick Links

- **[Phase I Implementation Plan](./PHASE-I-IMPLEMENTATION-PLAN.md)** - Complete 24-week roadmap
- **[Task Breakdown](./TASK-BREAKDOWN.md)** - Atomic task tracking (64 tasks for Week 1-2)
- **[Project Documentation](../../CLAUDE.md)** - AI assistant guide
- **[Technical Architecture](../../acm-tad.md)** - System design
- **[Threat Model](../../acm-threat-model.md)** - Security analysis

---

## Current Status

**Phase:** Week 1-2 (Foundation & Development Environment)
**Milestone:** M1 - Project Infrastructure
**Tasks Remaining:** 18 tasks (Week 1-2)

---

## Next 5 Tasks to Complete

1. **Initialize Go module**
   ```bash
   go mod init github.com/acm-project/acm
   ```

2. **Create .gitignore**
   ```bash
   # See template in task breakdown
   ```

3. **Create project directory structure**
   ```bash
   mkdir -p cmd/acm-service cmd/acm cmd/acm-setup
   mkdir -p internal/service internal/crypto internal/storage
   # ... (see TASK-BREAKDOWN.md for full list)
   ```

4. **Create tools.go**
   ```go
   //go:build tools
   // +build tools

   package tools

   import (
       _ "google.golang.org/protobuf/cmd/protoc-gen-go"
       _ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
   )
   ```

5. **Add core dependencies**
   ```bash
   go get google.golang.org/grpc@v1.60.0
   go get google.golang.org/protobuf@v1.31.0
   go get github.com/awnumar/memguard@v0.22.4
   # ... (see go.mod template)
   ```

---

## Week 1-2 Goals

### Must Complete (Critical Path)

- [ ] Go module initialized
- [ ] Project structure created
- [ ] Makefile working (`make build`, `make test`)
- [ ] Protocol buffers defined
- [ ] CI/CD pipeline running
- [ ] Certificate generation working
- [ ] Database schema created

### Should Complete

- [ ] golangci-lint configured
- [ ] Security scanning (gosec) set up
- [ ] Docker development environment
- [ ] Foundation documentation

### Nice to Have

- [ ] Pre-commit hooks
- [ ] VSCode workspace config
- [ ] EditorConfig

---

## Development Workflow

### Daily Workflow

```bash
# 1. Pull latest changes
git pull origin claude/setup-project-config-01FgH7kdAS6zu3C4e8py6P8f

# 2. Create feature branch (optional)
git checkout -b feature/my-feature

# 3. Make changes

# 4. Run tests
make test

# 5. Run linter
make lint

# 6. Run security scan
make security-scan

# 7. Commit changes
git add .
git commit -m "feat: description"

# 8. Push changes
git push -u origin feature/my-feature
```

### Build Commands

```bash
make build           # Build all binaries
make test            # Run all tests
make lint            # Run linters
make security-scan   # Run security scanners
make generate        # Generate protobuf code
make clean           # Clean build artifacts
make help            # Show all targets
```

### Certificate Management

```bash
# Generate certificates
./scripts/generate-certs.sh

# Renew certificates
./scripts/generate-certs.sh --renew

# Verify certificates
openssl x509 -in ~/.acm/certs/server.pem -text -noout
```

### Database Management

```bash
# Initialize database
./scripts/setup-sqlite.sh

# View database schema
sqlite3 ~/.acm/data/audit.db ".schema"

# Query audit logs
sqlite3 ~/.acm/data/audit.db "SELECT * FROM audit_events LIMIT 10"
```

---

## Project Structure (After Week 1-2)

```
automated-compromise-mitigation/
├── cmd/
│   ├── acm-service/      # Service daemon
│   ├── acm/              # OpenTUI client
│   └── acm-setup/        # Setup tool
├── internal/
│   ├── service/          # ACM service core
│   ├── crypto/           # Crypto operations
│   ├── storage/          # Data storage
│   ├── pwmanager/        # Password manager integrations
│   └── config/           # Configuration
├── pkg/
│   └── acmclient/        # gRPC client library
├── api/proto/acm/v1/     # Protobuf definitions
├── clients/tui/          # OpenTUI client
├── scripts/              # Automation scripts
├── configs/              # Configuration templates
├── docs/                 # Documentation
├── test/                 # Tests
├── .github/workflows/    # CI/CD
├── Makefile              # Build automation
├── go.mod                # Go dependencies
└── README.md             # Project readme
```

---

## Key Files to Create (Week 1-2)

### Build & Automation

- `Makefile` - Build automation
- `scripts/setup-dev.sh` - Dev environment setup
- `scripts/generate-certs.sh` - Certificate generation
- `scripts/setup-sqlite.sh` - Database initialization

### Configuration

- `configs/service.yaml.example` - Service config template
- `configs/client.yaml.example` - Client config template
- `configs/ca-config.json` - CA configuration
- `.golangci.yml` - Linter configuration

### Protocol Buffers

- `api/proto/acm/v1/common.proto` - Common types
- `api/proto/acm/v1/service.proto` - Main service
- `api/proto/acm/v1/crs.proto` - CRS service
- `api/proto/acm/v1/audit.proto` - Audit service
- `api/proto/acm/v1/him.proto` - HIM service

### Database

- `internal/storage/audit/schema.sql` - Audit log schema

### CI/CD

- `.github/workflows/ci.yml` - Continuous integration
- `.github/workflows/security.yml` - Security scanning
- `.github/workflows/release.yml` - Release automation
- `.github/dependabot.yml` - Dependency updates

### Documentation

- `README.md` - Project overview
- `SECURITY.md` - Security policy
- `CONTRIBUTING.md` - Contribution guide
- `docs/development/SETUP.md` - Dev setup
- `docs/development/ARCHITECTURE.md` - Architecture

---

## Common Commands Reference

### Go Commands

```bash
go mod init github.com/acm-project/acm
go mod tidy
go mod verify
go get <package>
go build ./cmd/...
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

### Protobuf Commands

```bash
# Generate Go code from protobuf
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/acm/v1/*.proto
```

### Linting Commands

```bash
golangci-lint run ./...
golangci-lint run --config .golangci.yml ./...
gosec ./...
```

### Security Commands

```bash
# Scan for secrets
git secrets --scan

# Check dependencies
go list -json -m all | nancy sleuth

# Vulnerability scanning
govulncheck ./...
```

---

## Success Metrics (Week 1-2)

### Must Pass

- [ ] `make build` succeeds
- [ ] `make test` succeeds (even with no tests yet)
- [ ] `make lint` shows zero errors
- [ ] `make generate` creates protobuf code
- [ ] CI/CD pipeline runs successfully
- [ ] Certificate generation works
- [ ] Database initialization works

### Quality Targets

- [ ] All code formatted with `gofmt`
- [ ] All imports organized with `goimports`
- [ ] No security findings from `gosec`
- [ ] Documentation complete for setup

---

## Troubleshooting

### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
rm go.sum
go mod tidy
```

### Protobuf Generation Issues

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installation
which protoc-gen-go
which protoc-gen-go-grpc
```

### Certificate Generation Issues

```bash
# Install cfssl
go install github.com/cloudflare/cfssl/cmd/cfssl@latest
go install github.com/cloudflare/cfssl/cmd/cfssljson@latest

# Verify installation
which cfssl
which cfssljson
```

---

## Resources

### Documentation

- [Go Documentation](https://go.dev/doc/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)

### Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
- [Cryptographic Right Answers](https://latacora.micro.blog/2018/04/03/cryptographic-right-answers.html)

### Tools

- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securego/gosec)
- [cfssl](https://github.com/cloudflare/cfssl)
- [SQLite Browser](https://sqlitebrowser.org/)

---

## Support & Communication

### Getting Help

- **Documentation:** Check `docs/` directory first
- **Issues:** Open GitHub issue
- **Questions:** GitHub Discussions
- **Security:** security@acm.dev (PGP key TBD)

### Team Sync

- **Daily Standup:** TBD
- **Weekly Review:** End of each week
- **Sprint Planning:** Start of each 2-week sprint

---

**Last Updated:** 2025-11-16
**Next Update:** End of Week 1 (2025-11-23)
