package audit

import (
	"context"
)

// Logger provides audit logging capabilities with cryptographic signatures.
type Logger interface {
	// LogEvent logs an event to the audit trail with a cryptographic signature.
	LogEvent(ctx context.Context, event Event) error

	// QueryEvents retrieves events matching the specified filter.
	QueryEvents(ctx context.Context, filter Filter) ([]Event, error)

	// VerifyIntegrity verifies the cryptographic signature of an event.
	VerifyIntegrity(ctx context.Context, eventID string) (bool, error)

	// ExportReport generates a compliance report for the specified time range.
	ExportReport(ctx context.Context, filter Filter, format ReportFormat) ([]byte, error)

	// Close closes the audit logger and releases resources.
	Close() error
}
