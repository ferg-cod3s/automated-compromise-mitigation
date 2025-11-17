package audit

import (
	"context"
	"testing"
	"time"
)

// TestLogEvent tests basic event logging functionality
func TestLogEvent(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	event := Event{
		Type:         EventTypeRotation,
		Status:       StatusSuccess,
		CredentialID: "test-credential-123",
		Site:         "example.com",
		Username:     "user@example.com",
		Message:      "Test rotation event",
	}

	err = logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	// Verify event was logged
	filter := Filter{
		CredentialID: "test-credential-123",
	}

	events, err := logger.QueryEvents(context.Background(), filter)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	retrieved := events[0]
	if retrieved.Type != EventTypeRotation {
		t.Errorf("Expected event type %s, got %s", EventTypeRotation, retrieved.Type)
	}
	if retrieved.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, retrieved.Status)
	}
	if retrieved.CredentialID != "test-credential-123" {
		t.Errorf("Expected credential ID test-credential-123, got %s", retrieved.CredentialID)
	}
	if retrieved.Signature == nil {
		t.Error("Expected signature to be set")
	}
}

// TestEventIDGeneration tests automatic event ID generation
func TestEventIDGeneration(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	event := Event{
		Type:    EventTypeRotation,
		Status:  StatusSuccess,
		Message: "Test event",
	}

	err = logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	events, err := logger.QueryEvents(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID == "" {
		t.Error("Expected event ID to be generated automatically")
	}
}

// TestTimestampGeneration tests automatic timestamp generation
func TestTimestampGeneration(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	before := time.Now()

	event := Event{
		Type:    EventTypeRotation,
		Status:  StatusSuccess,
		Message: "Test event",
	}

	err = logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	after := time.Now()

	events, err := logger.QueryEvents(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	timestamp := events[0].Timestamp
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("Timestamp %v is outside expected range [%v, %v]", timestamp, before, after)
	}
}

// TestSignatureVerification tests cryptographic signature verification
func TestSignatureVerification(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	event := Event{
		Type:         EventTypeRotation,
		Status:       StatusSuccess,
		CredentialID: "test-credential-456",
		Message:      "Test event for signature verification",
	}

	err = logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	// Query the event to get its ID
	events, err := logger.QueryEvents(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	eventID := events[0].ID

	// Verify the signature
	valid, err := logger.VerifyIntegrity(context.Background(), eventID)
	if err != nil {
		t.Fatalf("Failed to verify integrity: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed for valid event")
	}
}

// TestSignatureVerificationNonexistent tests verification of nonexistent event
func TestSignatureVerificationNonexistent(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	_, err = logger.VerifyIntegrity(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent event, got nil")
	}
}

// TestQueryEvents tests event filtering
func TestQueryEvents(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log multiple events
	events := []Event{
		{
			Type:         EventTypeRotation,
			Status:       StatusSuccess,
			CredentialID: "cred-1",
			Site:         "site1.com",
		},
		{
			Type:         EventTypeRotation,
			Status:       StatusFailure,
			CredentialID: "cred-2",
			Site:         "site2.com",
		},
		{
			Type:         EventTypeDetection,
			Status:       StatusSuccess,
			CredentialID: "cred-3",
			Site:         "site3.com",
		},
	}

	for _, event := range events {
		err := logger.LogEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	tests := []struct {
		name     string
		filter   Filter
		expected int
	}{
		{
			name:     "all events",
			filter:   Filter{},
			expected: 3,
		},
		{
			name:     "rotation events",
			filter:   Filter{EventType: EventTypeRotation},
			expected: 2,
		},
		{
			name:     "detection events",
			filter:   Filter{EventType: EventTypeDetection},
			expected: 1,
		},
		{
			name:     "success status",
			filter:   Filter{Status: StatusSuccess},
			expected: 2,
		},
		{
			name:     "failure status",
			filter:   Filter{Status: StatusFailure},
			expected: 1,
		},
		{
			name:     "specific credential",
			filter:   Filter{CredentialID: "cred-1"},
			expected: 1,
		},
		{
			name:     "limit results",
			filter:   Filter{Limit: 2},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := logger.QueryEvents(context.Background(), tt.filter)
			if err != nil {
				t.Fatalf("Failed to query events: %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("Expected %d events, got %d", tt.expected, len(results))
			}
		})
	}
}

// TestTimeRangeFilter tests filtering by time range
func TestTimeRangeFilter(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Log event with current time
	event := Event{
		Type:      EventTypeRotation,
		Status:    StatusSuccess,
		Timestamp: now,
	}

	err = logger.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	// Filter with time range that includes the event
	filter := Filter{
		StartTime: yesterday,
		EndTime:   tomorrow,
	}

	results, err := logger.QueryEvents(context.Background(), filter)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 event in time range, got %d", len(results))
	}

	// Filter with time range that excludes the event
	filterBefore := Filter{
		StartTime: yesterday.Add(-24 * time.Hour),
		EndTime:   yesterday,
	}

	results, err = logger.QueryEvents(context.Background(), filterBefore)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 events in excluded time range, got %d", len(results))
	}
}

// TestExportReport tests report export functionality
func TestExportReport(t *testing.T) {
	logger, err := NewMemoryLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log some events
	for i := 0; i < 5; i++ {
		event := Event{
			Type:    EventTypeRotation,
			Status:  StatusSuccess,
			Message: "Test event",
		}
		err := logger.LogEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	tests := []struct {
		name   string
		format ReportFormat
	}{
		{
			name:   "JSON export",
			format: ReportFormatJSON,
		},
		{
			name:   "CSV export",
			format: ReportFormatCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := logger.ExportReport(context.Background(), Filter{}, tt.format)
			if err != nil {
				t.Fatalf("Failed to export report: %v", err)
			}

			if len(data) == 0 {
				t.Error("Expected non-empty report data")
			}
		})
	}
}
