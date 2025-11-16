// Package bitwarden implements the PasswordManager interface for Bitwarden CLI.
//
// This implementation maintains zero-knowledge principles by invoking the Bitwarden
// CLI as a subprocess. The master password and vault encryption keys are never
// accessed by this code.
package bitwarden

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// Manager implements the PasswordManager interface for Bitwarden.
type Manager struct {
	cliPath string // Path to bw CLI executable
}

// New creates a new Bitwarden password manager instance.
// It verifies that the Bitwarden CLI is installed and accessible.
func New() (*Manager, error) {
	cliPath, err := exec.LookPath("bw")
	if err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrCLINotFound,
			Message:   "Bitwarden CLI (bw) not found in PATH",
			Cause:     err,
			Retryable: false,
		}
	}

	return &Manager{
		cliPath: cliPath,
	}, nil
}

// DetectCompromised queries Bitwarden for compromised credentials.
// Uses: bw list items --pretty
// Then filters for items that have weak/exposed passwords.
func (m *Manager) DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error) {
	// Check if vault is locked first
	locked, err := m.IsVaultLocked(ctx)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Bitwarden vault is locked. Please unlock with: bw unlock",
			Retryable: true,
		}
	}

	// List all items
	cmd := exec.CommandContext(ctx, m.cliPath, "list", "items")
	output, err := cmd.Output()
	if err != nil {
		return nil, m.wrapCLIError("list items", err)
	}

	var items []bitwardenItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse Bitwarden items JSON",
			Cause:   err,
		}
	}

	// Bitwarden doesn't have a built-in "compromised" flag in the free version,
	// so we need to check password history and use external breach detection.
	// For Phase I, we'll use a simplified approach: check for weak passwords
	// and rely on user's breach detection service if available.

	var compromised []pwmanager.CompromisedCredential
	for _, item := range items {
		if item.Type != 1 { // Type 1 = Login
			continue
		}

		// For now, we'll return items that might need rotation
		// In Phase II, integrate with HIBP or similar service
		if item.Login.Username != "" && item.Name != "" {
			cred := pwmanager.CompromisedCredential{
				ID:          item.ID,
				Site:        item.Name,
				Username:    item.Login.Username,
				BreachName:  "Unknown", // TODO: Integrate with HIBP
				BreachDate:  time.Time{},
				LastRotated: parseTime(item.RevisionDate),
				RequiresHIM: false,
			}

			// Only add if it hasn't been rotated recently (>90 days)
			if time.Since(cred.LastRotated) > 90*24*time.Hour {
				compromised = append(compromised, cred)
			}
		}
	}

	return compromised, nil
}

// GetCredential retrieves metadata for a specific credential.
func (m *Manager) GetCredential(ctx context.Context, id string) (*pwmanager.Credential, error) {
	locked, err := m.IsVaultLocked(ctx)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Bitwarden vault is locked",
			Retryable: true,
		}
	}

	cmd := exec.CommandContext(ctx, m.cliPath, "get", "item", id)
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "Not found") {
			return nil, &pwmanager.PasswordManagerError{
				Code:    pwmanager.ErrCredentialNotFound,
				Message: fmt.Sprintf("Credential with ID %s not found", id),
				Cause:   err,
			}
		}
		return nil, m.wrapCLIError("get item", err)
	}

	var item bitwardenItem
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse Bitwarden item JSON",
			Cause:   err,
		}
	}

	return &pwmanager.Credential{
		ID:           item.ID,
		Site:         item.Name,
		Username:     item.Login.Username,
		URL:          getFirstURI(item.Login.URIs),
		LastModified: parseTime(item.RevisionDate),
		Notes:        item.Notes,
		CustomFields: make(map[string]string), // TODO: Parse fields
	}, nil
}

