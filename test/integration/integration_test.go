package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/audit"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/crs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// MockPasswordManager implements pwmanager.PasswordManager for testing
type MockPasswordManager struct {
	credentials []pwmanager.CompromisedCredential
	locked      bool
	updateCalls int
}

func NewMockPasswordManager() *MockPasswordManager {
	return &MockPasswordManager{
		credentials: []pwmanager.CompromisedCredential{
			{
				ID:         "test-cred-1",
				Site:       "example.com",
				Username:   "user@example.com",
				BreachName: "Example Breach 2024",
				BreachDate: time.Now().Add(-30 * 24 * time.Hour),
			},
			{
				ID:         "test-cred-2",
				Site:       "testsite.com",
				Username:   "testuser@test.com",
				BreachName: "Test Breach 2023",
				BreachDate: time.Now().Add(-90 * 24 * time.Hour),
			},
		},
		locked: false,
	}
}

func (m *MockPasswordManager) DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error) {
	if m.locked {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Vault is locked",
			Retryable: true,
		}
	}
	return m.credentials, nil
}

func (m *MockPasswordManager) GetCredential(ctx context.Context, id string) (*pwmanager.Credential, error) {
	if m.locked {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Vault is locked",
			Retryable: true,
		}
	}

	for _, cred := range m.credentials {
		if cred.ID == id {
			return &pwmanager.Credential{
				ID:           cred.ID,
				Site:         cred.Site,
				Username:     cred.Username,
				LastModified: time.Now(), // Simulate recent update
			}, nil
		}
	}

	return nil, &pwmanager.PasswordManagerError{
		Code:    pwmanager.ErrCredentialNotFound,
		Message: "Credential not found",
	}
}

func (m *MockPasswordManager) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	if m.locked {
		return &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Vault is locked",
			Retryable: true,
		}
	}

	m.updateCalls++
	return nil
}

func (m *MockPasswordManager) VerifyUpdate(ctx context.Context, id string, expectedModifiedAfter time.Time) (bool, error) {
	return true, nil
}

func (m *MockPasswordManager) ListCredentials(ctx context.Context) ([]*pwmanager.Credential, error) {
	var creds []*pwmanager.Credential
	for _, c := range m.credentials {
		creds = append(creds, &pwmanager.Credential{
			ID:       c.ID,
			Site:     c.Site,
			Username: c.Username,
		})
	}
	return creds, nil
}

func (m *MockPasswordManager) IsVaultLocked(ctx context.Context) (bool, error) {
	return m.locked, nil
}

func (m *MockPasswordManager) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *MockPasswordManager) Type() string {
	return "mock"
}

// TestEndToEndRotationWorkflow tests the complete credential rotation workflow
func TestEndToEndRotationWorkflow(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockPM := NewMockPasswordManager()

	logger, err := audit.NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	service := crs.NewService(mockPM, logger)

	// Step 1: Detect compromised credentials
	t.Log("Step 1: Detecting compromised credentials...")
	compromised, err := service.DetectCompromised(ctx)
	if err != nil {
		t.Fatalf("Failed to detect compromised credentials: %v", err)
	}

	if len(compromised) != 2 {
		t.Fatalf("Expected 2 compromised credentials, got %d", len(compromised))
	}

	t.Logf("Found %d compromised credentials", len(compromised))

	// Step 2: Generate a new password
	t.Log("Step 2: Generating new password...")
	policy := pwmanager.PasswordPolicy{
		Length:           16,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
	}

	newPassword, err := service.GeneratePassword(ctx, policy)
	if err != nil {
		t.Fatalf("Failed to generate password: %v", err)
	}

	if len(newPassword) != 16 {
		t.Errorf("Expected password length 16, got %d", len(newPassword))
	}

	t.Logf("Generated password: %s", newPassword)

	// Step 3: Rotate the first credential
	t.Log("Step 3: Rotating first credential...")
	result, err := service.RotateCredential(ctx, compromised[0], newPassword)
	if err != nil {
		t.Fatalf("Failed to rotate credential: %v", err)
	}

	if result.Status != crs.RotationSuccess {
		t.Errorf("Expected rotation success, got %s: %v", result.Status, result.Error)
	}

	if mockPM.updateCalls != 1 {
		t.Errorf("Expected 1 password manager update call, got %d", mockPM.updateCalls)
	}

	t.Logf("Rotation completed in %v", result.Duration)

	// Step 4: Verify audit log entry was created
	t.Log("Step 4: Verifying audit log...")
	events, err := logger.QueryEvents(ctx, audit.Filter{
		EventType: audit.EventTypeRotation,
		Status:    audit.StatusSuccess,
	})
	if err != nil {
		t.Fatalf("Failed to query audit events: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected audit log entry for rotation, got none")
	}

	t.Logf("Audit log verified: %d rotation events logged", len(events))

	// Step 5: Verify event signature
	t.Log("Step 5: Verifying cryptographic signature...")
	if len(events) > 0 {
		valid, err := logger.VerifyIntegrity(ctx, events[0].ID)
		if err != nil {
			t.Fatalf("Failed to verify signature: %v", err)
		}
		if !valid {
			t.Error("Event signature verification failed")
		}
		t.Log("Signature verified successfully")
	}
}

