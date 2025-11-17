// Package acvsif provides interfaces for ACVS integration without import cycles.
// This package defines minimal interfaces that other packages can use to interact
// with ACVS without creating circular dependencies.
package acvsif

import (
	"context"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// ValidationResult represents the result of an ACVS validation.
type ValidationResult struct {
	Result     acmv1.ValidationResult
	CRCID      string
	Reasoning  string
	RuleIDs    []string
}

// EvidenceEntry represents an entry to add to the evidence chain.
type EvidenceEntry struct {
	EventType        acmv1.EvidenceEventType
	Site             string
	CredentialIDHash string
	Action           *acmv1.AutomationAction
	ValidationResult acmv1.ValidationResult
	CRCID            string
	AppliedRuleIDs   []string
	EvidenceData     map[string]interface{}
}

// CRCSummary provides a summary of a cached CRC.
// This type is shared between storage and acvs packages to avoid import cycles.
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
// This type is shared between storage and acvs packages to avoid import cycles.
type ExportRequest struct {
	CredentialID        string
	StartTime           time.Time
	EndTime             time.Time
	Format              acmv1.EvidenceExportFormat
	IncludeCRCSnapshots bool
}

// Service defines the minimal ACVS interface needed by rotation services.
type Service interface {
	// ValidateAction validates an automation action against a site's ToS.
	ValidateAction(ctx context.Context, site string, action *acmv1.AutomationAction) (*ValidationResult, error)

	// AddEvidenceEntry adds an entry to the evidence chain.
	AddEvidenceEntry(ctx context.Context, entry *EvidenceEntry) (string, error)

	// IsEnabled returns whether ACVS is currently enabled.
	IsEnabled() bool
}
