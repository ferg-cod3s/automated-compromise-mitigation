// Package pwmanager provides interfaces and implementations for interacting
// with password manager CLIs in a zero-knowledge manner.
//
// The pwmanager package abstracts interactions with various password managers
// (1Password, Bitwarden, LastPass, pass) without ever accessing master passwords
// or vault encryption keys. All operations are performed by executing the
// password manager's CLI in a subprocess and parsing structured output.
//
// # Zero-Knowledge Principles
//
// ACM maintains zero-knowledge security by:
//
//   - NEVER storing or transmitting master passwords
//   - NEVER accessing vault encryption keys
//   - Executing CLI in isolated subprocess with minimal environment
//   - Relying on user's authenticated CLI session
//   - Parsing only metadata from CLI output (not passwords)
//   - Clearing sensitive data from memory immediately after use
//
// # Supported Password Managers
//
// Phase I:
//   - 1Password CLI (op)
//   - Bitwarden CLI (bw)
//
// Future phases:
//   - LastPass CLI (lpass)
//   - pass (Unix Password Manager)
//   - KeePassXC CLI (keepassxc-cli)
//   - Dashlane CLI
//
// # Interface Design
//
// The PasswordManager interface provides a consistent abstraction across
// different password manager implementations:
//
//	type PasswordManager interface {
//	    DetectCompromised(ctx) ([]CompromisedCredential, error)
//	    GetCredential(ctx, id) (*Credential, error)
//	    UpdatePassword(ctx, id, newPassword) error
//	    VerifyUpdate(ctx, id, expectedModifiedAfter) (bool, error)
//	    IsAvailable(ctx) (bool, error)
//	    IsVaultLocked(ctx) (bool, error)
//	    Type() string
//	}
//
// # Example CLI Invocations
//
// 1Password:
//
//	# Detect compromised credentials
//	op item list --categories=Login --compromised --format=json
//
//	# Update password
//	op item edit <id> password="<new_password>"
//
// Bitwarden:
//
//	# Detect compromised credentials
//	bw list items --exposed
//
//	# Update password
//	bw edit item <id> --password "<new_password>"
//
// # Security Considerations
//
//   - CLI executed with minimal environment variables
//   - Subprocess stdout/stderr captured and parsed
//   - No shell interpretation (direct exec, not via sh/bash)
//   - Timeouts enforced to prevent hanging
//   - Memory locking for password buffers
//   - Explicit zeroing after use
//
// # Error Handling
//
// The package defines structured errors for common failure scenarios:
//
//   - ErrVaultLocked: Vault requires authentication
//   - ErrCLINotFound: Password manager CLI not installed
//   - ErrNetworkRequired: Operation requires network access
//   - ErrCredentialNotFound: Credential ID does not exist
//   - ErrUpdateFailed: Password update operation failed
//   - ErrPermissionDenied: Insufficient permissions
//
// # Example Usage
//
//	// Auto-detect password manager
//	pm, err := pwmanager.Detect()
//	if err != nil {
//	    log.Fatalf("No password manager found: %v", err)
//	}
//	log.Printf("Detected password manager: %s", pm.Type())
//
//	// Check if vault is unlocked
//	locked, _ := pm.IsVaultLocked(ctx)
//	if locked {
//	    log.Fatal("Please unlock your vault first")
//	}
//
//	// Detect compromised credentials
//	compromised, err := pm.DetectCompromised(ctx)
//	if err != nil {
//	    log.Fatalf("Detection failed: %v", err)
//	}
//	log.Printf("Found %d compromised credentials", len(compromised))
//
//	// Update a password
//	newPass := generateSecurePassword()
//	err = pm.UpdatePassword(ctx, "credential-id-123", newPass)
//	if err != nil {
//	    log.Fatalf("Update failed: %v", err)
//	}
//
//	// Verify update succeeded
//	verified, _ := pm.VerifyUpdate(ctx, "credential-id-123", time.Now().Add(-1*time.Minute))
//	if !verified {
//	    log.Fatal("Password update verification failed")
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - 1Password CLI integration (primary)
//   - Bitwarden CLI integration (secondary)
//   - Basic detect/update/verify operations
//   - Error handling and retry logic
//   - CLI availability detection
//
// Future phases will add:
//   - Additional password manager support
//   - Batch operations optimization
//   - Advanced error recovery
//   - Performance monitoring
package pwmanager
