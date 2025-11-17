// Package acvs implements the Automated Compliance Validation Service.
package acvs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/crc"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/evidence"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/nlp"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs/validator"
)

// ACVSService is the main ACVS service implementation.
type ACVSService struct {
	mu sync.RWMutex

	// Core components
	crcManager      CRCManager
	validator       Validator
	evidenceChain   EvidenceChainGenerator
	nlpEngine       NLPEngine
	tosFetcher      ToSFetcher

	// Configuration
	enabled            bool
	eulaVersion        string
	enabledAt          time.Time
	nlpModelVersion    string
	cacheTTLSeconds    int64
	evidenceChainEnabled bool
	defaultOnUncertain acmv1.ValidationResult
	modelPath          string

	// Statistics
	stats Statistics
}

// NewService creates a new ACVS service.
func NewService() (*ACVSService, error) {
	// Initialize components
	crcMgr := crc.NewManager()
	val := validator.NewValidator()
	evChain, err := evidence.NewChainGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create evidence chain: %w", err)
	}

	nlpEng := nlp.NewEngine("/var/lib/acm/models/legal-tos-v1")

	return &ACVSService{
		crcManager:           crcMgr,
		validator:            val,
		evidenceChain:        evChain,
		nlpEngine:            nlpEng,
		tosFetcher:           NewSimpleToSFetcher(),
		enabled:              false,
		eulaVersion:          "",
		nlpModelVersion:      nlpEng.GetModelVersion(),
		cacheTTLSeconds:      int64(crc.DefaultCacheTTL.Seconds()),
		evidenceChainEnabled: true,
		defaultOnUncertain:   acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED,
		modelPath:            nlpEng.GetModelPath(),
		stats:                Statistics{},
	}, nil
}

