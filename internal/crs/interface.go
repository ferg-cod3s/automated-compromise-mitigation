// Package crs implements the Credential Remediation Service (CRS).
//
// The CRS is responsible for detecting compromised credentials and performing
// safe, local vault updates. It maintains zero-knowledge security by never
// accessing master passwords or vault encryption keys.
package crs

import (
	"context"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// CredentialRemediationService is the core interface for credential rotation operations.
// All operations maintain zero-knowledge security and local-first principles.
type CredentialRemediationService interface {
	// DetectCompromised queries the password manager for credentials exposed in breaches.
	// Returns a list of compromised credentials that require rotation.
	DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error)

	// GeneratePassword creates a secure password using crypto/rand.
	// The password adheres to the provided policy constraints.
	GeneratePassword(ctx context.Context, policy pwmanager.PasswordPolicy) (string, error)

	// RotateCredential performs the complete rotation workflow for a single credential:
	//   1. Generate new password
	//   2. Update vault via password manager CLI
	//   3. Verify update success
	//   4. Log rotation event to audit trail
	RotateCredential(ctx context.Context, cred pwmanager.CompromisedCredential, newPassword string) (*RotationResult, error)

	// VerifyRotation confirms that a credential was successfully rotated by checking
	// the vault state and last modified timestamp.
	VerifyRotation(ctx context.Context, credentialID string) (bool, error)

	// GetRotationHistory returns the rotation history for a specific credential.
	GetRotationHistory(ctx context.Context, credentialID string) ([]RotationEvent, error)
}

// RotationResult represents the outcome of a credential rotation operation.
type RotationResult struct {
	// CredentialID is the ID of the rotated credential.
	CredentialID string

	// Status indicates the rotation outcome (success, failure, HIM required, etc.).
	Status RotationStatus

	// NewPasswordSet indicates whether the new password was successfully set.
	NewPasswordSet bool

	// Error contains error information if the rotation failed.
	Error *RotationError

	// StartTime is when the rotation began.
	StartTime time.Time

	// EndTime is when the rotation completed.
	EndTime time.Time

	// Duration is the total time taken for the rotation.
	Duration time.Duration

	// AuditEventID is the ID of the audit log entry for this rotation.
	AuditEventID string

	// ComplianceValidation contains ACVS validation results (if enabled).
	ComplianceValidation *ComplianceValidation
}

// RotationStatus indicates the outcome of a rotation operation.
type RotationStatus string

const (
	// RotationSuccess indicates the credential was successfully rotated.
	RotationSuccess RotationStatus = "success"

	// RotationFailure indicates the rotation failed.
	RotationFailure RotationStatus = "failure"

	// RotationHIMRequired indicates Human-in-the-Middle intervention is needed.
	RotationHIMRequired RotationStatus = "him_required"

	// RotationPending indicates the rotation is in progress.
	RotationPending RotationStatus = "pending"

	// RotationSkipped indicates the rotation was skipped by user request.
	RotationSkipped RotationStatus = "skipped"
)

// RotationError represents an error during credential rotation.
type RotationError struct {
	// Code is the error code (e.g., VAULT_LOCKED, HIM_REQUIRED).
	Code RotationErrorCode

	// Message is a human-readable error message.
	Message string

	// Cause is the underlying error, if any.
	Cause error

	// Retryable indicates if the operation can be retried.
	Retryable bool

	// HIMType indicates the type of HIM required (if code is HIM_REQUIRED).
	HIMType HIMType
}

func (e *RotationError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *RotationError) Unwrap() error {
	return e.Cause
}

// RotationErrorCode represents specific rotation error conditions.
type RotationErrorCode string

