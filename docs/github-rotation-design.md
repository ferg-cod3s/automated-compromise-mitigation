# GitHub PAT Rotation Architecture Design

**Version:** 1.0
**Date:** 2025-11-17
**Status:** Design Complete
**Task:** Phase III Task 3 - API-Based Rotation (GitHub)

---

## Executive Summary

This document describes the architecture for automated GitHub Personal Access Token (PAT) rotation using the GitHub REST API. The system implements a safe, atomic rotation workflow with state tracking, rollback support, and ACVS compliance validation.

## Goals

1. **Automated PAT Rotation** - Rotate GitHub PATs without manual intervention
2. **Zero-Downtime** - New token works before old token is deleted
3. **State Persistence** - Track rotation progress in SQLite
4. **Rollback Support** - Undo rotation if new token fails validation
5. **ACVS Integration** - Validate rotation against GitHub ToS
6. **Audit Trail** - Log all rotation events with evidence chain

---

## GitHub API Overview

### Authentication

GitHub PATs authenticate via HTTP Basic Auth or Bearer token:
```bash
curl -H "Authorization: token ghp_xxxx" https://api.github.com/user
```

### Relevant API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/user` | GET | Verify token validity and get user info |
| `/user/tokens` | GET | List all PATs (requires `read:user` scope) |
| `/user/tokens` | POST | Create new PAT (fine-grained tokens only) |
| `/user/tokens/{token_id}` | DELETE | Delete PAT by ID |

**Important:** Classic PATs cannot be created via API. Only **fine-grained tokens** support API creation.

### Token Scopes

PATs require specific scopes:
- `repo` - Repository access
- `workflow` - GitHub Actions access
- `read:user` - Read user profile
- `admin:org` - Organization management

**Rotation Requirement:** New token must have **same or superset** of scopes as old token.

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     ACM gRPC Service                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         CredentialService.RotateGitHubPAT()          │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                      │
│                       v                                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           GitHubRotator (Orchestrator)               │  │
│  │  - Pre-flight validation                             │  │
│  │  - State machine execution                           │  │
│  │  - Rollback on failure                               │  │
│  └─┬──────────┬──────────┬──────────┬───────────────────┘  │
│    │          │          │          │                       │
│    v          v          v          v                       │
│  ┌────┐   ┌─────┐   ┌──────┐   ┌──────┐                   │
│  │ GH │   │ACVS │   │State │   │Audit │                   │
│  │API │   │ Svc │   │ DB   │   │ Log  │                   │
│  └────┘   └─────┘   └──────┘   └──────┘                   │
│     │         │         │          │                        │
└─────┼─────────┼─────────┼──────────┼────────────────────────┘
      │         │         │          │
      v         v         v          v
  GitHub    ACVS     SQLite       Audit
   REST      CRC      State        Events
```

### Components

#### 1. **GitHubRotator** (`internal/rotation/github/rotator.go`)
Main orchestrator for rotation workflow.

**Responsibilities:**
- Execute rotation state machine
- Coordinate with GitHub API, ACVS, State DB
- Handle errors and rollback
- Update password manager

**State Machine:**
```
IDLE → VALIDATING → CREATING → TESTING → UPDATING_PM → DELETING → COMPLETE
         ↓              ↓          ↓           ↓           ↓
       ROLLBACK ← ─────────────────────────────────────────┘
```

#### 2. **GitHubAPIClient** (`internal/rotation/github/client.go`)
GitHub API wrapper.

**Methods:**
- `GetUser(token)` - Verify token and get user info
- `ListTokens(token)` - List all PATs
- `CreateToken(token, scopes, expiry)` - Create fine-grained token
- `DeleteToken(token, tokenID)` - Delete token by ID
- `TestToken(token, repo)` - Validate token permissions

#### 3. **RotationState** (`internal/rotation/state.go`)
SQLite-backed state persistence.

**Schema:**
```sql
CREATE TABLE rotation_state (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL,
    provider TEXT NOT NULL, -- 'github', 'aws', etc.
    state TEXT NOT NULL,    -- 'validating', 'creating', etc.
    old_token_id TEXT,
    new_token_id TEXT,
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    expires_at INTEGER,     -- Timeout for cleanup
    metadata_json TEXT,
    INDEX idx_credential_id (credential_id),
    INDEX idx_state (state),
    INDEX idx_expires_at (expires_at)
);
```

#### 4. **ACVS Integration**
Validate rotation against GitHub ToS before proceeding.

**Validation Points:**
- Pre-flight: Check if API-based rotation is allowed
- Post-creation: Log evidence of compliant rotation

---

## Rotation Workflow

### State Machine

```
1. IDLE
   ↓
