package logging

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "acm" {
		t.Errorf("Expected service name 'acm', got %s", config.ServiceName)
	}

	if config.Level == "" {
		t.Error("Expected non-empty log level")
	}

	if config.Format == "" {
		t.Error("Expected non-empty format")
	}

	if config.ComponentLevels == nil {
		t.Error("Expected non-nil component levels map")
	}
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	if config.Level != "debug" {
		t.Errorf("Expected debug level in development, got %s", config.Level)
	}

	if config.Format != FormatPretty {
		t.Errorf("Expected pretty format in development, got %s", config.Format)
	}

	if config.OutputMode != OutputStdout {
		t.Errorf("Expected stdout output in development, got %s", config.OutputMode)
	}
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	if config.Level != "info" {
		t.Errorf("Expected info level in production, got %s", config.Level)
	}

	if config.Format != FormatJSON {
		t.Errorf("Expected JSON format in production, got %s", config.Format)
	}

	if config.OutputMode != OutputStdout {
		t.Errorf("Expected stdout output in production, got %s", config.OutputMode)
	}
}

func TestConfig_GetComponentLevel(t *testing.T) {
	config := DefaultConfig()
	config.Level = "info"
	config.SetComponentLevel("test", "debug")

	// Component with specific level
	level := config.GetComponentLevel("test")
	if level != "debug" {
		t.Errorf("Expected 'debug' for test component, got %s", level)
	}

	// Component without specific level (should use global)
	level = config.GetComponentLevel("other")
	if level != "info" {
		t.Errorf("Expected 'info' for other component, got %s", level)
	}
}

func TestConfig_SetComponentLevel(t *testing.T) {
	config := DefaultConfig()
	config.SetComponentLevel("test1", "debug")
	config.SetComponentLevel("test2", "error")

	if config.ComponentLevels["test1"] != "debug" {
		t.Error("Failed to set component level for test1")
	}

	if config.ComponentLevels["test2"] != "error" {
		t.Error("Failed to set component level for test2")
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		level    string
		expected bool
	}{
		{"pretty format", FormatPretty, "info", true},
		{"debug level", FormatJSON, "debug", true},
		{"production", FormatJSON, "info", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Format: tt.format,
				Level:  tt.level,
			}

			result := config.IsDevelopment()
			if result != tt.expected {
				t.Errorf("IsDevelopment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		level    string
		expected bool
	}{
		{"production", FormatJSON, "info", true},
		{"debug level", FormatJSON, "debug", false},
		{"pretty format", FormatPretty, "info", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Format: tt.format,
				Level:  tt.level,
			}

			result := config.IsProduction()
			if result != tt.expected {
				t.Errorf("IsProduction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectError  bool
		expectedLevel string
		expectedFormat Format
		expectedMode OutputMode
	}{
		{
			name: "valid config",
			config: Config{
				Level:      "info",
				Format:     FormatJSON,
				OutputMode: OutputStdout,
			},
			expectError:   false,
			expectedLevel: "info",
			expectedFormat: FormatJSON,
			expectedMode:  OutputStdout,
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:      "invalid",
				Format:     FormatJSON,
				OutputMode: OutputStdout,
			},
			expectError:   false,
			expectedLevel: "info",
			expectedFormat: FormatJSON,
			expectedMode:  OutputStdout,
		},
		{
			name: "invalid format defaults to JSON",
			config: Config{
				Level:      "info",
				Format:     "invalid",
				OutputMode: OutputStdout,
			},
			expectError:   false,
			expectedLevel: "info",
			expectedFormat: FormatJSON,
			expectedMode:  OutputStdout,
		},
		{
			name: "invalid output mode defaults to stdout",
			config: Config{
				Level:      "info",
				Format:     FormatJSON,
				OutputMode: "invalid",
			},
			expectError:   false,
			expectedLevel: "info",
			expectedFormat: FormatJSON,
			expectedMode:  OutputStdout,
		},
		{
			name: "file output sets file path",
			config: Config{
				Level:      "info",
				Format:     FormatJSON,
				OutputMode: OutputFile,
				FilePath:   "",
			},
			expectError:   false,
			expectedLevel: "info",
			expectedFormat: FormatJSON,
			expectedMode:  OutputFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.config.Level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, tt.config.Level)
			}

			if tt.config.Format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, tt.config.Format)
			}

			if tt.config.OutputMode != tt.expectedMode {
				t.Errorf("Expected output mode %s, got %s", tt.expectedMode, tt.config.OutputMode)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "env var set",
			envKey:       "TEST_ENV_VAR",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "env var not set",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetEnvIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid int",
			envKey:       "TEST_INT_VAR",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "invalid int",
			envKey:       "TEST_INT_VAR",
			envValue:     "invalid",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "not set",
			envKey:       "NONEXISTENT_INT_VAR",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvIntOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetEnvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "1",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "yes",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "yes",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "not set",
			envKey:       "NONEXISTENT_BOOL_VAR",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvBoolOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()

	if info.GoVersion == "" {
		t.Error("Expected non-empty Go version")
	}

	if info.Platform == "" {
		t.Error("Expected non-empty platform")
	}

	// Version and BuildTime will be set by build flags in production
	// For now they're hardcoded
	if info.Version == "" {
		t.Error("Expected non-empty version")
	}
}