const (
	// ErrVaultLocked indicates the password manager vault is locked.
	ErrVaultLocked RotationErrorCode = "VAULT_LOCKED"

	// ErrCLINotFound indicates the password manager CLI is not installed.
	ErrCLINotFound RotationErrorCode = "CLI_NOT_FOUND"

	// ErrPasswordManagerUnavailable indicates no password manager is configured.
	ErrPasswordManagerUnavailable RotationErrorCode = "PASSWORD_MANAGER_UNAVAILABLE"

	// ErrNetworkRequired indicates network access is required.
	ErrNetworkRequired RotationErrorCode = "NETWORK_REQUIRED"

	// ErrHIMRequired indicates Human-in-the-Middle intervention is required.
	ErrHIMRequired RotationErrorCode = "HIM_REQUIRED"

	// ErrComplianceViolation indicates the rotation would violate ToS (ACVS blocked it).
	ErrComplianceViolation RotationErrorCode = "COMPLIANCE_VIOLATION"

	// ErrPasswordGenerationFailed indicates secure password generation failed.
	ErrPasswordGenerationFailed RotationErrorCode = "PASSWORD_GENERATION_FAILED"

	// ErrUpdateFailed indicates the vault update operation failed.
	ErrUpdateFailed RotationErrorCode = "UPDATE_FAILED"

	// ErrVerificationFailed indicates post-rotation verification failed.
	ErrVerificationFailed RotationErrorCode = "VERIFICATION_FAILED"
)

// HIMType indicates the type of Human-in-the-Middle intervention required.
type HIMType string

const (
	// HIMMFA indicates multi-factor authentication is required.
	HIMMFA HIMType = "mfa"

	// HIMCAPTCHA indicates CAPTCHA solving is required.
	HIMCAPTCHA HIMType = "captcha"

	// HIMManualRotation indicates manual rotation is required (no API available).
	HIMManualRotation HIMType = "manual_rotation"

	// HIMToSReview indicates Terms of Service review is required.
	HIMToSReview HIMType = "tos_review"
)

// RotationEvent represents a single rotation event in the history.
type RotationEvent struct {
	// EventID is the unique identifier for this event.
	EventID string

	// CredentialID is the ID of the credential (hashed for privacy).
	CredentialID string

	// Timestamp is when the rotation occurred.
	Timestamp time.Time

	// Status is the outcome of the rotation.
	Status RotationStatus

	// Method indicates how the rotation was performed (auto, manual, API).
	Method RotationMethod

	// InitiatedBy indicates who initiated the rotation (user, scheduled_task, etc.).
	InitiatedBy string

	// Duration is how long the rotation took.
	Duration time.Duration

	// ComplianceValidated indicates if ACVS validation was performed.
	ComplianceValidated bool
}

// RotationMethod indicates how a rotation was performed.
type RotationMethod string

const (
	// MethodAuto indicates automatic rotation via API or automation.
	MethodAuto RotationMethod = "auto"

	// MethodHIM indicates rotation with Human-in-the-Middle assistance.
	MethodHIM RotationMethod = "him"

	// MethodManual indicates fully manual rotation by the user.
	MethodManual RotationMethod = "manual"
)

// ComplianceValidation contains ACVS validation results for a rotation.
// This is only populated if ACVS is enabled.
type ComplianceValidation struct {
	// Enabled indicates if ACVS was enabled for this rotation.
	Enabled bool

	// CRCVersion is the version of the Compliance Rule Set used.
	CRCVersion string

	// Result is the validation result (allowed, blocked, him_required).
	Result ValidationResult

	// AppliedRules lists the CRC rules that were evaluated.
	AppliedRules []string

	// Reasoning explains why the validation resulted in this outcome.
	Reasoning string

	// EvidenceChainID links to the cryptographic evidence chain entry.
	EvidenceChainID string
}

// ValidationResult represents the outcome of ACVS compliance validation.
type ValidationResult string

const (
	// ValidationAllowed indicates automation is allowed by ToS.
	ValidationAllowed ValidationResult = "allowed"

	// ValidationBlocked indicates automation is prohibited by ToS.
	ValidationBlocked ValidationResult = "blocked"

	// ValidationHIMRequired indicates HIM workflow is required for compliance.
	ValidationHIMRequired ValidationResult = "him_required"

	// ValidationUnknown indicates ToS could not be parsed or is ambiguous.
	ValidationUnknown ValidationResult = "unknown"
)