2. VALIDATING (Pre-flight checks)
   - Verify current token is valid
   - Check password manager has token
   - Validate ACVS allows GitHub API rotation
   - Get current token scopes
   ↓
3. CREATING (Create new token)
   - Call GitHub API to create fine-grained token
   - Same scopes as current token
   - Set expiration (90 days default)
   - Store new_token_id in state
   ↓
4. TESTING (Validate new token)
   - Test new token against GitHub API
   - Verify scopes match
   - Test repository access (if applicable)
   ↓
5. UPDATING_PM (Update password manager)
   - Store new token in password manager
   - Keep old token temporarily
   ↓
6. DELETING (Delete old token)
   - Delete old token via GitHub API
   - Verify deletion
   ↓
7. COMPLETE
   - Clear rotation state
   - Log success to audit trail
   - Add evidence chain entry
```

### Rollback on Failure

If any step fails after CREATING:

```
ROLLBACK:
  1. Delete new token (if created)
  2. Restore old token in password manager
  3. Log failure to audit trail
  4. Mark state as FAILED
  5. Return error to user
```

### Detailed Steps

#### Step 1: VALIDATING (Pre-flight)

```go
// Verify current token
user, err := client.GetUser(currentToken)
if err != nil {
    return fmt.Errorf("current token invalid: %w", err)
}

// Get token scopes (from password manager metadata)
scopes := getTokenScopes(credentialID)

// ACVS validation
validation, err := acvs.ValidateAction(ctx, &ValidateActionRequest{
    Site: "github.com",
    Action: &AutomationAction{
        Type:   ACTION_TYPE_CREDENTIAL_ROTATION,
        Method: AUTOMATION_METHOD_API,
    },
})

if validation.Result == VALIDATION_RESULT_BLOCKED {
    return fmt.Errorf("GitHub ToS prohibits API rotation")
}

if validation.Result == VALIDATION_RESULT_HIM_REQUIRED {
    // Trigger Human-in-the-Middle workflow
    return requireHIM(ctx, credentialID)
}

// Store state
state := RotationState{
    ID:           generateID(),
    CredentialID: credentialID,
    Provider:     "github",
    State:        "validating",
    StartedAt:    time.Now(),
    Metadata: map[string]string{
        "username": user.Login,
        "scopes":   strings.Join(scopes, ","),
    },
}
db.SaveState(state)
```

#### Step 2: CREATING (Create new token)

```go
// Create fine-grained token
newToken, err := client.CreateToken(currentToken, CreateTokenRequest{
    Name:       fmt.Sprintf("ACM-Rotated-%d", time.Now().Unix()),
    Scopes:     scopes,
    ExpiresAt:  time.Now().Add(90 * 24 * time.Hour),
})

if err != nil {
    // Rollback: nothing to clean up yet
    return fmt.Errorf("failed to create token: %w", err)
}

// Update state
state.State = "creating"
state.NewTokenID = newToken.ID
state.Metadata["new_token_created_at"] = newToken.CreatedAt
db.SaveState(state)
```

#### Step 3: TESTING (Validate new token)

```go
// Test new token
user, err := client.GetUser(newToken.Token)
if err != nil {
    // Rollback: delete new token
    client.DeleteToken(currentToken, newToken.ID)
    return fmt.Errorf("new token validation failed: %w", err)
}

// Verify scopes (if API supports it)
actualScopes := newToken.Scopes
if !scopesMatch(scopes, actualScopes) {
    // Rollback
    client.DeleteToken(currentToken, newToken.ID)
    return fmt.Errorf("scope mismatch")
}

