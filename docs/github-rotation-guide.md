# GitHub Personal Access Token (PAT) Rotation Guide

**Version:** 1.0
**Last Updated:** November 2025
**Status:** Phase III Implementation

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [How It Works](#how-it-works)
4. [Step-by-Step Guide](#step-by-step-guide)
5. [ACVS Integration](#acvs-integration)
6. [Troubleshooting](#troubleshooting)
7. [API Reference](#api-reference)
8. [Security Considerations](#security-considerations)

---

## Overview

ACM's GitHub PAT Rotation feature provides a **semi-automated workflow** for rotating your GitHub Personal Access Tokens. Due to GitHub API limitations that prevent programmatic token creation/deletion, ACM guides you through the manual steps while automating validation, state tracking, and compliance checking.

### Why Semi-Automated?

GitHub's security model doesn't allow programmatic creation or deletion of Personal Access Tokens through the API. This is a security feature to prevent unauthorized token management. ACM works within these constraints by:

- **Validating** your current token before rotation
- **Guiding** you through manual token creation/deletion with step-by-step instructions
- **Verifying** the new token works correctly
- **Tracking** rotation state to ensure completion
- **Recording** evidence for compliance auditing (if ACVS is enabled)

---

## Prerequisites

### Required

1. **GitHub Personal Access Token** - Your current token that needs rotation
2. **ACM Service Running** - The ACM gRPC service must be running on localhost:8443
3. **mTLS Certificates** - Client certificates for secure communication
4. **GitHub Account Access** - Ability to access github.com/settings/tokens

### Optional

5. **ACVS Enabled** - For Terms of Service compliance validation
6. **Password Manager** - To store the new token securely after rotation

---

## How It Works

### The Three-Step Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                   GitHub PAT Rotation Workflow                   │
└─────────────────────────────────────────────────────────────────┘

Step 1: START ROTATION
├─→ Validate current token
├─→ Check ACVS compliance (if enabled)
├─→ Create rotation state
└─→ Provide creation instructions

Step 2: VERIFY NEW TOKEN
├─→ User creates new token manually
├─→ ACM validates new token
├─→ Verify same user/account
└─→ Provide deletion instructions

Step 3: CONFIRM DELETION
├─→ User deletes old token manually
├─→ Mark rotation complete
├─→ Add evidence chain entry (if ACVS enabled)
└─→ Clean up rotation state

```

### State Management

Each rotation creates a unique state ID that tracks progress through the workflow:

| State | Description | User Action Required |
|-------|-------------|---------------------|
| `validating` | Validating current token | None |
| `guiding` | Providing instructions | Follow instructions |
| `waiting_deletion` | Waiting for old token deletion | Delete old token |
| `verifying` | Verifying new token | None |
| `complete` | Rotation finished | None |
| `failed` | Rotation failed | Review error |
| `cancelled` | User cancelled | None |

---

## Step-by-Step Guide

### Step 1: Start Rotation

Use the `StartGitHubRotation` RPC call:

```go
import (
    acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

req := &acmv1.StartGitHubRotationRequest{
    CredentialId:  "my-github-token-001",  // Your identifier
    CurrentToken:  "ghp_xxxxxxxxxxxxxxxxxxxx",  // Current PAT
    Site:          "github.com",  // or GitHub Enterprise URL
}

resp, err := rotationClient.StartGitHubRotation(ctx, req)
if err != nil {
    log.Fatalf("Failed to start rotation: %v", err)
}

fmt.Printf("Rotation State ID: %s\n", resp.StateId)
fmt.Printf("Instructions:\n%s\n", resp.Instructions)
```

**Expected Output:**

```
Rotation State ID: rot-a1b2c3d4e5f67890

GitHub Personal Access Token Rotation Guide

Step 1: Create New Fine-Grained Token
--------------------------------------
1. Go to: https://github.com/settings/tokens/new
2. Click "Generate new token" → "Generate new token (fine-grained)"
3. Enter a descriptive name: "ACM Rotated Token - 2025-11-17"
4. Set expiration: 90 days (or your preference)
5. Select the SAME permissions as your current token:
   - Repository access
   - Permissions (match your current scopes)
6. Click "Generate token"
7. COPY THE TOKEN - you won't see it again!

Step 2: Verify New Token
-------------------------
Once you have the new token, return to ACM and provide it for verification.
```

**Important:** Save the `StateId` - you'll need it for the next steps.

### Step 2: Verify New Token

After manually creating the new token on GitHub:

```go
req := &acmv1.VerifyNewTokenRequest{
    StateId:  "rot-a1b2c3d4e5f67890",  // From Step 1
    NewToken: "ghp_yyyyyyyyyyyyyyyyyyyy",  // New token from GitHub
}

resp, err := rotationClient.VerifyNewToken(ctx, req)
if err != nil {
    log.Fatalf("Failed to verify token: %v", err)
}

fmt.Printf("Verification: %s\n", resp.Status.Message)
fmt.Printf("Next Instructions:\n%s\n", resp.Instructions)
```

**Expected Output:**

```
Verification: New token verified successfully

✅ New Token Verified Successfully!

Step 3: Delete Old Token
-------------------------
Now that your new token is working, you can safely delete the old one:

1. Go to: https://github.com/settings/tokens
2. Find your OLD token in the list
3. Click "Delete" next to the old token
4. Confirm deletion

Once deleted, return to ACM to confirm completion.

IMPORTANT: Make sure your password manager is updated with the NEW token
before deleting the old one!
```

### Step 3: Confirm Deletion

After manually deleting the old token:

```go
req := &acmv1.ConfirmDeletionRequest{
    StateId: "rot-a1b2c3d4e5f67890",  // Same state ID
}

resp, err := rotationClient.ConfirmDeletion(ctx, req)
if err != nil {
    log.Fatalf("Failed to confirm deletion: %v", err)
}

fmt.Printf("Rotation Complete!\n")
fmt.Printf("Completed at: %s\n", time.Unix(resp.CompletedAt, 0))
```

**Expected Output:**

```
Rotation Complete!
Completed at: 2025-11-17 14:30:45
```

---

## ACVS Integration

### Automatic Compliance Validation

If ACVS (Automated Compliance Validation Service) is enabled, ACM automatically:

1. **Pre-flight Validation** - Checks GitHub's ToS before rotation
2. **Evidence Chain** - Records cryptographically-signed audit trail
3. **Compliance Rules** - Applies cached CRC (Compliance Rule Set)

### ToS Compliance Check

Example ACVS validation during `StartGitHubRotation`:

```
ACVS Validation:
  Site: github.com
  CRC ID: crc-github-v2-2025-01
  Result: ALLOWED
  Reasoning: GitHub ToS permits automated token rotation for security purposes
  Rules Applied: github-automation-001, github-rate-limit-002
```

### Evidence Chain Entry

Upon completion, ACVS adds an evidence entry:

```json
{
  "event_type": "ROTATION",
  "site": "github.com",
  "credential_id_hash": "sha256:a1b2c3...",
  "action": {
    "type": "CREDENTIAL_ROTATION",
    "method": "MANUAL"
  },
  "validation_result": "ALLOWED",
  "crc_id": "crc-github-v2-2025-01",
  "evidence_data": {
    "rotation_id": "rot-a1b2c3d4e5f67890",
    "started_at": "2025-11-17T14:28:00Z",
    "completed_at": "2025-11-17T14:30:45Z",
    "duration_mins": 2.75,
    "username": "yourghuser"
  }
}
```

---

## Troubleshooting

### Common Issues

#### 1. Token Validation Fails

**Error:** `Current token validation failed: 401 Unauthorized`

**Solution:**
- Verify token is correct (no typos)
- Check token hasn't expired
- Ensure token has `user:read` permission minimum
- Verify GitHub API is accessible

#### 2. New Token Belongs to Different User

**Error:** `Token belongs to different user (expected: alice, got: bob)`

**Solution:**
- Ensure you're logged into the correct GitHub account
- Create the token while logged in as the same user

#### 3. ACVS Blocks Rotation

**Error:** `GitHub ToS prohibits this type of rotation`

**Solution:**
- Review ACVS reasoning in error message
- Check if GitHub has updated their ToS
- Invalidate CRC cache and retry: `InvalidateCRC("github.com")`
- Contact your compliance team if automated rotation is necessary

#### 4. Rotation State Not Found

**Error:** `Rotation state not found: rot-xyz123`

**Solution:**
- Check state ID is correct
- State may have expired (24-hour timeout)
- Start a new rotation

#### 5. Rotation Timeout

**Symptom:** State shows `expires_at` in the past

**Solution:**
- Rotation states expire after 24 hours for security
- Start a new rotation
- Complete rotations promptly to avoid expiration

### Viewing Rotation Status

Check the status of any rotation at any time:

```go
req := &acmv1.GetGitHubRotationStatusRequest{
    StateId: "rot-a1b2c3d4e5f67890",
}

resp, err := rotationClient.GetGitHubRotationStatus(ctx, req)
fmt.Printf("Current Step: %v\n", resp.CurrentStep)
fmt.Printf("Started: %s\n", time.Unix(resp.StartedAt, 0))
fmt.Printf("Updated: %s\n", time.Unix(resp.UpdatedAt, 0))
```

### Canceling a Rotation

If you need to cancel an in-progress rotation:

```go
req := &acmv1.CancelGitHubRotationRequest{
    StateId: "rot-a1b2c3d4e5f67890",
}

resp, err := rotationClient.CancelGitHubRotation(ctx, req)
fmt.Printf("Rotation cancelled\n")
```

**Note:** Cancellation doesn't delete any tokens you've created. You must manually clean up.

### Listing Active Rotations

View all incomplete rotations:

```go
req := &acmv1.ListActiveGitHubRotationsRequest{
    Site: "github.com",  // Optional filter
}

resp, err := rotationClient.ListActiveGitHubRotations(ctx, req)
for _, rotation := range resp.Rotations {
    fmt.Printf("State ID: %s, User: %s, Step: %v\n",
        rotation.StateId, rotation.Username, rotation.CurrentStep)
}
```

---

## API Reference

### gRPC Service: RotationService

#### StartGitHubRotation

**Request:**
```protobuf
message StartGitHubRotationRequest {
  string credential_id = 2;   // Required: Your identifier
  string current_token = 3;   // Required: Current PAT
  string site = 4;            // Optional: Default "github.com"
  string username = 5;        // Optional: Auto-detected
}
```

**Response:**
```protobuf
message StartGitHubRotationResponse {
  Status status = 1;
  string state_id = 2;         // Save this for next steps
  GitHubRotationStep next_step = 3;
  string instructions = 4;      // User-facing guide
  string username = 5;          // Detected GitHub username
  string crc_id = 6;           // CRC used (if ACVS enabled)
}
```

#### VerifyNewToken

**Request:**
```protobuf
message VerifyNewTokenRequest {
  string state_id = 2;        // Required: From StartGitHubRotation
  string new_token = 3;       // Required: Newly created PAT
}
```

**Response:**
```protobuf
message VerifyNewTokenResponse {
  Status status = 1;
  string state_id = 2;
  GitHubRotationStep next_step = 3;
  string instructions = 4;      // Deletion instructions
}
```

#### ConfirmDeletion

**Request:**
```protobuf
message ConfirmDeletionRequest {
  string state_id = 2;        // Required: Same state ID
}
```

**Response:**
```protobuf
message ConfirmDeletionResponse {
  Status status = 1;
  string state_id = 2;
  int64 completed_at = 3;     // Unix timestamp
  string evidence_id = 4;     // Evidence chain entry (if ACVS)
}
```

---

## Security Considerations

### Token Security

1. **Never log tokens** - Tokens are sensitive credentials
2. **Use mTLS** - All gRPC communication requires mutual TLS
3. **Secure transmission** - Tokens are only transmitted over encrypted channels
4. **Immediate cleanup** - Clear tokens from memory after use
5. **Credential hashing** - Evidence chain uses SHA-256 hashes, not plaintext IDs

### Best Practices

1. **Rotate regularly** - Rotate PATs every 90 days or when compromised
2. **Use fine-grained tokens** - Prefer fine-grained over classic PATs
3. **Minimum permissions** - Grant only necessary scopes
4. **Track rotation state** - Don't lose your state ID mid-rotation
5. **Update password manager** - Immediately save new token
6. **Verify before deleting** - Ensure new token works before deleting old
7. **Enable ACVS** - For compliance-critical environments

### Privacy

- **Credential IDs are hashed** - SHA-256 hash in evidence chain
- **Tokens never stored** - ACM never persists tokens
- **Local-first** - All processing on your device
- **Zero-knowledge** - ACM doesn't access your GitHub account

---

## Examples

### CLI Integration Example

```bash
#!/bin/bash
# rotate-github-token.sh - GitHub PAT rotation helper script

CRED_ID="github-main-pat"
CURRENT_TOKEN=$(pass github.com/token)  # Get from password manager

# Step 1: Start rotation
echo "Starting GitHub PAT rotation..."
STATE_ID=$(acm-cli rotation start-github \
    --credential-id "$CRED_ID" \
    --current-token "$CURRENT_TOKEN" \
    --site github.com | grep "State ID:" | awk '{print $3}')

echo "State ID: $STATE_ID"
echo "Please create a new token at: https://github.com/settings/tokens/new"
read -p "Enter new token: " NEW_TOKEN

# Step 2: Verify new token
echo "Verifying new token..."
acm-cli rotation verify-token \
    --state-id "$STATE_ID" \
    --new-token "$NEW_TOKEN"

# Update password manager
pass insert -e github.com/token <<< "$NEW_TOKEN"

echo "Please delete your old token at: https://github.com/settings/tokens"
read -p "Press enter after deleting old token..."

# Step 3: Confirm deletion
echo "Confirming deletion..."
acm-cli rotation confirm-deletion --state-id "$STATE_ID"

echo "✅ GitHub PAT rotation complete!"
```

### Go Client Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func rotateGitHubPAT(credID, currentToken string) error {
    // Setup mTLS connection
    creds, err := credentials.NewClientTLSFromFile("~/.acm/certs/client.crt", "")
    if err != nil {
        return fmt.Errorf("failed to load credentials: %w", err)
    }

    conn, err := grpc.Dial("localhost:8443", grpc.WithTransportCredentials(creds))
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer conn.Close()

    client := acmv1.NewRotationServiceClient(conn)
    ctx := context.Background()

    // Step 1: Start rotation
    startResp, err := client.StartGitHubRotation(ctx, &acmv1.StartGitHubRotationRequest{
        CredentialId: credID,
        CurrentToken: currentToken,
        Site:         "github.com",
    })
    if err != nil {
        return fmt.Errorf("start rotation failed: %w", err)
    }

    fmt.Printf("Instructions:\n%s\n", startResp.Instructions)

    // Prompt user for new token
    var newToken string
    fmt.Print("Enter new token: ")
    fmt.Scanln(&newToken)

    // Step 2: Verify new token
    verifyResp, err := client.VerifyNewToken(ctx, &acmv1.VerifyNewTokenRequest{
        StateId:  startResp.StateId,
        NewToken: newToken,
    })
    if err != nil {
        return fmt.Errorf("verify token failed: %w", err)
    }

    fmt.Printf("Instructions:\n%s\n", verifyResp.Instructions)

    // Prompt user to confirm deletion
    fmt.Print("Press enter after deleting old token...")
    fmt.Scanln()

    // Step 3: Confirm deletion
    _, err = client.ConfirmDeletion(ctx, &acmv1.ConfirmDeletionRequest{
        StateId: startResp.StateId,
    })
    if err != nil {
        return fmt.Errorf("confirm deletion failed: %w", err)
    }

    log.Println("✅ Rotation complete!")
    return nil
}
```

---

## Changelog

### Version 1.0 (2025-11-17)
- Initial GitHub PAT rotation feature
- Semi-automated workflow with user guidance
- ACVS integration for compliance validation
- Evidence chain recording
- State management and tracking
- Comprehensive error handling

---

## Support

For issues, questions, or feature requests:
- GitHub Issues: https://github.com/ferg-cod3s/automated-compromise-mitigation/issues
- Documentation: https://docs.acm-project.dev/rotation/github
- Security Issues: security@acm-project.dev

---

**End of GitHub PAT Rotation Guide**
