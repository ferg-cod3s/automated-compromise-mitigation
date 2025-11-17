package logging

import (
	"regexp"
	"testing"
)

func TestRedactString(t *testing.T) {
	config := DefaultRedactionConfig()

	tests := []struct {
		name     string
		input    string
		expected string
		mode     RedactionMode
	}{
		{
			name:     "GitHub classic token",
			input:    "Token: ghp_1234567890abcdefghijklmnopqrstuv",
			expected: "Token: [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "GitHub fine-grained PAT",
			input:    "PAT: github_pat_12345678901234567890_12345678901234567890123456789012345678901234567890123456789",
			expected: "PAT: [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "GitLab token",
			input:    "GitLab: glpat-abcdefghijklmnopqrst",
			expected: "GitLab: [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "AWS access key",
			input:    "AWS Key: AKIAIOSFODNN7EXAMPLE",
			expected: "AWS Key: [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "AWS secret key",
			input:    "aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			expected: "aws_secret_access_key=[REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature",
			expected: "Authorization: Bearer [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "Password in key-value",
			input:    "password=mysecretpassword123",
			expected: "password=[REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "URL with credentials",
			input:    "https://user:password123@github.com/repo.git",
			expected: "https://[REDACTED]@github.com/repo.git",
			mode:     RedactStandard,
		},
		{
			name:     "JWT token",
			input:    "JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expected: "JWT: [REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "API key",
			input:    "api_key: sk_test_1234567890abcdefghij",
			expected: "api_key=[REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "Private key",
			input:    "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQ...\n-----END RSA PRIVATE KEY-----",
			expected: "[REDACTED]",
			mode:     RedactStandard,
		},
		{
			name:     "Email (standard mode - not redacted)",
			input:    "Contact: user@example.com",
			expected: "Contact: user@example.com",
			mode:     RedactStandard,
		},
		{
			name:     "Email (aggressive mode - redacted)",
			input:    "Contact: user@example.com",
			expected: "Contact: [REDACTED]",
			mode:     RedactAggressive,
		},
		{
			name:     "Credit card (aggressive mode)",
			input:    "CC: 4532-1234-5678-9010",
			expected: "CC: [REDACTED]",
			mode:     RedactAggressive,
		},
		{
			name:     "SSN (aggressive mode)",
			input:    "SSN: 123-45-6789",
			expected: "SSN: [REDACTED]",
			mode:     RedactAggressive,
		},
		{
			name:     "Multiple sensitive patterns",
			input:    "password=secret123 api_key=abc123def456 token=xyz789",
			expected: "password=[REDACTED] api_key=[REDACTED] token=xyz789",
			mode:     RedactStandard,
		},
		{
			name:     "No sensitive data",
			input:    "This is a normal log message with no secrets",
			expected: "This is a normal log message with no secrets",
			mode:     RedactStandard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Mode = tt.mode
			result := RedactString(tt.input, config)
			if result != tt.expected {
				t.Errorf("RedactString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRedactValue(t *testing.T) {
	config := DefaultRedactionConfig()

	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "password key",
			key:      "password",
			value:    "mysecretpassword",
			expected: "[REDACTED]",
		},
		{
			name:     "api_key key",
			key:      "api_key",
			value:    "sk_test_1234567890",
			expected: "[REDACTED]",
		},
		{
			name:     "token key",
			key:      "token",
			value:    "abc123xyz789",
			expected: "[REDACTED]",
		},
		{
			name:     "authorization key",
			key:      "authorization",
			value:    "Bearer token123",
			expected: "[REDACTED]",
		},
		{
			name:     "non-sensitive key",
			key:      "username",
			value:    "john.doe",
			expected: "john.doe",
		},
		{
			name:     "non-sensitive integer",
			key:      "count",
			value:    42,
			expected: 42,
		},
		{
			name:     "key with hyphen",
			key:      "api-key",
			value:    "secret123",
			expected: "[REDACTED]",
		},
		{
			name:     "uppercase key",
			key:      "PASSWORD",
			value:    "secret",
			expected: "[REDACTED]",
		},
		{
			name:     "value with GitHub token",
			key:      "config",
			value:    "token=ghp_1234567890abcdefghijklmnopqrstuv",
			expected: "token=[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactValue(tt.key, tt.value, config)
			if result != tt.expected {
				t.Errorf("RedactValue(%q, %v) = %v, want %v", tt.key, tt.value, result, tt.expected)
			}
		})
	}
}

func TestRedactValueWithWhitelist(t *testing.T) {
	config := DefaultRedactionConfig()
	config.Whitelist = map[string]bool{
		"password": true, // Whitelist password key
	}

	result := RedactValue("password", "mysecret", config)
	if result != "mysecret" {
		t.Errorf("Whitelisted key should not be redacted, got %v", result)
	}

	// Non-whitelisted sensitive key should still be redacted
	result = RedactValue("api_key", "secret", config)
	if result != "[REDACTED]" {
		t.Errorf("Non-whitelisted sensitive key should be redacted, got %v", result)
	}
}

func TestRedactMap(t *testing.T) {
	config := DefaultRedactionConfig()

	input := map[string]interface{}{
		"username": "john.doe",
		"password": "secret123",
		"api_key":  "sk_test_abc123",
		"count":    42,
		"token":    "xyz789",
	}

	result := RedactMap(input, config)

	if result["username"] != "john.doe" {
		t.Errorf("Non-sensitive value should not be redacted")
	}
	if result["password"] != "[REDACTED]" {
		t.Errorf("Password should be redacted")
	}
	if result["api_key"] != "[REDACTED]" {
		t.Errorf("API key should be redacted")
	}
	if result["count"] != 42 {
		t.Errorf("Integer should not be redacted")
	}
	if result["token"] != "[REDACTED]" {
		t.Errorf("Token should be redacted")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"password", true},
		{"api_key", true},
		{"token", true},
		{"secret", true},
		{"username", false},
		{"email", false},
		{"count", false},
		{"PASSWORD", true}, // Case insensitive
		{"api-key", true},  // Hyphen handling
		{"x-api-key", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := IsSensitiveKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestRedactEmail(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		preserveDomain  bool
		expected        string
	}{
		{
			name:           "full redaction",
			email:          "user@example.com",
			preserveDomain: false,
			expected:       "[REDACTED]",
		},
		{
			name:           "preserve domain",
			email:          "user@example.com",
			preserveDomain: true,
			expected:       "[REDACTED]@example.com",
		},
		{
			name:           "invalid email",
			email:          "notanemail",
			preserveDomain: true,
			expected:       "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactEmail(tt.email, tt.preserveDomain)
			if result != tt.expected {
				t.Errorf("RedactEmail(%q, %v) = %q, want %q", tt.email, tt.preserveDomain, result, tt.expected)
			}
		})
	}
}

func TestRedactToken(t *testing.T) {
	token := "sk_test_1234567890abcdefghij"

	result := RedactToken(token, 7)
	expected := "sk_test...[REDACTED]"

	if result != expected {
		t.Errorf("RedactToken() = %q, want %q", result, expected)
	}

	// Short token
	shortToken := "abc"
	result = RedactToken(shortToken, 5)
	if result != "[REDACTED]" {
		t.Errorf("Short token should be fully redacted, got %q", result)
	}
}

func TestRedactAWSKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "valid AKIA key",
			key:      "AKIAIOSFODNN7EXAMPLE",
			expected: "AKIA...[REDACTED]",
		},
		{
			name:     "valid ASIA key",
			key:      "ASIAIOSFODNN7EXAMPLE",
			expected: "ASIA...[REDACTED]",
		},
		{
			name:     "short key",
			key:      "AKIA",
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactAWSKey(tt.key)
			if result != tt.expected {
				t.Errorf("RedactAWSKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestRedactGitHubToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "classic token",
			token:    "ghp_1234567890abcdefghijklmnopqrstuv",
			expected: "ghp_...[REDACTED]",
		},
		{
			name:     "fine-grained PAT",
			token:    "github_pat_12345678901234567890_1234567890",
			expected: "github_pat_...[REDACTED]",
		},
		{
			name:     "unknown token format",
			token:    "unknown_token_format",
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactGitHubToken(tt.token)
			if result != tt.expected {
				t.Errorf("RedactGitHubToken(%q) = %q, want %q", tt.token, result, tt.expected)
			}
		})
	}
}

func TestRedactNoneMode(t *testing.T) {
	config := DefaultRedactionConfig()
	config.Mode = RedactNone

	input := "password=secret api_key=123456"
	result := RedactString(input, config)

	if result != input {
		t.Errorf("RedactNone mode should not redact anything, got %q", result)
	}
}

func TestCustomPatterns(t *testing.T) {
	config := DefaultRedactionConfig()
	config.CustomPatterns = []*regexp.Regexp{
		regexp.MustCompile(`CUSTOM-\d{6}`),
	}

	input := "Custom ID: CUSTOM-123456"
	result := RedactString(input, config)
	expected := "Custom ID: [REDACTED]"

	if result != expected {
		t.Errorf("Custom pattern not redacted: got %q, want %q", result, expected)
	}
}

func BenchmarkRedactString(b *testing.B) {
	config := DefaultRedactionConfig()
	input := "password=secret api_key=sk_test_123 token=ghp_abc123 email=user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RedactString(input, config)
	}
}

func BenchmarkRedactStringNoMatch(b *testing.B) {
	config := DefaultRedactionConfig()
	input := "This is a normal log message with no sensitive data at all"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RedactString(input, config)
	}
}

func BenchmarkIsSensitiveKey(b *testing.B) {
	keys := []string{"password", "username", "api_key", "email", "token"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			_ = IsSensitiveKey(key)
		}
	}
}

func TestShouldRedactAttribute(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"password", true},
		{"token", true},
		{"api_key", true},
		{"secret", true},
		{"username", false},
		{"email", false},
		{"data", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := ShouldRedactAttribute(tt.key)
			if result != tt.expected {
				t.Errorf("ShouldRedactAttribute(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}
