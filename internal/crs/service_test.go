package crs

import (
	"context"
	"testing"

	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// TestGeneratePassword tests the password generation functionality
func TestGeneratePassword(t *testing.T) {
	// Create a service with nil password manager (not needed for password generation)
	service := NewService(nil, nil)

	tests := []struct {
		name        string
		policy      pwmanager.PasswordPolicy
		expectError bool
		validate    func(t *testing.T, password string)
	}{
		{
			name: "default policy",
			policy: pwmanager.PasswordPolicy{
				Length:           16,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
			},
			expectError: false,
			validate: func(t *testing.T, password string) {
				if len(password) != 16 {
					t.Errorf("Expected password length 16, got %d", len(password))
				}
				if !containsUppercase(password) {
					t.Error("Password should contain uppercase letters")
				}
				if !containsLowercase(password) {
					t.Error("Password should contain lowercase letters")
				}
				if !containsNumber(password) {
					t.Error("Password should contain numbers")
				}
				if !containsSymbol(password) {
					t.Error("Password should contain symbols")
				}
			},
		},
		{
			name: "long password",
			policy: pwmanager.PasswordPolicy{
				Length:           32,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
			},
			expectError: false,
			validate: func(t *testing.T, password string) {
				if len(password) != 32 {
					t.Errorf("Expected password length 32, got %d", len(password))
				}
			},
		},
		{
			name: "minimum length",
			policy: pwmanager.PasswordPolicy{
				Length:           12,
				RequireUppercase: false,
				RequireLowercase: true,
				RequireNumbers:   false,
				RequireSymbols:   false,
			},
			expectError: false,
			validate: func(t *testing.T, password string) {
				if len(password) != 12 {
					t.Errorf("Expected password length 12, got %d", len(password))
				}
			},
		},
		{
			name: "too short - should fail",
			policy: pwmanager.PasswordPolicy{
				Length:           8,
				RequireUppercase: true,
				RequireLowercase: true,
			},
			expectError: true,
			validate:    nil,
		},
		{
			name: "too long - should fail",
			policy: pwmanager.PasswordPolicy{
				Length:           256,
				RequireUppercase: true,
				RequireLowercase: true,
			},
			expectError: true,
			validate:    nil,
		},
		{
			name: "no character types - should fail",
			policy: pwmanager.PasswordPolicy{
				Length:           16,
				RequireUppercase: false,
				RequireLowercase: false,
				RequireNumbers:   false,
				RequireSymbols:   false,
			},
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := service.GeneratePassword(context.Background(), tt.policy)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, password)
			}
		})
	}
}

// TestPasswordUniqueness ensures generated passwords are unique
func TestPasswordUniqueness(t *testing.T) {
	service := NewService(nil, nil)

	policy := pwmanager.PasswordPolicy{
		Length:           16,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
	}

	passwords := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		password, err := service.GeneratePassword(context.Background(), policy)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}

		if passwords[password] {
			t.Errorf("Generated duplicate password: %s", password)
		}
		passwords[password] = true
	}

	if len(passwords) != iterations {
		t.Errorf("Expected %d unique passwords, got %d", iterations, len(passwords))
	}
}

// TestPasswordStrength tests that generated passwords meet complexity requirements
func TestPasswordStrength(t *testing.T) {
	service := NewService(nil, nil)

	policy := pwmanager.PasswordPolicy{
		Length:           20,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
	}

	// Generate multiple passwords and verify they all meet requirements
	for i := 0; i < 50; i++ {
		password, err := service.GeneratePassword(context.Background(), policy)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}

		// Verify all requirements are met
		hasUpper := false
		hasLower := false
		hasNumber := false
		hasSymbol := false

		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
			} else if char >= 'a' && char <= 'z' {
				hasLower = true
			} else if char >= '0' && char <= '9' {
				hasNumber = true
			} else {
				hasSymbol = true
			}
		}

		if policy.RequireUppercase && !hasUpper {
			t.Errorf("Password missing uppercase: %s", password)
		}
		if policy.RequireLowercase && !hasLower {
			t.Errorf("Password missing lowercase: %s", password)
		}
		if policy.RequireNumbers && !hasNumber {
			t.Errorf("Password missing number: %s", password)
		}
		if policy.RequireSymbols && !hasSymbol {
			t.Errorf("Password missing symbol: %s", password)
		}
	}
}

// Helper functions for validation
func containsUppercase(s string) bool {
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, char := range s {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

func containsSymbol(s string) bool {
	symbols := "!@#$%^&*()-_=+[]{}|;:,.<>?/"
	for _, char := range s {
		for _, symbol := range symbols {
			if char == symbol {
				return true
			}
		}
	}
	return false
}
