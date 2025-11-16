package audit

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MemoryLogger implements Logger using in-memory storage.
// This is suitable for Phase I testing. Phase II will use SQLite.
type MemoryLogger struct {
	mu         sync.RWMutex
	events     []Event
	signingKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewMemoryLogger creates a new in-memory audit logger.
func NewMemoryLogger() (*MemoryLogger, error) {
	// Generate signing keys
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return &MemoryLogger{
		events:     make([]Event, 0),
		signingKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// LogEvent logs an event to the audit trail with a cryptographic signature.
func (l *MemoryLogger) LogEvent(ctx context.Context, event Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Create signature
	message := fmt.Sprintf("%s|%d|%s|%s|%s",
		event.ID,
		event.Timestamp.Unix(),
		event.CredentialID,
		event.Type,
		event.Status)
	event.Signature = ed25519.Sign(l.signingKey, []byte(message))

	// Append to events
	l.events = append(l.events, event)

	return nil
}

// QueryEvents retrieves events matching the specified filter.
func (l *MemoryLogger) QueryEvents(ctx context.Context, filter Filter) ([]Event, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var results []Event
	for _, event := range l.events {
		if matchesFilter(event, filter) {
			results = append(results, event)
		}
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results, nil
}

// VerifyIntegrity verifies the cryptographic signature of an event.
func (l *MemoryLogger) VerifyIntegrity(ctx context.Context, eventID string) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, event := range l.events {
		if event.ID == eventID {
			// Recreate the message
			message := fmt.Sprintf("%s|%d|%s|%s|%s",
				event.ID,
				event.Timestamp.Unix(),
				event.CredentialID,
				event.Type,
				event.Status)

			// Verify signature
			return ed25519.Verify(l.publicKey, []byte(message), event.Signature), nil
		}
	}

	return false, fmt.Errorf("event not found: %s", eventID)
}

// ExportReport generates a compliance report for the specified time range.
func (l *MemoryLogger) ExportReport(ctx context.Context, filter Filter, format ReportFormat) ([]byte, error) {
	events, err := l.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case ReportFormatJSON:
		return json.MarshalIndent(events, "", "  ")
	case ReportFormatCSV:
		return exportCSV(events)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Close closes the audit logger.
func (l *MemoryLogger) Close() error {
	// Nothing to close for memory logger
	return nil
}

// Helper functions

func matchesFilter(event Event, filter Filter) bool {
	if filter.EventType != "" && event.Type != filter.EventType {
		return false
	}
	if filter.Status != "" && event.Status != filter.Status {
		return false
	}
	if filter.CredentialID != "" && event.CredentialID != filter.CredentialID {
		return false
	}
	if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
		return false
	}
	return true
}

func generateEventID() string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:16])
}

func exportCSV(events []Event) ([]byte, error) {
	csv := "ID,Timestamp,Type,Status,CredentialID,Site,Username,Message\n"
	for _, event := range events {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%q\n",
			event.ID,
			event.Timestamp.Format(time.RFC3339),
			event.Type,
			event.Status,
			event.CredentialID,
			event.Site,
			event.Username,
			event.Message,
		)
	}
	return []byte(csv), nil
}
