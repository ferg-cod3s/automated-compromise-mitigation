# ACM Phase III Planning Document

**Date:** 2025-11-17
**Status:** Planning Phase
**Prerequisites:** Phase I & II 100% Complete
**Target Timeline:** 6-8 weeks

---

## Executive Summary

Phase III focuses on enhancing the ACM system with production-ready features deferred from Phase II, including SQLite persistence, production NLP models, API-based credential rotation, OpenTUI interface, and enhanced Human-in-the-Middle (HIM) workflows.

**Phase III Goals:**
1. Replace in-memory storage with SQLite persistence
2. Integrate production Legal NLP engine with Python spaCy
3. Implement API-based rotation for GitHub and AWS IAM
4. Build OpenTUI terminal interface with Bubbletea
5. Enhance HIM workflows with TOTP/MFA and CAPTCHA support

**Success Criteria:**
- All data persists across service restarts
- Real ToS analysis with production NLP models
- Automated GitHub PAT and AWS IAM credential rotation
- Fully functional terminal UI for all operations
- MFA/CAPTCHA handling via HIM workflows

---

## Table of Contents

1. [Task 1: SQLite Persistence](#task-1-sqlite-persistence)
2. [Task 2: Production Legal NLP Engine](#task-2-production-legal-nlp-engine)
3. [Task 3: API-Based Rotation - GitHub](#task-3-api-based-rotation---github)
4. [Task 4: API-Based Rotation - AWS IAM](#task-4-api-based-rotation---aws-iam)
5. [Task 5: OpenTUI Interface](#task-5-opentui-interface)
6. [Task 6: Enhanced HIM Workflows](#task-6-enhanced-him-workflows)
7. [Dependencies & Sequencing](#dependencies--sequencing)
8. [Testing Strategy](#testing-strategy)
9. [Documentation Requirements](#documentation-requirements)

---

## Task 1: SQLite Persistence

**Objective:** Replace in-memory storage with SQLite for CRC cache, evidence chains, and audit logs.

**Priority:** HIGH
**Estimated Time:** 1 week
**Dependencies:** None (can start immediately)

### Atomic Subtasks

#### 1.1 Design SQLite Schema
- [ ] Design `crcs` table schema (id, site, tos_url, tos_hash, parsed_at, expires_at, recommendation, rules_json, signature)
- [ ] Design `evidence_entries` table schema (id, timestamp, event_type, site, credential_id_hash, action_json, validation_result, crc_id, applied_rule_ids, evidence_data_json, previous_entry_id, chain_hash, signature)
- [ ] Design `audit_events` table schema (already exists in Phase I, verify compatibility)
- [ ] Design indexes for performance (site, expires_at, timestamp, credential_id_hash)
- [ ] Create migration strategy for schema versioning
- **Files:** `internal/acvs/storage/schema.sql`, `internal/acvs/storage/migrations/`
- **Success Criteria:** SQL schema reviewed and validated

#### 1.2 Implement SQLite CRC Storage
- [ ] Create `internal/acvs/storage/sqlite_crc.go` implementing CRCManager interface
- [ ] Implement `Store()` method with JSON serialization of rules
- [ ] Implement `Get()` method with JSON deserialization
- [ ] Implement `List()` method with site filtering
- [ ] Implement `Invalidate()` method
- [ ] Implement `Clear()` method
- [ ] Implement `GetCacheStats()` method with SQL aggregation
- [ ] Add connection pooling and transaction support
- **Files:** `internal/acvs/storage/sqlite_crc.go` (~300 lines)
- **Success Criteria:** All CRCManager interface methods implemented

#### 1.3 Implement SQLite Evidence Chain Storage
- [ ] Create `internal/acvs/storage/sqlite_evidence.go` implementing EvidenceChainGenerator interface
- [ ] Implement `AddEntry()` method with atomic insertion
- [ ] Implement `GetEntry()` method
- [ ] Implement `Export()` method with filtering
- [ ] Implement `Verify()` method
- [ ] Implement `VerifyChain()` method with SQL joins
- [ ] Implement `GetChainHead()` method
- [ ] Implement `GetChainLength()` method
- [ ] Add transaction support for chain integrity
- **Files:** `internal/acvs/storage/sqlite_evidence.go` (~400 lines)
- **Success Criteria:** All EvidenceChainGenerator interface methods implemented

#### 1.4 Database Initialization and Migration
- [ ] Create database initialization function
- [ ] Implement schema migration system (version tracking)
- [ ] Add automatic migration on service startup
- [ ] Implement rollback support for failed migrations
- [ ] Create database backup functionality
- **Files:** `internal/acvs/storage/init.go`, `internal/acvs/storage/migrate.go`
- **Success Criteria:** Database auto-initializes and migrates on first run

#### 1.5 Update ACVS Service Integration
- [ ] Add configuration flag for storage backend (memory vs. sqlite)
- [ ] Update `NewService()` to accept storage backend
- [ ] Implement graceful fallback to in-memory if SQLite fails
- [ ] Add database health checks
- [ ] Update service shutdown to close database connections
- **Files:** `internal/acvs/service.go` (modifications)
- **Success Criteria:** Service runs with both memory and SQLite backends

#### 1.6 Testing
- [ ] Create unit tests for SQLite CRC storage (10 test cases)
- [ ] Create unit tests for SQLite evidence storage (12 test cases)
- [ ] Create integration tests for data persistence across restarts
- [ ] Test migration system with schema changes
- [ ] Test concurrent access with multiple goroutines
- [ ] Test database corruption recovery
- **Files:** `internal/acvs/storage/sqlite_crc_test.go`, `internal/acvs/storage/sqlite_evidence_test.go`
- **Success Criteria:** All tests pass, 100% persistence verified

#### 1.7 Documentation
- [ ] Document SQLite schema in `docs/storage-schema.md`
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md with storage details
- [ ] Add database configuration examples
- [ ] Document backup and restore procedures
- **Success Criteria:** Complete storage documentation

---

## Task 2: Production Legal NLP Engine

**Objective:** Replace mock ToS analysis with production Python spaCy-based Legal NLP engine.

**Priority:** MEDIUM
**Estimated Time:** 2 weeks
**Dependencies:** None (parallel with Task 1)

### Atomic Subtasks

#### 2.1 Design NLP Service Architecture
- [ ] Design gRPC interface for Python NLP service
- [ ] Define proto messages for ToS analysis requests/responses
- [ ] Plan entity recognition schema (clauses, restrictions, permissions)
- [ ] Design confidence scoring algorithm
- [ ] Plan model training data requirements
- **Files:** `api/proto/acm/v1/nlp.proto`, `docs/nlp-architecture.md`
- **Success Criteria:** Architecture document complete, proto defined

#### 2.2 Create Python NLP Service Scaffold
- [ ] Set up Python project structure (`services/legal-nlp/`)
- [ ] Create Dockerfile for Python service
- [ ] Add requirements.txt (spacy, grpcio, transformers)
- [ ] Create gRPC server stub
- [ ] Add health check endpoint
- [ ] Configure logging
- **Files:** `services/legal-nlp/main.py`, `services/legal-nlp/Dockerfile`, `services/legal-nlp/requirements.txt`
- **Success Criteria:** Python service runs and responds to health checks

#### 2.3 Implement spaCy NLP Pipeline
- [ ] Download and configure spaCy legal model (en_legal_ner)
- [ ] Implement document preprocessing (HTML stripping, section splitting)
- [ ] Implement named entity recognition for legal terms
- [ ] Implement sentence classification (prohibition, permission, requirement)
- [ ] Implement clause extraction
- [ ] Add confidence scoring based on pattern matching
- **Files:** `services/legal-nlp/nlp_engine.py` (~500 lines)
- **Success Criteria:** NLP pipeline processes ToS documents

#### 2.4 Implement Rule Extraction Logic
- [ ] Create rule templates for automation restrictions
- [ ] Implement pattern matching for API usage terms
- [ ] Implement detection of rate limiting clauses
- [ ] Implement MFA/security requirement extraction
- [ ] Implement credential management policy extraction
- [ ] Map extracted clauses to ComplianceRule proto format
- **Files:** `services/legal-nlp/rule_extractor.py` (~400 lines)
- **Success Criteria:** Rules extracted match expected format

#### 2.5 Implement gRPC Service Handlers
- [ ] Implement AnalyzeToS RPC handler
- [ ] Add request validation and sanitization
- [ ] Implement caching of analyzed ToS (optional)
- [ ] Add timeout handling (30s max)
- [ ] Implement error handling and logging
- **Files:** `services/legal-nlp/grpc_server.py` (~200 lines)
- **Success Criteria:** gRPC service responds to Go client

#### 2.6 Integrate with Go ACVS Service
- [ ] Create Go gRPC client for Python NLP service
- [ ] Update `nlp.Engine` to call Python service instead of mock
- [ ] Add retry logic for Python service failures
- [ ] Implement fallback to mock on NLP service unavailable
- [ ] Add NLP service health monitoring
- **Files:** `internal/acvs/nlp/python_client.go` (~250 lines)
- **Success Criteria:** Go service successfully calls Python NLP service

#### 2.7 Model Training and Tuning
- [ ] Collect sample ToS documents for training (20+ sites)
- [ ] Manually annotate ToS clauses for training data
- [ ] Fine-tune spaCy model on legal domain
- [ ] Evaluate model accuracy (target: >80% precision/recall)
- [ ] Optimize confidence thresholds
- **Files:** `services/legal-nlp/training/`, `services/legal-nlp/models/`
- **Success Criteria:** Model achieves >80% accuracy on test set

#### 2.8 Deployment Configuration
- [ ] Create docker-compose.yml for multi-service deployment
- [ ] Configure service discovery (Go → Python)
- [ ] Add environment configuration for model paths
- [ ] Create startup script for combined service
- [ ] Document Python service deployment
- **Files:** `docker-compose.yml`, `scripts/start-services.sh`
- **Success Criteria:** Both services start together with docker-compose

#### 2.9 Testing
- [ ] Create unit tests for NLP pipeline components
- [ ] Create integration tests with sample ToS documents
- [ ] Test edge cases (empty ToS, malformed HTML, very long documents)
- [ ] Benchmark analysis performance (target: <5s per ToS)
- [ ] Test gRPC communication reliability
- **Files:** `services/legal-nlp/tests/`, `test/integration/nlp_integration_test.go`
- **Success Criteria:** All tests pass, performance targets met

#### 2.10 Documentation
- [ ] Document NLP service architecture
- [ ] Document model training process
- [ ] Create API documentation for gRPC interface
- [ ] Add troubleshooting guide
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md
- **Files:** `docs/nlp-service.md`, `docs/nlp-training.md`
- **Success Criteria:** Complete NLP service documentation

---

## Task 3: API-Based Rotation - GitHub

**Objective:** Implement automated GitHub Personal Access Token (PAT) rotation using GitHub API.

**Priority:** MEDIUM
**Estimated Time:** 1 week
**Dependencies:** Task 1 (SQLite for tracking rotation state)

### Atomic Subtasks

#### 3.1 Design GitHub Rotation Architecture
- [ ] Research GitHub PAT API endpoints and authentication
- [ ] Design rotation workflow (create new → test → delete old)
- [ ] Plan credential storage in password manager
- [ ] Design state tracking for multi-step rotation
- [ ] Plan error handling and rollback strategy
- **Files:** `docs/github-rotation-design.md`
- **Success Criteria:** Architecture reviewed and approved

#### 3.2 Implement GitHub API Client
- [ ] Create GitHub API client struct
- [ ] Implement authentication with existing PAT
- [ ] Implement CreatePAT endpoint
- [ ] Implement DeletePAT endpoint
- [ ] Implement ListPATs endpoint
- [ ] Implement TestPAT endpoint (validate permissions)
- [ ] Add rate limit handling
- **Files:** `internal/rotation/github/client.go` (~300 lines)
- **Success Criteria:** All GitHub API endpoints functional

#### 3.3 Implement Rotation Workflow
- [ ] Create GitHubRotator struct implementing rotation interface
- [ ] Implement pre-flight checks (validate current PAT)
- [ ] Implement new PAT creation with same scopes
- [ ] Implement PAT validation (test against GitHub API)
- [ ] Implement password manager update
- [ ] Implement old PAT deletion
- [ ] Add rollback on failure
- **Files:** `internal/rotation/github/rotator.go` (~400 lines)
- **Success Criteria:** Complete rotation workflow implemented

#### 3.4 Implement State Tracking
- [ ] Create rotation state table in SQLite
- [ ] Implement state persistence for multi-step rotation
- [ ] Implement state recovery on service restart
- [ ] Add state expiration for abandoned rotations
- [ ] Implement state cleanup
- **Files:** `internal/rotation/state.go` (~200 lines)
- **Success Criteria:** Rotation state persists across restarts

#### 3.5 Add ACVS Integration
- [ ] Create ACVS validation for GitHub rotation actions
- [ ] Implement evidence chain logging for rotation events
- [ ] Add compliance checking before rotation
- [ ] Implement HIM fallback for restricted sites
- **Files:** Modifications to `internal/acvs/service.go`
- **Success Criteria:** GitHub rotation validates against ToS

#### 3.6 Create gRPC Service Handlers
- [ ] Add RotateGitHubPAT RPC to credential.proto
- [ ] Implement handler for GitHub rotation
- [ ] Add progress reporting for long-running rotations
- [ ] Implement cancellation support
- **Files:** `api/proto/acm/v1/credential.proto` (modifications), `internal/server/credential_service.go` (modifications)
- **Success Criteria:** gRPC RPC works end-to-end

#### 3.7 Testing
- [ ] Create unit tests for GitHub API client (mocked)
- [ ] Create unit tests for rotation workflow
- [ ] Create integration test with real GitHub account (test mode)
- [ ] Test rollback scenarios
- [ ] Test state recovery
- **Files:** `internal/rotation/github/client_test.go`, `internal/rotation/github/rotator_test.go`
- **Success Criteria:** All tests pass including integration test

#### 3.8 Documentation
- [ ] Document GitHub rotation workflow
- [ ] Create user guide for GitHub PAT setup
- [ ] Document required GitHub scopes
- [ ] Add troubleshooting guide
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md
- **Files:** `docs/github-rotation.md`
- **Success Criteria:** Complete GitHub rotation documentation

---

## Task 4: API-Based Rotation - AWS IAM

**Objective:** Implement automated AWS IAM Access Key rotation using AWS SDK.

**Priority:** MEDIUM
**Estimated Time:** 1 week
**Dependencies:** Task 1 (SQLite), Task 3 (rotation framework)

### Atomic Subtasks

#### 4.1 Design AWS IAM Rotation Architecture
- [ ] Research AWS IAM API and authentication methods
- [ ] Design rotation workflow (create new key → test → deactivate → delete old)
- [ ] Plan AWS credential storage (access key + secret key)
- [ ] Design error handling for AWS-specific issues
- [ ] Plan permission requirements (IAM:CreateAccessKey, IAM:DeleteAccessKey)
- **Files:** `docs/aws-iam-rotation-design.md`
- **Success Criteria:** Architecture document complete

#### 4.2 Implement AWS IAM Client
- [ ] Add AWS SDK dependency (github.com/aws/aws-sdk-go-v2)
- [ ] Create AWS IAM client wrapper
- [ ] Implement CreateAccessKey API call
- [ ] Implement DeleteAccessKey API call
- [ ] Implement ListAccessKeys API call
- [ ] Implement GetUser API call (validation)
- [ ] Add retry logic for AWS throttling
- **Files:** `internal/rotation/aws/client.go` (~350 lines)
- **Success Criteria:** All AWS IAM API calls functional

#### 4.3 Implement AWS Rotation Workflow
- [ ] Create AWSRotator struct
- [ ] Implement pre-flight validation (check current credentials)
- [ ] Implement new access key creation
- [ ] Implement credential validation (test with AWS STS)
- [ ] Implement password manager update with both keys
- [ ] Implement old key deactivation
- [ ] Implement old key deletion (after grace period)
- [ ] Add rollback logic
- **Files:** `internal/rotation/aws/rotator.go` (~450 lines)
- **Success Criteria:** Complete AWS rotation workflow

#### 4.4 Implement Grace Period Handling
- [ ] Add deactivated key tracking in SQLite
- [ ] Implement scheduled deletion after grace period (24-48 hours)
- [ ] Add background job for cleanup
- [ ] Implement manual deletion trigger
- **Files:** `internal/rotation/aws/cleanup.go` (~150 lines)
- **Success Criteria:** Old keys deleted after grace period

#### 4.5 Add ACVS Integration
- [ ] Create ACVS validation for AWS rotation
- [ ] Log evidence chain entries for AWS operations
- [ ] Check AWS Terms of Service compliance
- [ ] Implement HIM workflow if needed
- **Files:** Modifications to ACVS service
- **Success Criteria:** AWS rotation integrated with ACVS

#### 4.6 Create gRPC Service Handlers
- [ ] Add RotateAWSAccessKey RPC
- [ ] Implement handler with progress reporting
- [ ] Add support for multiple AWS accounts
- [ ] Implement cancellation
- **Files:** `api/proto/acm/v1/credential.proto` (modifications)
- **Success Criteria:** gRPC RPC functional

#### 4.7 Testing
- [ ] Create unit tests for AWS client (mocked)
- [ ] Create unit tests for rotation workflow
- [ ] Create integration test with test AWS account
- [ ] Test grace period and cleanup
- [ ] Test rollback scenarios
- **Files:** `internal/rotation/aws/client_test.go`, `internal/rotation/aws/rotator_test.go`
- **Success Criteria:** All tests pass

#### 4.8 Documentation
- [ ] Document AWS rotation workflow
- [ ] Create user guide for AWS IAM setup
- [ ] Document required IAM permissions
- [ ] Add security best practices
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md
- **Files:** `docs/aws-iam-rotation.md`
- **Success Criteria:** Complete AWS rotation documentation

---

## Task 5: OpenTUI Interface

**Objective:** Build terminal UI with Bubbletea for interactive ACM operations.

**Priority:** MEDIUM
**Estimated Time:** 2 weeks
**Dependencies:** Task 1 (data to display), Task 3/4 (rotation features)

### Atomic Subtasks

#### 5.1 Design TUI Architecture
- [ ] Design screen/view hierarchy (main menu, credential list, rotation status, settings)
- [ ] Design navigation model (keyboard shortcuts, tab navigation)
- [ ] Plan data refresh strategy (polling vs. streaming)
- [ ] Design color scheme and styling
- [ ] Create wireframes for each screen
- **Files:** `docs/tui-design.md`, `docs/tui-wireframes/`
- **Success Criteria:** Design approved with wireframes

#### 5.2 Set Up Bubbletea Project Structure
- [ ] Create `cmd/acm-tui/` directory
- [ ] Add Bubbletea dependencies
- [ ] Create main TUI entry point
- [ ] Set up model-update-view pattern
- [ ] Configure terminal detection and fallbacks
- **Files:** `cmd/acm-tui/main.go`
- **Success Criteria:** Basic Bubbletea app runs

#### 5.3 Implement Main Menu Screen
- [ ] Create main menu model
- [ ] Implement menu rendering with Lipgloss styling
- [ ] Add keyboard navigation (arrows, enter)
- [ ] Implement menu item selection
- [ ] Add status bar with service connection status
- **Files:** `cmd/acm-tui/screens/main_menu.go` (~200 lines)
- **Success Criteria:** Main menu functional and styled

#### 5.4 Implement Credential List Screen
- [ ] Create credential list model
- [ ] Implement table view for credentials
- [ ] Add sorting (by site, last rotated, status)
- [ ] Add filtering/search
- [ ] Implement pagination for large lists
- [ ] Add detail view for selected credential
- **Files:** `cmd/acm-tui/screens/credential_list.go` (~350 lines)
- **Success Criteria:** Credential list displays and navigates

#### 5.5 Implement Rotation Workflow Screen
- [ ] Create rotation wizard model
- [ ] Implement step-by-step rotation UI
- [ ] Add progress indicators
- [ ] Implement real-time status updates
- [ ] Add error display with retry options
- [ ] Implement HIM interaction screens
- **Files:** `cmd/acm-tui/screens/rotation.go` (~400 lines)
- **Success Criteria:** Rotation workflow works via TUI

#### 5.6 Implement ACVS Status Screen
- [ ] Create ACVS dashboard model
- [ ] Display ACVS enabled/disabled status
- [ ] Show cached CRCs in table
- [ ] Display evidence chain statistics
- [ ] Add CRC invalidation UI
- [ ] Implement evidence chain export
- **Files:** `cmd/acm-tui/screens/acvs.go` (~300 lines)
- **Success Criteria:** ACVS status visible and manageable

#### 5.7 Implement Settings Screen
- [ ] Create settings model
- [ ] Display service configuration
- [ ] Add ACVS enable/disable toggle
- [ ] Show password manager status
- [ ] Add certificate info display
- [ ] Implement configuration save/load
- **Files:** `cmd/acm-tui/screens/settings.go` (~250 lines)
- **Success Criteria:** Settings viewable and editable

#### 5.8 Implement HIM Interaction Screens
- [ ] Create MFA prompt screen
- [ ] Create CAPTCHA interaction screen
- [ ] Implement TOTP code entry
- [ ] Add manual intervention prompts
- [ ] Implement timeout handling
- **Files:** `cmd/acm-tui/screens/him.go` (~300 lines)
- **Success Criteria:** HIM workflows interactive

#### 5.9 Add Real-Time Updates
- [ ] Implement gRPC streaming for status updates
- [ ] Add event-driven UI updates
- [ ] Implement refresh intervals
- [ ] Add auto-scroll for logs
- [ ] Implement notification system
- **Files:** `cmd/acm-tui/client/streaming.go` (~200 lines)
- **Success Criteria:** UI updates in real-time

#### 5.10 Implement Help and Documentation
- [ ] Create help screen with keyboard shortcuts
- [ ] Add context-sensitive help (F1 key)
- [ ] Implement command palette (Ctrl+P)
- [ ] Add tooltips for complex features
- **Files:** `cmd/acm-tui/screens/help.go` (~150 lines)
- **Success Criteria:** Help accessible from all screens

#### 5.11 Testing
- [ ] Create unit tests for screen models
- [ ] Create integration tests for navigation
- [ ] Test keyboard input handling
- [ ] Test terminal resize handling
- [ ] Test with different terminal emulators
- **Files:** `cmd/acm-tui/screens/*_test.go`
- **Success Criteria:** All screens tested, no panics

#### 5.12 Documentation
- [ ] Create user guide for TUI
- [ ] Document keyboard shortcuts
- [ ] Add screenshots/asciinema recordings
- [ ] Create troubleshooting guide
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md
- **Files:** `docs/tui-user-guide.md`
- **Success Criteria:** Complete TUI documentation with screenshots

---

## Task 6: Enhanced HIM Workflows

**Objective:** Add TOTP/MFA and CAPTCHA support to Human-in-the-Middle workflows.

**Priority:** LOW
**Estimated Time:** 1 week
**Dependencies:** Task 5 (TUI for interaction)

### Atomic Subtasks

#### 6.1 Design Enhanced HIM Architecture
- [ ] Design TOTP integration architecture
- [ ] Design CAPTCHA solver integration (2captcha, anti-captcha)
- [ ] Plan biometric authentication prompts
- [ ] Design session persistence for long-running HIM
- [ ] Plan timeout and retry strategies
- **Files:** `docs/enhanced-him-design.md`
- **Success Criteria:** Architecture document complete

#### 6.2 Implement TOTP Support
- [ ] Add TOTP library dependency (github.com/pquerna/otp)
- [ ] Create TOTP prompt in HIM service
- [ ] Implement TOTP code validation
- [ ] Add QR code display for TOTP setup
- [ ] Implement emergency backup codes
- **Files:** `internal/him/totp.go` (~200 lines)
- **Success Criteria:** TOTP codes accepted and validated

#### 6.3 Implement SMS/Email MFA Support
- [ ] Create SMS code prompt
- [ ] Create email code prompt
- [ ] Implement code entry validation
- [ ] Add resend code functionality
- [ ] Implement timeout handling
- **Files:** `internal/him/mfa.go` (~250 lines)
- **Success Criteria:** MFA codes entered via HIM

#### 6.4 Implement CAPTCHA Integration
- [ ] Add 2captcha API client
- [ ] Implement CAPTCHA image submission
- [ ] Implement solution retrieval
- [ ] Add fallback to manual solving
- [ ] Implement balance checking
- **Files:** `internal/him/captcha.go` (~300 lines)
- **Success Criteria:** CAPTCHAs solved via API or manual

#### 6.5 Implement Push Notification MFA
- [ ] Create push notification prompt
- [ ] Implement wait-for-approval logic
- [ ] Add timeout with retry
- [ ] Show approval status
- **Files:** `internal/him/push_mfa.go` (~150 lines)
- **Success Criteria:** Push MFA handled

#### 6.6 Implement Biometric Prompts
- [ ] Create biometric authentication prompt
- [ ] Add OS-specific biometric integration (TouchID, Windows Hello)
- [ ] Implement fallback to password
- [ ] Add timeout handling
- **Files:** `internal/him/biometric.go` (~200 lines)
- **Success Criteria:** Biometric prompts shown

#### 6.7 Update HIM Service
- [ ] Extend HIM session with new interaction types
- [ ] Add state persistence for multi-step MFA
- [ ] Implement session recovery
- [ ] Add audit logging for HIM interactions
- **Files:** `internal/him/service.go` (modifications)
- **Success Criteria:** All HIM types supported

#### 6.8 Add gRPC Support
- [ ] Extend HIMService proto with new message types
- [ ] Implement streaming for long-running HIM
- [ ] Add progress reporting
- [ ] Implement cancellation
- **Files:** `api/proto/acm/v1/him.proto` (modifications)
- **Success Criteria:** New HIM types available via gRPC

#### 6.9 Integrate with TUI
- [ ] Add TOTP entry screen to TUI
- [ ] Add CAPTCHA display screen
- [ ] Add push notification waiting screen
- [ ] Add biometric prompt screen
- [ ] Implement timeout countdowns
- **Files:** `cmd/acm-tui/screens/him.go` (modifications)
- **Success Criteria:** All HIM types work in TUI

#### 6.10 Testing
- [ ] Create unit tests for each HIM type
- [ ] Create integration tests for HIM workflows
- [ ] Test timeout and retry logic
- [ ] Test CAPTCHA API integration (with test API key)
- [ ] Test biometric integration on multiple platforms
- **Files:** `internal/him/*_test.go`
- **Success Criteria:** All HIM tests pass

#### 6.11 Documentation
- [ ] Document each HIM type
- [ ] Create setup guide for CAPTCHA API
- [ ] Document biometric setup requirements
- [ ] Add troubleshooting guide
- [ ] Update PHASE3_IMPLEMENTATION_SUMMARY.md
- **Files:** `docs/enhanced-him.md`
- **Success Criteria:** Complete HIM documentation

---

## Dependencies & Sequencing

### Task Dependencies

```
Task 1 (SQLite)
├─> Task 3 (GitHub Rotation) - requires state tracking
└─> Task 4 (AWS Rotation) - requires state tracking

Task 2 (NLP) - Independent, can run in parallel

Task 3 (GitHub Rotation)
└─> Task 4 (AWS Rotation) - reuses rotation framework

Task 1, 3, 4
└─> Task 5 (OpenTUI) - needs data and features to display

Task 5 (OpenTUI)
└─> Task 6 (Enhanced HIM) - needs TUI for user interaction
```

### Recommended Sequence

**Week 1-2:**
- Start Task 1 (SQLite Persistence) - HIGH priority
- Start Task 2 (Production NLP) - parallel track

**Week 3:**
- Complete Task 1
- Continue Task 2
- Start Task 3 (GitHub Rotation)

**Week 4:**
- Complete Task 2
- Complete Task 3
- Start Task 4 (AWS Rotation)

**Week 5-6:**
- Complete Task 4
- Start Task 5 (OpenTUI) - major effort

**Week 7-8:**
- Complete Task 5
- Start Task 6 (Enhanced HIM)
- Complete Task 6

### Parallel Work Opportunities

- Task 1 and Task 2 can run fully in parallel
- Task 3 and remaining Task 2 work can overlap
- Task 5 can start once Tasks 1, 3, 4 have basic functionality

---

## Testing Strategy

### Unit Testing Requirements

Each task must include:
- Unit tests for all public functions
- Mock external dependencies (GitHub API, AWS SDK, SQLite)
- Edge case coverage
- Error path testing
- Target: 80%+ code coverage

### Integration Testing Requirements

- End-to-end workflow tests for each feature
- Cross-component integration tests
- Database migration tests
- Service restart and recovery tests
- Performance benchmarks

### Manual Testing Checklist

- [ ] SQLite data persists across service restarts
- [ ] NLP service analyzes real ToS documents accurately
- [ ] GitHub PAT rotation works with real GitHub account
- [ ] AWS IAM rotation works with real AWS account
- [ ] TUI runs in multiple terminal emulators (iTerm2, Windows Terminal, GNOME Terminal)
- [ ] HIM workflows complete successfully with real MFA
- [ ] All services work together in docker-compose
- [ ] Documentation is accurate and complete

### Performance Testing

- [ ] CRC cache operations < 10ms
- [ ] Evidence chain queries < 50ms for 10,000 entries
- [ ] ToS analysis < 5s per document
- [ ] GitHub rotation completes < 30s
- [ ] AWS rotation completes < 30s
- [ ] TUI renders < 16ms per frame (60 FPS)
- [ ] Database migrations < 5s for 100,000 entries

---

## Documentation Requirements

### Per-Task Documentation

Each task must produce:
1. **Architecture Document** - Design decisions and rationale
2. **User Guide** - How to use the feature
3. **API Documentation** - gRPC/function signatures
4. **Troubleshooting Guide** - Common issues and solutions

### Phase III Summary Document

Create `PHASE3_IMPLEMENTATION_SUMMARY.md` including:
- All implemented features
- Code metrics and statistics
- Test coverage results
- Performance benchmarks
- Known limitations
- Migration guide from Phase II
- Future Phase IV recommendations

### Updated Documents

- [ ] Update `CLAUDE.md` with Phase III status
- [ ] Update `README.md` with new features
- [ ] Update `acm-tad.md` with architecture changes
- [ ] Update `acm-security-planning.md` with new security measures
- [ ] Create `CHANGELOG.md` with version history

---

## Success Metrics

### Functional Metrics

- [ ] 100% of SQLite persistence tests passing
- [ ] NLP accuracy >80% on test ToS documents
- [ ] GitHub rotation success rate >95%
- [ ] AWS rotation success rate >95%
- [ ] TUI works in all major terminal emulators
- [ ] All HIM types implemented and tested

### Performance Metrics

- [ ] Database operations meet performance targets
- [ ] ToS analysis completes in <5s
- [ ] API rotations complete in <30s each
- [ ] TUI maintains 60 FPS
- [ ] Service memory usage <500MB

### Quality Metrics

- [ ] Code coverage >80%
- [ ] Zero critical bugs in production
- [ ] All documentation complete and reviewed
- [ ] Security audit passed
- [ ] User acceptance testing passed

---

## Risk Assessment

### High-Risk Items

1. **NLP Model Accuracy**
   - **Risk:** Production NLP may not be accurate enough
   - **Mitigation:** Start with mock fallback, iterate on model
   - **Contingency:** Keep mock analysis as fallback

2. **API Rate Limits**
   - **Risk:** GitHub/AWS APIs may rate limit during rotation
   - **Mitigation:** Implement exponential backoff and retries
   - **Contingency:** Queue rotations and process slowly

3. **Database Migrations**
   - **Risk:** Migration failures could corrupt data
   - **Mitigation:** Implement rollback and backup
   - **Contingency:** Manual recovery procedures

### Medium-Risk Items

1. **TUI Compatibility** - Test extensively across terminals
2. **CAPTCHA API Costs** - Monitor usage and implement limits
3. **Biometric Integration** - OS-specific, may not work everywhere

---

## Definition of Done

Phase III is complete when:

- [ ] All 6 tasks completed with subtasks checked off
- [ ] All unit tests passing (target: 150+ new tests)
- [ ] All integration tests passing
- [ ] All manual testing completed
- [ ] Performance benchmarks met
- [ ] All documentation written and reviewed
- [ ] Security review completed
- [ ] User acceptance testing passed
- [ ] PHASE3_IMPLEMENTATION_SUMMARY.md published
- [ ] All code committed and pushed
- [ ] Ready for Phase IV planning

---

**Document Status:** Planning Complete
**Next Step:** Begin Task 1 (SQLite Persistence)
**Target Completion:** 8 weeks from start date

