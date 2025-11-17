// Package nlp provides Legal NLP integration for ToS analysis.
// Phase II implements a stub that returns mock CRCs for testing.
// Phase III will integrate with spaCy/Transformers for real NLP analysis.
package nlp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Engine implements the NLPEngine interface.
// Phase II: Stub implementation with mock analysis.
// Phase III: Will integrate with Python spaCy via subprocess/gRPC.
type Engine struct {
	modelVersion string
	modelPath    string
	available    bool
}

// NewEngine creates a new NLP engine.
func NewEngine(modelPath string) *Engine {
	return &Engine{
		modelVersion: "stub-v1.0.0",
		modelPath:    modelPath,
		available:    true, // Stub is always available
	}
}

// AnalyzeToS analyzes a ToS document and generates a CRC.
// Phase II: Returns a mock CRC based on heuristics.
// Phase III: Will call Python NLP engine for real analysis.
func (e *Engine) AnalyzeToS(ctx context.Context, tosContent string, site string, tosURL string) (*acmv1.ComplianceRuleSet, error) {
	if tosContent == "" {
		return nil, fmt.Errorf("ToS content cannot be empty")
	}

	if site == "" {
		return nil, fmt.Errorf("site cannot be empty")
	}

	// Phase II: Mock analysis using simple keyword detection
	// This is a placeholder until real NLP is integrated
	crc := e.mockAnalysis(tosContent, site, tosURL)

	return crc, nil
}

// IsAvailable checks if the NLP engine is available.
func (e *Engine) IsAvailable(ctx context.Context) bool {
	// Phase II: Stub is always available
	// Phase III: Will check if Python process is running
	return e.available
}

// GetModelVersion returns the NLP model version.
func (e *Engine) GetModelVersion() string {
	return e.modelVersion
}

// GetModelPath returns the path to the NLP model.
func (e *Engine) GetModelPath() string {
	return e.modelPath
}

// SetModelPath sets the path to the NLP model.
func (e *Engine) SetModelPath(path string) {
	e.modelPath = path
}