// UpdatePassword updates the password for a credential in the vault.
func (m *Manager) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	locked, err := m.IsVaultLocked(ctx)
	if err != nil {
		return err
	}
	if locked {
		return &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "Bitwarden vault is locked",
			Retryable: true,
		}
	}

	// First, get the current item
	cmd := exec.CommandContext(ctx, m.cliPath, "get", "item", id)
	output, err := cmd.Output()
	if err != nil {
		return m.wrapCLIError("get item for update", err)
	}

	var item bitwardenItem
	if err := json.Unmarshal(output, &item); err != nil {
		return &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse item for update",
			Cause:   err,
		}
	}

	// Update the password
	item.Login.Password = newPassword

	// Encode back to JSON
	updatedJSON, err := json.Marshal(item)
	if err != nil {
		return &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to encode updated item",
			Cause:   err,
		}
	}

	// Use bw edit to update the item
	// Note: This requires the item JSON to be passed via stdin or encoded
	cmd = exec.CommandContext(ctx, m.cliPath, "edit", "item", id, string(updatedJSON))
	if err := cmd.Run(); err != nil {
		return &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrUpdateFailed,
			Message:   fmt.Sprintf("Failed to update credential %s", id),
			Cause:     err,
			Retryable: true,
		}
	}

	// Sync to remote vault (optional but recommended)
	syncCmd := exec.CommandContext(ctx, m.cliPath, "sync")
	_ = syncCmd.Run() // Ignore sync errors

	return nil
}

// VerifyUpdate confirms that a password was successfully updated.
func (m *Manager) VerifyUpdate(ctx context.Context, id string, expectedModifiedAfter time.Time) (bool, error) {
	cred, err := m.GetCredential(ctx, id)
	if err != nil {
		return false, err
	}

	return cred.LastModified.After(expectedModifiedAfter), nil
}

// IsAvailable checks if the Bitwarden CLI is installed and accessible.
func (m *Manager) IsAvailable(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, m.cliPath, "--version")
	err := cmd.Run()
	return err == nil, nil
}

// IsVaultLocked checks if the Bitwarden vault is currently locked.
func (m *Manager) IsVaultLocked(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, m.cliPath, "status")
	output, err := cmd.Output()
	if err != nil {
		return true, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to check vault status",
			Cause:   err,
		}
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(output, &status); err != nil {
		return true, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse status JSON",
			Cause:   err,
		}
	}

	// Status can be: "unlocked", "locked", "unauthenticated"
	return status.Status != "unlocked", nil
}

// Type returns the type identifier for this password manager.
func (m *Manager) Type() string {
	return "bitwarden"
}

// wrapCLIError wraps a CLI error into a PasswordManagerError.
func (m *Manager) wrapCLIError(operation string, err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := string(exitErr.Stderr)

		if strings.Contains(stderr, "locked") {
			return &pwmanager.PasswordManagerError{
				Code:      pwmanager.ErrVaultLocked,
				Message:   "Vault is locked",
				Cause:     err,
				Retryable: true,
			}
		}

		if strings.Contains(stderr, "not found") {
			return &pwmanager.PasswordManagerError{
				Code:    pwmanager.ErrCredentialNotFound,
				Message: "Credential not found",
				Cause:   err,
			}
		}
	}

	return &pwmanager.PasswordManagerError{
		Code:      pwmanager.ErrUpdateFailed,
		Message:   fmt.Sprintf("Bitwarden CLI operation '%s' failed", operation),
		Cause:     err,
		Retryable: true,
	}
}

// bitwardenItem represents a Bitwarden vault item.
type bitwardenItem struct {
	ID           string `json:"id"`
	OrganizationID string `json:"organizationId,omitempty"`
	FolderID     string `json:"folderId,omitempty"`
	Type         int    `json:"type"` // 1 = Login, 2 = Note, 3 = Card, 4 = Identity
	Name         string `json:"name"`
	Notes        string `json:"notes,omitempty"`
	Favorite     bool   `json:"favorite"`
	Login        struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		TOTP     string `json:"totp,omitempty"`
		URIs     []struct {
			Match int    `json:"match,omitempty"`
			URI   string `json:"uri"`
		} `json:"uris,omitempty"`
	} `json:"login,omitempty"`
	RevisionDate string `json:"revisionDate"`
}

// Helper functions

func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func getFirstURI(uris []struct {
	Match int
	URI   string
}) string {
	if len(uris) > 0 {
		return uris[0].URI
	}
	return ""
}
