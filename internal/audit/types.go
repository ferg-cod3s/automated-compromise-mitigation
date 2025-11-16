package audit

import "time"

// Event represents a single audit log entry.
type Event struct {
	// ID is the unique identifier for this event.
	ID string

	// Timestamp is when the event occurred.
	Timestamp time.Time

	// Type categorizes the event (rotation, detection, compliance, etc.).
	Type EventType

	// Status indicates the outcome (success, failure, pending).
	Status EventStatus

	// CredentialID is the hashed credential identifier.
	CredentialID string

	// Site is the website or service associated with this event.
	Site string

	// Username is the username associated with this event.
	Username string

	// Message provides additional context about the event.
	Message string

	// Metadata contains additional key-value data.
	Metadata map[string]string

	// Signature is the Ed25519 signature of this event.
	Signature []byte
}

// EventType categorizes audit events.
type EventType string

const (
	// EventTypeRotation indicates a credential rotation event.
	EventTypeRotation EventType = "rotation"

	// EventTypeDetection indicates a breach detection event.
	EventTypeDetection EventType = "detection"

	// EventTypeCompliance indicates an ACVS compliance validation event.
	EventTypeCompliance EventType = "compliance"

	// EventTypeHIM indicates a Human-in-the-Middle interaction event.
	EventTypeHIM EventType = "him"

	// EventTypeAuth indicates an authentication event.
	EventTypeAuth EventType = "auth"

	// EventTypeSystem indicates a system-level event.
	EventTypeSystem EventType = "system"
)

// EventStatus indicates the outcome of an event.
type EventStatus string

const (
	// StatusSuccess indicates the operation succeeded.
	StatusSuccess EventStatus = "success"

	// StatusFailure indicates the operation failed.
	StatusFailure EventStatus = "failure"

	// StatusPending indicates the operation is in progress.
	StatusPending EventStatus = "pending"

	// StatusSkipped indicates the operation was skipped.
	StatusSkipped EventStatus = "skipped"
)

// Filter specifies criteria for querying audit events.
type Filter struct {
	// EventType filters by event type.
	EventType EventType

	// Status filters by event status.
	Status EventStatus

	// CredentialID filters by credential ID.
	CredentialID string

	// Site filters by site name.
	Site string

	// StartTime filters events after this time.
	StartTime time.Time

	// EndTime filters events before this time.
	EndTime time.Time

	// Limit restricts the number of results returned.
	Limit int
}

// ReportFormat specifies the format for exported reports.
type ReportFormat string

const (
	// ReportFormatJSON exports as JSON.
	ReportFormatJSON ReportFormat = "json"

	// ReportFormatCSV exports as CSV.
	ReportFormatCSV ReportFormat = "csv"

	// ReportFormatPDF exports as PDF (Phase II).
	ReportFormatPDF ReportFormat = "pdf"

	// ReportFormatHTML exports as HTML.
	ReportFormatHTML ReportFormat = "html"
)