// TestRotationWithLockedVault tests handling of locked vault scenario
func TestRotationWithLockedVault(t *testing.T) {
	ctx := context.Background()
	mockPM := NewMockPasswordManager()
	mockPM.locked = true

	logger, err := audit.NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	service := crs.NewService(mockPM, logger)

	// Attempt detection with locked vault
	_, err = service.DetectCompromised(ctx)
	if err == nil {
		t.Error("Expected error for locked vault, got nil")
	}

	pmErr, ok := err.(*pwmanager.PasswordManagerError)
	if !ok {
		t.Errorf("Expected PasswordManagerError, got %T", err)
	} else if pmErr.Code != pwmanager.ErrVaultLocked {
		t.Errorf("Expected ErrVaultLocked, got %s", pmErr.Code)
	}

	// Verify failure was logged
	events, err := logger.QueryEvents(ctx, audit.Filter{
		EventType: audit.EventTypeDetection,
		Status:    audit.StatusFailure,
	})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected audit log entry for failed detection")
	}
}

// TestMultipleRotations tests rotating multiple credentials
func TestMultipleRotations(t *testing.T) {
	ctx := context.Background()
	mockPM := NewMockPasswordManager()

	logger, err := audit.NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	service := crs.NewService(mockPM, logger)

	// Detect compromised credentials
	compromised, err := service.DetectCompromised(ctx)
	if err != nil {
		t.Fatalf("Failed to detect: %v", err)
	}

	// Rotate all credentials
	policy := pwmanager.DefaultPasswordPolicy()
	successCount := 0

	for _, cred := range compromised {
		newPassword, err := service.GeneratePassword(ctx, policy)
		if err != nil {
			t.Errorf("Failed to generate password for %s: %v", cred.ID, err)
			continue
		}

		result, err := service.RotateCredential(ctx, cred, newPassword)
		if err != nil {
			t.Errorf("Failed to rotate %s: %v", cred.ID, err)
			continue
		}

		if result.Status == crs.RotationSuccess {
			successCount++
		}
	}

	if successCount != len(compromised) {
		t.Errorf("Expected %d successful rotations, got %d", len(compromised), successCount)
	}

	if mockPM.updateCalls != len(compromised) {
		t.Errorf("Expected %d update calls, got %d", len(compromised), mockPM.updateCalls)
	}

	// Verify all rotations were logged
	events, err := logger.QueryEvents(ctx, audit.Filter{
		EventType: audit.EventTypeRotation,
	})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) != len(compromised) {
		t.Errorf("Expected %d audit events, got %d", len(compromised), len(events))
	}
}

// TestAuditReportGeneration tests end-to-end audit report generation
func TestAuditReportGeneration(t *testing.T) {
	ctx := context.Background()
	mockPM := NewMockPasswordManager()

	logger, err := audit.NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	service := crs.NewService(mockPM, logger)

	// Perform some rotations
	compromised, _ := service.DetectCompromised(ctx)
	policy := pwmanager.DefaultPasswordPolicy()

	for _, cred := range compromised {
		newPassword, _ := service.GeneratePassword(ctx, policy)
		service.RotateCredential(ctx, cred, newPassword)
	}

	// Generate JSON report
	jsonReport, err := logger.ExportReport(ctx, audit.Filter{}, audit.ReportFormatJSON)
	if err != nil {
		t.Fatalf("Failed to generate JSON report: %v", err)
	}

	if len(jsonReport) == 0 {
		t.Error("JSON report is empty")
	}

	// Generate CSV report
	csvReport, err := logger.ExportReport(ctx, audit.Filter{}, audit.ReportFormatCSV)
	if err != nil {
		t.Fatalf("Failed to generate CSV report: %v", err)
	}

	if len(csvReport) == 0 {
		t.Error("CSV report is empty")
	}

	t.Logf("Generated reports: JSON=%d bytes, CSV=%d bytes", len(jsonReport), len(csvReport))
}

// TestPasswordPolicyEnforcement tests that different policies are enforced
func TestPasswordPolicyEnforcement(t *testing.T) {
	ctx := context.Background()
	service := crs.NewService(nil, nil)

	policies := []struct {
		name   string
		policy pwmanager.PasswordPolicy
	}{
		{
			name: "strict policy",
			policy: pwmanager.PasswordPolicy{
				Length:           32,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
			},
		},
		{
			name: "moderate policy",
			policy: pwmanager.PasswordPolicy{
				Length:           16,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
			},
		},
		{
			name: "basic policy",
			policy: pwmanager.PasswordPolicy{
				Length:           12,
				RequireUppercase: false,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
			},
		},
	}

	for _, tt := range policies {
		t.Run(tt.name, func(t *testing.T) {
			password, err := service.GeneratePassword(ctx, tt.policy)
			if err != nil {
				t.Fatalf("Failed to generate password: %v", err)
			}

			if len(password) != tt.policy.Length {
				t.Errorf("Expected length %d, got %d", tt.policy.Length, len(password))
			}

			// Verify policy requirements
			hasUpper, hasLower, hasNumber, hasSymbol := false, false, false, false
			for _, char := range password {
				if char >= 'A' && char <= 'Z' {
					hasUpper = true
				} else if char >= 'a' && char <= 'z' {
					hasLower = true
				} else if char >= '0' && char <= '9' {
					hasNumber = true
				} else {
					hasSymbol = true
				}
			}

			if tt.policy.RequireUppercase && !hasUpper {
				t.Error("Policy requires uppercase but none found")
			}
			if tt.policy.RequireLowercase && !hasLower {
				t.Error("Policy requires lowercase but none found")
			}
			if tt.policy.RequireNumbers && !hasNumber {
				t.Error("Policy requires numbers but none found")
			}
			if tt.policy.RequireSymbols && !hasSymbol {
				t.Error("Policy requires symbols but none found")
			}
		})
	}
}

// TestMain can be used for setup/teardown
func TestMain(m *testing.M) {
	// Setup
	home := os.Getenv("HOME")
	if home == "" {
		home = "/tmp"
	}
	testDir := filepath.Join(home, ".acm-test")
	os.MkdirAll(testDir, 0700)

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testDir)

	os.Exit(code)
}
