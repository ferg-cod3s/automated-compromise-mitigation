// Package audit implements the cryptographically signed audit logging system.
//
// The Audit Logger provides tamper-evident logging of all credential rotation
// operations, compliance checks, and HIM events. All entries are signed with
// Ed25519 signatures and stored in a local SQLite database.
package audit

import (
	"context"
	"io"
	"time"
)

// AuditLogger manages the cryptographically signed audit trail.
type AuditLogger interface {
	// Log records an audit event with cryptographic signature.
	// Returns the event ID and signature for verification.
	Log(ctx context.Context, event AuditEvent) (eventID string, signature string, error error)

	// Query retrieves audit events matching the provided filter.
	Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)

	// ExportReport generates a compliance report in the specified format.
	// Returns an io.Reader for streaming the report data.
	ExportReport(ctx context.Context, format ReportFormat, filter AuditFilter) (io.Reader, error)

	// VerifyIntegrity verifies the cryptographic signatures of all audit entries
	// within the specified time range. Returns true if all signatures are valid.
	VerifyIntegrity(ctx context.Context, from, to time.Time) (valid bool, errors []string, err error)

	// GetStatistics returns aggregate statistics for the audit log.
	GetStatistics(ctx context.Context, filter AuditFilter) (*AuditStatistics, error)

	// Cleanup removes audit entries older than the retention period.
	// Returns the number of entries deleted.
	Cleanup(ctx context.Context, retentionDays int) (int, error)
}

// AuditEvent represents a single event in the audit trail.
type AuditEvent struct {
	// ID is the unique identifier for this event (auto-generated).
	ID string

	// Timestamp is when the event occurred (ISO 8601 format).
	Timestamp time.Time

	// EventType categorizes the event (rotation, detection, compliance_check, etc.).
	EventType EventType

	// CredentialIDHash is the SHA-256 hash of the credential ID.
	// The credential ID itself is NEVER stored in plaintext for privacy.
	CredentialIDHash string

	// Action describes what action was performed (detected, rotated, skipped, failed).
	Action Action

	// Status indicates the outcome (success, failure, pending).
	Status Status

	// Details contains additional context as a JSON blob.
	// Sensitive data in this field should be encrypted.
	Details map[string]interface{}

	// CRCRuleApplied references the Compliance Rule Set rule ID (if ACVS enabled).
	CRCRuleApplied string

	// EvidenceChainID links to the evidence chain entry (if ACVS enabled).
	EvidenceChainID string

	// Signature is the Ed25519 signature of this event.
	// Signature is computed over: ID || Timestamp || CredentialIDHash || Action
	Signature string

	// CreatedAt is the Unix timestamp when this entry was created.
	CreatedAt int64
}

// EventType categorizes audit events.
type EventType string

const (
	// EventRotation indicates a credential rotation operation.
	EventRotation EventType = "rotation"

	// EventDetection indicates a compromised credential detection operation.
	EventDetection EventType = "detection"

	// EventComplianceCheck indicates an ACVS compliance validation.
	EventComplianceCheck EventType = "compliance_check"

	// EventHIMPrompt indicates a Human-in-the-Middle prompt was issued.
	EventHIMPrompt EventType = "him_prompt"

	// EventHIMResponse indicates a user responded to a HIM prompt.
	EventHIMResponse EventType = "him_response"

	// EventConfiguration indicates a configuration change.
	EventConfiguration EventType = "configuration"

	// EventCertificateRenewal indicates a client certificate was renewed.
	EventCertificateRenewal EventType = "certificate_renewal"

	// EventServiceStartup indicates the ACM service started.
	EventServiceStartup EventType = "service_startup"

	// EventServiceShutdown indicates the ACM service shut down.
	EventServiceShutdown EventType = "service_shutdown"
)

// Action describes what action was performed.
type Action string

const (
	// ActionDetected indicates a compromised credential was detected.
	ActionDetected Action = "detected"

	// ActionRotated indicates a credential was rotated.
	ActionRotated Action = "rotated"

	// ActionSkipped indicates a rotation was skipped by the user.
	ActionSkipped Action = "skipped"

	// ActionFailed indicates an operation failed.
	ActionFailed Action = "failed"

	// ActionValidated indicates a compliance check was performed.
	ActionValidated Action = "validated"

	// ActionPrompted indicates a HIM prompt was issued.
	ActionPrompted Action = "prompted"

	// ActionResponded indicates a user responded to a prompt.
	ActionResponded Action = "responded"

	// ActionUpdated indicates a configuration was updated.
	ActionUpdated Action = "updated"
)

// Status indicates the outcome of an operation.
type Status string

