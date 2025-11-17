// Package github provides GitHub Personal Access Token rotation functionality.
package github

import (
	"context"
	"fmt"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/rotation"
)

// Rotator orchestrates GitHub PAT rotation workflows.
// Since GitHub's API doesn't support programmatic PAT creation/deletion,
// this provides a guided, semi-automated workflow.
type Rotator struct {
	client     *Client
	stateStore rotation.StateStore
}

// NewRotator creates a new GitHub PAT rotator.
func NewRotator(stateStore rotation.StateStore) *Rotator {
	return &Rotator{
		client:     NewClient(),
		stateStore: stateStore,
	}
}

// RotationRequest represents a request to rotate a GitHub PAT.
type RotationRequest struct {
	CredentialID string
	CurrentToken string
	Site         string // e.g., "github.com" or custom GitHub Enterprise URL
	Username     string
}

// RotationResult represents the result of a rotation attempt.
type RotationResult struct {
	Success       bool
	State         rotation.RotationState
	NextStep      RotationStep
	Instructions  string
	Error         error
	CompletedAt   time.Time
}

// RotationStep represents a step in the semi-automated rotation process.
type RotationStep string

const (
	StepValidating   RotationStep = "validating"
	StepGuiding      RotationStep = "guiding"
	StepWaitingToken RotationStep = "waiting_for_token"
	StepVerifying    RotationStep = "verifying"
	StepComplete     RotationStep = "complete"
	StepFailed       RotationStep = "failed"
)

// StartRotation initiates a GitHub PAT rotation workflow.
// This performs pre-flight validation and returns instructions for the user.
func (r *Rotator) StartRotation(ctx context.Context, req RotationRequest) (*RotationResult, error) {
	// Validate request
	if req.CredentialID == "" {
		return nil, fmt.Errorf("credential ID is required")
	}
	if req.CurrentToken == "" {
		return nil, fmt.Errorf("current token is required")
	}

	// Step 1: Validate current token
	user, err := r.client.GetUser(ctx, req.CurrentToken)
	if err != nil {
		return &RotationResult{
			Success:  false,
			NextStep: StepFailed,
			Error:    fmt.Errorf("current token validation failed: %w", err),
		}, nil
	}

	// Create rotation state
	state := rotation.RotationState{
		ID:           rotation.GenerateStateID(),
		CredentialID: req.CredentialID,
		Provider:     "github",
		State:        string(StepValidating),
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour timeout
		Metadata: map[string]string{
			"username": user.Login,
			"site":     req.Site,
		},
	}

	// Save initial state
	if err := r.stateStore.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to save rotation state: %w", err)
	}

	// Return instructions for user
	instructions := r.generateCreationInstructions(user.Login, req.Site)

	return &RotationResult{
		Success:      true,
		State:        state,
		NextStep:     StepGuiding,
		Instructions: instructions,
	}, nil
}

// VerifyNewToken verifies a newly created token and completes the rotation.
func (r *Rotator) VerifyNewToken(ctx context.Context, stateID, newToken string) (*RotationResult, error) {
	// Load rotation state
	state, err := r.stateStore.GetState(ctx, stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load rotation state: %w", err)
	}

	// Update state to verifying
	state.State = string(StepVerifying)
	state.UpdatedAt = time.Now()
	if err := r.stateStore.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	// Verify new token works
	user, err := r.client.GetUser(ctx, newToken)
	if err != nil {
		state.State = string(StepFailed)
		state.Metadata["error"] = err.Error()
		r.stateStore.SaveState(ctx, state)

		return &RotationResult{
			Success:  false,
			State:    state,
			NextStep: StepFailed,
			Error:    fmt.Errorf("new token validation failed: %w", err),
		}, nil
	}

	// Verify it's the same user
	expectedUsername := state.Metadata["username"]
	if user.Login != expectedUsername {
		state.State = string(StepFailed)
		state.Metadata["error"] = "token belongs to different user"
		r.stateStore.SaveState(ctx, state)

		return &RotationResult{
			Success:  false,
			State:    state,
			NextStep: StepFailed,
			Error:    fmt.Errorf("token belongs to different user (expected: %s, got: %s)", expectedUsername, user.Login),
		}, nil
	}

	// Generate deletion instructions
	instructions := r.generateDeletionInstructions(state.Metadata["site"])

	// Update state to waiting for deletion
	state.State = "waiting_deletion"
	state.UpdatedAt = time.Now()
	state.Metadata["new_token_verified_at"] = time.Now().Format(time.RFC3339)
	if err := r.stateStore.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	return &RotationResult{
		Success:      true,
		State:        state,
		NextStep:     StepGuiding,
		Instructions: instructions,
	}, nil
}

