// Package storage provides SQLite-based persistent storage for ACVS.
package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

//go:embed schema.sql
var schemaSQL string

const (
	// DefaultDatabasePath is the default location for the ACM database.
	DefaultDatabasePath = "~/.acm/data/acm.db"

	// DefaultCacheTTL is the default cache lifetime for CRCs (30 days).
	DefaultCacheTTL = 30 * 24 * time.Hour

	// CurrentSchemaVersion is the current schema version.
	CurrentSchemaVersion = 1
)

// Config holds database configuration.
type Config struct {
	// Path to the SQLite database file.
	Path string

	// CacheTTL for CRCs.
	CacheTTL time.Duration

	// BackupPath for automatic backups.
	BackupPath string

	// EnableWAL enables Write-Ahead Logging (recommended).
	EnableWAL bool

	// MaxOpenConns sets the maximum number of open connections.
	MaxOpenConns int

	// BusyTimeout sets the busy timeout in milliseconds.
	BusyTimeout int
}

// DefaultConfig returns the default database configuration.
func DefaultConfig() Config {
	return Config{
		Path:         expandPath(DefaultDatabasePath),
		CacheTTL:     DefaultCacheTTL,
		BackupPath:   expandPath("~/.acm/data/backups"),
		EnableWAL:    true,
		MaxOpenConns: 10,
		BusyTimeout:  5000, // 5 seconds
	}
}

// Initialize initializes the database and returns a connection.
// This function:
// 1. Creates the database file and parent directories if needed
// 2. Opens the database connection
// 3. Configures SQLite pragmas
// 4. Creates the initial schema if database is new
// 5. Runs pending migrations
func Initialize(ctx context.Context, config Config) (*sql.DB, error) {
	// Create parent directories if needed
	dbDir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Create backup directory if needed
	if config.BackupPath != "" {
		if err := os.MkdirAll(config.BackupPath, 0700); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}
	}

	// Check if database file exists (new database)
	isNewDB := !fileExists(config.Path)

	// Open database connection
	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxOpenConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Configure SQLite pragmas
	if err := configurePragmas(ctx, db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure pragmas: %w", err)
	}

	// Initialize schema for new database
	if isNewDB {
		if err := createInitialSchema(ctx, db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create initial schema: %w", err)
		}
	}

	// Run pending migrations
	if err := runMigrations(ctx, db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Verify database integrity
	if err := verifyIntegrity(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("database integrity check failed: %w", err)
	}

	return db, nil
}

// configurePragmas sets SQLite pragmas for optimal performance and safety.
func configurePragmas(ctx context.Context, db *sql.DB, config Config) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
		fmt.Sprintf("PRAGMA busy_timeout = %d", config.BusyTimeout),
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA auto_vacuum = INCREMENTAL",
		"PRAGMA temp_store = MEMORY",
	}

	// Enable WAL mode if configured
	if config.EnableWAL {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	return nil
}

// createInitialSchema creates the initial database schema.
func createInitialSchema(ctx context.Context, db *sql.DB) error {
	// Execute schema SQL
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	// Record schema version
	checksum := calculateChecksum(schemaSQL)
	query := `
		INSERT INTO schema_version (version, applied_at, description, checksum)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(version) DO NOTHING
	`

	_, err := db.ExecContext(ctx, query,
		CurrentSchemaVersion,
		time.Now().Unix(),
		"Initial Phase III schema with CRCs, Evidence Chain, and Audit Events",
		checksum,
	)

	if err != nil {
		return fmt.Errorf("failed to record schema version: %w", err)
	}

	return nil
}

// verifyIntegrity runs SQLite integrity checks on the database.
func verifyIntegrity(ctx context.Context, db *sql.DB) error {
	var result string
	err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}

	return nil
}

// GetSchemaVersion returns the current schema version from the database.
func GetSchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	err := db.QueryRowContext(ctx, `
		SELECT MAX(version) FROM schema_version
	`).Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// CreateBackup creates a backup of the database.
func CreateBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Use SQLite backup API
	query := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Verify backup integrity
	backupDB, err := sql.Open("sqlite", backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup for verification: %w", err)
	}
	defer backupDB.Close()

	if err := verifyIntegrity(ctx, backupDB); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("backup integrity check failed: %w", err)
	}

	return nil
}

// GetDatabaseStats returns statistics about the database.
func GetDatabaseStats(ctx context.Context, db *sql.DB) (DatabaseStats, error) {
	stats := DatabaseStats{}

	// Get page count and page size
	err := db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&stats.PageCount)
	if err != nil {
		return stats, fmt.Errorf("failed to get page count: %w", err)
	}

	err = db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&stats.PageSize)
	if err != nil {
		return stats, fmt.Errorf("failed to get page size: %w", err)
	}

	stats.SizeBytes = int64(stats.PageCount * stats.PageSize)

	// Get table counts
	tables := []struct {
		name  string
		count *int64
	}{
		{"crcs", &stats.CRCCount},
		{"evidence_entries", &stats.EvidenceCount},
		{"audit_events", &stats.AuditEventCount},
	}

	for _, table := range tables {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table.name)
		if err := db.QueryRowContext(ctx, query).Scan(table.count); err != nil {
			return stats, fmt.Errorf("failed to count %s: %w", table.name, err)
		}
	}

	return stats, nil
}

// DatabaseStats provides statistics about the database.
type DatabaseStats struct {
	SizeBytes       int64
	PageCount       int
	PageSize        int
	CRCCount        int64
	EvidenceCount   int64
	AuditEventCount int64
}

// Helper functions

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// calculateChecksum calculates SHA-256 checksum of a string.
func calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
