package crc

import (
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// Summary provides a summary of a cached CRC.
type Summary struct {
	ID             string
	Site           string
	ParsedAt       time.Time
	ExpiresAt      time.Time
	Recommendation acmv1.ComplianceRecommendation
	RuleCount      int32
	Expired        bool
}
