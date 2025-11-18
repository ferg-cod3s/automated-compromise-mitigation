package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRotatingFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := RotationConfig{
		MaxSizeMB:  10,
		MaxAgeDays: 7,
		MaxBackups: 3,
		Compress:   true,
	}

	writer, err := NewRotatingFileWriter(config, logFile)
	if err != nil {
		t.Fatalf("Failed to create rotating file writer: %v", err)
	}

	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}

	// Write some data
	data := []byte("test log entry\n")
	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to log file: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Verify file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestNewRotatingFileWriter_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "nested", "dir", "test.log")

	config := RotationConfig{
		MaxSizeMB:  10,
		MaxAgeDays: 7,
		MaxBackups: 3,
		Compress:   false,
	}

	writer, err := NewRotatingFileWriter(config, logFile)
	if err != nil {
		t.Fatalf("Failed to create rotating file writer: %v", err)
	}

	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}

	// Verify directory was created
	dir := filepath.Dir(logFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
}

func TestNewMultiWriter(t *testing.T) {
	var buf1, buf2 strings.Builder

	writer := NewMultiWriter(&buf1, &buf2)
	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}

	data := []byte("test data")
	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Verify both writers received the data
	if buf1.String() != string(data) {
		t.Errorf("Expected buf1 to contain %q, got %q", data, buf1.String())
	}

	if buf2.String() != string(data) {
		t.Errorf("Expected buf2 to contain %q, got %q", data, buf2.String())
	}
}

func TestSetupOutput_Stdout(t *testing.T) {
	config := &Config{
		OutputMode: OutputStdout,
	}

	writer, cleanup, err := SetupOutput(config)
	if err != nil {
		t.Fatalf("Failed to setup output: %v", err)
	}
	defer cleanup()

	if writer != os.Stdout {
		t.Error("Expected stdout writer")
	}
}

func TestSetupOutput_File(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &Config{
		OutputMode: OutputFile,
		FilePath:   logFile,
		Rotation: RotationConfig{
			MaxSizeMB:  10,
			MaxAgeDays: 7,
			MaxBackups: 3,
			Compress:   false,
		},
	}

	writer, cleanup, err := SetupOutput(config)
	if err != nil {
		t.Fatalf("Failed to setup output: %v", err)
	}
	defer cleanup()

	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}

	// Write some data
	data := []byte("test log entry\n")
	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Cleanup should close the file
	if err := cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
}

func TestSetupOutput_Both(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := &Config{
		OutputMode: OutputBoth,
		FilePath:   logFile,
		Rotation: RotationConfig{
			MaxSizeMB:  10,
			MaxAgeDays: 7,
			MaxBackups: 3,
			Compress:   false,
		},
	}

	writer, cleanup, err := SetupOutput(config)
	if err != nil {
		t.Fatalf("Failed to setup output: %v", err)
	}
	defer cleanup()

	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}

	// Write some data
	data := []byte("test log entry\n")
	n, err := writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Cleanup should close the file
	if err := cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Verify file exists and contains data
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("Expected file to contain %q, got %q", data, content)
	}
}

func TestGetLogFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Non-existent file should return 0
	size, err := GetLogFileSize(logFile)
	if err != nil {
		t.Fatalf("Failed to get file size: %v", err)
	}
	if size != 0 {
		t.Errorf("Expected size 0 for non-existent file, got %d", size)
	}

	// Create file with known size
	data := []byte("test data")
	if err := os.WriteFile(logFile, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	size, err = GetLogFileSize(logFile)
	if err != nil {
		t.Fatalf("Failed to get file size: %v", err)
	}

	if size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), size)
	}
}

func TestCleanupRotatedLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create some rotated log files
	files := []string{
		logFile,
		logFile + ".1",
		logFile + ".2.gz",
		logFile + ".old",
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// CleanupRotatedLogs currently relies on lumberjack's built-in cleanup
	// This test just verifies it doesn't error
	err := CleanupRotatedLogs(logFile, 30)
	if err != nil {
		t.Errorf("CleanupRotatedLogs failed: %v", err)
	}
}

func TestRotateLogNow(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := RotationConfig{
		MaxSizeMB:  10,
		MaxAgeDays: 7,
		MaxBackups: 3,
		Compress:   false,
	}

	writer, err := NewRotatingFileWriter(config, logFile)
	if err != nil {
		t.Fatalf("Failed to create rotating file writer: %v", err)
	}

	// Write some data
	data := []byte("test log entry\n")
	_, err = writer.Write(data)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Test rotation
	err = RotateLogNow(writer)
	if err != nil {
		t.Errorf("RotateLogNow failed: %v", err)
	}

	// Test with non-rotatable writer (should not error)
	var buf strings.Builder
	err = RotateLogNow(&buf)
	if err != nil {
		t.Errorf("RotateLogNow with non-rotatable writer failed: %v", err)
	}
}
