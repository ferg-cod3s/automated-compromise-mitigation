// Package crs implements the Credential Remediation Service.
package crs

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/audit"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// Service implements the CredentialRemediationService interface.
type Service struct {
	pwManager     pwmanager.PasswordManager
	auditLogger   audit.Logger
	defaultPolicy pwmanager.PasswordPolicy
}

// NewService creates a new CRS instance with the specified password manager and audit logger.
func NewService(pm pwmanager.PasswordManager, auditer audit.Logger) *Service {
	return &Service{
		pwManager:     pm,
		auditLogger:   auditer,
		defaultPolicy: pwmanager.DefaultPasswordPolicy(),
	}
}

// DetectCompromised queries the password manager for credentials exposed in breaches.
func (s *Service) DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error) {
	creds, err := s.pwManager.DetectCompromised(ctx)
	if err != nil {
		// Log detection failure
		_ = s.auditLogger.LogEvent(ctx, audit.Event{
			Type:      audit.EventTypeDetection,
			Status:    audit.StatusFailure,
			Message:   fmt.Sprintf("Detection failed: %v", err),
			Timestamp: time.Now(),
		})
		return nil, err
	}

	// Log successful detection
	_ = s.auditLogger.LogEvent(ctx, audit.Event{
		Type:      audit.EventTypeDetection,
		Status:    audit.StatusSuccess,
		Message:   fmt.Sprintf("Detected %d compromised credentials", len(creds)),
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"count":           fmt.Sprintf("%d", len(creds)),
			"password_manager": s.pwManager.Type(),
		},
	})

	return creds, nil
}

// GeneratePassword creates a secure password using crypto/rand.
func (s *Service) GeneratePassword(ctx context.Context, policy pwmanager.PasswordPolicy) (string, error) {
	if policy.Length == 0 {
		policy = s.defaultPolicy
	}

	// Validate policy
	if policy.Length < 12 {
		return "", &RotationError{
			Code:      ErrPasswordGenerationFailed,
			Message:   "Password length must be at least 12 characters",
			Retryable: false,
		}
	}

	if policy.Length > 128 {
		return "", &RotationError{
			Code:      ErrPasswordGenerationFailed,
			Message:   "Password length must be at most 128 characters",
			Retryable: false,
		}
	}

	// Build character set
	charset := ""
	if policy.CustomCharset != "" {
		charset = policy.CustomCharset
	} else {
		if policy.RequireLowercase {
			charset += "abcdefghijklmnopqrstuvwxyz"
		}
		if policy.RequireUppercase {
			charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		}
		if policy.RequireNumbers {
			charset += "0123456789"
		}
		if policy.RequireSymbols {
			charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
		}

		if policy.ExcludeAmbiguous {
			// Remove ambiguous characters: 0, O, o, l, 1, I
			charset = removeChars(charset, "0Ool1I")
		}
	}

	if charset == "" {
		return "", &RotationError{
			Code:      ErrPasswordGenerationFailed,
			Message:   "Password policy must allow at least one character type",
			Retryable: false,
		}
	}

	// Generate password using crypto/rand
	password := make([]byte, policy.Length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < policy.Length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", &RotationError{
				Code:      ErrPasswordGenerationFailed,
				Message:   "Failed to generate secure random number",
				Cause:     err,
				Retryable: true,
			}
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}

