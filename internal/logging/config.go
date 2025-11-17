// Package logging provides structured logging configuration.
package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// Format represents the log output format.
type Format string

const (
	// FormatJSON outputs logs in JSON format (production).
	FormatJSON Format = "json"

	// FormatPretty outputs logs in human-readable format (development).
	FormatPretty Format = "pretty"

	// FormatText outputs logs in standard text format.
	FormatText Format = "text"
)

// OutputMode represents where logs are written.
type OutputMode string

const (
	// OutputStdout writes logs to stdout.
	OutputStdout OutputMode = "stdout"

	// OutputFile writes logs to a file.
	OutputFile OutputMode = "file"

	// OutputBoth writes logs to both stdout and file.
	OutputBoth OutputMode = "both"
)

// Config holds logging configuration.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level string

	// Format is the output format (json, pretty, text)
	Format Format

	// OutputMode determines where logs are written
	OutputMode OutputMode

	// Output is the writer for logs (derived from OutputMode)
	Output io.Writer

	// FilePath is the path to the log file (if OutputMode includes file)
	FilePath string

	// ServiceName is the name of the service (added to all logs)
	ServiceName string

	// Version is the service version (added to all logs)
	Version string

	// Hostname is the server hostname (added to all logs)
	Hostname string

	// PID is the process ID (added to all logs)
	PID int

	// ComponentLevels maps component names to their log levels
	ComponentLevels map[string]string

	// Rotation settings (for file output)
	Rotation RotationConfig
}

// RotationConfig holds log rotation configuration.
type RotationConfig struct {
	// MaxSizeMB is the maximum size in megabytes before rotation
	MaxSizeMB int

	// MaxAgeDays is the maximum age in days before deletion
	MaxAgeDays int

	// MaxBackups is the maximum number of old log files to keep
	MaxBackups int

	// Compress determines whether rotated logs are compressed
	Compress bool
}

// DefaultConfig returns a default logging configuration.
func DefaultConfig() Config {
	hostname, _ := os.Hostname()

	return Config{
		Level:       getEnvOrDefault("ACM_LOG_LEVEL", "info"),
		Format:      Format(getEnvOrDefault("ACM_LOG_FORMAT", "json")),
		OutputMode:  OutputMode(getEnvOrDefault("ACM_LOG_OUTPUT", "stdout")),
		Output:      os.Stdout,
		FilePath:    getEnvOrDefault("ACM_LOG_FILE", "/var/log/acm/acm.log"),
		ServiceName: "acm",
		Version:     "0.3.0", // TODO: Get from build flags
		Hostname:    hostname,
		PID:         os.Getpid(),
		ComponentLevels: make(map[string]string),
		Rotation: RotationConfig{
			MaxSizeMB:  getEnvIntOrDefault("ACM_LOG_MAX_SIZE", 100),
			MaxAgeDays: getEnvIntOrDefault("ACM_LOG_MAX_AGE", 30),
			MaxBackups: getEnvIntOrDefault("ACM_LOG_MAX_BACKUPS", 10),
			Compress:   getEnvBoolOrDefault("ACM_LOG_COMPRESS", true),
		},
	}
}

// DevelopmentConfig returns a configuration optimized for development.
func DevelopmentConfig() Config {
	config := DefaultConfig()
	config.Level = "debug"
	config.Format = FormatPretty
	config.OutputMode = OutputStdout
	config.Output = os.Stdout
	return config
}

// ProductionConfig returns a configuration optimized for production.
func ProductionConfig() Config {
	config := DefaultConfig()
	config.Level = "info"
	config.Format = FormatJSON
	config.OutputMode = OutputStdout
	config.Output = os.Stdout
	return config
}

// GetComponentLevel returns the log level for a specific component.
// Falls back to the global level if no component-specific level is set.
func (c *Config) GetComponentLevel(component string) string {
	if level, ok := c.ComponentLevels[component]; ok {
		return level
	}
	return c.Level
}

// SetComponentLevel sets the log level for a specific component.
func (c *Config) SetComponentLevel(component, level string) {
	if c.ComponentLevels == nil {
		c.ComponentLevels = make(map[string]string)
	}
	c.ComponentLevels[component] = level
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Format == FormatPretty || c.Level == "debug"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Format == FormatJSON && c.Level == "info"
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Level] {
		c.Level = "info" // Default to info if invalid
	}

	// Validate format
	validFormats := map[Format]bool{
		FormatJSON:   true,
		FormatPretty: true,
		FormatText:   true,
	}
	if !validFormats[c.Format] {
		c.Format = FormatJSON // Default to JSON if invalid
	}

	// Set output based on mode
	switch c.OutputMode {
	case OutputStdout:
		c.Output = os.Stdout
	case OutputFile:
		// File output will be created by rotation setup
		if c.FilePath == "" {
			c.FilePath = "/var/log/acm/acm.log"
		}
	case OutputBoth:
		// Will be handled by multi-writer in rotation setup
	default:
		c.OutputMode = OutputStdout
		c.Output = os.Stdout
	}

	return nil
}

// getEnvOrDefault gets an environment variable or returns a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault gets an integer environment variable or returns a default value.
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault gets a boolean environment variable or returns a default value.
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// buildInfo returns build information for logging.
type buildInfo struct {
	Version   string
	BuildTime string
	GoVersion string
	Platform  string
}

// GetBuildInfo returns current build information.
func GetBuildInfo() buildInfo {
	return buildInfo{
		Version:   "0.3.0", // TODO: Get from build flags
		BuildTime: "unknown", // TODO: Get from build flags
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
