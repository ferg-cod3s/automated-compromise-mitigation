package crc

import (
	"context"
	"testing"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestStoreAndGet tests basic CRC caching
func TestStoreAndGet(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	// Create a test CRC
	crc := &acmv1.ComplianceRuleSet{
		Id:             "CRC-001",
		Site:           "example.com",
		TosUrl:         "https://example.com/terms",
		TosVersion:     "2024-01-01",
		TosHash:        "abc123",
		ParsedAt:       timestamppb.Now(),
		ExpiresAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}

	// Store the CRC
	err := manager.Store(ctx, crc)
	if err != nil {
		t.Fatalf("Failed to store CRC: %v", err)
	}

	// Retrieve the CRC
	retrieved, found, err := manager.Get(ctx, "example.com")
	if err != nil {
		t.Fatalf("Failed to get CRC: %v", err)
	}

	if !found {
		t.Fatal("CRC should have been found")
	}

	if retrieved.Id != crc.Id {
		t.Errorf("Expected CRC ID %s, got %s", crc.Id, retrieved.Id)
	}

	if retrieved.Site != crc.Site {
		t.Errorf("Expected site %s, got %s", crc.Site, retrieved.Site)
	}
}

// TestGetNonExistent tests retrieving a non-existent CRC
func TestGetNonExistent(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	retrieved, found, err := manager.Get(ctx, "nonexistent.com")
	if err != nil {
		t.Fatalf("Get should not error for non-existent CRC: %v", err)
	}

	if found {
		t.Error("CRC should not have been found")
	}

	if retrieved != nil {
		t.Error("Retrieved CRC should be nil for non-existent site")
	}
}

// TestExpiration tests CRC expiration handling
func TestExpiration(t *testing.T) {
	manager := NewManager()

	// Test IsExpired with an already-expired CRC
	expiredCRC := &acmv1.ComplianceRuleSet{
		Id:             "CRC-002",
		Site:           "expired.com",
		TosUrl:         "https://expired.com/terms",
		TosVersion:     "2023-01-01",
		ParsedAt:       timestamppb.New(time.Now().Add(-60 * 24 * time.Hour)),
		ExpiresAt:      timestamppb.New(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}

	// Check if it's expired (without storing - Store() overwrites ExpiresAt)
	if !manager.IsExpired(expiredCRC) {
		t.Error("CRC with past ExpiresAt should be marked as expired")
	}

	// Test that non-expired CRC is not expired
	validCRC := &acmv1.ComplianceRuleSet{
		Id:             "CRC-valid",
		Site:           "valid.com",
		ExpiresAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}

	if manager.IsExpired(validCRC) {
		t.Error("CRC with future ExpiresAt should not be marked as expired")
	}
}

// TestList tests listing CRCs with filtering
func TestList(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	// Create multiple CRCs
	sites := []string{"example.com", "test.com", "sample.org"}
	for _, site := range sites {
		crc := &acmv1.ComplianceRuleSet{
			Id:             "CRC-" + site,
			Site:           site,
			TosUrl:         "https://" + site + "/terms",
			ParsedAt:       timestamppb.Now(),
			ExpiresAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
			Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
		}
		err := manager.Store(ctx, crc)
		if err != nil {
			t.Fatalf("Failed to store CRC for %s: %v", site, err)
		}
	}

	// List all CRCs
	all, err := manager.List(ctx, "", false)
	if err != nil {
		t.Fatalf("Failed to list CRCs: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 CRCs, got %d", len(all))
	}

	// List with exact filter (filter uses exact match, not substring)
	filtered, err := manager.List(ctx, "example.com", false)
	if err != nil {
		t.Fatalf("Failed to list filtered CRCs: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered CRC, got %d", len(filtered))
	}

	if len(filtered) > 0 && filtered[0].Site != "example.com" {
		t.Errorf("Expected filtered site example.com, got %s", filtered[0].Site)
	}
}

// TestInvalidate tests CRC invalidation
func TestInvalidate(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	// Create and store a CRC
	crc := &acmv1.ComplianceRuleSet{
		Id:             "CRC-003",
		Site:           "invalidate.com",
		TosUrl:         "https://invalidate.com/terms",
		ParsedAt:       timestamppb.Now(),
		ExpiresAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}

	err := manager.Store(ctx, crc)
	if err != nil {
		t.Fatalf("Failed to store CRC: %v", err)
	}

	// Verify it exists
	_, found, _ := manager.Get(ctx, "invalidate.com")
	if !found {
		t.Fatal("CRC should exist before invalidation")
	}

	// Invalidate it
	err = manager.Invalidate(ctx, "invalidate.com")
	if err != nil {
		t.Fatalf("Failed to invalidate CRC: %v", err)
	}

	// Verify it's gone
	_, found, _ = manager.Get(ctx, "invalidate.com")
	if found {
		t.Error("CRC should not exist after invalidation")
	}
}

// TestClear tests clearing all CRCs
func TestClear(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	// Create multiple CRCs
	for i := 1; i <= 5; i++ {
		site := "site" + string(rune('0'+i)) + ".com"
		crc := &acmv1.ComplianceRuleSet{
			Id:             "CRC-clear-test-" + site,
			Site:           site,
			ParsedAt:       timestamppb.Now(),
			ExpiresAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
			Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
		}
		manager.Store(ctx, crc)
	}

	// Verify we have 5 CRCs
	all, _ := manager.List(ctx, "", false)
	if len(all) != 5 {
		t.Fatalf("Expected 5 CRCs before clear, got %d", len(all))
	}

	// Clear all
	manager.Clear()

	// Verify cache is empty
	all, _ = manager.List(ctx, "", false)
	if len(all) != 0 {
		t.Errorf("Expected 0 CRCs after clear, got %d", len(all))
	}
}

// TestCacheStats tests cache statistics
func TestCacheStats(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	// Initially empty
	stats := manager.GetCacheStats()
	if stats.ValidEntries != 0 || stats.ExpiredEntries != 0 {
		t.Error("Cache should be empty initially")
	}

	// Add valid CRCs (Store() always sets ExpiresAt to now + cacheTTL, so both will be valid)
	crc1 := &acmv1.ComplianceRuleSet{
		Id:             "CRC-001",
		Site:           "site1.com",
		ParsedAt:       timestamppb.Now(),
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}
	manager.Store(ctx, crc1)

	crc2 := &acmv1.ComplianceRuleSet{
		Id:             "CRC-002",
		Site:           "site2.com",
		ParsedAt:       timestamppb.Now(),
		Recommendation: acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
	}
	manager.Store(ctx, crc2)

	// Check stats - both should be valid since Store() sets future expiration
	stats = manager.GetCacheStats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 total entries, got %d", stats.TotalEntries)
	}
	if stats.ValidEntries != 2 {
		t.Errorf("Expected 2 valid entries, got %d", stats.ValidEntries)
	}
	if stats.ExpiredEntries != 0 {
		t.Errorf("Expected 0 expired entries, got %d", stats.ExpiredEntries)
	}
}

// TestSetCacheTTL tests cache TTL configuration
func TestSetCacheTTL(t *testing.T) {
	manager := NewManager()

	// Default TTL should be 30 days
	defaultTTL := manager.GetCacheTTL()
	if defaultTTL != DefaultCacheTTL {
		t.Errorf("Expected default TTL %v, got %v", DefaultCacheTTL, defaultTTL)
	}

	// Set custom TTL
	customTTL := 7 * 24 * time.Hour // 7 days
	manager.SetCacheTTL(customTTL)

	newTTL := manager.GetCacheTTL()
	if newTTL != customTTL {
		t.Errorf("Expected custom TTL %v, got %v", customTTL, newTTL)
	}
}
