package evidence

import (
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// Entry represents data for a new evidence chain entry.
type Entry struct {
	EventType          acmv1.EvidenceEventType
	Site               string
	CredentialIDHash   string
	Action             *acmv1.AutomationAction
	ValidationResult   acmv1.ValidationResult
	CRCID              string
	AppliedRuleIDs     []string
	EvidenceData       map[string]interface{}
}

// ExportRequest specifies parameters for evidence chain export.
type ExportRequest struct {
	CredentialID        string
	StartTime           time.Time
	EndTime             time.Time
	Format              acmv1.EvidenceExportFormat
	IncludeCRCSnapshots bool
}
