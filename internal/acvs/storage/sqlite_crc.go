// Package storage provides SQLite-based persistent storage for ACVS.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvsif"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "modernc.org/sqlite" // SQLite driver
)

// SQLiteCRCManager implements the CRCManager interface with SQLite persistence.
type SQLiteCRCManager struct {
	db       *sql.DB
	cacheTTL time.Duration
}

// NewSQLiteCRCManager creates a new SQLite-backed CRC manager.
func NewSQLiteCRCManager(db *sql.DB, cacheTTL time.Duration) *SQLiteCRCManager {
	return &SQLiteCRCManager{
		db:       db,
		cacheTTL: cacheTTL,
	}
}

// Store saves a CRC to the database.
func (m *SQLiteCRCManager) Store(ctx context.Context, crc *acmv1.ComplianceRuleSet) error {
	if crc == nil {
		return fmt.Errorf("cannot store nil CRC")
	}

	if crc.Site == "" {
		return fmt.Errorf("CRC site cannot be empty")
	}

	now := time.Now()
	expiresAt := now.Add(m.cacheTTL)

	// Generate ID if not set
	if crc.Id == "" {
		crc.Id = generateCRCID(crc)
	}

	// Set timestamps if not set
	if crc.ParsedAt == nil {
		crc.ParsedAt = timestamppb.New(now)
	}

	// Set expiration
	crc.ExpiresAt = timestamppb.New(expiresAt)

	// Serialize rules to JSON
	rulesJSON, err := json.Marshal(crc.Rules)
	if err != nil {
		return fmt.Errorf("failed to serialize rules: %w", err)
	}

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert or replace CRC
	query := `
		INSERT OR REPLACE INTO crcs (
			id, site, tos_url, tos_version, tos_hash,
			parsed_at, expires_at, stored_at,
			recommendation, reasoning, rules_json, signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.ExecContext(ctx, query,
		crc.Id,
		crc.Site,
		crc.TosUrl,
		crc.TosVersion,
		crc.TosHash,
		crc.ParsedAt.AsTime().Unix(),
		crc.ExpiresAt.AsTime().Unix(),
		now.Unix(),
		int32(crc.Recommendation),
		crc.Reasoning,
		string(rulesJSON),
		crc.Signature,
	)

	if err != nil {
		return fmt.Errorf("failed to insert CRC: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Get retrieves a CRC from the database.
// Returns (crc, found, error). If found=false, crc will be nil.
func (m *SQLiteCRCManager) Get(ctx context.Context, site string) (*acmv1.ComplianceRuleSet, bool, error) {
	if site == "" {
		return nil, false, fmt.Errorf("site cannot be empty")
	}

	query := `
		SELECT
			id, site, tos_url, tos_version, tos_hash,
			parsed_at, expires_at,
			recommendation, reasoning, rules_json, signature
		FROM crcs
		WHERE site = ?
		  AND expires_at > ?
		ORDER BY parsed_at DESC
		LIMIT 1
	`

	now := time.Now().Unix()

	var (
		id, tosURL, tosVersion, tosHash, reasoning, rulesJSON, signature string
		parsedAt, expiresAt                                              int64
		recommendation                                                   int32
	)

	err := m.db.QueryRowContext(ctx, query, site, now).Scan(
		&id, &site, &tosURL, &tosVersion, &tosHash,
		&parsedAt, &expiresAt,
		&recommendation, &reasoning, &rulesJSON, &signature,
	)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to query CRC: %w", err)
	}

	// Deserialize rules
	var rules []*acmv1.ComplianceRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return nil, false, fmt.Errorf("failed to deserialize rules: %w", err)
	}

	// Construct CRC
	crc := &acmv1.ComplianceRuleSet{
		Id:             id,
		Site:           site,
		TosUrl:         tosURL,
		TosVersion:     tosVersion,
		TosHash:        tosHash,
		ParsedAt:       timestamppb.New(time.Unix(parsedAt, 0)),
		ExpiresAt:      timestamppb.New(time.Unix(expiresAt, 0)),
		Rules:          rules,
		Recommendation: acmv1.ComplianceRecommendation(recommendation),
		Reasoning:      reasoning,
		Signature:      signature,
	}

	return crc, true, nil
}

// List returns all cached CRCs matching the filter.
func (m *SQLiteCRCManager) List(ctx context.Context, siteFilter string, includeExpired bool) ([]acvsif.CRCSummary, error) {
	query := `
		SELECT
			id, site, parsed_at, expires_at, recommendation,
			(SELECT COUNT(*) FROM json_each(rules_json)) as rule_count
		FROM crcs
		WHERE 1=1
	`

	args := []interface{}{}

	// Apply site filter
	if siteFilter != "" {
		query += " AND site = ?"
		args = append(args, siteFilter)
	}

	// Apply expiration filter
	if !includeExpired {
		query += " AND expires_at > ?"
		args = append(args, time.Now().Unix())
	}

	query += " ORDER BY parsed_at DESC"

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query CRCs: %w", err)
	}
	defer rows.Close()

	var summaries []acvsif.CRCSummary
	now := time.Now()

	for rows.Next() {
		var (
			id, site                string
			parsedAt, expiresAt     int64
			recommendation, ruleCount int32
		)

		if err := rows.Scan(&id, &site, &parsedAt, &expiresAt, &recommendation, &ruleCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		expiresAtTime := time.Unix(expiresAt, 0)
		expired := now.After(expiresAtTime)

		summary := acvsif.CRCSummary{
			ID:             id,
			Site:           site,
			ParsedAt:       time.Unix(parsedAt, 0),
			ExpiresAt:      expiresAtTime,
			Recommendation: acmv1.ComplianceRecommendation(recommendation),
			RuleCount:      ruleCount,
			Expired:        expired,
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return summaries, nil
}

// Invalidate removes a CRC from the database.
func (m *SQLiteCRCManager) Invalidate(ctx context.Context, site string) error {
	if site == "" {
		return fmt.Errorf("site cannot be empty")
	}

	query := "DELETE FROM crcs WHERE site = ?"

	result, err := m.db.ExecContext(ctx, query, site)
	if err != nil {
		return fmt.Errorf("failed to invalidate CRC: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Not an error, just no CRC found
		return nil
	}

	return nil
}

// IsExpired checks if a CRC has expired.
func (m *SQLiteCRCManager) IsExpired(crc *acmv1.ComplianceRuleSet) bool {
	if crc == nil || crc.ExpiresAt == nil {
		return true
	}

	return time.Now().After(crc.ExpiresAt.AsTime())
}

// GetCacheTTL returns the configured cache TTL.
func (m *SQLiteCRCManager) GetCacheTTL() time.Duration {
	return m.cacheTTL
}

// SetCacheTTL sets the cache TTL.
// Note: This does not affect existing CRCs, only new ones.
func (m *SQLiteCRCManager) SetCacheTTL(ttl time.Duration) {
	m.cacheTTL = ttl
}

// GetCacheStats returns statistics about the CRC cache.
func (m *SQLiteCRCManager) GetCacheStats(ctx context.Context) (CacheStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN expires_at > ? THEN 1 ELSE 0 END) as valid,
			SUM(CASE WHEN expires_at <= ? THEN 1 ELSE 0 END) as expired
		FROM crcs
	`

	now := time.Now().Unix()

	var total, valid, expired int

	err := m.db.QueryRowContext(ctx, query, now, now).Scan(&total, &valid, &expired)
	if err != nil {
		return CacheStats{}, fmt.Errorf("failed to get cache stats: %w", err)
	}

	return CacheStats{
		TotalEntries:   total,
		ValidEntries:   valid,
		ExpiredEntries: expired,
	}, nil
}

