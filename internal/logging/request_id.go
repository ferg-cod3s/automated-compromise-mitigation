// Package logging provides request ID generation and propagation.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"
)

const (
	// RequestIDMetadataKey is the gRPC metadata key for request IDs.
	RequestIDMetadataKey = "x-request-id"

	// RequestIDHeader is the HTTP header name for request IDs.
	RequestIDHeader = "X-Request-ID"
)

// GenerateRequestID generates a new UUID v7 request ID.
// UUID v7 is time-sortable and collision-resistant.
//
// Format: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx
// - First 48 bits: Unix timestamp in milliseconds
// - Next 12 bits: Random data
// - Next 2 bits: Version (0111 = 7)
// - Next 2 bits: Variant (10 = RFC 4122)
// - Last 62 bits: Random data
func GenerateRequestID() string {
	var uuid [16]byte

	// Get current timestamp in milliseconds
	timestamp := time.Now().UnixMilli()

	// Write timestamp to first 48 bits (6 bytes)
	binary.BigEndian.PutUint64(uuid[0:8], uint64(timestamp)<<16)

	// Fill remaining bytes with random data
	if _, err := rand.Read(uuid[6:]); err != nil {
		// Fallback to timestamp-based randomness if crypto/rand fails
		timestamp := time.Now().UnixNano()
		binary.BigEndian.PutUint64(uuid[8:], uint64(timestamp))
	}

	// Set version to 7 (0111xxxx)
	uuid[6] = (uuid[6] & 0x0f) | 0x70

	// Set variant to RFC 4122 (10xxxxxx)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// Format as UUID string
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:16],
	)
}

// ExtractRequestID extracts the request ID from gRPC metadata.
// Returns empty string if not found.
func ExtractRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(RequestIDMetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// InjectRequestID injects a request ID into gRPC metadata for outgoing calls.
// If requestID is empty, a new one is generated.
func InjectRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = GenerateRequestID()
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	md.Set(RequestIDMetadataKey, requestID)
	return metadata.NewOutgoingContext(ctx, md)
}

// GetOrGenerateRequestID gets the request ID from context or generates a new one.
// This is useful for ensuring every request has an ID.
func GetOrGenerateRequestID(ctx context.Context) (context.Context, string) {
	// First try to extract from incoming metadata
	requestID := ExtractRequestID(ctx)
	if requestID != "" {
		// Store in context for later retrieval
		ctx = SetRequestIDInContext(ctx, requestID)
		return ctx, requestID
	}

	// Try to get from context value
	requestID = getRequestIDFromContext(ctx)
	if requestID != "" {
		return ctx, requestID
	}

	// Generate new request ID
	requestID = GenerateRequestID()
	ctx = SetRequestIDInContext(ctx, requestID)
	return ctx, requestID
}

// IsValidRequestID validates a request ID format.
// Currently checks for non-empty and reasonable length.
func IsValidRequestID(requestID string) bool {
	// UUID v7 format: 8-4-4-4-12 = 36 characters with dashes
	return len(requestID) == 36 && requestID[8] == '-' && requestID[13] == '-' &&
		requestID[18] == '-' && requestID[23] == '-'
}
