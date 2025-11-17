// Package rotation provides common types and utilities for credential rotation.
package rotation

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// RotationState represents the persistent state of a credential rotation.
type RotationState struct {
	ID           string            `json:"id"`
	CredentialID string            `json:"credential_id"`
	Provider     string            `json:"provider"` // "github", "aws", etc.
	State        string            `json:"state"`    // current state in the workflow
	StartedAt    time.Time         `json:"started_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ExpiresAt    time.Time         `json:"expires_at"` // for cleanup
	Metadata     map[string]string `json:"metadata"`   // provider-specific data
}

// StateFilter represents filter criteria for querying rotation states.
type StateFilter struct {
	CredentialID  string
	Provider      string
	State         string
	ExcludeStates []string
	OnlyExpired   bool
}

// GenerateStateID generates a unique rotation state ID.
func GenerateStateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "rot-" + hex.EncodeToString(b)
}