// RotateCredential performs the complete rotation workflow for a single credential.
func (s *Service) RotateCredential(ctx context.Context, cred pwmanager.CompromisedCredential, newPassword string) (*RotationResult, error) {
	startTime := time.Now()

	result := &RotationResult{
		CredentialID: hashCredentialID(cred.ID),
		Status:       RotationPending,
		StartTime:    startTime,
	}

	// Step 1: Validate inputs
	if newPassword == "" {
		result.Status = RotationFailure
		result.Error = &RotationError{
			Code:    ErrPasswordGenerationFailed,
			Message: "New password cannot be empty",
		}
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, result.Error
	}

	// Step 2: Update vault via password manager CLI
	if err := s.pwManager.UpdatePassword(ctx, cred.ID, newPassword); err != nil {
		result.Status = RotationFailure
		result.Error = &RotationError{
			Code:    ErrUpdateFailed,
			Message: fmt.Sprintf("Failed to update password in vault: %v", err),
			Cause:   err,
		}

		// Check if HIM is required
		if pmErr, ok := err.(*pwmanager.PasswordManagerError); ok {
			if pmErr.Code == pwmanager.ErrVaultLocked {
				result.Status = RotationHIMRequired
				result.Error.Code = ErrHIMRequired
				result.Error.HIMType = HIMMFA
			}
		}

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)

		// Log failure
		_ = s.auditLogger.LogEvent(ctx, audit.Event{
			Type:         audit.EventTypeRotation,
			Status:       audit.StatusFailure,
			CredentialID: result.CredentialID,
			Site:         cred.Site,
			Message:      result.Error.Message,
			Timestamp:    time.Now(),
			Metadata: map[string]string{
				"error_code": string(result.Error.Code),
			},
		})

		return result, result.Error
	}

	result.NewPasswordSet = true

	// Step 3: Verify update success
	verified, err := s.VerifyRotation(ctx, cred.ID)
	if err != nil || !verified {
		result.Status = RotationFailure
		result.Error = &RotationError{
			Code:    ErrVerificationFailed,
			Message: "Failed to verify password update",
			Cause:   err,
		}
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)

		// Log verification failure
		_ = s.auditLogger.LogEvent(ctx, audit.Event{
			Type:         audit.EventTypeRotation,
			Status:       audit.StatusFailure,
			CredentialID: result.CredentialID,
			Site:         cred.Site,
			Message:      "Verification failed",
			Timestamp:    time.Now(),
		})

		return result, result.Error
	}

	// Step 4: Log rotation event to audit trail
	auditEvent := audit.Event{
		Type:         audit.EventTypeRotation,
		Status:       audit.StatusSuccess,
		CredentialID: result.CredentialID,
		Site:         cred.Site,
		Username:     cred.Username,
		Message:      "Password rotated successfully",
		Timestamp:    time.Now(),
		Metadata: map[string]string{
			"password_manager": s.pwManager.Type(),
			"breach_name":      cred.BreachName,
		},
	}

	if err := s.auditLogger.LogEvent(ctx, auditEvent); err != nil {
		// Log event creation failed, but rotation was successful
		// Don't fail the rotation, just log a warning
		result.Status = RotationSuccess
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, nil
	}

	result.AuditEventID = auditEvent.ID
	result.Status = RotationSuccess
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	return result, nil
}

// VerifyRotation confirms that a credential was successfully rotated.
func (s *Service) VerifyRotation(ctx context.Context, credentialID string) (bool, error) {
	// Get the credential's current modification time
	cred, err := s.pwManager.GetCredential(ctx, credentialID)
	if err != nil {
		return false, err
	}

	// Check if it was modified within the last minute
	return time.Since(cred.LastModified) < time.Minute, nil
}

// GetRotationHistory returns the rotation history for a specific credential.
func (s *Service) GetRotationHistory(ctx context.Context, credentialID string) ([]RotationEvent, error) {
	hashedID := hashCredentialID(credentialID)

	// Query audit log for rotation events for this credential
	events, err := s.auditLogger.QueryEvents(ctx, audit.Filter{
		CredentialID: hashedID,
		EventType:    audit.EventTypeRotation,
	})
	if err != nil {
		return nil, err
	}

	// Convert audit events to rotation events
	var history []RotationEvent
	for _, event := range events {
		rotEvent := RotationEvent{
			EventID:      event.ID,
			CredentialID: event.CredentialID,
			Timestamp:    event.Timestamp,
			InitiatedBy:  "user", // TODO: Track who initiated
			Method:       MethodAuto,
		}

		if event.Status == audit.StatusSuccess {
			rotEvent.Status = RotationSuccess
		} else {
			rotEvent.Status = RotationFailure
		}

		// Parse duration from metadata if available
		if durationStr, ok := event.Metadata["duration"]; ok {
			if d, err := time.ParseDuration(durationStr); err == nil {
				rotEvent.Duration = d
			}
		}

		history = append(history, rotEvent)
	}

	return history, nil
}

// hashCredentialID creates a SHA-256 hash of the credential ID for privacy.
func hashCredentialID(id string) string {
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:])
}

// removeChars removes all occurrences of chars from s.
func removeChars(s string, chars string) string {
	result := ""
	for _, c := range s {
		if !containsChar(chars, c) {
			result += string(c)
		}
	}
	return result
}

// containsChar checks if s contains rune c.
func containsChar(s string, c rune) bool {
	for _, char := range s {
		if char == c {
			return true
		}
	}
	return false
}
