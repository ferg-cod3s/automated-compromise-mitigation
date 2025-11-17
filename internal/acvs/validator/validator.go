// Package validator provides compliance validation logic for ACVS.
package validator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// Validator implements the Validator interface for compliance validation.
type Validator struct {
	mu                 sync.RWMutex
	rateLimitTracker   map[string]*rateLimitEntry
	defaultOnUncertain acmv1.ValidationResult
}

// rateLimitEntry tracks rate limit usage for a site.
type rateLimitEntry struct {
	Site         string
	RequestCount int
	WindowStart  time.Time
	WindowEnd    time.Time
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{
		rateLimitTracker:   make(map[string]*rateLimitEntry),
		defaultOnUncertain: acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED,
	}
}

// NewValidatorWithDefaults creates a new Validator with custom defaults.
func NewValidatorWithDefaults(defaultOnUncertain acmv1.ValidationResult) *Validator {
	return &Validator{
		rateLimitTracker:   make(map[string]*rateLimitEntry),
		defaultOnUncertain: defaultOnUncertain,
	}
}

// Validate performs pre-flight validation of an automation action against CRC rules.
func (v *Validator) Validate(ctx context.Context, crc *acmv1.ComplianceRuleSet, action *acmv1.AutomationAction) (*Result, error) {
	if crc == nil {
		return nil, fmt.Errorf("CRC cannot be nil")
	}

	if action == nil {
		return nil, fmt.Errorf("action cannot be nil")
	}

	// Get applicable rules
	applicableRules := v.GetApplicableRules(crc, action)

	// If no applicable rules, use overall recommendation
	if len(applicableRules) == 0 {
		return v.validateWithOverallRecommendation(crc, action), nil
	}

	// Determine recommendation from applicable rules
	recommendation, reasoning := v.DetermineRecommendation(applicableRules)

	// Recommend automation method
	method := v.RecommendMethod(recommendation, action)

	// Convert recommendation to validation result
	result := v.recommendationToValidationResult(recommendation)

	// Check rate limits if applicable
	rateLimited, err := v.CheckRateLimit(ctx, crc.Site, action)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	if rateLimited {
		result = acmv1.ValidationResult_VALIDATION_RESULT_RATE_LIMITED
		reasoning = fmt.Sprintf("Rate limit exceeded. %s", reasoning)
	}

	// Extract rule IDs
	ruleIDs := make([]string, len(applicableRules))
	for i, rule := range applicableRules {
		ruleIDs[i] = rule.Id
	}

	return &Result{
		Result:            result,
		RecommendedMethod: method,
		ApplicableRuleIDs: ruleIDs,
		Reasoning:         reasoning,
	}, nil
}

// CheckRateLimit checks if an action would exceed rate limits.
func (v *Validator) CheckRateLimit(ctx context.Context, site string, action *acmv1.AutomationAction) (bool, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	entry, exists := v.rateLimitTracker[site]
	if !exists {
		return false, nil
	}

	now := time.Now()

	// Check if we're still in the rate limit window
	if now.After(entry.WindowEnd) {
		// Window expired, reset
		delete(v.rateLimitTracker, site)
		return false, nil
	}

	// For now, we'll implement a simple check
	// In a real implementation, this would use the CRC's rate limit rules
	// and track actual API calls
	return false, nil
}

// GetApplicableRules returns rules that apply to the given action.
func (v *Validator) GetApplicableRules(crc *acmv1.ComplianceRuleSet, action *acmv1.AutomationAction) []*acmv1.ComplianceRule {
	if crc == nil || len(crc.Rules) == 0 {
		return nil
	}

	applicable := make([]*acmv1.ComplianceRule, 0)

	for _, rule := range crc.Rules {
		if v.isRuleApplicable(rule, action) {
			applicable = append(applicable, rule)
		}
	}

	return applicable
}

// isRuleApplicable checks if a rule applies to the given action.
func (v *Validator) isRuleApplicable(rule *acmv1.ComplianceRule, action *acmv1.AutomationAction) bool {
	// Check by action type
	switch action.Type {
	case acmv1.ActionType_ACTION_TYPE_CREDENTIAL_ROTATION,
		acmv1.ActionType_ACTION_TYPE_PASSWORD_CHANGE:
		// These actions care about automation, API, and credentials rules
		return rule.Category == acmv1.RuleCategory_RULE_CATEGORY_AUTOMATION ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_API_USAGE ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_CREDENTIALS ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_BOTS

	case acmv1.ActionType_ACTION_TYPE_MFA_SETUP:
		return rule.Category == acmv1.RuleCategory_RULE_CATEGORY_AUTOMATION ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_API_USAGE

	default:
		// For unspecified or other actions, consider all automation-related rules
		return rule.Category == acmv1.RuleCategory_RULE_CATEGORY_AUTOMATION ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_API_USAGE ||
			rule.Category == acmv1.RuleCategory_RULE_CATEGORY_BOTS
	}
}

