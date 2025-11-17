// Package github provides GitHub Personal Access Token rotation functionality.
package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvsif"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/rotation"
)

// testEnv holds common test fixtures.
type testEnv struct {
	server     *httptest.Server
	client     *Client
	stateStore *mockStateStore
	acvs       *mockACVSService
	rotator    *Rotator
}

// createTestEnv creates a test environment with mock server and services.
func createTestEnv(acvsEnabled bool) *testEnv {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"login":"testuser","id":12345,"name":"Test User","email":"test@example.com"}`)
		case "/rate_limit":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"rate":{"limit":5000,"remaining":4999,"reset":%d}}`, time.Now().Add(1*time.Hour).Unix())
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client := NewClientWithHTTP(&http.Client{})
	client.baseURL = server.URL

	stateStore := newMockStateStore()

	// Always create mock ACVS, but set enabled flag based on parameter
	acvs := &mockACVSService{
		enabled: acvsEnabled,
		validationResult: &acvsif.ValidationResult{
			Result:    acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED,
			CRCID:     "crc-test-123",
			Reasoning: "Test automation allowed",
		},
		evidenceID: "evidence-123",
	}

	rotator := &Rotator{
		client:     client,
		stateStore: stateStore,
		acvs:       acvs,
	}

	return &testEnv{
		server:     server,
		client:     client,
		stateStore: stateStore,
		acvs:       acvs,
		rotator:    rotator,
	}
}

func (env *testEnv) Close() {
	env.server.Close()
}

// mockACVSService is a mock ACVS service for testing.
type mockACVSService struct {
	enabled          bool
	validationResult *acvsif.ValidationResult
	validationError  error
	evidenceID       string
	evidenceError    error
}

func (m *mockACVSService) IsEnabled() bool {
	return m.enabled
}

func (m *mockACVSService) ValidateAction(ctx context.Context, site string, action *acmv1.AutomationAction) (*acvsif.ValidationResult, error) {
	if m.validationError != nil {
		return nil, m.validationError
	}
	return m.validationResult, nil
}

func (m *mockACVSService) AddEvidenceEntry(ctx context.Context, entry *acvsif.EvidenceEntry) (string, error) {
	if m.evidenceError != nil {
		return "", m.evidenceError
	}
	return m.evidenceID, nil
}

// mockStateStore is a mock state store for testing.
type mockStateStore struct {
	states map[string]rotation.RotationState
}

func newMockStateStore() *mockStateStore {
	return &mockStateStore{
		states: make(map[string]rotation.RotationState),
	}
}

func (m *mockStateStore) SaveState(ctx context.Context, state rotation.RotationState) error {
	m.states[state.ID] = state
	return nil
}

func (m *mockStateStore) GetState(ctx context.Context, id string) (rotation.RotationState, error) {
	state, ok := m.states[id]
	if !ok {
		return rotation.RotationState{}, rotation.ErrStateNotFound
	}
	return state, nil
}

