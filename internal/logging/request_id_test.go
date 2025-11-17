package logging

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

func TestGenerateRequestID(t *testing.T) {
	// Test that we can generate a request ID
	id := GenerateRequestID()
	if id == "" {
		t.Fatal("GenerateRequestID returned empty string")
	}

	// Test UUID v7 format (8-4-4-4-12)
	if !IsValidRequestID(id) {
		t.Errorf("GenerateRequestID returned invalid format: %s", id)
	}

	// Test that IDs are unique
	id2 := GenerateRequestID()
	if id == id2 {
		t.Error("GenerateRequestID returned duplicate IDs")
	}

	// Test that ID contains dashes in correct positions
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Errorf("Expected 5 parts, got %d: %s", len(parts), id)
	}

	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 ||
		len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("Invalid UUID format: %s", id)
	}

	// Test that version is 7 (should be 7xxx in third group)
	if parts[2][0] != '7' {
		t.Errorf("Expected UUID v7, got version %c: %s", parts[2][0], id)
	}
}

func TestIsValidRequestID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid UUID v7",
			input:    GenerateRequestID(),
			expected: true,
		},
		{
			name:     "valid format",
			input:    "01234567-89ab-7def-0123-456789abcdef",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too short",
			input:    "123",
			expected: false,
		},
		{
			name:     "missing dashes",
			input:    "0123456789ab7def0123456789abcdef",
			expected: false,
		},
		{
			name:     "wrong dash positions",
			input:    "01234567-89ab-7def-01234-56789abcdef",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRequestID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidRequestID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "no metadata",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name: "with request ID",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(RequestIDMetadataKey, "test-request-id"),
			),
			expected: "test-request-id",
		},
		{
			name: "empty request ID",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(RequestIDMetadataKey, ""),
			),
			expected: "",
		},
		{
			name: "multiple values (takes first)",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(
					RequestIDMetadataKey, "first-id",
					RequestIDMetadataKey, "second-id",
				),
			),
			expected: "first-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRequestID(tt.ctx)
			if result != tt.expected {
				t.Errorf("ExtractRequestID() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInjectRequestID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		requestID string
		validate  func(t *testing.T, ctx context.Context)
	}{
		{
			name:      "inject into empty context",
			ctx:       context.Background(),
			requestID: "test-id",
			validate: func(t *testing.T, ctx context.Context) {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					t.Fatal("No metadata in context")
				}
				values := md.Get(RequestIDMetadataKey)
				if len(values) != 1 || values[0] != "test-id" {
					t.Errorf("Expected request ID 'test-id', got %v", values)
				}
			},
		},
		{
			name: "inject into existing metadata",
			ctx: metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("other-key", "other-value"),
			),
			requestID: "new-id",
			validate: func(t *testing.T, ctx context.Context) {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					t.Fatal("No metadata in context")
				}
				// Check request ID
				values := md.Get(RequestIDMetadataKey)
				if len(values) != 1 || values[0] != "new-id" {
					t.Errorf("Expected request ID 'new-id', got %v", values)
				}
				// Check other metadata preserved
				other := md.Get("other-key")
				if len(other) != 1 || other[0] != "other-value" {
					t.Errorf("Expected other metadata preserved, got %v", other)
				}
			},
		},
		{
			name:      "generate ID if empty",
			ctx:       context.Background(),
			requestID: "",
			validate: func(t *testing.T, ctx context.Context) {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					t.Fatal("No metadata in context")
				}
				values := md.Get(RequestIDMetadataKey)
				if len(values) != 1 {
					t.Fatal("Expected one request ID")
				}
				if !IsValidRequestID(values[0]) {
					t.Errorf("Generated invalid request ID: %s", values[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InjectRequestID(tt.ctx, tt.requestID)
			tt.validate(t, result)
		})
	}
}

func TestGetOrGenerateRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		validate func(t *testing.T, ctx context.Context, id string)
	}{
		{
			name: "extract from incoming metadata",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(RequestIDMetadataKey, "metadata-id"),
			),
			validate: func(t *testing.T, ctx context.Context, id string) {
				if id != "metadata-id" {
					t.Errorf("Expected 'metadata-id', got %q", id)
				}
				// Check it's stored in context
				storedID := getRequestIDFromContext(ctx)
				if storedID != "metadata-id" {
					t.Errorf("ID not stored in context: got %q", storedID)
				}
			},
		},
		{
			name: "get from context value",
			ctx:  SetRequestIDInContext(context.Background(), "context-id"),
			validate: func(t *testing.T, ctx context.Context, id string) {
				if id != "context-id" {
					t.Errorf("Expected 'context-id', got %q", id)
				}
			},
		},
		{
			name: "generate new ID",
			ctx:  context.Background(),
			validate: func(t *testing.T, ctx context.Context, id string) {
				if id == "" {
					t.Error("Expected generated ID, got empty string")
				}
				if !IsValidRequestID(id) {
					t.Errorf("Generated invalid ID: %s", id)
				}
				// Check it's stored in context
				storedID := getRequestIDFromContext(ctx)
				if storedID != id {
					t.Errorf("ID not stored in context: got %q, want %q", storedID, id)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, id := GetOrGenerateRequestID(tt.ctx)
			tt.validate(t, ctx, id)
		})
	}
}

func TestRequestIDUniqueness(t *testing.T) {
	// Generate many IDs to test for collisions
	count := 10000
	ids := make(map[string]bool, count)

	for i := 0; i < count; i++ {
		id := GenerateRequestID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestRequestIDTimeSortable(t *testing.T) {
	// Generate IDs with time gaps and verify they sort chronologically
	id1 := GenerateRequestID()

	// Sleep to ensure different timestamp
	time.Sleep(2 * time.Millisecond)
	id2 := GenerateRequestID()

	time.Sleep(2 * time.Millisecond)
	id3 := GenerateRequestID()

	// UUIDs should be lexicographically sortable by time
	// when generated with sufficient time gaps
	if !(id1 < id2 && id2 < id3) {
		t.Errorf("IDs not time-sortable: %s, %s, %s", id1, id2, id3)
	}

	// Extract timestamp portions (first 8 chars) and verify ordering
	ts1 := id1[:8]
	ts2 := id2[:8]
	ts3 := id3[:8]

	if !(ts1 <= ts2 && ts2 <= ts3) {
		t.Errorf("Timestamp portions not ordered: %s, %s, %s", ts1, ts2, ts3)
	}
}
