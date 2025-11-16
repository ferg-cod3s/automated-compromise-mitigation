// Package crs implements the Credential Remediation Service (CRS).
//
// The CRS is the core module responsible for detecting compromised credentials
// and performing safe, local vault updates. It maintains zero-knowledge security
// by never accessing master passwords or vault encryption keys.
//
// # Architecture
//
// The CRS follows a five-phase workflow for credential rotation:
//
//  1. Detection Phase: Query password manager CLI for breach reports
//  2. Analysis Phase: Extract metadata and determine rotation strategy
//  3. Generation Phase: Generate secure passwords using crypto/rand
//  4. Update Phase: Execute vault update via CLI and verify success
//  5. Audit Phase: Log rotation event to cryptographically signed audit trail
//
// # Security Controls
//
// The CRS implements multiple security layers:
//
//   - Memory Protection: syscall.Mlock() on password buffers to prevent swap to disk
//   - Explicit Zeroing: Sensitive data cleared from RAM after use
//   - Subprocess Isolation: CLI executed in isolated context with minimal environment
//   - Atomic Transactions: Vault state verified before and after updates
//   - Rollback Capability: Encrypted backup of old password until verification succeeds
//
// # Example Usage
//
//	ctx := context.Background()
//	crs := crs.NewService(pwManager, auditLogger)
//
//	// Detect compromised credentials
//	compromised, err := crs.DetectCompromised(ctx)
//	if err != nil {
//	    log.Fatalf("Detection failed: %v", err)
//	}
//
//	// Rotate each credential
//	for _, cred := range compromised {
//	    policy := pwmanager.DefaultPasswordPolicy()
//	    newPass, _ := crs.GeneratePassword(ctx, policy)
//	    result, err := crs.RotateCredential(ctx, cred, newPass)
//	    if err != nil {
//	        log.Printf("Rotation failed for %s: %v", cred.Site, err)
//	        continue
//	    }
//	    log.Printf("Rotated %s successfully", cred.Site)
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - 1Password CLI integration
//   - Basic rotation workflow
//   - Local audit logging
//   - Error handling and retry logic
//
// Future phases will add:
//   - Multi-password manager support (Bitwarden, LastPass)
//   - ACVS compliance validation integration
//   - Advanced HIM workflows
//   - Batch rotation optimization
package crs
