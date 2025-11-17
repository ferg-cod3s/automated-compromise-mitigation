// Package acvs provides the Automated Compliance Validation Service.
// ACVS validates automation actions against target site Terms of Service,
// generates evidence chains for compliance proof, and manages legal NLP analysis.
package acvs

import (
	"context"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// Service defines the main ACVS service interface.
type Service interface {
	// IsEnabled returns whether ACVS is currently enabled (user has opted in).
	IsEnabled() bool

	// Enable opts the user into ACVS (requires EULA acceptance).
	Enable(ctx context.Context, eulaVersion string, consent bool) error

	// Disable opts the user out of ACVS.
	Disable(ctx context.Context, clearCache bool, preserveEvidence bool) error

	// AnalyzeToS fetches and analyzes a website's Terms of Service.
	AnalyzeToS(ctx context.Context, site string, tosURL string, forceRefresh bool, timeoutSecs int32) (*acmv1.ComplianceRuleSet, error)

	// ValidateAction performs pre-flight validation of an automation action.
	ValidateAction(ctx context.Context, site string, action *acmv1.AutomationAction, credentialID string, forceRefresh bool) (*ValidationResult, error)

	// GetCRC retrieves a cached Compliance Rule Set for a site.
	GetCRC(ctx context.Context, site string) (*acmv1.ComplianceRuleSet, bool, error)

	// ListCRCs lists all cached CRCs.
	ListCRCs(ctx context.Context, siteFilter string, includeExpired bool) ([]CRCSummary, error)

	// InvalidateCRC removes a cached CRC.
	InvalidateCRC(ctx context.Context, site string) error

	// ExportEvidenceChain exports evidence chain entries for a time range.
	ExportEvidenceChain(ctx context.Context, req *ExportRequest) ([]*acmv1.EvidenceChainEntry, error)

	// GetStatus returns current ACVS status and configuration.
	GetStatus(ctx context.Context) (*Status, error)

	// GetStatistics returns ACVS usage statistics.
	GetStatistics(ctx context.Context) (*Statistics, error)
}

// ValidationResult contains the result of an ACVS validation check.
type ValidationResult struct {
	// Result of validation
	Result acmv1.ValidationResult

	// Recommended automation method
	RecommendedMethod acmv1.AutomationMethod

	// CRC rules that apply to this action
	ApplicableRuleIDs []string

	// Detailed reasoning
	Reasoning string

	// Evidence entry ID (for audit trail)
	EvidenceEntryID string

	// Error message if validation failed
	ErrorMessage string
}

// CRCSummary provides a summary of a cached CRC.
type CRCSummary struct {
	ID             string
	Site           string
	ParsedAt       time.Time
	ExpiresAt      time.Time
	Recommendation acmv1.ComplianceRecommendation
	RuleCount      int32
	Expired        bool
}

// ExportRequest specifies parameters for evidence chain export.
type ExportRequest struct {
	CredentialID     string
	StartTime        time.Time
	EndTime          time.Time
	Format           acmv1.EvidenceExportFormat
	IncludeCRCSnapshots bool
}

// Status represents the current ACVS operational status.
type Status struct {
	Enabled       bool
	EULAVersion   string
	EnabledAt     time.Time
	Configuration Configuration
	Statistics    Statistics
}

// Configuration holds ACVS configuration parameters.
type Configuration struct {
	NLPModelVersion    string
	CacheTTLSeconds    int64
	EvidenceChainEnabled bool
	DefaultOnUncertain acmv1.ValidationResult
	ModelPath          string
}

// Statistics tracks ACVS usage metrics.
type Statistics struct {
	TotalAnalyses        int64
	TotalValidations     int64
	ValidationsAllowed   int64
	ValidationsHIMRequired int64
	ValidationsBlocked   int64
	CRCsCached          int32
	EvidenceEntries     int64
}

// CRCManager manages Compliance Rule Set caching and versioning.
type CRCManager interface {
	// Store saves a CRC to the cache.
	Store(ctx context.Context, crc *acmv1.ComplianceRuleSet) error

	// Get retrieves a CRC from cache.
	Get(ctx context.Context, site string) (*acmv1.ComplianceRuleSet, bool, error)

	// List returns all cached CRCs matching the filter.
	List(ctx context.Context, siteFilter string, includeExpired bool) ([]CRCSummary, error)

	// Invalidate removes a CRC from cache.
	Invalidate(ctx context.Context, site string) error

	// IsExpired checks if a CRC has expired.
	IsExpired(crc *acmv1.ComplianceRuleSet) bool

	// GetCacheTTL returns the configured cache TTL.
	GetCacheTTL() time.Duration

	// SetCacheTTL sets the cache TTL.
	SetCacheTTL(ttl time.Duration)
}

// Validator validates automation actions against CRC rules.
type Validator interface {
	// Validate performs pre-flight validation.
	Validate(ctx context.Context, crc *acmv1.ComplianceRuleSet, action *acmv1.AutomationAction) (*ValidationResult, error)

	// CheckRateLimit checks if action would exceed rate limits.
	CheckRateLimit(ctx context.Context, site string, action *acmv1.AutomationAction) (bool, error)

	// GetApplicableRules returns rules that apply to the given action.
	GetApplicableRules(crc *acmv1.ComplianceRuleSet, action *acmv1.AutomationAction) []*acmv1.ComplianceRule

	// DetermineRecommendation determines overall recommendation from rules.
	DetermineRecommendation(rules []*acmv1.ComplianceRule) (acmv1.ComplianceRecommendation, string)

	// RecommendMethod recommends the best automation method.
	RecommendMethod(recommendation acmv1.ComplianceRecommendation, action *acmv1.AutomationAction) acmv1.AutomationMethod
}

// EvidenceChainGenerator generates cryptographically-signed evidence chains.
type EvidenceChainGenerator interface {
	// AddEntry adds a new entry to the evidence chain.
	AddEntry(ctx context.Context, entry *EvidenceEntry) (string, error)

	// GetEntry retrieves an evidence entry by ID.
	GetEntry(ctx context.Context, entryID string) (*acmv1.EvidenceChainEntry, error)

	// Export exports evidence entries for a time range.
	Export(ctx context.Context, req *ExportRequest) ([]*acmv1.EvidenceChainEntry, error)

	// Verify verifies the integrity of an evidence entry.
	Verify(ctx context.Context, entry *acmv1.EvidenceChainEntry) (bool, error)

	// VerifyChain verifies the integrity of the entire chain.
	VerifyChain(ctx context.Context) (bool, []string, error)

	// GetChainHead returns the most recent evidence entry ID.
	GetChainHead(ctx context.Context) (string, error)
}

// EvidenceEntry represents data for a new evidence chain entry.
type EvidenceEntry struct {
	EventType          acmv1.EvidenceEventType
	Site               string
	CredentialIDHash   string
	Action             *acmv1.AutomationAction
	ValidationResult   acmv1.ValidationResult
	CRCID              string
	AppliedRuleIDs     []string
	EvidenceData       map[string]interface{}
}

// NLPEngine interfaces with the Legal NLP engine for ToS analysis.
type NLPEngine interface {
	// AnalyzeToS analyzes a ToS document and generates a CRC.
	AnalyzeToS(ctx context.Context, tosContent string, site string, tosURL string) (*acmv1.ComplianceRuleSet, error)

	// IsAvailable checks if the NLP engine is available.
	IsAvailable(ctx context.Context) bool

	// GetModelVersion returns the NLP model version.
	GetModelVersion() string

	// GetModelPath returns the path to the NLP model.
	GetModelPath() string

	// SetModelPath sets the path to the NLP model.
	SetModelPath(path string)
}

// ToSFetcher fetches Terms of Service content from websites.
type ToSFetcher interface {
	// FetchToS fetches ToS content from a URL.
	FetchToS(ctx context.Context, url string) (string, error)

	// DiscoverToSURL attempts to discover the ToS URL for a site.
	DiscoverToSURL(ctx context.Context, site string) (string, error)
}
