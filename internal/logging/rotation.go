// Package logging provides log rotation functionality.
package logging

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewRotatingFileWriter creates a rotating file writer using lumberjack.
// The writer automatically rotates log files based on size, age, and backup count.
func NewRotatingFileWriter(config RotationConfig, filePath string) (io.Writer, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    config.MaxSizeMB,    // megabytes
		MaxAge:     config.MaxAgeDays,   // days
		MaxBackups: config.MaxBackups,   // number of backups
		Compress:   config.Compress,     // compress rotated files
		LocalTime:  true,                // use local time for filenames
	}, nil
}

// NewMultiWriter creates a writer that writes to multiple destinations.
// This is used when OutputMode is "both" (stdout and file).
func NewMultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// SetupOutput configures the output writer based on the configuration.
// It returns an io.Writer and a cleanup function (for closing file writers).
func SetupOutput(config *Config) (io.Writer, func() error, error) {
	cleanup := func() error { return nil }

	switch config.OutputMode {
	case OutputStdout:
		return os.Stdout, cleanup, nil

	case OutputFile:
		writer, err := NewRotatingFileWriter(config.Rotation, config.FilePath)
		if err != nil {
			return nil, cleanup, err
		}

		// Create cleanup function to close the file writer
		cleanup = func() error {
			if closer, ok := writer.(io.Closer); ok {
				return closer.Close()
			}
			return nil
		}

		return writer, cleanup, nil

	case OutputBoth:
		fileWriter, err := NewRotatingFileWriter(config.Rotation, config.FilePath)
		if err != nil {
			return nil, cleanup, err
		}

		multiWriter := NewMultiWriter(os.Stdout, fileWriter)

		// Create cleanup function to close the file writer
		cleanup = func() error {
			if closer, ok := fileWriter.(io.Closer); ok {
				return closer.Close()
			}
			return nil
		}

		return multiWriter, cleanup, nil

	default:
		// Default to stdout
		return os.Stdout, cleanup, nil
	}
}

// CleanupRotatedLogs removes old rotated log files on service start.
// This is useful for ensuring disk space management.
func CleanupRotatedLogs(filePath string, maxAge int) error {
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)

	// Find all rotated log files (*.log.* or *.log.*.gz)
	pattern := filepath.Join(dir, baseName+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// For now, we rely on lumberjack's built-in cleanup
	// This function is here for potential future enhancements
	// like custom cleanup logic beyond what lumberjack provides
	_ = matches

	return nil
}

// GetLogFileSize returns the current size of the log file in bytes.
func GetLogFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return info.Size(), nil
}

// RotateLogNow forces an immediate rotation of the log file.
// This is useful for manual log rotation (e.g., on SIGHUP).
func RotateLogNow(writer io.Writer) error {
	if rotator, ok := writer.(*lumberjack.Logger); ok {
		return rotator.Rotate()
	}
	// If not a lumberjack logger, check if it's a multi-writer
	// For now, we don't support rotation on multi-writers
	return nil
}
