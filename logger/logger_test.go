package logger

import (
	"testing"
)

func TestNewSimpleLogger(t *testing.T) {
	logger := NewSimpleLogger()
	if logger == nil {
		t.Fatal("NewSimpleLogger() returned nil")
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}
}

func TestLoggerDebug(t *testing.T) {
	logger := NewSimpleLogger()

	// Log a debug message
	logger.Debug("Test debug message")

	// Debug message is only output to stdout, no file check needed
}

func TestLoggerInfo(t *testing.T) {
	logger := NewSimpleLogger()

	// Log an info message
	logger.Info("Test info message")

	// Info message is only output to stdout, no file check needed
}

func TestLoggerWarn(t *testing.T) {
	logger := NewSimpleLogger()

	// Log a warn message
	logger.Warn("Test warn message")

	// Warn message is only output to stdout, no file check needed
}

func TestLoggerError(t *testing.T) {
	logger := NewSimpleLogger()

	// Log an error message
	logger.Error("Test error message")

	// Error message is only output to stderr, no file check needed
}

func TestLoggerBatchStart(t *testing.T) {
	logger := NewSimpleLogger()

	// Log batch start
	logger.BatchStart(5, 2)

	// Batch start message is only output to stdout, no file check needed
}

func TestLoggerBatchComplete(t *testing.T) {
	logger := NewSimpleLogger()

	// Log batch complete
	logger.BatchComplete(3, 1, 1, 5)

	// Batch complete message is only output to stdout, no file check needed
}

func TestLoggerDownloadSuccess(t *testing.T) {
	logger := NewSimpleLogger()

	// Log download success
	logger.DownloadSuccess("video123", "Test Video", 0, 1024*1024)

	// Download success message is only output to stdout, no file check needed
}

func TestLoggerDownloadFail(t *testing.T) {
	logger := NewSimpleLogger()

	// Log download fail
	logger.DownloadFail("video123", "Test Video", nil, 0)

	// Download fail message is only output to stderr, no file check needed
}

func TestLoggerDownloadSkip(t *testing.T) {
	logger := NewSimpleLogger()

	// Log download skip
	logger.DownloadSkip("video123", "Test Video")

	// Download skip message is only output to stdout, no file check needed
}

func TestLoggerClose(t *testing.T) {
	logger := NewSimpleLogger()

	// Close logger
	logger.Close()

	// Close is a no-op for SimpleLogger, just ensure it doesn't panic
}

func TestFormatMessage(t *testing.T) {
	// Test formatMessage function indirectly through logger methods
	logger := NewSimpleLogger()

	// This should not panic
	logger.Info("Test formatted message: %s", "value")
	logger.Debug("Test debug formatted message: %d", 123)
}