const (
	// StatusSuccess indicates the operation completed successfully.
	StatusSuccess Status = "success"

	// StatusFailure indicates the operation failed.
	StatusFailure Status = "failure"

	// StatusPending indicates the operation is in progress.
	StatusPending Status = "pending"

	// StatusCancelled indicates the operation was cancelled.
	StatusCancelled Status = "cancelled"
)

// AuditFilter specifies criteria for querying audit events.
type AuditFilter struct {
	// FromTime filters events after this time (inclusive).
	FromTime time.Time

	// ToTime filters events before this time (inclusive).
	ToTime time.Time

	// EventTypes filters by event type (empty means all types).
	EventTypes []EventType

	// Actions filters by action (empty means all actions).
	Actions []Action

	// Statuses filters by status (empty means all statuses).
	Statuses []Status

	// CredentialIDHash filters by specific credential (empty means all credentials).
	CredentialIDHash string

	// Limit limits the number of results (0 means no limit).
	Limit int

	// Offset skips the first N results (for pagination).
	Offset int

	// OrderBy specifies the field to order by (default: timestamp).
	OrderBy string

	// Descending indicates descending order (default: true for newest first).
	Descending bool
}

// ReportFormat specifies the format for exported compliance reports.
type ReportFormat string

const (
	// FormatJSON exports the report as JSON.
	FormatJSON ReportFormat = "json"

	// FormatPDF exports the report as PDF (requires ACVS evidence chain).
	FormatPDF ReportFormat = "pdf"

	// FormatCSV exports the report as CSV.
	FormatCSV ReportFormat = "csv"

	// FormatHTML exports the report as HTML.
	FormatHTML ReportFormat = "html"
)

// AuditStatistics contains aggregate statistics from the audit log.
type AuditStatistics struct {
	// TotalEvents is the total number of events in the specified range.
	TotalEvents int

	// EventsByType breaks down events by type.
	EventsByType map[EventType]int

	// EventsByStatus breaks down events by status.
	EventsByStatus map[Status]int

	// RotationStats contains rotation-specific statistics.
	RotationStats *RotationStatistics

	// ComplianceStats contains ACVS-specific statistics (if enabled).
	ComplianceStats *ComplianceStatistics

	// TimeRange is the time range covered by these statistics.
	TimeRange struct {
		From time.Time
		To   time.Time
	}
}

// RotationStatistics contains statistics about credential rotations.
type RotationStatistics struct {
	// TotalRotations is the total number of rotation attempts.
	TotalRotations int

	// SuccessfulRotations is the number of successful rotations.
	SuccessfulRotations int

	// FailedRotations is the number of failed rotations.
	FailedRotations int

	// HIMRequiredRotations is the number of rotations requiring HIM.
	HIMRequiredRotations int

	// AverageDuration is the average time for a rotation operation.
	AverageDuration time.Duration

	// UniqueCredentialsRotated is the number of unique credentials rotated.
	UniqueCredentialsRotated int
}

// ComplianceStatistics contains statistics about ACVS compliance checks.
type ComplianceStatistics struct {
	// TotalChecks is the total number of compliance checks performed.
	TotalChecks int

	// AllowedActions is the number of actions allowed by ToS.
	AllowedActions int

	// BlockedActions is the number of actions blocked by ToS.
	BlockedActions int

	// HIMRequiredActions is the number of actions requiring HIM for compliance.
	HIMRequiredActions int

	// UnknownComplianceActions is the number of actions with unknown compliance.
	UnknownComplianceActions int

	// EvidenceChainsGenerated is the number of evidence chains created.
	EvidenceChainsGenerated int
}

// AuditError represents an error from audit operations.
type AuditError struct {
	// Code is the error code.
	Code AuditErrorCode

	// Message is a human-readable error message.
	Message string

	// Cause is the underlying error, if any.
	Cause error
}

func (e *AuditError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *AuditError) Unwrap() error {
	return e.Cause
}

// AuditErrorCode represents specific audit error conditions.
type AuditErrorCode string

const (
	// ErrDatabaseError indicates a database operation failed.
	ErrDatabaseError AuditErrorCode = "DATABASE_ERROR"

	// ErrSigningFailed indicates cryptographic signing failed.
	ErrSigningFailed AuditErrorCode = "SIGNING_FAILED"

	// ErrVerificationFailed indicates signature verification failed.
	ErrVerificationFailed AuditErrorCode = "VERIFICATION_FAILED"

	// ErrInvalidFilter indicates the provided filter is invalid.
	ErrInvalidFilter AuditErrorCode = "INVALID_FILTER"

	// ErrExportFailed indicates report export failed.
	ErrExportFailed AuditErrorCode = "EXPORT_FAILED"
)
