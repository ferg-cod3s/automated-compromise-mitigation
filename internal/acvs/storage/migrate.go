// Package storage provides SQLite-based persistent storage for ACVS.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a database migration.
type Migration struct {
	Version     int
	Description string
	SQL         string
	Checksum    string
}

// runMigrations applies all pending migrations to the database.
func runMigrations(ctx context.Context, db *sql.DB, config Config) error {
	// Get current schema version
	currentVersion, err := GetSchemaVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Find all available migrations
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Filter pending migrations
	var pending []Migration
	for _, m := range migrations {
		if m.Version > currentVersion {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		// No pending migrations
		return nil
	}

	// Create backup before migrations
	if config.BackupPath != "" {
		backupFile := filepath.Join(config.BackupPath, fmt.Sprintf("acm-migration-v%d-%d.db",
			currentVersion, time.Now().Unix()))
		if err := CreateBackup(ctx, db, backupFile); err != nil {
			return fmt.Errorf("failed to create pre-migration backup: %w", err)
		}
	}

	// Apply each pending migration
	for _, migration := range pending {
		if err := applyMigration(ctx, db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// applyMigration applies a single migration to the database.
func applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in schema_version table
	query := `
		INSERT INTO schema_version (version, applied_at, description, checksum)
		VALUES (?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, query,
		migration.Version,
		time.Now().Unix(),
		migration.Description,
		migration.Checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// loadMigrations loads all migration files from the migrations directory.
func loadMigrations() ([]Migration, error) {
	// Get the migrations directory path
	// In a real deployment, this would be embedded or use a fixed path
	// For now, we'll look in internal/acvs/storage/migrations/
	migrationsDir := "internal/acvs/storage/migrations"

	// Check if directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// No migrations directory, return empty list
		return []Migration{}, nil
	}

	var migrations []Migration

	// Walk migrations directory
	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-SQL files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		// Parse migration filename (format: 001_description.sql)
		parts := strings.SplitN(d.Name(), "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration filename: %s", d.Name())
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid migration version in %s: %w", d.Name(), err)
		}

		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		// Read migration SQL
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", d.Name(), err)
		}

		// Calculate checksum
		checksum := calculateChecksum(string(content))

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
			Checksum:    checksum,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedMigrations returns all applied migrations from the database.
func GetAppliedMigrations(ctx context.Context, db *sql.DB) ([]Migration, error) {
	query := `
		SELECT version, applied_at, description, checksum
		FROM schema_version
		ORDER BY version ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var migrations []Migration

	for rows.Next() {
		var m Migration
		var appliedAt int64

		if err := rows.Scan(&m.Version, &appliedAt, &m.Description, &m.Checksum); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}

		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migrations: %w", err)
	}

	return migrations, nil
}

// RollbackMigration rolls back the most recent migration.
// WARNING: This is a destructive operation and should only be used in development.
func RollbackMigration(ctx context.Context, db *sql.DB, targetVersion int) error {
	// Get current version
	currentVersion, err := GetSchemaVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion <= targetVersion {
		return fmt.Errorf("already at or below target version %d", targetVersion)
	}

	// WARNING: In a production system, you would need to implement proper
	// rollback SQL for each migration. For Phase III, we'll just delete
	// the schema_version entry and warn the user.

	return fmt.Errorf("migration rollback not fully implemented in Phase III - manual rollback required")
}

// VerifyMigrations verifies that all applied migrations match their checksums.
func VerifyMigrations(ctx context.Context, db *sql.DB) (bool, []string, error) {
	applied, err := GetAppliedMigrations(ctx, db)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	available, err := loadMigrations()
	if err != nil {
		return false, nil, fmt.Errorf("failed to load available migrations: %w", err)
	}

	// Create map of available migrations
	availableMap := make(map[int]Migration)
	for _, m := range available {
		availableMap[m.Version] = m
	}

	var errors []string

	for _, appliedMigration := range applied {
		availableMigration, exists := availableMap[appliedMigration.Version]
		if !exists {
			errors = append(errors, fmt.Sprintf("Migration %d: applied but not found in available migrations",
				appliedMigration.Version))
			continue
		}

		if appliedMigration.Checksum != availableMigration.Checksum {
			errors = append(errors, fmt.Sprintf("Migration %d: checksum mismatch (expected %s, got %s)",
				appliedMigration.Version, availableMigration.Checksum, appliedMigration.Checksum))
		}
	}

	return len(errors) == 0, errors, nil
}
