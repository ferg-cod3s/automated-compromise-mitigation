// Package onepassword implements the PasswordManager interface for 1Password CLI.
//
// This implementation maintains zero-knowledge principles by invoking the 1Password
// CLI (op) as a subprocess. The master password and vault encryption keys are never
// accessed by this code.
package onepassword

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// Manager implements the PasswordManager interface for 1Password.
type Manager struct {
	cliPath string // Path to op CLI executable
}

// New creates a new 1Password password manager instance.
// It verifies that the 1Password CLI is installed and accessible.
func New() (*Manager, error) {
	cliPath, err := exec.LookPath("op")
	if err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrCLINotFound,
			Message:   "1Password CLI (op) not found in PATH",
			Cause:     err,
			Retryable: false,
		}
	}

	return &Manager{
		cliPath: cliPath,
	}, nil
}

// DetectCompromised queries 1Password for compromised credentials.
// Uses: op item list --categories Login --format json
func (m *Manager) DetectCompromised(ctx context.Context) ([]pwmanager.CompromisedCredential, error) {
	// Check if signed in
	signedIn, err := m.IsAvailable(ctx)
	if err != nil || !signedIn {
		return nil, &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrVaultLocked,
			Message:   "1Password CLI not signed in. Please sign in with: op signin",
			Retryable: true,
		}
	}

	// List all login items
	cmd := exec.CommandContext(ctx, m.cliPath, "item", "list", "--categories", "Login", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, m.wrapCLIError("list items", err)
	}

	var items []onePasswordItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse 1Password items JSON",
			Cause:   err,
		}
	}

	// 1Password has built-in Watchtower for breach detection
	// For Phase I, we'll use a simplified approach similar to Bitwarden
	var compromised []pwmanager.CompromisedCredential

	for _, item := range items {
		// Get detailed item info to check for tags or watchtower flags
		detailCmd := exec.CommandContext(ctx, m.cliPath, "item", "get", item.ID, "--format", "json")
		detailOutput, err := detailCmd.Output()
		if err != nil {
			continue // Skip items we can't access
		}

		var detailedItem onePasswordDetailedItem
		if err := json.Unmarshal(detailOutput, &detailedItem); err != nil {
			continue
		}

		// Check if item has been updated recently
		lastModified := parseTime(detailedItem.UpdatedAt)
		if time.Since(lastModified) > 90*24*time.Hour {
			username := getFieldValue(detailedItem.Fields, "username")
			if username != "" {
				cred := pwmanager.CompromisedCredential{
					ID:          item.ID,
					Site:        item.Title,
					Username:    username,
					BreachName:  "Unknown", // TODO: Integrate with 1Password Watchtower API
					BreachDate:  time.Time{},
					LastRotated: lastModified,
					RequiresHIM: false,
				}
				compromised = append(compromised, cred)
			}
		}
	}

	return compromised, nil
}

// GetCredential retrieves metadata for a specific credential.
func (m *Manager) GetCredential(ctx context.Context, id string) (*pwmanager.Credential, error) {
	cmd := exec.CommandContext(ctx, m.cliPath, "item", "get", id, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, &pwmanager.PasswordManagerError{
				Code:    pwmanager.ErrCredentialNotFound,
				Message: fmt.Sprintf("Credential with ID %s not found", id),
				Cause:   err,
			}
		}
		return nil, m.wrapCLIError("get item", err)
	}

	var item onePasswordDetailedItem
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, &pwmanager.PasswordManagerError{
			Code:    pwmanager.ErrUpdateFailed,
			Message: "Failed to parse 1Password item JSON",
			Cause:   err,
		}
	}

	return &pwmanager.Credential{
		ID:           item.ID,
		Site:         item.Title,
		Username:     getFieldValue(item.Fields, "username"),
		URL:          getURL(item.URLs),
		LastModified: parseTime(item.UpdatedAt),
		Notes:        getNotesSection(item.Fields),
		CustomFields: make(map[string]string), // TODO: Parse custom fields
	}, nil
}

