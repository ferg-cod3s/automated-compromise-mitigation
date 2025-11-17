// Package storage provides SQLite-based persistent storage for ACVS.
package storage

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

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SQLiteEvidenceChainGenerator implements the EvidenceChainGenerator interface with SQLite persistence.
type SQLiteEvidenceChainGenerator struct {
	db         *sql.DB
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// NewSQLiteEvidenceChainGenerator creates a new SQLite-backed evidence chain generator.
func NewSQLiteEvidenceChainGenerator(db *sql.DB) (*SQLiteEvidenceChainGenerator, error) {
	// Generate signing keys
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key: %w", err)
	}

	return &SQLiteEvidenceChainGenerator{
		db:         db,
		publicKey:  pub,
		privateKey: priv,
	}, nil
}

// NewSQLiteEvidenceChainGeneratorWithKeys creates a generator with existing keys.
func NewSQLiteEvidenceChainGeneratorWithKeys(db *sql.DB, privateKey ed25519.PrivateKey) (*SQLiteEvidenceChainGenerator, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	pub := privateKey.Public().(ed25519.PublicKey)

	return &SQLiteEvidenceChainGenerator{
		db:         db,
		publicKey:  pub,
		privateKey: privateKey,
	}, nil
}

// AddEntry adds a new entry to the evidence chain.
func (g *SQLiteEvidenceChainGenerator) AddEntry(ctx context.Context, entry *acvs.EvidenceEntry) (string, error) {
	if entry == nil {
		return "", fmt.Errorf("entry cannot be nil")
	}

	// Begin transaction
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get previous entry ID (chain head)
	var previousID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM evidence_entries ORDER BY timestamp DESC LIMIT 1
	`).Scan(&previousID)

	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to get chain head: %w", err)
	}

	// Generate entry ID
	entryID := generateEntryID(entry)

	// Compute chain hash
	var prevIDStr string
	if previousID.Valid {
		prevIDStr = previousID.String
	}
	chainHash := computeChainHash(entryID, prevIDStr)

	// Serialize action to JSON
	var actionContextJSON string
	if entry.Action != nil && entry.Action.Context != nil {
		contextBytes, err := json.Marshal(entry.Action.Context)
		if err != nil {
			return "", fmt.Errorf("failed to serialize action context: %w", err)
		}
		actionContextJSON = string(contextBytes)
	}

	// Serialize applied rule IDs
	appliedRuleIDsJSON, err := json.Marshal(entry.AppliedRuleIDs)
	if err != nil {
		return "", fmt.Errorf("failed to serialize applied rule IDs: %w", err)
	}

	// Serialize evidence data
	evidenceDataJSON, err := json.Marshal(entry.EvidenceData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize evidence data: %w", err)
	}

	// Create proto entry for signing
	protoEntry := &acmv1.EvidenceChainEntry{
		Id:                entryID,
		Timestamp:         timestamppb.New(time.Now()),
		EventType:         entry.EventType,
		Site:              entry.Site,
		CredentialIdHash:  entry.CredentialIDHash,
		ValidationResult:  entry.ValidationResult,
		CrcId:             entry.CRCID,
		AppliedRuleIds:    entry.AppliedRuleIDs,
		EvidenceData:      string(evidenceDataJSON),
		PreviousEntryId:   prevIDStr,
		ChainHash:         chainHash,
	}

	// Add action to proto entry
	if entry.Action != nil {
		protoEntry.Action = entry.Action
	}

	// Sign the entry
	signature, err := g.signEntry(protoEntry)
	if err != nil {
		return "", fmt.Errorf("failed to sign entry: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO evidence_entries (
			id, timestamp, event_type, site, credential_id_hash,
			action_type, action_method, action_context_json,
			validation_result, crc_id, applied_rule_ids_json,
			evidence_data_json, previous_entry_id, chain_hash, signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var actionType, actionMethod int32
	if entry.Action != nil {
		actionType = int32(entry.Action.Type)
		actionMethod = int32(entry.Action.Method)
	}

	_, err = tx.ExecContext(ctx, query,
		entryID,
		protoEntry.Timestamp.AsTime().Unix(),
		int32(entry.EventType),
		entry.Site,
		entry.CredentialIDHash,
		actionType,
		actionMethod,
		actionContextJSON,
		int32(entry.ValidationResult),
		nullString(entry.CRCID),
		string(appliedRuleIDsJSON),
		string(evidenceDataJSON),
		nullString(prevIDStr),
		chainHash,
		signature,
	)

	if err != nil {
		return "", fmt.Errorf("failed to insert evidence entry: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return entryID, nil
}

// GetEntry retrieves an evidence entry by ID.
func (g *SQLiteEvidenceChainGenerator) GetEntry(ctx context.Context, entryID string) (*acmv1.EvidenceChainEntry, error) {
	query := `
		SELECT
			id, timestamp, event_type, site, credential_id_hash,
			action_type, action_method, action_context_json,
			validation_result, crc_id, applied_rule_ids_json,
			evidence_data_json, previous_entry_id, chain_hash, signature
		FROM evidence_entries
		WHERE id = ?
	`

	var (
		id, site, credentialIDHash, evidenceDataJSON, chainHash, signature string
		actionContextJSON, appliedRuleIDsJSON                              string
		crcID, previousEntryID                                             sql.NullString
		timestamp                                                          int64
		eventType, actionType, actionMethod, validationResult              int32
	)

	err := g.db.QueryRowContext(ctx, query, entryID).Scan(
		&id, &timestamp, &eventType, &site, &credentialIDHash,
		&actionType, &actionMethod, &actionContextJSON,
		&validationResult, &crcID, &appliedRuleIDsJSON,
		&evidenceDataJSON, &previousEntryID, &chainHash, &signature,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry not found: %s", entryID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query entry: %w", err)
	}

	// Deserialize applied rule IDs
	var appliedRuleIDs []string
	if err := json.Unmarshal([]byte(appliedRuleIDsJSON), &appliedRuleIDs); err != nil {
		return nil, fmt.Errorf("failed to deserialize applied rule IDs: %w", err)
	}

	// Construct action
	action := &acmv1.AutomationAction{
		Type:   acmv1.ActionType(actionType),
		Method: acmv1.AutomationMethod(actionMethod),
	}

	if actionContextJSON != "" {
		var context map[string]string
		if err := json.Unmarshal([]byte(actionContextJSON), &context); err != nil {
			return nil, fmt.Errorf("failed to deserialize action context: %w", err)
		}
		action.Context = context
	}

	// Construct evidence entry
	entry := &acmv1.EvidenceChainEntry{
		Id:               id,
		Timestamp:        timestamppb.New(time.Unix(timestamp, 0)),
		EventType:        acmv1.EvidenceEventType(eventType),
		Site:             site,
		CredentialIdHash: credentialIDHash,
		Action:           action,
		ValidationResult: acmv1.ValidationResult(validationResult),
		CrcId:            stringValue(crcID),
		AppliedRuleIds:   appliedRuleIDs,
		EvidenceData:     evidenceDataJSON,
		PreviousEntryId:  stringValue(previousEntryID),
		ChainHash:        chainHash,
		Signature:        signature,
	}

	return entry, nil
}

// Export exports evidence entries for a time range or credential.
func (g *SQLiteEvidenceChainGenerator) Export(ctx context.Context, req *acvs.ExportRequest) ([]*acmv1.EvidenceChainEntry, error) {
	query := `
		SELECT
			id, timestamp, event_type, site, credential_id_hash,
			action_type, action_method, action_context_json,
			validation_result, crc_id, applied_rule_ids_json,
			evidence_data_json, previous_entry_id, chain_hash, signature
		FROM evidence_entries
		WHERE 1=1
	`

	args := []interface{}{}

	// Apply filters
	if req.CredentialID != "" {
		query += " AND credential_id_hash = ?"
		args = append(args, req.CredentialID)
	}

	if !req.StartTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, req.StartTime.Unix())
	}

	if !req.EndTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, req.EndTime.Unix())
	}

	query += " ORDER BY timestamp ASC"

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query evidence entries: %w", err)
	}
	defer rows.Close()

	var results []*acmv1.EvidenceChainEntry

	for rows.Next() {
		var (
			id, site, credentialIDHash, evidenceDataJSON, chainHash, signature string
			actionContextJSON, appliedRuleIDsJSON                              string
			crcID, previousEntryID                                             sql.NullString
			timestamp                                                          int64
			eventType, actionType, actionMethod, validationResult              int32
		)

		if err := rows.Scan(
			&id, &timestamp, &eventType, &site, &credentialIDHash,
			&actionType, &actionMethod, &actionContextJSON,
			&validationResult, &crcID, &appliedRuleIDsJSON,
			&evidenceDataJSON, &previousEntryID, &chainHash, &signature,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Deserialize applied rule IDs
		var appliedRuleIDs []string
		if err := json.Unmarshal([]byte(appliedRuleIDsJSON), &appliedRuleIDs); err != nil {
			return nil, fmt.Errorf("failed to deserialize applied rule IDs: %w", err)
		}

		// Construct action
		action := &acmv1.AutomationAction{
			Type:   acmv1.ActionType(actionType),
			Method: acmv1.AutomationMethod(actionMethod),
		}

		if actionContextJSON != "" {
			var context map[string]string
			if err := json.Unmarshal([]byte(actionContextJSON), &context); err != nil {
				return nil, fmt.Errorf("failed to deserialize action context: %w", err)
			}
			action.Context = context
		}

		// Construct evidence entry
		entry := &acmv1.EvidenceChainEntry{
			Id:               id,
			Timestamp:        timestamppb.New(time.Unix(timestamp, 0)),
			EventType:        acmv1.EvidenceEventType(eventType),
			Site:             site,
			CredentialIdHash: credentialIDHash,
			Action:           action,
			ValidationResult: acmv1.ValidationResult(validationResult),
			CrcId:            stringValue(crcID),
			AppliedRuleIds:   appliedRuleIDs,
			EvidenceData:     evidenceDataJSON,
			PreviousEntryId:  stringValue(previousEntryID),
			ChainHash:        chainHash,
			Signature:        signature,
		}

		results = append(results, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Verify verifies the integrity of an evidence entry.
func (g *SQLiteEvidenceChainGenerator) Verify(ctx context.Context, entry *acmv1.EvidenceChainEntry) (bool, error) {
	if entry == nil {
		return false, fmt.Errorf("entry cannot be nil")
	}

	// Verify signature
	sigBytes, err := hex.DecodeString(entry.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature encoding: %w", err)
	}

	message := g.constructSignatureMessage(entry)
	valid := ed25519.Verify(g.publicKey, message, sigBytes)

	return valid, nil
}

// VerifyChain verifies the integrity of the entire evidence chain.
func (g *SQLiteEvidenceChainGenerator) VerifyChain(ctx context.Context) (bool, []string, error) {
	// Get all entries in chronological order
	query := `
		SELECT
			id, timestamp, event_type, site, credential_id_hash,
			action_type, action_method, action_context_json,
			validation_result, crc_id, applied_rule_ids_json,
			evidence_data_json, previous_entry_id, chain_hash, signature
		FROM evidence_entries
		ORDER BY timestamp ASC
	`

	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return false, nil, fmt.Errorf("failed to query chain: %w", err)
	}
	defer rows.Close()

	var errors []string
	var previousID string

	for rows.Next() {
		var (
			id, site, credentialIDHash, evidenceDataJSON, chainHash, signature string
			actionContextJSON, appliedRuleIDsJSON                              string
			crcID, previousEntryID                                             sql.NullString
			timestamp                                                          int64
			eventType, actionType, actionMethod, validationResult              int32
		)

		if err := rows.Scan(
			&id, &timestamp, &eventType, &site, &credentialIDHash,
			&actionType, &actionMethod, &actionContextJSON,
			&validationResult, &crcID, &appliedRuleIDsJSON,
			&evidenceDataJSON, &previousEntryID, &chainHash, &signature,
		); err != nil {
			return false, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Deserialize applied rule IDs
		var appliedRuleIDs []string
		if err := json.Unmarshal([]byte(appliedRuleIDsJSON), &appliedRuleIDs); err != nil {
			errors = append(errors, fmt.Sprintf("Entry %s: failed to deserialize rule IDs: %v", id, err))
			continue
		}

		// Construct action
		action := &acmv1.AutomationAction{
			Type:   acmv1.ActionType(actionType),
			Method: acmv1.AutomationMethod(actionMethod),
		}

		if actionContextJSON != "" {
			var context map[string]string
			if err := json.Unmarshal([]byte(actionContextJSON), &context); err != nil {
				errors = append(errors, fmt.Sprintf("Entry %s: failed to deserialize action context: %v", id, err))
				continue
			}
			action.Context = context
		}

		// Construct entry for verification
		entry := &acmv1.EvidenceChainEntry{
			Id:               id,
			Timestamp:        timestamppb.New(time.Unix(timestamp, 0)),
			EventType:        acmv1.EvidenceEventType(eventType),
			Site:             site,
			CredentialIdHash: credentialIDHash,
			Action:           action,
			ValidationResult: acmv1.ValidationResult(validationResult),
			CrcId:            stringValue(crcID),
			AppliedRuleIds:   appliedRuleIDs,
			EvidenceData:     evidenceDataJSON,
			PreviousEntryId:  stringValue(previousEntryID),
			ChainHash:        chainHash,
			Signature:        signature,
		}

		// Verify signature
		valid, err := g.Verify(ctx, entry)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Entry %s: signature verification error: %v", id, err))
			continue
		}

		if !valid {
			errors = append(errors, fmt.Sprintf("Entry %s: invalid signature", id))
		}

		// Verify chain linkage
		if previousID != "" {
			if entry.PreviousEntryId != previousID {
				errors = append(errors, fmt.Sprintf("Entry %s: broken chain link (expected %s, got %s)",
					id, previousID, entry.PreviousEntryId))
			}

			// Verify chain hash
			expectedHash := computeChainHash(id, previousID)
			if entry.ChainHash != expectedHash {
				errors = append(errors, fmt.Sprintf("Entry %s: invalid chain hash", id))
			}
		}

		previousID = id
	}

	if err := rows.Err(); err != nil {
		return false, nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return len(errors) == 0, errors, nil
}

// GetChainHead returns the most recent evidence entry ID.
func (g *SQLiteEvidenceChainGenerator) GetChainHead(ctx context.Context) (string, error) {
	var id string
	err := g.db.QueryRowContext(ctx, `
		SELECT id FROM evidence_entries ORDER BY timestamp DESC LIMIT 1
	`).Scan(&id)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("chain is empty")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get chain head: %w", err)
	}

	return id, nil
}

// GetChainLength returns the number of entries in the chain.
func (g *SQLiteEvidenceChainGenerator) GetChainLength(ctx context.Context) (int, error) {
	var count int
	err := g.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM evidence_entries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain length: %w", err)
	}
	return count, nil
}

// GetPublicKey returns the public key for signature verification.
func (g *SQLiteEvidenceChainGenerator) GetPublicKey() []byte {
	return g.publicKey
}

// Clear removes all entries from the chain.
// WARNING: This is destructive and should only be used for testing or when
// the user explicitly disables ACVS with clear_cache=true.
func (g *SQLiteEvidenceChainGenerator) Clear(ctx context.Context) error {
	_, err := g.db.ExecContext(ctx, "DELETE FROM evidence_entries")
	if err != nil {
		return fmt.Errorf("failed to clear evidence chain: %w", err)
	}
	return nil
}

// Helper functions

// generateEntryID generates a unique ID for an evidence entry.
func generateEntryID(entry *acvs.EvidenceEntry) string {
	now := time.Now().Unix()
	data := fmt.Sprintf("%d:%s:%s:%v", now, entry.Site, entry.CredentialIDHash, entry.EventType)
	hash := sha256.Sum256([]byte(data))
	shortHash := hex.EncodeToString(hash[:8])
	return fmt.Sprintf("EVD-%d-%s", now, shortHash)
}

// computeChainHash computes the chain hash linking this entry to the previous one.
func computeChainHash(entryID, previousID string) string {
	if previousID == "" {
		// First entry in chain
		return hashString(entryID)
	}

	// Hash(currentID + previousID)
	data := entryID + previousID
	return hashString(data)
}

// hashString computes SHA-256 hash of a string.
func hashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// signEntry signs an evidence entry using Ed25519.
func (g *SQLiteEvidenceChainGenerator) signEntry(entry *acmv1.EvidenceChainEntry) (string, error) {
	message := g.constructSignatureMessage(entry)
	signature := ed25519.Sign(g.privateKey, message)
	return hex.EncodeToString(signature), nil
}

// constructSignatureMessage constructs the message to be signed.
func (g *SQLiteEvidenceChainGenerator) constructSignatureMessage(entry *acmv1.EvidenceChainEntry) []byte {
	message := fmt.Sprintf("%s|%d|%s|%s|%s|%s|%s|%s",
		entry.Id,
		entry.Timestamp.AsTime().Unix(),
		entry.Site,
		entry.CredentialIdHash,
		entry.EventType.String(),
		entry.ValidationResult.String(),
		entry.CrcId,
		entry.ChainHash,
	)
	return []byte(message)
}

// nullString converts a string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// stringValue converts sql.NullString to string.
func stringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