// IsEnabled returns whether ACVS is currently enabled.
func (s *ACVSService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Enable opts the user into ACVS.
func (s *ACVSService) Enable(ctx context.Context, eulaVersion string, consent bool) error {
	if !consent {
		return fmt.Errorf("user consent required to enable ACVS")
	}

	if eulaVersion == "" {
		return fmt.Errorf("EULA version required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = true
	s.eulaVersion = eulaVersion
	s.enabledAt = time.Now()

	// Log evidence event
	if s.evidenceChainEnabled {
		_, err := s.evidenceChain.AddEntry(ctx, &evidence.Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_ACVS_ENABLED,
			Site:             "acm-service",
			CredentialIDHash: "",
			EvidenceData: map[string]interface{}{
				"eula_version": eulaVersion,
				"enabled_at":   s.enabledAt,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to log evidence: %w", err)
		}
	}

	return nil
}

// Disable opts the user out of ACVS.
func (s *ACVSService) Disable(ctx context.Context, clearCache bool, preserveEvidence bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = false

	if clearCache {
		s.crcManager.Clear()
	}

	if !preserveEvidence && s.evidenceChainEnabled {
		s.evidenceChain.Clear()
	}

	// Log evidence event (if preserving)
	if preserveEvidence && s.evidenceChainEnabled {
		_, err := s.evidenceChain.AddEntry(ctx, &evidence.Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_ACVS_DISABLED,
			Site:             "acm-service",
			CredentialIDHash: "",
			EvidenceData: map[string]interface{}{
				"disabled_at":      time.Now(),
				"cleared_cache":    clearCache,
				"preserved_evidence": preserveEvidence,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to log evidence: %w", err)
		}
	}

	return nil
}

// AnalyzeToS fetches and analyzes a website's Terms of Service.
func (s *ACVSService) AnalyzeToS(ctx context.Context, site string, tosURL string, forceRefresh bool, timeoutSecs int32) (*acmv1.ComplianceRuleSet, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("ACVS not enabled")
	}

	// Check cache first (unless force refresh)
	if !forceRefresh {
		cached, found, err := s.crcManager.Get(ctx, site)
		if err != nil {
			return nil, fmt.Errorf("cache lookup failed: %w", err)
		}

		if found && !s.crcManager.IsExpired(cached) {
			s.incrementStat("total_analyses")
			return cached, nil
		}
	}

	// Discover ToS URL if not provided
	if tosURL == "" {
		discovered, err := s.tosFetcher.DiscoverToSURL(ctx, site)
		if err != nil {
			return nil, fmt.Errorf("failed to discover ToS URL: %w", err)
		}
		tosURL = discovered
	}

	// Fetch ToS content
	tosContent, err := s.tosFetcher.FetchToS(ctx, tosURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ToS: %w", err)
	}

	// Analyze with NLP engine
	if !s.nlpEngine.IsAvailable(ctx) {
		return nil, fmt.Errorf("NLP engine not available")
	}

	crcResult, err := s.nlpEngine.AnalyzeToS(ctx, tosContent, site, tosURL)
	if err != nil {
		return nil, fmt.Errorf("NLP analysis failed: %w", err)
	}

	// Store in cache
	err = s.crcManager.Store(ctx, crcResult)
	if err != nil {
		return nil, fmt.Errorf("failed to cache CRC: %w", err)
	}

	// Log evidence
	if s.evidenceChainEnabled {
		_, err = s.evidenceChain.AddEntry(ctx, &evidence.Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_CRC_UPDATE,
			Site:             site,
			CredentialIDHash: "",
			CRCID:            crcResult.Id,
			EvidenceData: map[string]interface{}{
				"tos_url":        tosURL,
				"tos_version":    crcResult.TosVersion,
				"tos_hash":       crcResult.TosHash,
				"rule_count":     len(crcResult.Rules),
				"recommendation": crcResult.Recommendation.String(),
			},
		})
		if err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to log evidence: %v\n", err)
		}
	}

	s.incrementStat("total_analyses")

	return crcResult, nil
}

// ValidateAction performs pre-flight validation of an automation action.
func (s *ACVSService) ValidateAction(ctx context.Context, site string, action *acmv1.AutomationAction, credentialID string, forceRefresh bool) (*ValidationResult, error) {
	if !s.IsEnabled() {
		return &ValidationResult{
			Result:            acmv1.ValidationResult_VALIDATION_RESULT_DISABLED,
			RecommendedMethod: acmv1.AutomationMethod_AUTOMATION_METHOD_MANUAL,
			Reasoning:         "ACVS is not enabled",
		}, nil
	}

	// Get CRC (analyze if needed)
	crcResult, err := s.AnalyzeToS(ctx, site, "", forceRefresh, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get CRC: %w", err)
	}

	// Validate action
	result, err := s.validator.Validate(ctx, crcResult, action)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Hash credential ID for evidence chain
	credHash := s.hashCredentialID(credentialID)

	// Log evidence
	if s.evidenceChainEnabled {
		evidenceID, err := s.evidenceChain.AddEntry(ctx, &evidence.Entry{
			EventType:        acmv1.EvidenceEventType_EVIDENCE_EVENT_TYPE_VALIDATION,
			Site:             site,
			CredentialIDHash: credHash,
			Action:           action,
			ValidationResult: result.Result,
			CRCID:            crcResult.Id,
			AppliedRuleIDs:   result.ApplicableRuleIDs,
			EvidenceData: map[string]interface{}{
				"recommendation": result.RecommendedMethod.String(),
				"reasoning":      result.Reasoning,
			},
		})
		if err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to log evidence: %v\n", err)
		} else {
			result.EvidenceEntryID = evidenceID
		}
	}

	// Update statistics
	s.incrementStat("total_validations")
	switch result.Result {
	case acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED:
		s.incrementStat("validations_allowed")
	case acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED:
		s.incrementStat("validations_him_required")
	case acmv1.ValidationResult_VALIDATION_RESULT_BLOCKED:
		s.incrementStat("validations_blocked")
	}

	return result, nil
}

// GetCRC retrieves a cached CRC.
func (s *ACVSService) GetCRC(ctx context.Context, site string) (*acmv1.ComplianceRuleSet, bool, error) {
	if !s.IsEnabled() {
		return nil, false, fmt.Errorf("ACVS not enabled")
	}

	return s.crcManager.Get(ctx, site)
}

// ListCRCs lists all cached CRCs.
func (s *ACVSService) ListCRCs(ctx context.Context, siteFilter string, includeExpired bool) ([]CRCSummary, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("ACVS not enabled")
	}

	summaries, err := s.crcManager.List(ctx, siteFilter, includeExpired)
	if err != nil {
		return nil, err
	}

	// Convert from crc.Summary to acvs.CRCSummary
	result := make([]CRCSummary, len(summaries))
	for i, sum := range summaries {
		result[i] = CRCSummary{
			ID:             sum.ID,
			Site:           sum.Site,
			ParsedAt:       sum.ParsedAt,
			ExpiresAt:      sum.ExpiresAt,
			Recommendation: sum.Recommendation,
			RuleCount:      sum.RuleCount,
			Expired:        sum.Expired,
		}
	}
	return result, nil
}