// CleanExpired removes all expired CRCs from the database.
// Returns the number of entries removed.
func (m *SQLiteCRCManager) CleanExpired(ctx context.Context) (int, error) {
	query := "DELETE FROM crcs WHERE expires_at <= ?"

	result, err := m.db.ExecContext(ctx, query, time.Now().Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired CRCs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// Clear removes all CRCs from the database.
// WARNING: This is destructive and should only be used for testing or when
// the user explicitly disables ACVS with clear_cache=true.
func (m *SQLiteCRCManager) Clear(ctx context.Context) error {
	query := "DELETE FROM crcs"

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clear CRCs: %w", err)
	}

	return nil
}

// Size returns the number of CRCs in the database.
func (m *SQLiteCRCManager) Size(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM crcs"

	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get size: %w", err)
	}

	return count, nil
}

// CacheStats provides cache statistics.
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	ExpiredEntries int
}

// generateCRCID generates a unique ID for a CRC based on site and ToS hash.
// This is extracted from the in-memory implementation.
func generateCRCID(crc *acmv1.ComplianceRuleSet) string {
	// Format: CRC-{site}-{short-hash}
	// Use ToS hash as the unique identifier
	shortHash := crc.TosHash
	if len(shortHash) > 16 {
		shortHash = shortHash[:16]
	}
	return fmt.Sprintf("CRC-%s-%s", crc.Site, shortHash)
}
