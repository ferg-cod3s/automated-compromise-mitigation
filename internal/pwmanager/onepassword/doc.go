// Package onepassword implements the PasswordManager interface for 1Password CLI.
//
// This package provides integration with 1Password's command-line interface (op),
// allowing ACM to detect and remediate compromised credentials stored in 1Password
// vaults while maintaining zero-knowledge security.
//
// # 1Password CLI Requirements
//
// Installation:
//
//	# macOS
//	brew install 1password-cli
//
//	# Linux
//	wget https://downloads.1password.com/linux/tar/stable/x86_64/1password-cli-latest-linux-amd64.tar.gz
//	tar -xzf 1password-cli-latest-linux-amd64.tar.gz
//	sudo mv op /usr/local/bin/
//
//	# Windows
//	choco install 1password-cli
//
// Authentication:
//
// The user must authenticate with 1Password before using ACM:
//
//	eval $(op signin)
//
// Or use biometric unlock if configured.
//
// # CLI Commands Used
//
// Detect Compromised:
//
//	op item list --categories=Login --format=json | jq 'map(select(.overview.tags // [] | contains(["compromised"])))'
//
// Get Credential:
//
//	op item get <uuid> --format=json
//
// Update Password:
//
//	op item edit <uuid> password="<new_password>"
//
// Verify Update:
//
//	op item get <uuid> --fields label=password
//
// # Known Limitations
//
//   - Requires network connectivity for some operations (vault sync)
//   - Session tokens expire after 30 minutes of inactivity
//   - Biometric unlock only available on supported platforms
//   - No native breach detection (relies on Watchtower feature)
//
// # Example Usage
//
//	import "github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager/onepassword"
//
//	ctx := context.Background()
//	op := onepassword.New("/usr/local/bin/op")
//
//	// Check if CLI is available
//	available, err := op.IsAvailable(ctx)
//	if !available {
//	    log.Fatal("1Password CLI not found")
//	}
//
//	// Detect compromised credentials
//	compromised, err := op.DetectCompromised(ctx)
//	if err != nil {
//	    log.Fatalf("Detection failed: %v", err)
//	}
//
//	// Update password for compromised credential
//	for _, cred := range compromised {
//	    newPass := generateSecurePassword()
//	    err := op.UpdatePassword(ctx, cred.ID, newPass)
//	    if err != nil {
//	        log.Printf("Update failed for %s: %v", cred.Site, err)
//	    }
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - Basic 1Password CLI integration
//   - Detect compromised via Watchtower
//   - Password update operations
//   - Session management
//   - Error handling
//
// Future phases will add:
//   - Service account token support
//   - Connect server integration
//   - Advanced filtering and search
//   - Batch operations
package onepassword