// Update state
state.State = "testing"
db.SaveState(state)
```

#### Step 4: UPDATING_PM (Update password manager)

```go
// Update password manager
err = passwordManager.UpdateCredential(credentialID, Credential{
    Username: user.Login,
    Password: newToken.Token,
    Metadata: map[string]string{
        "token_id":   newToken.ID,
        "scopes":     strings.Join(scopes, ","),
        "created_at": newToken.CreatedAt,
        "expires_at": newToken.ExpiresAt,
    },
})

if err != nil {
    // Rollback: delete new token, restore old token
    client.DeleteToken(currentToken, newToken.ID)
    return fmt.Errorf("failed to update password manager: %w", err)
}

// Update state
state.State = "updating_pm"
db.SaveState(state)
```

#### Step 5: DELETING (Delete old token)

```go
// Delete old token from GitHub
err = client.DeleteToken(newToken.Token, oldTokenID)
if err != nil {
    // Log warning but don't fail - new token is already active
    log.Warnf("Failed to delete old token: %v", err)
    // User can manually delete via GitHub UI
}

// Update state
state.State = "deleting"
state.OldTokenID = oldTokenID
db.SaveState(state)
```

#### Step 6: COMPLETE (Finalize)

```go
// Log to audit trail
audit.LogEvent(AuditEvent{
    Type:         "rotation",
    Status:       "success",
    CredentialID: credentialID,
    Site:         "github.com",
    Message:      "GitHub PAT rotated successfully",
})

// Add evidence chain entry
acvs.AddEvidenceEntry(EvidenceEntry{
    EventType:        EVIDENCE_EVENT_TYPE_ROTATION,
    Site:             "github.com",
    CredentialIDHash: hashCredentialID(credentialID),
    Action:           action,
    ValidationResult: VALIDATION_RESULT_ALLOWED,
    CRCID:            crcID,
})

// Clear state
db.DeleteState(state.ID)

return RotationResult{
    Success:      true,
    NewTokenID:   newToken.ID,
    OldTokenID:   oldTokenID,
    RotatedAt:    time.Now(),
}
```

---

## Error Handling

### Categorized Errors

| Error Type | Example | Rollback Strategy |
|------------|---------|-------------------|
| **Pre-flight** | Current token invalid | No rollback needed, fail early |
| **Creation failed** | GitHub API rate limit | No rollback needed, retry later |
| **Testing failed** | New token has wrong scopes | Delete new token |
| **PM update failed** | Password manager unavailable | Delete new token, restore old |
| **Deletion failed** | Old token already deleted | Log warning, continue |

### Retry Logic

- **Transient errors:** Retry with exponential backoff (3 attempts)
  - Network errors
  - GitHub API rate limits (respect `Retry-After` header)
  - Database locks

- **Permanent errors:** Fail immediately
  - Invalid credentials
  - Insufficient permissions
  - ACVS blocks rotation

### State Recovery

On service restart:

```go
// Find incomplete rotations
incompleteStates := db.FindStates(where: "state != 'complete'")

for _, state := range incompleteStates {
    if time.Since(state.UpdatedAt) > 1 * time.Hour {
        // Rotation timed out, attempt rollback
        rollbackRotation(state)
    } else {
        // Resume rotation
        resumeRotation(state)
    }
}
```

---

## GitHub Token Types

### Classic PAT (Not Supported)

- **Prefix:** `ghp_`
- **Creation:** Manual via GitHub UI only
- **Scopes:** Broad (repo, workflow, etc.)
- **ACM Support:** ❌ Cannot rotate via API

### Fine-Grained PAT (Supported)

- **Prefix:** `github_pat_`
- **Creation:** API or UI
- **Scopes:** Granular (per-repository)
- **Expiration:** Required (max 1 year)
- **ACM Support:** ✅ Full rotation support

**User Guidance:** ACM will detect token type and guide users to migrate from Classic → Fine-Grained.

---

## Security Considerations

### Token Storage

- **Old token:** Kept in password manager until deletion confirmed
- **New token:** Returned to caller, never logged
- **State DB:** Token IDs stored, never raw tokens

### Least Privilege

- New token created with **minimum required scopes**
- If old token has excessive scopes, recommend scope reduction

### Audit Trail

Every rotation generates:
1. **Audit event:** Rotation start, success/failure
2. **Evidence chain entry:** ACVS validation proof
3. **State record:** Full rotation history

### Rate Limiting

GitHub API limits:
- **5,000 requests/hour** (authenticated)
- **60 requests/hour** (unauthenticated)

**ACM Strategy:**
- Batch rotations with delays
- Respect `X-RateLimit-Remaining` header
- Queue rotations if rate-limited

---

## ACVS Integration

### Pre-Flight Validation

```go
// Check GitHub ToS compliance
validation := acvs.ValidateAction(ctx, ValidateActionRequest{
    Site: "github.com",
    Action: AutomationAction{
        Type:   ACTION_TYPE_CREDENTIAL_ROTATION,
        Method: AUTOMATION_METHOD_API,
        Context: map[string]string{
            "token_type": "fine_grained_pat",
            "scopes":     "repo,workflow",
        },
    },
})