// mockAnalysis performs mock ToS analysis using simple keyword detection.
// This is a stub implementation for Phase II testing.
func (e *Engine) mockAnalysis(tosContent string, site string, tosURL string) *acmv1.ComplianceRuleSet {
	lower := strings.ToLower(tosContent)
	now := time.Now()

	// Compute ToS hash
	hash := sha256.Sum256([]byte(tosContent))
	tosHash := hex.EncodeToString(hash[:])

	// Extract version (mock - use current date)
	tosVersion := now.Format("2006-01-02")

	rules := make([]*acmv1.ComplianceRule, 0)
	recommendation := acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN
	reasoning := "Mock analysis - Phase II stub implementation"

	// Rule 1: Check for automation prohibitions
	if e.containsKeywords(lower, []string{"prohibit", "automated", "bot", "scraping"}) {
		rules = append(rules, &acmv1.ComplianceRule{
			Id:       "CRC-001",
			Category: acmv1.RuleCategory_RULE_CATEGORY_AUTOMATION,
			Severity: acmv1.RuleSeverity_RULE_SEVERITY_HIGH,
			Rule:     "Prohibits automated access and bot usage",
			ExtractedText: e.extractContext(tosContent, "automated"),
			Confidence: 0.85,
			Implications: &acmv1.RuleImplications{
				AllowsApiAutomation:       false,
				RequiresHumanInteraction:  true,
				RateLimit:                 nil,
			},
		})
		recommendation = acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_HIM_REQUIRED
		reasoning = "ToS contains automation prohibitions; human interaction recommended"
	}

	// Rule 2: Check for API allowances
	if e.containsKeywords(lower, []string{"api", "programmatic access", "developer"}) {
		rules = append(rules, &acmv1.ComplianceRule{
			Id:       "CRC-002",
			Category: acmv1.RuleCategory_RULE_CATEGORY_API_USAGE,
			Severity: acmv1.RuleSeverity_RULE_SEVERITY_MEDIUM,
			Rule:     "Allows API usage with rate limits",
			ExtractedText: e.extractContext(tosContent, "api"),
			Confidence: 0.90,
			Implications: &acmv1.RuleImplications{
				AllowsApiAutomation:      true,
				RequiresHumanInteraction: false,
				RateLimit: &acmv1.RateLimit{
					Requests: 60,
					Window:   "1h",
					Scope:    "user",
				},
			},
		})

		if recommendation == acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN {
			recommendation = acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED_WITH_API
			reasoning = "ToS allows API usage; use official API for automation"
		}
	}

	// Rule 3: Check for rate limits
	if e.containsKeywords(lower, []string{"rate limit", "requests per", "throttle"}) {
		rules = append(rules, &acmv1.ComplianceRule{
			Id:       "CRC-003",
			Category: acmv1.RuleCategory_RULE_CATEGORY_RATE_LIMITING,
			Severity: acmv1.RuleSeverity_RULE_SEVERITY_MEDIUM,
			Rule:     "Enforces rate limiting on API calls",
			ExtractedText: e.extractContext(tosContent, "rate limit"),
			Confidence: 0.80,
			Implications: &acmv1.RuleImplications{
				AllowsApiAutomation: true,
				RateLimit: &acmv1.RateLimit{
					Requests: 100,
					Window:   "1h",
					Scope:    "user",
				},
			},
		})
	}

	// Rule 4: Check for credential management policies
	if e.containsKeywords(lower, []string{"password", "credentials", "authentication"}) {
		rules = append(rules, &acmv1.ComplianceRule{
			Id:       "CRC-004",
			Category: acmv1.RuleCategory_RULE_CATEGORY_CREDENTIALS,
			Severity: acmv1.RuleSeverity_RULE_SEVERITY_INFO,
			Rule:     "Provides guidance on credential management",
			ExtractedText: e.extractContext(tosContent, "password"),
			Confidence: 0.75,
			Implications: &acmv1.RuleImplications{
				MentionsCredentialRotation: true,
			},
		})
	}

	// If no specific rules matched, default to ALLOWED for common sites
	if len(rules) == 0 {
		recommendation = acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED
		reasoning = "No automation restrictions found in ToS; automation appears to be allowed"
	}

	crc := &acmv1.ComplianceRuleSet{
		Id:             "", // Will be set by CRC Manager
		Site:           site,
		TosUrl:         tosURL,
		TosVersion:     tosVersion,
		TosHash:        tosHash,
		ParsedAt:       timestamppb.New(now),
		ExpiresAt:      timestamppb.New(now.Add(30 * 24 * time.Hour)),
		Rules:          rules,
		Recommendation: recommendation,
		Reasoning:      reasoning,
		Signature:      "", // Will be set by CRC Manager
	}

	return crc
}

// containsKeywords checks if the text contains any of the keywords.
func (e *Engine) containsKeywords(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// extractContext extracts a snippet of context around a keyword.
func (e *Engine) extractContext(text string, keyword string) string {
	lower := strings.ToLower(text)
	keywordLower := strings.ToLower(keyword)

	index := strings.Index(lower, keywordLower)
	if index == -1 {
		return ""
	}

	// Extract 100 characters before and after
	start := index - 100
	if start < 0 {
		start = 0
	}

	end := index + len(keyword) + 100
	if end > len(text) {
		end = len(text)
	}

	context := text[start:end]
	return "..." + strings.TrimSpace(context) + "..."
}

// AnalyzeBatch analyzes multiple ToS documents in batch.
// Useful for testing or bulk analysis.
func (e *Engine) AnalyzeBatch(ctx context.Context, requests []*AnalysisRequest) ([]*acmv1.ComplianceRuleSet, error) {
	results := make([]*acmv1.ComplianceRuleSet, 0, len(requests))

	for _, req := range requests {
		crc, err := e.AnalyzeToS(ctx, req.TosContent, req.Site, req.TosURL)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", req.Site, err)
		}
		results = append(results, crc)
	}

	return results, nil
}

// AnalysisRequest represents a ToS analysis request.
type AnalysisRequest struct {
	Site       string
	TosURL     string
	TosContent string
}

// TrainingData represents training data for the NLP model.
// This will be used in Phase III for model training.
type TrainingData struct {
	TosContent string
	Site       string
	Labels     []Label
}

// Label represents an annotated label for training.
type Label struct {
	Category   acmv1.RuleCategory
	Text       string
	Start      int
	End        int
	Confidence float32
}

// TODO Phase III: Implement real NLP integration
// - Subprocess to Python spaCy engine
// - gRPC service for NLP analysis
// - Model training pipeline
// - Fine-tuning on legal ToS corpus
