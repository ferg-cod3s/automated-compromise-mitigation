// Package pwmanager provides interfaces and implementations for interacting
// with password manager CLIs in a zero-knowledge manner.
//
// The PasswordManager interface abstracts interactions with various password
// managers (1Password, Bitwarden, LastPass, etc.) without ever accessing the
// master password or vault encryption keys.
package pwmanager

import (
	"context"
	"time"
)

// PasswordManager defines the interface for interacting with password manager CLIs.
// Implementations must maintain zero-knowledge principles: they execute the CLI
// in a subprocess and parse JSON output without ever handling master passwords
// or encryption keys.
type PasswordManager interface {
	// DetectCompromised queries the password manager for compromised credentials.
	// Returns a list of credentials that have been exposed in known data breaches.
	//
	// Example CLI invocations:
	//   - Bitwarden: `bw list items --exposed`
	//   - 1Password: `op item list --categories=Login --compromised`
	DetectCompromised(ctx context.Context) ([]CompromisedCredential, error)

	// GetCredential retrieves metadata for a specific credential by ID.
	// Does NOT return the password itself, only metadata needed for rotation.
	GetCredential(ctx context.Context, id string) (*Credential, error)

	// UpdatePassword updates the password for a specific credential in the vault.
	// The password manager CLI handles all encryption/decryption internally.
	//
	// Example CLI invocation:
	//   - Bitwarden: `bw edit item <id> --password "<new_password>"`
	UpdatePassword(ctx context.Context, id string, newPassword string) error

	// VerifyUpdate confirms that a password was successfully updated by comparing
	// the last modified timestamp or retrieving the item again.
	VerifyUpdate(ctx context.Context, id string, expectedModifiedAfter time.Time) (bool, error)

	// IsAvailable checks if the password manager CLI is installed and accessible.
	IsAvailable(ctx context.Context) (bool, error)

	// IsVaultLocked checks if the vault is currently locked and requires authentication.
	IsVaultLocked(ctx context.Context) (bool, error)

	// Type returns the type of password manager (e.g., "bitwarden", "1password").
	Type() string
}

// CompromisedCredential represents a credential that has been exposed in a data breach.
type CompromisedCredential struct {
	// ID is the unique identifier for this credential in the password manager.
	// This ID should be hashed before logging to protect user privacy.
	ID string

	// Site is the website or service associated with this credential.
	Site string

	// Username is the username or email associated with this credential.
	Username string

	// BreachName is the name of the data breach where this credential was found.
	BreachName string

	// BreachDate is when the breach occurred.
	BreachDate time.Time

	// LastRotated is when the password was last changed.
	// Zero value indicates the password has never been rotated by ACM.
	LastRotated time.Time

	// RequiresHIM indicates whether this credential requires Human-in-the-Middle
	// intervention (e.g., due to MFA, CAPTCHA, or ToS restrictions).
	RequiresHIM bool
}

// Credential represents metadata about a password manager entry.
// It does NOT include the actual password (zero-knowledge principle).
type Credential struct {
	// ID is the unique identifier for this credential.
	ID string

	// Site is the website or service name.
	Site string

	// Username is the username or email.
	Username string

	// URL is the login URL for the service.
	URL string

	// LastModified is when this credential was last modified.
	LastModified time.Time

	// Notes contains any notes associated with this credential.
	Notes string

	// CustomFields contains any custom fields defined for this credential.
	CustomFields map[string]string
}

// PasswordPolicy defines the requirements for generated passwords.
type PasswordPolicy struct {
	// Length is the desired password length (default: 32).
	Length int

	// RequireUppercase indicates if uppercase letters are required.
	RequireUppercase bool

	// RequireLowercase indicates if lowercase letters are required.
	RequireLowercase bool

	// RequireNumbers indicates if numbers are required.
	RequireNumbers bool

	// RequireSymbols indicates if special symbols are required.
	RequireSymbols bool

	// ExcludeAmbiguous indicates if ambiguous characters (0, O, l, 1) should be excluded.
	ExcludeAmbiguous bool

	// CustomCharset allows specifying a custom character set (overrides other settings).
	CustomCharset string
}

// DefaultPasswordPolicy returns the default password policy used by ACM.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		Length:           32,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
		ExcludeAmbiguous: false,
		CustomCharset:    "",
	}
}

// PasswordManagerError represents an error from password manager operations.
type PasswordManagerError struct {
	// Code is the error code (e.g., VAULT_LOCKED, CLI_NOT_FOUND).
	Code ErrorCode

	// Message is a human-readable error message.
	Message string

	// Cause is the underlying error, if any.
	Cause error

	// Retryable indicates if the operation can be retried.
	Retryable bool
}

func (e *PasswordManagerError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *PasswordManagerError) Unwrap() error {
	return e.Cause
}

// ErrorCode represents specific error conditions.
type ErrorCode string

const (
	// ErrVaultLocked indicates the vault is locked and requires authentication.
	ErrVaultLocked ErrorCode = "VAULT_LOCKED"

	// ErrCLINotFound indicates the password manager CLI is not installed.
	ErrCLINotFound ErrorCode = "CLI_NOT_FOUND"

	// ErrNetworkRequired indicates network access is required (e.g., for vault sync).
	ErrNetworkRequired ErrorCode = "NETWORK_REQUIRED"

	// ErrCredentialNotFound indicates the specified credential ID does not exist.
	ErrCredentialNotFound ErrorCode = "CREDENTIAL_NOT_FOUND"

	// ErrUpdateFailed indicates the password update operation failed.
	ErrUpdateFailed ErrorCode = "UPDATE_FAILED"

	// ErrPermissionDenied indicates insufficient permissions to perform the operation.
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"
)