// UpdatePassword updates the password for a credential in the vault.
func (m *Manager) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	// 1Password CLI v2 uses: op item edit <id> password=<new_password>
	cmd := exec.CommandContext(ctx, m.cliPath, "item", "edit", id, fmt.Sprintf("password=%s", newPassword))
	if err := cmd.Run(); err != nil {
		return &pwmanager.PasswordManagerError{
			Code:      pwmanager.ErrUpdateFailed,
			Message:   fmt.Sprintf("Failed to update credential %s", id),
			Cause:     err,
			Retryable: true,
		}
	}

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

// IsAvailable checks if the 1Password CLI is installed and the user is signed in.
func (m *Manager) IsAvailable(ctx context.Context) (bool, error) {
	// Try to list accounts to verify we're signed in
	cmd := exec.CommandContext(ctx, m.cliPath, "account", "list", "--format", "json")
	err := cmd.Run()
	return err == nil, nil
}

// IsVaultLocked checks if the 1Password vault requires authentication.
// For 1Password CLI v2, the session is managed automatically with biometric unlock.
func (m *Manager) IsVaultLocked(ctx context.Context) (bool, error) {
	available, err := m.IsAvailable(ctx)
	if err != nil {
		return true, err
	}
	// If CLI is available and signed in, vault is not locked
	return !available, nil
}

// Type returns the type identifier for this password manager.
func (m *Manager) Type() string {
	return "1password"
}

// wrapCLIError wraps a CLI error into a PasswordManagerError.
func (m *Manager) wrapCLIError(operation string, err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := string(exitErr.Stderr)

		if strings.Contains(stderr, "not signed in") || strings.Contains(stderr, "authentication") {
			return &pwmanager.PasswordManagerError{
				Code:      pwmanager.ErrVaultLocked,
				Message:   "Not signed in to 1Password",
				Cause:     err,
				Retryable: true,
			}
		}

		if strings.Contains(stderr, "not found") {
			return &pwmanager.PasswordManagerError{
				Code:    pwmanager.ErrCredentialNotFound,
				Message: "Item not found",
				Cause:   err,
			}
		}
	}

	return &pwmanager.PasswordManagerError{
		Code:      pwmanager.ErrUpdateFailed,
		Message:   fmt.Sprintf("1Password CLI operation '%s' failed", operation),
		Cause:     err,
		Retryable: true,
	}
}

// onePasswordItem represents a 1Password vault item from list operation.
type onePasswordItem struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Version   int      `json:"version"`
	Vault     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vault"`
	Category  string    `json:"category"`
	UpdatedAt string    `json:"updated_at"`
	Tags      []string  `json:"tags,omitempty"`
}

// onePasswordDetailedItem represents a detailed 1Password item from get operation.
type onePasswordDetailedItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	UpdatedAt string `json:"updated_at"`
	URLs      []struct {
		Primary bool   `json:"primary"`
		Href    string `json:"href"`
	} `json:"urls,omitempty"`
	Fields []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Purpose string `json:"purpose,omitempty"`
		Label   string `json:"label"`
		Value   string `json:"value,omitempty"`
	} `json:"fields,omitempty"`
}

// Helper functions

func parseTime(timeStr string) time.Time {
	// 1Password uses RFC3339 format
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func getFieldValue(fields []struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
	Label   string `json:"label"`
	Value   string `json:"value,omitempty"`
}, label string) string {
	for _, field := range fields {
		if strings.EqualFold(field.Label, label) || strings.EqualFold(field.Purpose, label) {
			return field.Value
		}
	}
	return ""
}

func getURL(urls []struct {
	Primary bool
	Href    string
}) string {
	for _, url := range urls {
		if url.Primary {
			return url.Href
		}
	}
	if len(urls) > 0 {
		return urls[0].Href
	}
	return ""
}

func getNotesSection(fields []struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
	Label   string `json:"label"`
	Value   string `json:"value,omitempty"`
}) string {
	for _, field := range fields {
		if field.Type == "CONCEALED" && strings.Contains(strings.ToLower(field.Label), "note") {
			return field.Value
		}
	}
	return ""
}
