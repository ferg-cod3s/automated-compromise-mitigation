// Package bitwarden implements the PasswordManager interface for Bitwarden CLI.
//
// This package provides integration with Bitwarden's command-line interface (bw),
// allowing ACM to detect and remediate compromised credentials stored in Bitwarden
// vaults while maintaining zero-knowledge security.
//
// # Bitwarden CLI Requirements
//
// Installation:
//
//	# macOS
//	brew install bitwarden-cli
//
//	# Linux (via npm)
//	npm install -g @bitwarden/cli
//
//	# Linux (via snap)
//	snap install bw
//
//	# Windows (via Chocolatey)
//	choco install bitwarden-cli
//
// Authentication:
//
// The user must authenticate with Bitwarden before using ACM:
//
//	bw login
//	export BW_SESSION="$(bw unlock --raw)"
//
// Or use biometric unlock if configured.
//
// # CLI Commands Used
//
// Detect Compromised:
//
//	bw list items --exposed
//
// Get Credential:
//
//	bw get item <uuid>
//
// Update Password:
//
//	# Get item, modify password, encode as base64, update
//	bw get item <uuid> | jq '.login.password = "<new_password>"' | bw encode | bw edit item <uuid>
//
// Verify Update:
//
//	bw get item <uuid> --output json | jq -r '.revisionDate'
//
// Sync Vault:
//
//	bw sync
//
// # Known Limitations
//
//   - Requires BW_SESSION environment variable to be set
//   - Session expires after vault lock timeout (default: 15 min)
//   - Update operations require JSON encoding/decoding
//   - Sync may be required after remote changes
//   - Exposed items API requires premium subscription
//
// # Example Usage
//
//	import "github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager/bitwarden"
//
//	ctx := context.Background()
//	bw := bitwarden.New("/usr/local/bin/bw")
//
//	// Check if CLI is available
//	available, err := bw.IsAvailable(ctx)
//	if !available {
//	    log.Fatal("Bitwarden CLI not found")
//	}
//
//	// Check if vault is unlocked
//	locked, _ := bw.IsVaultLocked(ctx)
//	if locked {
//	    log.Fatal("Please unlock your vault with: bw unlock")
//	}
//
//	// Detect compromised credentials
//	compromised, err := bw.DetectCompromised(ctx)
//	if err != nil {
//	    log.Fatalf("Detection failed: %v", err)
//	}
//
//	// Update password for compromised credential
//	for _, cred := range compromised {
//	    newPass := generateSecurePassword()
//	    err := bw.UpdatePassword(ctx, cred.ID, newPass)
//	    if err != nil {
//	        log.Printf("Update failed for %s: %v", cred.Site, err)
//	    }
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - Basic Bitwarden CLI integration
//   - Detect compromised via --exposed flag
//   - Password update operations
//   - Session detection and validation
//   - Error handling
//
// Future phases will add:
//   - Self-hosted server support
//   - Organization vault support
//   - Advanced filtering and search
//   - Batch operations
//   - Automatic session renewal
package bitwarden
