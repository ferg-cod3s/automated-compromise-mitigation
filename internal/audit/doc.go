// Package audit implements the cryptographically signed audit logging system.
//
// The Audit Logger provides tamper-evident logging of all credential rotation
// operations, compliance checks, and HIM events. All entries are signed with
// Ed25519 signatures and stored in a local SQLite database.
//
// # Database Schema
//
// The audit log uses SQLite with the following schema:
//
//	CREATE TABLE audit_events (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    timestamp TEXT NOT NULL,
//	    event_type TEXT NOT NULL,
//	    credential_id_hash TEXT NOT NULL,
//	    action TEXT NOT NULL,
//	    status TEXT NOT NULL,
//	    details TEXT,
//	    crc_rule_applied TEXT,
//	    evidence_chain_id TEXT,
//	    signature TEXT NOT NULL,
//	    created_at INTEGER NOT NULL
//	);
//
// # Cryptographic Signing
//
// Each audit entry is signed using Ed25519:
//
//   - Signature computed over: ID || Timestamp || CredentialIDHash || Action
//   - Private key stored in OS keychain / secure enclave
//   - Public key used for verification during integrity checks
//   - Merkle tree structure for efficient batch verification (future)
//
// # Privacy Protection
//
//   - Credential IDs are always SHA-256 hashed before storage
//   - Passwords and tokens are NEVER logged (even encrypted)
//   - Sensitive details fields are encrypted with AES-256-GCM
//   - User input is sanitized and masked
//
// # Compliance Reporting
//
// The audit log supports exporting compliance reports in multiple formats:
//
//   - JSON: Machine-readable format for automation
//   - PDF: Human-readable format with evidence chain (requires ACVS)
//   - CSV: Spreadsheet-compatible format for analysis
//   - HTML: Web-viewable format with filtering
//
// # Example Usage
//
//	ctx := context.Background()
//	logger := audit.NewLogger("/home/user/.acm/data/audit.db", signingKey)
//
//	// Log a rotation event
//	event := audit.AuditEvent{
//	    Timestamp:        time.Now(),
//	    EventType:        audit.EventRotation,
//	    CredentialIDHash: hashCredentialID(credID),
//	    Action:           audit.ActionRotated,
//	    Status:           audit.StatusSuccess,
//	    Details: map[string]interface{}{
//	        "site":     "github.com",
//	        "method":   "auto",
//	        "duration": "2.3s",
//	    },
//	}
//	eventID, sig, err := logger.Log(ctx, event)
//	if err != nil {
//	    log.Fatalf("Audit log failed: %v", err)
//	}
//	log.Printf("Logged event %s with signature %s", eventID, sig)
//
//	// Query recent events
//	filter := audit.AuditFilter{
//	    FromTime:   time.Now().Add(-7 * 24 * time.Hour),
//	    EventTypes: []audit.EventType{audit.EventRotation},
//	    Statuses:   []audit.Status{audit.StatusSuccess},
//	    Limit:      100,
//	}
//	events, _ := logger.Query(ctx, filter)
//	log.Printf("Found %d rotation events in last 7 days", len(events))
//
//	// Verify integrity
//	valid, errors, _ := logger.VerifyIntegrity(ctx, sevenDaysAgo, time.Now())
//	if !valid {
//	    log.Fatalf("Audit log integrity check failed: %v", errors)
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - SQLite storage with Ed25519 signatures
//   - Basic event logging and querying
//   - JSON and CSV export
//   - Integrity verification
//
// Future phases will add:
//   - PDF report generation with evidence chain
//   - Merkle tree for efficient batch verification
//   - Encrypted storage of sensitive details
//   - Real-time log streaming to clients
package audit