switch validation.Result {
case VALIDATION_RESULT_ALLOWED:
    // Proceed with rotation
case VALIDATION_RESULT_HIM_REQUIRED:
    // Require human intervention (e.g., MFA)
case VALIDATION_RESULT_BLOCKED:
    // Abort rotation
}
```

### Evidence Chain

```go
// Log rotation event to evidence chain
evidenceEntry := EvidenceEntry{
    EventType:        EVIDENCE_EVENT_TYPE_ROTATION,
    Site:             "github.com",
    CredentialIDHash: sha256(credentialID),
    Action:           action,
    ValidationResult: validation.Result,
    CRCID:            crc.ID,
    AppliedRuleIDs:   validation.ApplicableRuleIDs,
    EvidenceData: map[string]interface{}{
        "old_token_id":  oldTokenID,
        "new_token_id":  newToken.ID,
        "rotated_at":    time.Now(),
        "rotation_type": "api",
    },
}

acvs.AddEvidenceEntry(ctx, evidenceEntry)
```

---

## Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| Full rotation | < 30s | GitHub API latency ~500ms |
| State persistence | < 10ms | SQLite write |
| Rollback | < 10s | Delete new token + restore |
| Concurrent rotations | 10/minute | Respect GitHub rate limits |

---

## Testing Strategy

### Unit Tests

- GitHub API client (mocked responses)
- Rotation state machine
- Error handling and rollback
- ACVS integration

### Integration Tests

- End-to-end rotation with test GitHub account
- Rollback scenarios
- State recovery after service restart
- Rate limit handling

### Manual Testing

- Rotate real GitHub PAT
- Verify new token works in git operations
- Verify old token deleted
- Check audit logs and evidence chain

---

## User Experience

### CLI Command

```bash
# Rotate specific GitHub credential
acm rotate github --credential-id abc123

# Rotate all GitHub credentials
acm rotate github --all

# Dry-run (validate only, don't rotate)
acm rotate github --credential-id abc123 --dry-run
```

### TUI Workflow (Phase III Task 5)

```
┌─────────────────────────────────────────────┐
│ GitHub PAT Rotation                         │
├─────────────────────────────────────────────┤
│                                             │
│ Credential: github.com (user@example.com)  │
│ Token ID:   gho_abc123xyz                  │
│ Scopes:     repo, workflow                 │
│ Expires:    2025-12-31                     │
│                                             │
│ [✓] Validating current token               │
│ [→] Creating new token...                  │
│ [ ] Testing new token                      │
│ [ ] Updating password manager              │
│ [ ] Deleting old token                     │
│                                             │
│ [Cancel]            [Continue]             │
└─────────────────────────────────────────────┘
```

---

## Future Enhancements (Phase IV)

1. **OAuth App Integration** - Rotate OAuth tokens (not just PATs)
2. **GitHub Enterprise Support** - Custom API endpoints
3. **Organization Tokens** - Rotate org-level PATs
4. **Deploy Key Rotation** - Rotate SSH deploy keys
5. **Scheduled Rotation** - Automatic rotation before expiration
6. **Scope Optimization** - Recommend minimal scopes

---

## References

- [GitHub REST API - Authentication](https://docs.github.com/en/rest/authentication)
- [GitHub PAT Documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [Fine-Grained PATs](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token)

---

**Document Status:** Design Complete ✅
**Next Step:** Implement GitHub API Client (Task 3.2)
