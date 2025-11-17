// Package crc provides Compliance Rule Set (CRC) caching and management.
package crc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Manager implements the CRCManager interface with in-memory caching.
// In Phase II, this uses an in-memory cache. Phase III will add SQLite persistence.
type Manager struct {
	mu       sync.RWMutex
	cache    map[string]*cachedCRC
	cacheTTL time.Duration
}

// cachedCRC wraps a CRC with metadata.
type cachedCRC struct {
	CRC       *acmv1.ComplianceRuleSet
	StoredAt  time.Time
	ExpiresAt time.Time
}

// DefaultCacheTTL is the default cache lifetime (30 days).
const DefaultCacheTTL = 30 * 24 * time.Hour

// NewManager creates a new CRC Manager.
func NewManager() *Manager {
	return &Manager{
		cache:    make(map[string]*cachedCRC),
		cacheTTL: DefaultCacheTTL,
	}
}

// NewManagerWithTTL creates a new CRC Manager with a custom TTL.
func NewManagerWithTTL(ttl time.Duration) *Manager {
	return &Manager{
		cache:    make(map[string]*cachedCRC),
		cacheTTL: ttl,
	}
}

// Store saves a CRC to the cache.
func (m *Manager) Store(ctx context.Context, crc *acmv1.ComplianceRuleSet) error {
	if crc == nil {
		return fmt.Errorf("cannot store nil CRC")
	}

	if crc.Site == "" {
		return fmt.Errorf("CRC site cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(m.cacheTTL)

	// Generate ID if not set
	if crc.Id == "" {
		crc.Id = m.generateCRCID(crc)
	}

	// Set timestamps if not set
	if crc.ParsedAt == nil {
		crc.ParsedAt = timestamppb.New(now)
	}

	// Set expiration
	crc.ExpiresAt = timestamppb.New(expiresAt)

	m.cache[crc.Site] = &cachedCRC{
		CRC:       crc,
		StoredAt:  now,
		ExpiresAt: expiresAt,
	}

	return nil
}

// Get retrieves a CRC from cache.
// Returns (crc, found, error). If found=false, crc will be nil.
func (m *Manager) Get(ctx context.Context, site string) (*acmv1.ComplianceRuleSet, bool, error) {
	if site == "" {
		return nil, false, fmt.Errorf("site cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	cached, exists := m.cache[site]
	if !exists {
		return nil, false, nil
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false, nil
	}

	return cached.CRC, true, nil
}

// List returns all cached CRCs matching the filter.
func (m *Manager) List(ctx context.Context, siteFilter string, includeExpired bool) ([]*Summary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]*Summary, 0, len(m.cache))
	now := time.Now()

	for _, cached := range m.cache {
		crc := cached.CRC

		// Apply site filter if provided
		if siteFilter != "" && crc.Site != siteFilter {
			continue
		}

		expired := now.After(cached.ExpiresAt)

		// Skip expired if not requested
		if expired && !includeExpired {
			continue
		}

		summary := &Summary{
			ID:             crc.Id,
			Site:           crc.Site,
			ParsedAt:       crc.ParsedAt.AsTime(),
			ExpiresAt:      cached.ExpiresAt,
			Recommendation: crc.Recommendation,
			RuleCount:      int32(len(crc.Rules)),
			Expired:        expired,
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Invalidate removes a CRC from cache.
func (m *Manager) Invalidate(ctx context.Context, site string) error {
	if site == "" {
		return fmt.Errorf("site cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, site)
	return nil
}

// IsExpired checks if a CRC has expired.
func (m *Manager) IsExpired(crc *acmv1.ComplianceRuleSet) bool {
	if crc == nil || crc.ExpiresAt == nil {
		return true
	}

	return time.Now().After(crc.ExpiresAt.AsTime())
}

// GetCacheTTL returns the configured cache TTL.
func (m *Manager) GetCacheTTL() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cacheTTL
}

// SetCacheTTL sets the cache TTL.
func (m *Manager) SetCacheTTL(ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheTTL = ttl
}

// generateCRCID generates a unique ID for a CRC based on site and ToS hash.
func (m *Manager) generateCRCID(crc *acmv1.ComplianceRuleSet) string {
	// Format: CRC-{site}-{short-hash}
	data := fmt.Sprintf("%s:%s:%s", crc.Site, crc.TosVersion, crc.TosHash)
	hash := sha256.Sum256([]byte(data))
	shortHash := hex.EncodeToString(hash[:8])
	return fmt.Sprintf("CRC-%s-%s", crc.Site, shortHash)
}

// GetCacheStats returns statistics about the cache.
func (m *Manager) GetCacheStats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := CacheStats{
		TotalEntries: len(m.cache),
	}

	now := time.Now()
	for _, cached := range m.cache {
		if now.After(cached.ExpiresAt) {
			stats.ExpiredEntries++
		} else {
			stats.ValidEntries++
		}
	}

	return stats
}

// CacheStats provides cache statistics.
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	ExpiredEntries int
}

// CleanExpired removes all expired entries from the cache.
// Returns the number of entries removed.
func (m *Manager) CleanExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	now := time.Now()

	for site, cached := range m.cache {
		if now.After(cached.ExpiresAt) {
			delete(m.cache, site)
			removed++
		}
	}

	return removed
}

// Clear removes all entries from the cache.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]*cachedCRC)
}

// Size returns the number of entries in the cache.
func (m *Manager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cache)
}
