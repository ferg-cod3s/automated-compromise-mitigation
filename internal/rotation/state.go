// Package rotation provides common types and utilities for credential rotation.
package rotation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// StateStore provides persistence for rotation state.
type StateStore interface {
	SaveState(ctx context.Context, state RotationState) error
	GetState(ctx context.Context, id string) (RotationState, error)
	ListStates(ctx context.Context, filter StateFilter) ([]RotationState, error)
	DeleteState(ctx context.Context, id string) error
	CleanupExpired(ctx context.Context) (int, error)
}

// SQLiteStateStore implements StateStore using SQLite.
type SQLiteStateStore struct {
	db *sql.DB
}

// NewSQLiteStateStore creates a new SQLite-backed state store.
func NewSQLiteStateStore(db *sql.DB) (*SQLiteStateStore, error) {
	store := &SQLiteStateStore{db: db}

	// Initialize schema if needed
	if err := store.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the rotation_state table if it doesn't exist.
func (s *SQLiteStateStore) initSchema(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS rotation_state (
    id TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    state TEXT NOT NULL,
    started_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    metadata_json TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_rotation_credential_id ON rotation_state(credential_id);
CREATE INDEX IF NOT EXISTS idx_rotation_provider ON rotation_state(provider);
CREATE INDEX IF NOT EXISTS idx_rotation_state ON rotation_state(state);
CREATE INDEX IF NOT EXISTS idx_rotation_expires_at ON rotation_state(expires_at);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// SaveState saves or updates a rotation state.
func (s *SQLiteStateStore) SaveState(ctx context.Context, state RotationState) error {
	// Serialize metadata
	metadataJSON, err := json.Marshal(state.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
INSERT OR REPLACE INTO rotation_state (
    id, credential_id, provider, state,
    started_at, updated_at, expires_at, metadata_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		state.ID,
		state.CredentialID,
		state.Provider,
		state.State,
		state.StartedAt.Unix(),
		state.UpdatedAt.Unix(),
		state.ExpiresAt.Unix(),
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to save rotation state: %w", err)
	}

	return nil
}

// GetState retrieves a rotation state by ID.
func (s *SQLiteStateStore) GetState(ctx context.Context, id string) (RotationState, error) {
	query := `
SELECT id, credential_id, provider, state,
       started_at, updated_at, expires_at, metadata_json
FROM rotation_state
WHERE id = ?
	`

	var state RotationState
	var metadataJSON string
	var startedAt, updatedAt, expiresAt int64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&state.ID,
		&state.CredentialID,
		&state.Provider,
		&state.State,
		&startedAt,
		&updatedAt,
		&expiresAt,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return state, fmt.Errorf("rotation state not found: %s", id)
	}

	if err != nil {
		return state, fmt.Errorf("failed to query rotation state: %w", err)
	}

	// Deserialize timestamps
	state.StartedAt = time.Unix(startedAt, 0)
	state.UpdatedAt = time.Unix(updatedAt, 0)
	state.ExpiresAt = time.Unix(expiresAt, 0)

	// Deserialize metadata
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &state.Metadata); err != nil {
			return state, fmt.Errorf("failed to deserialize metadata: %w", err)
		}
	} else {
		state.Metadata = make(map[string]string)
	}

	return state, nil
}

// ListStates retrieves rotation states matching the filter.
func (s *SQLiteStateStore) ListStates(ctx context.Context, filter StateFilter) ([]RotationState, error) {
	query := `
SELECT id, credential_id, provider, state,
       started_at, updated_at, expires_at, metadata_json
FROM rotation_state
WHERE 1=1
	`

	args := []interface{}{}

	// Apply filters
	if filter.CredentialID != "" {
		query += " AND credential_id = ?"
		args = append(args, filter.CredentialID)
	}

	if filter.Provider != "" {
		query += " AND provider = ?"
		args = append(args, filter.Provider)
	}

	if filter.State != "" {
		query += " AND state = ?"
		args = append(args, filter.State)
	}

	if len(filter.ExcludeStates) > 0 {
		placeholders := ""
		for i, state := range filter.ExcludeStates {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, state)
		}
		query += fmt.Sprintf(" AND state NOT IN (%s)", placeholders)
	}

	if filter.OnlyExpired {
		query += " AND expires_at <= ?"
		args = append(args, time.Now().Unix())
	}

	query += " ORDER BY started_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rotation states: %w", err)
	}
	defer rows.Close()

	var states []RotationState

	for rows.Next() {
		var state RotationState
		var metadataJSON string
		var startedAt, updatedAt, expiresAt int64

		if err := rows.Scan(
			&state.ID,
			&state.CredentialID,
			&state.Provider,
			&state.State,
			&startedAt,
			&updatedAt,
			&expiresAt,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Deserialize timestamps
		state.StartedAt = time.Unix(startedAt, 0)
		state.UpdatedAt = time.Unix(updatedAt, 0)
		state.ExpiresAt = time.Unix(expiresAt, 0)

		// Deserialize metadata
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &state.Metadata); err != nil {
				return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
			}
		} else {
			state.Metadata = make(map[string]string)
		}

		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return states, nil
}

// DeleteState deletes a rotation state by ID.
func (s *SQLiteStateStore) DeleteState(ctx context.Context, id string) error {
	query := "DELETE FROM rotation_state WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rotation state: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("rotation state not found: %s", id)
	}

	return nil
}

// CleanupExpired removes all expired rotation states.
// Returns the number of states deleted.
func (s *SQLiteStateStore) CleanupExpired(ctx context.Context) (int, error) {
	query := "DELETE FROM rotation_state WHERE expires_at <= ?"

	result, err := s.db.ExecContext(ctx, query, time.Now().Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired states: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rows), nil
}

// GetStats returns statistics about rotation states.
func (s *SQLiteStateStore) GetStats(ctx context.Context) (StateStats, error) {
	query := `
SELECT
    COUNT(*) as total,
    SUM(CASE WHEN state = 'complete' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN state = 'failed' THEN 1 ELSE 0 END) as failed,
    SUM(CASE WHEN expires_at <= ? THEN 1 ELSE 0 END) as expired
FROM rotation_state
	`

	var stats StateStats
	err := s.db.QueryRowContext(ctx, query, time.Now().Unix()).Scan(
		&stats.Total,
		&stats.Completed,
		&stats.Failed,
		&stats.Expired,
	)

	if err != nil {
		return stats, fmt.Errorf("failed to get stats: %w", err)
	}

	stats.InProgress = stats.Total - stats.Completed - stats.Failed

	return stats, nil
}

// StateStats provides statistics about rotation states.
type StateStats struct {
	Total      int
	InProgress int
	Completed  int
	Failed     int
	Expired    int
}