// ConfirmDeletion confirms that the old token has been deleted and completes rotation.
func (r *Rotator) ConfirmDeletion(ctx context.Context, stateID string) (*RotationResult, error) {
	// Load rotation state
	state, err := r.stateStore.GetState(ctx, stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load rotation state: %w", err)
	}

	// Mark as complete
	state.State = string(StepComplete)
	state.UpdatedAt = time.Now()
	state.Metadata["completed_at"] = time.Now().Format(time.RFC3339)

	if err := r.stateStore.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	// Clean up state after 7 days (keep for audit purposes)
	state.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	r.stateStore.SaveState(ctx, state)

	return &RotationResult{
		Success:     true,
		State:       state,
		NextStep:    StepComplete,
		CompletedAt: time.Now(),
	}, nil
}

// GetRotationStatus gets the current status of a rotation.
func (r *Rotator) GetRotationStatus(ctx context.Context, stateID string) (*RotationResult, error) {
	state, err := r.stateStore.GetState(ctx, stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load rotation state: %w", err)
	}

	var nextStep RotationStep
	switch state.State {
	case string(StepValidating):
		nextStep = StepGuiding
	case "waiting_deletion":
		nextStep = StepGuiding
	case string(StepComplete):
		nextStep = StepComplete
	case string(StepFailed):
		nextStep = StepFailed
	default:
		nextStep = RotationStep(state.State)
	}

	return &RotationResult{
		Success:  state.State == string(StepComplete),
		State:    state,
		NextStep: nextStep,
	}, nil
}

// CancelRotation cancels an in-progress rotation.
func (r *Rotator) CancelRotation(ctx context.Context, stateID string) error {
	state, err := r.stateStore.GetState(ctx, stateID)
	if err != nil {
		return fmt.Errorf("failed to load rotation state: %w", err)
	}

	state.State = "cancelled"
	state.UpdatedAt = time.Now()
	state.Metadata["cancelled_at"] = time.Now().Format(time.RFC3339)

	return r.stateStore.SaveState(ctx, state)
}

// generateCreationInstructions generates step-by-step instructions for creating a new PAT.
func (r *Rotator) generateCreationInstructions(username, site string) string {
	if site == "" {
		site = "github.com"
	}

	return fmt.Sprintf(`GitHub Personal Access Token Rotation Guide

Step 1: Create New Fine-Grained Token
--------------------------------------
1. Go to: https://%s/settings/tokens/new
2. Click "Generate new token" → "Generate new token (fine-grained)"
3. Enter a descriptive name: "ACM Rotated Token - %s"
4. Set expiration: 90 days (or your preference)
5. Select the SAME permissions as your current token:
   - Repository access
   - Permissions (match your current scopes)
6. Click "Generate token"
7. COPY THE TOKEN - you won't see it again!

Step 2: Verify New Token
-------------------------
Once you have the new token, return to ACM and provide it for verification.
ACM will test the token to ensure it works before proceeding.

IMPORTANT: Keep your current token active until ACM confirms the new token works!
`, site, time.Now().Format("2006-01-02"))
}

// generateDeletionInstructions generates instructions for deleting the old token.
func (r *Rotator) generateDeletionInstructions(site string) string {
	if site == "" {
		site = "github.com"
	}

	return fmt.Sprintf(`✅ New Token Verified Successfully!

Step 3: Delete Old Token
-------------------------
Now that your new token is working, you can safely delete the old one:

1. Go to: https://%s/settings/tokens
2. Find your OLD token in the list
3. Click "Delete" next to the old token
4. Confirm deletion

Once deleted, return to ACM to confirm completion.

IMPORTANT: Make sure your password manager is updated with the NEW token
before deleting the old one!
`, site)
}

// ListActiveRotations returns all active (incomplete) rotations.
func (r *Rotator) ListActiveRotations(ctx context.Context) ([]rotation.RotationState, error) {
	return r.stateStore.ListStates(ctx, rotation.StateFilter{
		Provider:     "github",
		ExcludeStates: []string{string(StepComplete), "cancelled"},
	})
}

// CleanupExpiredStates removes expired rotation states.
func (r *Rotator) CleanupExpiredStates(ctx context.Context) (int, error) {
	return r.stateStore.CleanupExpired(ctx)
}
