package validator

import (
	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// Result contains the result of a compliance validation check.
type Result struct {
	// Result of validation
	Result acmv1.ValidationResult

	// Recommended automation method
	RecommendedMethod acmv1.AutomationMethod

	// CRC rules that apply to this action
	ApplicableRuleIDs []string

	// Detailed reasoning
	Reasoning string

	// Error message if validation failed
	ErrorMessage string
}