// InvalidateCRC removes a cached CRC.
func (s *ACVSService) InvalidateCRC(ctx context.Context, site string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("ACVS not enabled")
	}

	return s.crcManager.Invalidate(ctx, site)
}

// ExportEvidenceChain exports evidence chain entries.
func (s *ACVSService) ExportEvidenceChain(ctx context.Context, req *ExportRequest) ([]*acmv1.EvidenceChainEntry, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("ACVS not enabled")
	}

	if !s.evidenceChainEnabled {
		return nil, fmt.Errorf("evidence chain not enabled")
	}

	return s.evidenceChain.Export(ctx, req)
}

// GetStatus returns current ACVS status.
func (s *ACVSService) GetStatus(ctx context.Context) (*Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Status{
		Enabled:     s.enabled,
		EULAVersion: s.eulaVersion,
		EnabledAt:   s.enabledAt,
		Configuration: Configuration{
			NLPModelVersion:      s.nlpModelVersion,
			CacheTTLSeconds:      s.cacheTTLSeconds,
			EvidenceChainEnabled: s.evidenceChainEnabled,
			DefaultOnUncertain:   s.defaultOnUncertain,
			ModelPath:            s.modelPath,
		},
		Statistics: s.stats,
	}, nil
}

// GetStatistics returns ACVS usage statistics.
func (s *ACVSService) GetStatistics(ctx context.Context) (*Statistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get cache stats
	cacheStats := s.crcManager.(*crc.Manager).GetCacheStats()

	s.stats.CRCsCached = int32(cacheStats.ValidEntries)
	s.stats.EvidenceEntries = int64(s.evidenceChain.(*evidence.ChainGenerator).GetChainLength())

	return &s.stats, nil
}

// hashCredentialID hashes a credential ID for privacy.
func (s *ACVSService) hashCredentialID(credentialID string) string {
	if credentialID == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(credentialID))
	return hex.EncodeToString(hash[:])
}

// incrementStat atomically increments a statistic.
func (s *ACVSService) incrementStat(stat string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch stat {
	case "total_analyses":
		s.stats.TotalAnalyses++
	case "total_validations":
		s.stats.TotalValidations++
	case "validations_allowed":
		s.stats.ValidationsAllowed++
	case "validations_him_required":
		s.stats.ValidationsHIMRequired++
	case "validations_blocked":
		s.stats.ValidationsBlocked++
	}
}

// SetConfiguration updates ACVS configuration.
func (s *ACVSService) SetConfiguration(config Configuration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cacheTTLSeconds = config.CacheTTLSeconds
	s.evidenceChainEnabled = config.EvidenceChainEnabled
	s.defaultOnUncertain = config.DefaultOnUncertain

	if config.ModelPath != "" {
		s.modelPath = config.ModelPath
		s.nlpEngine.SetModelPath(config.ModelPath)
	}

	// Update CRC manager TTL
	ttl := time.Duration(config.CacheTTLSeconds) * time.Second
	s.crcManager.SetCacheTTL(ttl)

	// Update validator default
	s.validator.(*validator.Validator).SetDefaultOnUncertain(config.DefaultOnUncertain)
}
