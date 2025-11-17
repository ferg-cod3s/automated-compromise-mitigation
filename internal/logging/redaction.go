// Package logging provides sensitive data redaction for logs.
package logging

import (
	"regexp"
	"strings"
)

// SensitivePatterns holds compiled regular expressions for detecting sensitive data.
var SensitivePatterns = struct {
	// API keys and tokens
	GitHubToken    *regexp.Regexp
	GitLabToken    *regexp.Regexp
	GenericAPIKey  *regexp.Regexp
	BearerToken    *regexp.Regexp
	AWSAccessKey   *regexp.Regexp
	AWSSecretKey   *regexp.Regexp

	// Passwords and secrets
	Password       *regexp.Regexp
	Secret         *regexp.Regexp
	PrivateKey     *regexp.Regexp

	// Personal information
	Email          *regexp.Regexp
	CreditCard     *regexp.Regexp
	SSN            *regexp.Regexp

	// URLs with embedded credentials
	URLWithCreds   *regexp.Regexp

	// JWT tokens
	JWT            *regexp.Regexp
}{
	// GitHub tokens (classic: ghp_, fine-grained: github_pat_)
	GitHubToken:   regexp.MustCompile(`(ghp_[a-zA-Z0-9]{30,}|github_pat_[a-zA-Z0-9_]{22,})`),

	// GitLab tokens (glpat-)
	GitLabToken:   regexp.MustCompile(`glpat-[a-zA-Z0-9_-]{20,}`),

	// Generic API keys (common patterns)
	GenericAPIKey: regexp.MustCompile(`(?i)(api[_-]?key|apikey|access[_-]?token)[\s:=]+['"` + "`" + `]?([a-zA-Z0-9_\-]{12,})['"` + "`" + `]?`),

	// Bearer tokens in Authorization headers
	BearerToken:   regexp.MustCompile(`(?i)bearer\s+([a-zA-Z0-9_\-\.]+)`),

	// AWS credentials
	AWSAccessKey:  regexp.MustCompile(`(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`),
	AWSSecretKey:  regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key)[\s:=]+['"` + "`" + `]?[a-zA-Z0-9/+=]{40}['"` + "`" + `]?`),

	// Passwords in various formats
	Password:      regexp.MustCompile(`(?i)(password|passwd|pwd)[\s:=]+['"` + "`" + `]?([^\s'"` + "`" + `]{6,})['"` + "`" + `]?`),

	// Generic secrets
	Secret:        regexp.MustCompile(`(?i)(secret|private[_-]?key)[\s:=]+['"` + "`" + `]?([a-zA-Z0-9_\-+=/.]{20,})['"` + "`" + `]?`),

	// Private keys (PEM format)
	PrivateKey:    regexp.MustCompile(`-----BEGIN\s+(?:RSA|EC|OPENSSH|DSA)?\s*PRIVATE KEY-----[^-]+-----END\s+(?:RSA|EC|OPENSSH|DSA)?\s*PRIVATE KEY-----`),

	// Email addresses
	Email:         regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),

	// Credit card numbers (simple pattern, 13-19 digits with optional spaces/dashes)
	CreditCard:    regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4,7}\b`),

	// SSN (US Social Security Number)
	SSN:           regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),

	// URLs with embedded credentials (http://user:pass@host)
	URLWithCreds:  regexp.MustCompile(`(https?://)[^:]+:([^@]+)@`),

	// JWT tokens (header.payload.signature)
	JWT:           regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
}

// SensitiveKeys lists attribute keys that should be redacted if they contain values.
// Note: Keys should use underscores (not hyphens) as they are normalized before lookup.
var SensitiveKeys = map[string]bool{
	"password":          true,
	"passwd":            true,
	"pwd":               true,
	"secret":            true,
	"api_key":           true,
	"apikey":            true,
	"access_token":      true,
	"refresh_token":     true,
	"token":             true,
	"authorization":     true,
	"auth":              true,
	"private_key":       true,
	"master_password":   true,
	"vault_key":         true,
	"encryption_key":    true,
	"session_id":        true,
	"cookie":            true,
	"x_api_key":         true, // Normalized from x-api-key
	"x_auth_token":      true, // Normalized from x-auth-token
	"x_session_token":   true, // Normalized from x-session-token
}

// RedactionMode determines how aggressively to redact data.
type RedactionMode int

const (
	// RedactNone disables redaction (DANGEROUS - only for testing)
	RedactNone RedactionMode = iota

	// RedactStandard redacts known patterns (default)
	RedactStandard

	// RedactAggressive redacts standard patterns plus emails and IPs
	RedactAggressive

	// RedactParanoid redacts everything that looks remotely sensitive
	RedactParanoid
)

const (
	// RedactedPlaceholder is the text used to replace redacted values
	RedactedPlaceholder = "[REDACTED]"

	// RedactedHashPlaceholder shows a hash of the redacted value for correlation
	RedactedHashPlaceholder = "[REDACTED:%s]"
)

// RedactionConfig configures the redaction behavior.
type RedactionConfig struct {
	// Mode sets the redaction aggressiveness
	Mode RedactionMode

	// ShowHashes includes SHA256 hash of redacted value for correlation
	ShowHashes bool

	// Whitelist of keys that should NOT be redacted (even if they match patterns)
	Whitelist map[string]bool

	// CustomPatterns for domain-specific sensitive data
	CustomPatterns []*regexp.Regexp
}

