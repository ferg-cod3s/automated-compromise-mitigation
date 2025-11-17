package evidence

import (
	"context"
	"testing"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// TestNewChainGenerator tests chain generator creation
func TestNewChainGenerator(t *testing.T) {
	gen, err := NewChainGenerator()
	if err != nil {
		t.Fatalf("Failed to create chain generator: %v", err)
	}

	if gen == nil {
		t.Fatal("Chain generator should not be nil")
	}

	if gen.publicKey == nil || len(gen.publicKey) == 0 {
		t.Error("Public key should be generated")
	}

	if gen.privateKey == nil || len(gen.privateKey) == 0 {
		t.Error("Private key should be generated")
	}

	if gen.chainHead != "" {
		t.Error("Chain head should be empty for new generator")
	}

	if len(gen.chain) != 0 {
		t.Error("Chain should be empty for new generator")
	}
}

// TestAddEntry tests adding evidence entries
func TestAddEntry(t *testing.T) {
	gen, err := NewChainGenerator()
	if err != nil {
		t.Fatalf("Failed to create chain generator: %v", err)
	}

	ctx := context.Background()

	// Add first entry
	entry1 := &Entry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_ACVS_ENABLED,
		Site:             "example.com",
		CredentialIDHash: "hash123",
		EvidenceData: map[string]interface{}{
			"eula_version": "1.0",
		},
	}

	entryID1, err := gen.AddEntry(ctx, entry1)
	if err != nil {
		t.Fatalf("Failed to add first entry: %v", err)
	}

	if entryID1 == "" {
		t.Error("Entry ID should not be empty")
	}

	if gen.chainHead != entryID1 {
		t.Errorf("Chain head should be %s, got %s", entryID1, gen.chainHead)
	}

	if gen.GetChainLength() != 1 {
		t.Errorf("Chain length should be 1, got %d", gen.GetChainLength())
	}

	// Add second entry
	entry2 := &Entry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
		Site:             "example.com",
		CredentialIDHash: "hash123",
		ValidationResult: acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED,
		EvidenceData:     map[string]interface{}{},
	}

	entryID2, err := gen.AddEntry(ctx, entry2)
	if err != nil {
		t.Fatalf("Failed to add second entry: %v", err)
	}

	if gen.chainHead != entryID2 {
		t.Errorf("Chain head should be updated to %s, got %s", entryID2, gen.chainHead)
	}

	if gen.GetChainLength() != 2 {
		t.Errorf("Chain length should be 2, got %d", gen.GetChainLength())
	}

	// Verify second entry links to first
	retrieved, err := gen.GetEntry(ctx, entryID2)
	if err != nil {
		t.Fatalf("Failed to retrieve second entry: %v", err)
	}

	if retrieved.PreviousEntryId != entryID1 {
		t.Errorf("Second entry should link to first (%s), got %s", entryID1, retrieved.PreviousEntryId)
	}
}

// TestGetEntry tests retrieving evidence entries
func TestGetEntry(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	entry := &Entry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_CRC_UPDATE,
		Site:             "test.com",
		CredentialIDHash: "",
		CRCID:            "CRC-001",
		EvidenceData:     map[string]interface{}{"tos_url": "https://test.com/terms"},
	}

	entryID, err := gen.AddEntry(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Retrieve the entry
	retrieved, err := gen.GetEntry(ctx, entryID)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	if retrieved.Id != entryID {
		t.Errorf("Expected entry ID %s, got %s", entryID, retrieved.Id)
	}

	if retrieved.Site != "test.com" {
		t.Errorf("Expected site test.com, got %s", retrieved.Site)
	}

	if retrieved.EventType != acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_CRC_UPDATE {
		t.Errorf("Expected event type CRC_UPDATE, got %s", retrieved.EventType.String())
	}

	// Try to get non-existent entry
	_, err = gen.GetEntry(ctx, "nonexistent")
	if err == nil {
		t.Error("Should error when getting non-existent entry")
	}
}

// TestVerifySignature tests signature verification
func TestVerifySignature(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	entry := &Entry{
		EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
		Site:             "verify.com",
		CredentialIDHash: "hash456",
		ValidationResult: acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED,
		EvidenceData:     map[string]interface{}{},
	}

	entryID, _ := gen.AddEntry(ctx, entry)
	retrieved, _ := gen.GetEntry(ctx, entryID)

	// Verify valid signature
	valid, err := gen.Verify(ctx, retrieved)
	if err != nil {
		t.Fatalf("Verify should not error: %v", err)
	}

	if !valid {
		t.Error("Signature should be valid")
	}

	// Tamper with the entry
	tampered := *retrieved
	tampered.Site = "tampered.com"

	// Verify should fail
	valid, err = gen.Verify(ctx, &tampered)
	if err != nil {
		t.Fatalf("Verify should not error: %v", err)
	}

	if valid {
		t.Error("Signature should be invalid for tampered entry")
	}
}

// TestVerifyChain tests full chain verification
func TestVerifyChain(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	// Add multiple entries
	for i := 1; i <= 5; i++ {
		entry := &Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:             "chain.com",
			CredentialIDHash: "hash",
			ValidationResult: acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED,
			EvidenceData:     map[string]interface{}{"iteration": i},
		}
		_, err := gen.AddEntry(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to add entry %d: %v", i, err)
		}
	}

	// Verify chain integrity
	valid, errors, err := gen.VerifyChain(ctx)
	if err != nil {
		t.Fatalf("VerifyChain should not error: %v", err)
	}

	if !valid {
		t.Errorf("Chain should be valid, got errors: %v", errors)
	}

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}
}

