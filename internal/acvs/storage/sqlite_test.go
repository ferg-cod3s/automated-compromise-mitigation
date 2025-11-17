package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs"
)

// setupTestDB creates a temporary test database.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize database
	config := Config{
		Path:         dbPath,
		CacheTTL:     time.Hour,
		EnableWAL:    true,
		MaxOpenConns: 5,
		BusyTimeout:  1000,
	}

	db, err := Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// TestSQLiteCRCManager_StoreAndGet tests storing and retrieving CRCs.
func TestSQLiteCRCManager_StoreAndGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	manager := NewSQLiteCRCManager(db, time.Hour)
	ctx := context.Background()

	// Create test CRC
	crc := &acmv1.ComplianceRuleSet{
		Site:       "github.com",
		TosUrl:     "https://github.com/site/terms",
		TosVersion: "2025-01-01",
		TosHash:    "abc123def456",
		Rules: []*acmv1.ComplianceRule{
			{
				Id:       "RULE-001",
				Category: acmv1.RuleCategory_RULE_CATEGORY_API_USAGE,
				Severity: acmv1.RuleSeverity_RULE_SEVERITY_INFO,
				Rule:     "API usage is permitted",
			},
		},
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
		Reasoning:      "API is explicitly allowed",
		Signature:      "signature123",
	}

	// Store CRC
	err := manager.Store(ctx, crc)
	if err != nil {
		t.Fatalf("failed to store CRC: %v", err)
	}

	// Retrieve CRC
	retrieved, found, err := manager.Get(ctx, "github.com")
	if err != nil {
		t.Fatalf("failed to get CRC: %v", err)
	}

	if !found {
		t.Fatal("CRC not found")
	}

	// Verify retrieved CRC
	if retrieved.Site != crc.Site {
		t.Errorf("site mismatch: got %s, want %s", retrieved.Site, crc.Site)
	}

	if len(retrieved.Rules) != len(crc.Rules) {
		t.Errorf("rules count mismatch: got %d, want %d", len(retrieved.Rules), len(crc.Rules))
	}
}

// TestSQLiteCRCManager_List tests listing CRCs.
func TestSQLiteCRCManager_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	manager := NewSQLiteCRCManager(db, time.Hour)
	ctx := context.Background()

	// Store multiple CRCs
	sites := []string{"github.com", "gitlab.com", "bitbucket.org"}
	for _, site := range sites {
		crc := &acmv1.ComplianceRuleSet{
			Site:           site,
			TosUrl:         "https://" + site + "/terms",
			TosVersion:     "2025-01-01",
			TosHash:        site + "-hash",
			Rules:          []*acmv1.ComplianceRule{},
			Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
		}

		if err := manager.Store(ctx, crc); err != nil {
			t.Fatalf("failed to store CRC for %s: %v", site, err)
		}
	}

	// List all CRCs
	summaries, err := manager.List(ctx, "", false)
	if err != nil {
		t.Fatalf("failed to list CRCs: %v", err)
	}

	if len(summaries) != len(sites) {
		t.Errorf("CRC count mismatch: got %d, want %d", len(summaries), len(sites))
	}
}

// TestSQLiteEvidenceChain_AddAndGet tests adding and retrieving evidence entries.
func TestSQLiteEvidenceChain_AddAndGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	generator, err := NewSQLiteEvidenceChainGenerator(db)
	if err != nil {
		t.Fatalf("failed to create evidence chain generator: %v", err)
	}

	ctx := context.Background()

	// Create test entry
	entry := &acvs.EvidenceEntry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
		Site:             "github.com",
		CredentialIDHash: "hash123",
		Action: &acmv1.AutomationAction{
			Type:   acmv1.ActionType_ACTION_TYPE_CREDENTIAL_ROTATION,
			Method: acmv1.AutomationMethod_AUTOMATION_METHOD_API,
		},
		ValidationResult: acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED,
		CRCID:            "CRC-001",
		AppliedRuleIDs:   []string{"RULE-001"},
		EvidenceData: map[string]interface{}{
			"test": "data",
		},
	}

	// Add entry
	entryID, err := generator.AddEntry(ctx, entry)
	if err != nil {
		t.Fatalf("failed to add evidence entry: %v", err)
	}

	// Retrieve entry
	retrieved, err := generator.GetEntry(ctx, entryID)
	if err != nil {
		t.Fatalf("failed to get evidence entry: %v", err)
	}

	// Verify retrieved entry
	if retrieved.Site != entry.Site {
		t.Errorf("site mismatch: got %s, want %s", retrieved.Site, entry.Site)
	}

	if retrieved.EventType != entry.EventType {
		t.Errorf("event type mismatch: got %v, want %v", retrieved.EventType, entry.EventType)
	}
}

// TestSQLiteEvidenceChain_Verify tests signature verification.
func TestSQLiteEvidenceChain_Verify(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	generator, err := NewSQLiteEvidenceChainGenerator(db)
	if err != nil {
		t.Fatalf("failed to create evidence chain generator: %v", err)
	}

	ctx := context.Background()

	// Add test entry
	entry := &acvs.EvidenceEntry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_ROTATION,
		Site:             "github.com",
		CredentialIDHash: "hash456",
		Action: &acmv1.AutomationAction{
			Type:   acmv1.ActionType_ACTION_TYPE_CREDENTIAL_ROTATION,
			Method: acmv1.AutomationMethod_AUTOMATION_METHOD_API,
		},
		ValidationResult: acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED,
		EvidenceData:     map[string]interface{}{},
	}

	entryID, err := generator.AddEntry(ctx, entry)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	// Retrieve and verify
	retrieved, err := generator.GetEntry(ctx, entryID)
	if err != nil {
		t.Fatalf("failed to get entry: %v", err)
	}

	valid, err := generator.Verify(ctx, retrieved)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	if !valid {
		t.Error("signature verification failed")
	}
}

// TestDatabaseInitialization tests database initialization and migration.
func TestDatabaseInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DefaultConfig()
	config.Path = dbPath

	db, err := Initialize(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify schema version
	version, err := GetSchemaVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}

	if version != CurrentSchemaVersion {
		t.Errorf("schema version mismatch: got %d, want %d", version, CurrentSchemaVersion)
	}

	// Verify integrity
	if err := verifyIntegrity(context.Background(), db); err != nil {
		t.Fatalf("integrity check failed: %v", err)
	}
}

// TestDatabaseStats tests database statistics retrieval.
func TestDatabaseStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := GetDatabaseStats(context.Background(), db)
	if err != nil {
		t.Fatalf("failed to get database stats: %v", err)
	}

	if stats.PageCount == 0 {
		t.Error("page count should not be zero")
	}

	if stats.PageSize == 0 {
		t.Error("page size should not be zero")
	}

	// Initially, counts should be zero
	if stats.CRCCount != 0 {
		t.Errorf("initial CRC count should be 0, got %d", stats.CRCCount)
	}
}

// TestBackupAndRestore tests database backup functionality.
func TestBackupAndRestore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Add some data
	manager := NewSQLiteCRCManager(db, time.Hour)
	crc := &acmv1.ComplianceRuleSet{
		Site:           "github.com",
		TosUrl:         "https://github.com/terms",
		TosVersion:     "2025-01-01",
		TosHash:        "test-hash",
		Rules:          []*acmv1.ComplianceRule{},
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}

	if err := manager.Store(ctx, crc); err != nil {
		t.Fatalf("failed to store CRC: %v", err)
	}

	// Create backup
	backupPath := filepath.Join(t.TempDir(), "backup.db")
	if err := CreateBackup(ctx, db, backupPath); err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file does not exist")
	}
}