func (m *mockStateStore) ListStates(ctx context.Context, filter rotation.StateFilter) ([]rotation.RotationState, error) {
	var result []rotation.RotationState
	for _, state := range m.states {
		if filter.Provider != "" && state.Provider != filter.Provider {
			continue
		}
		if filter.State != "" && state.State != filter.State {
			continue
		}
		if len(filter.ExcludeStates) > 0 {
			excluded := false
			for _, excludeState := range filter.ExcludeStates {
				if state.State == excludeState {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		result = append(result, state)
	}
	return result, nil
}

func (m *mockStateStore) DeleteState(ctx context.Context, id string) error {
	delete(m.states, id)
	return nil
}

func (m *mockStateStore) CleanupExpired(ctx context.Context) (int, error) {
	count := 0
	now := time.Now()
	for id, state := range m.states {
		if state.ExpiresAt.Before(now) {
			delete(m.states, id)
			count++
		}
	}
	return count, nil
}

// TestStartRotation tests the StartRotation workflow.
func TestStartRotation(t *testing.T) {
	tests := []struct {
		name        string
		req         RotationRequest
		acvsEnabled bool
		wantSuccess bool
	}{
		{
			name: "successful_start_without_acvs",
			req: RotationRequest{
				CredentialID: "cred-123",
				CurrentToken: "ghp_valid_token",
				Site:         "github.com",
			},
			acvsEnabled: false,
			wantSuccess: true,
		},
		{
			name: "successful_start_with_acvs",
			req: RotationRequest{
				CredentialID: "cred-456",
				CurrentToken: "ghp_valid_token",
				Site:         "github.com",
			},
			acvsEnabled: true,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := createTestEnv(tt.acvsEnabled)
			defer env.Close()

			ctx := context.Background()
			result, err := env.rotator.StartRotation(ctx, tt.req)

			if tt.wantSuccess {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if !result.Success {
					t.Fatalf("expected success, got failure: %v", result.Error)
				}
				if result.Instructions == "" {
					t.Error("expected instructions, got empty")
				}
				if result.State.ID == "" {
					t.Error("expected state ID, got empty")
				}
			} else {
				if result != nil && result.Success {
					t.Error("expected failure, got success")
				}
			}
		})
	}
}

// TestFullRotationWorkflow tests the complete rotation workflow.
func TestFullRotationWorkflow(t *testing.T) {
	env := createTestEnv(true)
	defer env.Close()

	ctx := context.Background()

	// Step 1: Start rotation
	startReq := RotationRequest{
		CredentialID: "cred-123",
		CurrentToken: "ghp_old_token",
		Site:         "github.com",
	}

	startResult, err := env.rotator.StartRotation(ctx, startReq)
	if err != nil || !startResult.Success {
		t.Fatalf("StartRotation failed: %v", err)
	}

	stateID := startResult.State.ID
	if stateID == "" {
		t.Fatal("expected state ID, got empty")
	}

	// Step 2: Verify new token
	verifyResult, err := env.rotator.VerifyNewToken(ctx, stateID, "ghp_new_token")
	if err != nil || !verifyResult.Success {
		t.Fatalf("VerifyNewToken failed: %v, result: %v", err, verifyResult)
	}

	// Step 3: Confirm deletion
	confirmResult, err := env.rotator.ConfirmDeletion(ctx, stateID)
	if err != nil || !confirmResult.Success {
		t.Fatalf("ConfirmDeletion failed: %v", err)
	}

	if confirmResult.NextStep != StepComplete {
		t.Errorf("expected next step %v, got %v", StepComplete, confirmResult.NextStep)
	}

	if confirmResult.CompletedAt.IsZero() {
		t.Error("expected completion timestamp, got zero")
	}

	// Verify final state
	state, err := env.stateStore.GetState(ctx, stateID)
	if err != nil {
		t.Fatalf("failed to get state: %v", err)
	}

	if state.State != string(StepComplete) {
		t.Errorf("expected state %v, got %v", StepComplete, state.State)
	}
}

// TestListActiveRotations tests listing active rotations.
func TestListActiveRotations(t *testing.T) {
	env := createTestEnv(false)
	defer env.Close()

	ctx := context.Background()

	// Start multiple rotations
	for i := 0; i < 3; i++ {
		req := RotationRequest{
			CredentialID: fmt.Sprintf("cred-%d", i),
			CurrentToken: "ghp_token",
			Site:         "github.com",
		}
		_, err := env.rotator.StartRotation(ctx, req)
		if err != nil {
			t.Fatalf("failed to start rotation %d: %v", i, err)
		}
	}

	// List active rotations
	states, err := env.rotator.ListActiveRotations(ctx)
	if err != nil {
		t.Fatalf("ListActiveRotations failed: %v", err)
	}

	if len(states) != 3 {
		t.Errorf("expected 3 active rotations, got %d", len(states))
	}
}

// TestCancelRotation tests canceling a rotation.
func TestCancelRotation(t *testing.T) {
	env := createTestEnv(false)
	defer env.Close()

	ctx := context.Background()

	// Start rotation
	req := RotationRequest{
		CredentialID: "cred-123",
		CurrentToken: "ghp_token",
		Site:         "github.com",
	}
	result, err := env.rotator.StartRotation(ctx, req)
	if err != nil || !result.Success {
		t.Fatalf("StartRotation failed: %v", err)
	}

	stateID := result.State.ID

	// Cancel rotation
	err = env.rotator.CancelRotation(ctx, stateID)
	if err != nil {
		t.Fatalf("CancelRotation failed: %v", err)
	}

	// Verify state is cancelled
	state, err := env.stateStore.GetState(ctx, stateID)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	if state.State != "cancelled" {
		t.Errorf("expected state 'cancelled', got %v", state.State)
	}
}

// TestACVSBlocked tests that rotation is blocked when ACVS returns BLOCKED.
func TestACVSBlocked(t *testing.T) {
	env := createTestEnv(true)
	defer env.Close()

	// Override ACVS to return BLOCKED
	env.acvs.validationResult.Result = acmv1.ValidationResult_VALIDATION_RESULT_BLOCKED
	env.acvs.validationResult.Reasoning = "ToS prohibits automated token rotation"

	ctx := context.Background()
	req := RotationRequest{
		CredentialID: "cred-123",
		CurrentToken: "ghp_token",
		Site:         "github.com",
	}

	result, err := env.rotator.StartRotation(ctx, req)

	// Should fail due to ACVS block
	if result != nil && result.Success {
		t.Error("expected rotation to be blocked by ACVS, got success")
	}

	if err == nil && (result == nil || result.Error == nil) {
		t.Error("expected error due to ACVS block")
	}
}

// TestGetRotationStatus tests querying rotation status.
func TestGetRotationStatus(t *testing.T) {
	env := createTestEnv(false)
	defer env.Close()

	ctx := context.Background()

	// Start rotation
	req := RotationRequest{
		CredentialID: "cred-123",
		CurrentToken: "ghp_token",
		Site:         "github.com",
	}
	startResult, err := env.rotator.StartRotation(ctx, req)
	if err != nil || !startResult.Success {
		t.Fatalf("StartRotation failed: %v", err)
	}

	stateID := startResult.State.ID

	// Get status
	statusResult, err := env.rotator.GetRotationStatus(ctx, stateID)
	if err != nil {
		t.Fatalf("GetRotationStatus failed: %v", err)
	}

	if statusResult.State.ID != stateID {
		t.Errorf("expected state ID %v, got %v", stateID, statusResult.State.ID)
	}
}