// DetermineRecommendation determines overall recommendation from a set of rules.
func (v *Validator) DetermineRecommendation(rules []*acmv1.ComplianceRule) (acmv1.ComplianceRecommendation, string) {
	if len(rules) == 0 {
		return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN,
			"No applicable rules found"
	}

	// Priority: BLOCKED > HIM_REQUIRED > ALLOWED_WITH_API > ALLOWED > UNCERTAIN
	hasBlocking := false
	hasHIMRequired := false
	hasAPIAllowed := false
	hasAllowed := false
	hasUncertain := false

	var reasons []string

	for _, rule := range rules {
		implications := rule.Implications

		// Critical severity rules with explicit prohibitions = BLOCKED
		if rule.Severity == acmv1.RuleSeverity_RULE_SEVERITY_CRITICAL {
			if implications != nil && implications.RequiresHumanInteraction {
				hasBlocking = true
				reasons = append(reasons, fmt.Sprintf("Critical rule %s prohibits automation", rule.Id))
			}
		}

		// High severity rules requiring human interaction = HIM_REQUIRED
		if rule.Severity >= acmv1.RuleSeverity_RULE_SEVERITY_HIGH {
			if implications != nil && implications.RequiresHumanInteraction {
				hasHIMRequired = true
				reasons = append(reasons, fmt.Sprintf("Rule %s requires human interaction", rule.Id))
			}
		}

		// Rules explicitly allowing API = ALLOWED_WITH_API
		if implications != nil && implications.AllowsApiAutomation {
			hasAPIAllowed = true
			reasons = append(reasons, fmt.Sprintf("Rule %s allows API automation", rule.Id))
		}

		// Low/Medium severity with no prohibitions = ALLOWED
		if rule.Severity <= acmv1.RuleSeverity_RULE_SEVERITY_MEDIUM {
			if implications == nil || !implications.RequiresHumanInteraction {
				hasAllowed = true
			}
		}

		// Low confidence rules = UNCERTAIN
		if rule.Confidence < 0.70 {
			hasUncertain = true
		}
	}

	// Determine final recommendation
	if hasBlocking {
		return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_BLOCKED,
			strings.Join(reasons, "; ")
	}

	if hasHIMRequired {
		return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_HIM_REQUIRED,
			strings.Join(reasons, "; ")
	}

	if hasAPIAllowed {
		return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED_WITH_API,
			strings.Join(reasons, "; ")
	}

	if hasAllowed && !hasUncertain {
		return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
			"Automation appears to be allowed based on ToS analysis"
	}

	return acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN,
		"Unable to determine compliance with sufficient confidence"
}

// RecommendMethod recommends the best automation method based on the recommendation.
func (v *Validator) RecommendMethod(recommendation acmv1.ComplianceRecommendation, action *acmv1.AutomationAction) acmv1.AutomationMethod {
	switch recommendation {
	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED_WITH_API:
		return acmv1.AutomationMethod_AUTOMATION_METHOD_API

	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED:
		// If action already specifies a method, keep it
		if action.Method != acmv1.AutomationMethod_AUTOMATION_METHOD_UNSPECIFIED {
			return action.Method
		}
		// Default to API if available, otherwise CLI
		return acmv1.AutomationMethod_AUTOMATION_METHOD_API

	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_HIM_REQUIRED,
		acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_BLOCKED,
		acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN:
		return acmv1.AutomationMethod_AUTOMATION_METHOD_MANUAL

	default:
		return acmv1.AutomationMethod_AUTOMATION_METHOD_MANUAL
	}
}

// recommendationToValidationResult converts a ComplianceRecommendation to ValidationResult.
func (v *Validator) recommendationToValidationResult(rec acmv1.ComplianceRecommendation) acmv1.ValidationResult {
	switch rec {
	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED,
		acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_ALLOWED_WITH_API:
		return acmv1.ValidationResult_VALIDATION_RESULT_ALLOWED

	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_HIM_REQUIRED:
		return acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED

	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_BLOCKED:
		return acmv1.ValidationResult_VALIDATION_RESULT_BLOCKED

	case acmv1.ComplianceRecommendation_COMPLIANCE_RECOMMENDATION_UNCERTAIN:
		return v.defaultOnUncertain

	default:
		return acmv1.ValidationResult_VALIDATION_RESULT_HIM_REQUIRED
	}
}

// validateWithOverallRecommendation uses the CRC's overall recommendation when no specific rules apply.
func (v *Validator) validateWithOverallRecommendation(crc *acmv1.ComplianceRuleSet, action *acmv1.AutomationAction) *Result {
	result := v.recommendationToValidationResult(crc.Recommendation)
	method := v.RecommendMethod(crc.Recommendation, action)

	return &Result{
		Result:            result,
		RecommendedMethod: method,
		ApplicableRuleIDs: []string{},
		Reasoning:         crc.Reasoning,
	}
}

// SetDefaultOnUncertain sets the default validation result for uncertain cases.
func (v *Validator) SetDefaultOnUncertain(result acmv1.ValidationResult) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.defaultOnUncertain = result
}

// GetDefaultOnUncertain returns the default validation result for uncertain cases.
func (v *Validator) GetDefaultOnUncertain() acmv1.ValidationResult {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.defaultOnUncertain
}

// TrackRateLimit records a rate limit for tracking purposes.
// This is called after a successful API call to track usage.
func (v *Validator) TrackRateLimit(site string, rateLimit *acmv1.RateLimit) {
	if rateLimit == nil {
		return
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	window := v.parseWindow(rateLimit.Window)

	entry := &rateLimitEntry{
		Site:         site,
		RequestCount: 1,
		WindowStart:  now,
		WindowEnd:    now.Add(window),
	}

	if existing, exists := v.rateLimitTracker[site]; exists {
		if now.Before(existing.WindowEnd) {
			// Still in window, increment
			existing.RequestCount++
		} else {
			// Window expired, reset
			v.rateLimitTracker[site] = entry
		}
	} else {
		v.rateLimitTracker[site] = entry
	}
}

// parseWindow parses a window string like "1h", "24h", "60s" into a duration.
func (v *Validator) parseWindow(window string) time.Duration {
	duration, err := time.ParseDuration(window)
	if err != nil {
		// Default to 1 hour if parsing fails
		return time.Hour
	}
	return duration
}

// ClearRateLimitTracking clears all rate limit tracking data.
func (v *Validator) ClearRateLimitTracking() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.rateLimitTracker = make(map[string]*rateLimitEntry)
}
