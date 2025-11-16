// Package audit implements cryptographically-signed audit logging for ACM.
//
// All ACM operations are logged to a local SQLite database with Ed25519 signatures
// to provide tamper-evidence. The audit log serves as the evidence chain for
// compliance validation (Phase II ACVS).
package audit

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Logger provides audit logging capabilities with cryptographic signatures.
type Logger interface {
	// LogEvent logs an event to the audit trail with a cryptographic signature.
	LogEvent(ctx context.Context, event Event) error

	// QueryEvents retrieves events matching the specified filter.
	QueryEvents(ctx context.Context, filter Filter) ([]Event, error)

	// VerifyIntegrity verifies the cryptographic signature of an event.
	VerifyIntegrity(ctx context.Context, eventID string) (bool, error)

	// ExportReport generates a compliance report for the specified time range.
	ExportReport(ctx context.Context, filter Filter, format ReportFormat) ([]byte, error)

	// Close closes the audit logger and releases resources.
	Close() error
}

// SQLiteLogger implements Logger using SQLite for storage.
type SQLiteLogger struct {
	db         *sql.DB
	signingKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewSQLiteLogger creates a new SQLite-backed audit logger.
// If dbPath is empty, uses an in-memory database.
func NewSQLiteLogger(dbPath string) (*SQLiteLogger, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Generate or load signing keys
	publicKey, privateKey, err := loadOrGenerateKeys(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}

	return &SQLiteLogger{
		db:         db,
		signingKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// LogEvent logs an event to the audit trail with a cryptographic signature.
func (l *SQLiteLogger) LogEvent(ctx context.Context, event Event) error {
	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create signature
	message := fmt.Sprintf("%s|%d|%s|%s|%s",
		event.ID,
		event.Timestamp.Unix(),
		event.CredentialID,
		event.Type,
		event.Status)
	signature := ed25519.Sign(l.signingKey, []byte(message))

	// Insert into database
	query := `
		INSERT INTO audit_events (
			id, timestamp, event_type, status, credential_id,
			site, username, message, metadata, signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = l.db.ExecContext(ctx, query,
		event.ID,
		event.Timestamp.Unix(),
		event.Type,
		event.Status,
		event.CredentialID,
		event.Site,
		event.Username,
		event.Message,
		string(metadataJSON),
		hex.EncodeToString(signature),
	)

	if err != nil {
		return fmt.Errorf("failed to insert audit event: %w", err)
	}

	return nil
}

// QueryEvents retrieves events matching the specified filter.
func (l *SQLiteLogger) QueryEvents(ctx context.Context, filter Filter) ([]Event, error) {
	query := "SELECT id, timestamp, event_type, status, credential_id, site, username, message, metadata, signature FROM audit_events WHERE 1=1"
	args := []interface{}{}

	if filter.EventType != "" {
		query += " AND event_type = ?"
		args = append(args, filter.EventType)
	}

	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if !filter.StartTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime.Unix())
	}

	if !filter.EndTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime.Unix())
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var timestamp int64
		var metadataJSON string
		var signatureHex string

		err := rows.Scan(
			&event.ID,
			&timestamp,
			&event.Type,
			&event.Status,
			&event.CredentialID,
			&event.Site,
			&event.Username,
			&event.Message,
			&metadataJSON,
			&signatureHex,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		event.Timestamp = time.Unix(timestamp, 0)

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		signature, err := hex.DecodeString(signatureHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature: %w", err)
		}
		event.Signature = signature

		events = append(events, event)
	}

	return events, nil
}

// VerifyIntegrity verifies the cryptographic signature of an event.
func (l *SQLiteLogger) VerifyIntegrity(ctx context.Context, eventID string) (bool, error) {
	events, err := l.QueryEvents(ctx, Filter{})
	if err != nil {
		return false, err
	}

	for _, event := range events {
		if event.ID == eventID {
			// Recreate the message
			message := fmt.Sprintf("%s|%d|%s|%s|%s",
				event.ID,
				event.Timestamp.Unix(),
				event.CredentialID,
				event.Type,
				event.Status)

			// Verify signature
			return ed25519.Verify(l.publicKey, []byte(message), event.Signature), nil
		}
	}

	return false, fmt.Errorf("event not found: %s", eventID)
}

// ExportReport generates a compliance report for the specified time range.
func (l *SQLiteLogger) ExportReport(ctx context.Context, filter Filter, format ReportFormat) ([]byte, error) {
	events, err := l.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case ReportFormatJSON:
		return json.MarshalIndent(events, "", "  ")
	case ReportFormatCSV:
		return exportCSV(events)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Close closes the audit logger.
func (l *SQLiteLogger) Close() error {
	return l.db.Close()
}

// Helper functions

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS audit_events (
		id TEXT PRIMARY KEY,
		timestamp INTEGER NOT NULL,
		event_type TEXT NOT NULL,
		status TEXT NOT NULL,
		credential_id TEXT,
		site TEXT,
		username TEXT,
		message TEXT,
		metadata TEXT,
		signature TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON audit_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_credential_id ON audit_events(credential_id);
	CREATE INDEX IF NOT EXISTS idx_event_type ON audit_events(event_type);

	CREATE TABLE IF NOT EXISTS signing_keys (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		public_key TEXT NOT NULL,
		private_key TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);
	`

	_, err := db.Exec(schema)
	return err
}

func loadOrGenerateKeys(db *sql.DB) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// Try to load existing keys
	var publicKeyHex, privateKeyHex string
	err := db.QueryRow("SELECT public_key, private_key FROM signing_keys WHERE id = 1").
		Scan(&publicKeyHex, &privateKeyHex)

	if err == nil {
		// Keys exist, decode them
		publicKey, err := hex.DecodeString(publicKeyHex)
		if err != nil {
			return nil, nil, err
		}

		privateKey, err := hex.DecodeString(privateKeyHex)
		if err != nil {
			return nil, nil, err
		}

		return ed25519.PublicKey(publicKey), ed25519.PrivateKey(privateKey), nil
	}

	// Generate new keys
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Store keys
	_, err = db.Exec(
		"INSERT INTO signing_keys (id, public_key, private_key, created_at) VALUES (1, ?, ?, ?)",
		hex.EncodeToString(publicKey),
		hex.EncodeToString(privateKey),
		time.Now().Unix(),
	)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

func generateEventID() string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:16])
}

func exportCSV(events []Event) ([]byte, error) {
	csv := "ID,Timestamp,Type,Status,CredentialID,Site,Username,Message\n"
	for _, event := range events {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%q\n",
			event.ID,
			event.Timestamp.Format(time.RFC3339),
			event.Type,
			event.Status,
			event.CredentialID,
			event.Site,
			event.Username,
			event.Message,
		)
	}
	return []byte(csv), nil
}