// DefaultRedactionConfig returns the default redaction configuration.
func DefaultRedactionConfig() RedactionConfig {
	return RedactionConfig{
		Mode:           RedactStandard,
		ShowHashes:     false,
		Whitelist:      make(map[string]bool),
		CustomPatterns: nil,
	}
}

// RedactString redacts sensitive data from a string.
func RedactString(s string, config RedactionConfig) string {
	if config.Mode == RedactNone {
		return s
	}

	result := s

	// Redact GitHub tokens
	result = SensitivePatterns.GitHubToken.ReplaceAllString(result, RedactedPlaceholder)

	// Redact GitLab tokens
	result = SensitivePatterns.GitLabToken.ReplaceAllString(result, RedactedPlaceholder)

	// Redact AWS credentials
	result = SensitivePatterns.AWSAccessKey.ReplaceAllString(result, RedactedPlaceholder)
	result = SensitivePatterns.AWSSecretKey.ReplaceAllString(result, "$1="+RedactedPlaceholder)

	// Redact Bearer tokens
	result = SensitivePatterns.BearerToken.ReplaceAllString(result, "Bearer "+RedactedPlaceholder)

	// Redact passwords
	result = SensitivePatterns.Password.ReplaceAllString(result, "$1="+RedactedPlaceholder)

	// Redact generic secrets
	result = SensitivePatterns.Secret.ReplaceAllString(result, "$1="+RedactedPlaceholder)

	// Redact private keys
	result = SensitivePatterns.PrivateKey.ReplaceAllString(result, RedactedPlaceholder)

	// Redact URLs with credentials
	result = SensitivePatterns.URLWithCreds.ReplaceAllString(result, "$1"+RedactedPlaceholder+"@")

	// Redact JWT tokens
	result = SensitivePatterns.JWT.ReplaceAllString(result, RedactedPlaceholder)

	// Redact generic API keys
	result = SensitivePatterns.GenericAPIKey.ReplaceAllString(result, "$1="+RedactedPlaceholder)

	// Mode-specific redactions
	if config.Mode >= RedactAggressive {
		// Redact email addresses
		result = SensitivePatterns.Email.ReplaceAllString(result, RedactedPlaceholder)

		// Redact credit cards
		result = SensitivePatterns.CreditCard.ReplaceAllString(result, RedactedPlaceholder)

		// Redact SSNs
		result = SensitivePatterns.SSN.ReplaceAllString(result, RedactedPlaceholder)
	}

	// Custom patterns
	for _, pattern := range config.CustomPatterns {
		result = pattern.ReplaceAllString(result, RedactedPlaceholder)
	}

	return result
}

// RedactValue redacts a value based on its key name.
func RedactValue(key string, value interface{}, config RedactionConfig) interface{} {
	if config.Mode == RedactNone {
		return value
	}

	// Check whitelist
	if config.Whitelist[key] {
		return value
	}

	// Normalize key (lowercase, replace - with _)
	normalizedKey := strings.ToLower(strings.ReplaceAll(key, "-", "_"))

	// Check if key is sensitive
	if SensitiveKeys[normalizedKey] {
		return RedactedPlaceholder
	}

	// Check if value is a string and contains sensitive patterns
	if strValue, ok := value.(string); ok {
		redacted := RedactString(strValue, config)
		if redacted != strValue {
			return redacted
		}
	}

	return value
}

// IsSensitiveKey returns true if the key name suggests sensitive data.
func IsSensitiveKey(key string) bool {
	normalizedKey := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	return SensitiveKeys[normalizedKey]
}

// RedactMap redacts sensitive values in a map.
func RedactMap(m map[string]interface{}, config RedactionConfig) map[string]interface{} {
	if config.Mode == RedactNone {
		return m
	}

	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = RedactValue(k, v, config)
	}
	return result
}

// ShouldRedactAttribute determines if a log attribute should be redacted.
func ShouldRedactAttribute(key string) bool {
	return IsSensitiveKey(key)
}

// RedactEmail redacts an email address, optionally preserving the domain.
func RedactEmail(email string, preserveDomain bool) string {
	if !preserveDomain {
		return RedactedPlaceholder
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return RedactedPlaceholder
	}

	return RedactedPlaceholder + "@" + parts[1]
}

// RedactToken redacts a token but preserves a prefix for identification.
func RedactToken(token string, prefixLen int) string {
	if len(token) <= prefixLen {
		return RedactedPlaceholder
	}

	return token[:prefixLen] + "..." + RedactedPlaceholder
}

// RedactAWSKey redacts an AWS access key but preserves the prefix for correlation.
func RedactAWSKey(key string) string {
	if len(key) < 8 {
		return RedactedPlaceholder
	}

	// Preserve first 4 chars (e.g., "AKIA") for key type identification
	return key[:4] + "..." + RedactedPlaceholder
}

// RedactGitHubToken redacts a GitHub token but preserves the prefix.
func RedactGitHubToken(token string) string {
	if strings.HasPrefix(token, "ghp_") {
		return "ghp_..." + RedactedPlaceholder
	}
	if strings.HasPrefix(token, "github_pat_") {
		return "github_pat_..." + RedactedPlaceholder
	}
	return RedactedPlaceholder
}