// TestChainLinking tests that entries are properly linked
func TestChainLinking(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	var previousID string
	entryIDs := make([]string, 0)

	// Add 3 entries
	for i := 1; i <= 3; i++ {
		entry := &Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:             "linking.com",
			CredentialIDHash: "hash",
			EvidenceData:     map[string]interface{}{"step": i},
		}

		entryID, _ := gen.AddEntry(ctx, entry)
		entryIDs = append(entryIDs, entryID)

		retrieved, _ := gen.GetEntry(ctx, entryID)

		if i == 1 {
			// First entry should have empty previous ID
			if retrieved.PreviousEntryId != "" {
				t.Errorf("First entry should have empty PreviousEntryId, got %s", retrieved.PreviousEntryId)
			}
		} else {
			// Subsequent entries should link to previous
			if retrieved.PreviousEntryId != previousID {
				t.Errorf("Entry %d should link to %s, got %s", i, previousID, retrieved.PreviousEntryId)
			}
		}

		previousID = entryID
	}

	// Verify chain order
	if len(gen.chain) != 3 {
		t.Errorf("Chain should have 3 entries, got %d", len(gen.chain))
	}

	for i, entryID := range gen.chain {
		if entryID != entryIDs[i] {
			t.Errorf("Chain position %d should be %s, got %s", i, entryIDs[i], entryID)
		}
	}
}

// TestExport tests exporting evidence entries
func TestExport(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	// Add entries with different credentials and times
	now := time.Now()

	entries := []struct {
		credentialHash string
		timestamp      time.Time
	}{
		{"hash1", now.Add(-2 * time.Hour)},
		{"hash2", now.Add(-1 * time.Hour)},
		{"hash1", now},
	}

	for _, e := range entries {
		entry := &Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:             "export.com",
			CredentialIDHash: e.credentialHash,
			EvidenceData:     map[string]interface{}{},
		}
		gen.AddEntry(ctx, entry)
		// Sleep to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Export all
	req := &ExportRequest{}
	exported, err := gen.Export(ctx, req)
	if err != nil {
		t.Fatalf("Export should not error: %v", err)
	}

	if len(exported) != 3 {
		t.Errorf("Expected 3 exported entries, got %d", len(exported))
	}

	// Export filtered by credential
	req = &ExportRequest{
		CredentialID: "hash1",
	}
	exported, err = gen.Export(ctx, req)
	if err != nil {
		t.Fatalf("Export with filter should not error: %v", err)
	}

	if len(exported) != 2 {
		t.Errorf("Expected 2 entries for hash1, got %d", len(exported))
	}

	// Export filtered by time - use wider range since timestamps may vary slightly
	req = &ExportRequest{
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(1 * time.Hour), // Include all entries
	}
	exported, err = gen.Export(ctx, req)
	if err != nil {
		t.Fatalf("Export with time filter should not error: %v", err)
	}

	// Should get all 3 entries within this wide time range
	if len(exported) != 3 {
		t.Errorf("Expected 3 entries in time range, got %d", len(exported))
	}
}

// TestGetChainHead tests retrieving the chain head
func TestGetChainHead(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	// Empty chain
	_, err := gen.GetChainHead(ctx)
	if err == nil {
		t.Error("Should error when chain is empty")
	}

	// Add entries
	var lastID string
	for i := 1; i <= 3; i++ {
		entry := &Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:             "head.com",
			EvidenceData:     map[string]interface{}{},
		}
		lastID, _ = gen.AddEntry(ctx, entry)
	}

	// Get head
	head, err := gen.GetChainHead(ctx)
	if err != nil {
		t.Fatalf("GetChainHead should not error: %v", err)
	}

	if head != lastID {
		t.Errorf("Chain head should be %s, got %s", lastID, head)
	}
}

// TestClear tests clearing the evidence chain
func TestClear(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	// Add entries
	for i := 1; i <= 5; i++ {
		entry := &Entry{
			EventType:    acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:         "clear.com",
			EvidenceData: map[string]interface{}{},
		}
		gen.AddEntry(ctx, entry)
	}

	if gen.GetChainLength() != 5 {
		t.Fatalf("Chain should have 5 entries before clear")
	}

	// Clear the chain
	gen.Clear()

	if gen.GetChainLength() != 0 {
		t.Errorf("Chain should be empty after clear, got %d entries", gen.GetChainLength())
	}

	if gen.chainHead != "" {
		t.Error("Chain head should be empty after clear")
	}

	// Should error when getting head of empty chain
	_, err := gen.GetChainHead(ctx)
	if err == nil {
		t.Error("Should error when getting head of empty chain")
	}
}

// TestExportToJSON tests JSON export functionality
func TestExportToJSON(t *testing.T) {
	gen, _ := NewChainGenerator()
	ctx := context.Background()

	// Add a few entries
	for i := 1; i <= 3; i++ {
		entry := &Entry{
			EventType:    acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:         "json.com",
			EvidenceData: map[string]interface{}{"index": i},
		}
		gen.AddEntry(ctx, entry)
	}

	// Export to JSON
	jsonData, err := gen.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON should not error: %v", err)
	}

	if jsonData == "" {
		t.Error("JSON data should not be empty")
	}

	// Should contain public key
	if !containsString(jsonData, "public_key") {
		t.Error("JSON should contain public_key field")
	}

	// Should contain entries
	if !containsString(jsonData, "entries") {
		t.Error("JSON should contain entries field")
	}

	// Should contain length
	if !containsString(jsonData, "\"length\": 3") {
		t.Error("JSON should contain correct length")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
