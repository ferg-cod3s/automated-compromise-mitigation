// Package logging provides structured logging for ACM using Go's log/slog.
package logging

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// contextKey is the type for context keys in this package.
type contextKey string

const (
	// requestIDKey is the context key for request IDs.
	requestIDKey contextKey = "request_id"

	// componentKey is the context key for component names.
	componentKey contextKey = "component"
)

// Logger wraps slog.Logger with ACM-specific functionality.
type Logger struct {
	*slog.Logger
	component string
	attrs     []slog.Attr
	mu        sync.RWMutex
}

// Global logger instance (default)
var (
	globalLogger     *Logger
	globalLogCleanup func() error
	once             sync.Once
)

// Initialize initializes the global logging system.
// This should be called once at application startup.
func Initialize(config Config) error {
	var initErr error
	once.Do(func() {
		// Validate configuration
		if err := config.Validate(); err != nil {
			initErr = err
			return
		}

		// Setup output with rotation if configured
		output, cleanup, err := SetupOutput(&config)
		if err != nil {
			initErr = err
			return
		}
		globalLogCleanup = cleanup
		config.Output = output

		handler, err := createHandler(config)
		if err != nil {
			initErr = err
			return
		}

		globalLogger = &Logger{
			Logger:    slog.New(handler),
			component: "acm",
		}

		// Set as default slog logger
		slog.SetDefault(globalLogger.Logger)
	})

	return initErr
}

// Shutdown gracefully shuts down the logging system.
// This should be called during application shutdown to flush and close log files.
func Shutdown() error {
	if globalLogCleanup != nil {
		return globalLogCleanup()
	}
	return nil
}

// Default returns the global logger instance.
func Default() *Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		_ = Initialize(DefaultConfig())
	}
	return globalLogger
}

// NewLogger creates a component-specific logger.
func NewLogger(component string) *Logger {
	if globalLogger == nil {
		_ = Initialize(DefaultConfig())
	}

	baseLogger := globalLogger.Logger.With("component", component)
	return &Logger{
		Logger:    baseLogger,
		component: component,
	}
}

// WithContext extracts request ID from context and adds it to the logger.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	attrs := make([]slog.Attr, 0, 2)

	// Extract request ID from context
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Extract component from context (if overridden)
	if component := getComponentFromContext(ctx); component != "" && component != l.component {
		attrs = append(attrs, slog.String("component", component))
	}

	if len(attrs) == 0 {
		return l
	}

	// Create new logger with context attributes
	handler := l.Logger.Handler()
	for _, attr := range attrs {
		handler = handler.WithAttrs([]slog.Attr{attr})
	}

	return &Logger{
		Logger:    slog.New(handler),
		component: l.component,
		attrs:     attrs,
	}
}

// With returns a new logger with additional attributes.
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		Logger:    l.Logger.With(args...),
		component: l.component,
	}
}

// WithAttrs returns a new logger with additional slog.Attr attributes.
func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	if len(attrs) == 0 {
		return l
	}

	handler := l.Logger.Handler().WithAttrs(attrs)
	return &Logger{
		Logger:    slog.New(handler),
		component: l.component,
		attrs:     append(l.attrs, attrs...),
	}
}

// WithError returns a new logger with an error attribute.
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithAttrs(slog.String("error", err.Error()))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
}

// TimedOperation executes a function and logs its duration.
// This is useful for performance monitoring.
func (l *Logger) TimedOperation(ctx context.Context, operation string, fn func() error) error {
	logger := l.WithContext(ctx)
	start := time.Now()

	logger.Debug("operation started", "operation", operation)

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error("operation failed",
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
		return err
	}

	// Warn if operation is slow (>100ms)
	level := slog.LevelInfo
	if duration > 100*time.Millisecond {
		level = slog.LevelWarn
	}

	logger.Log(ctx, level, "operation completed",
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}

// TimedOperationWithResult executes a function, logs its duration, and returns a result.
func (l *Logger) TimedOperationWithResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	logger := l.WithContext(ctx)
	start := time.Now()

	logger.Debug("operation started", "operation", operation)

	result, err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error("operation failed",
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
		return result, err
	}

	level := slog.LevelInfo
	if duration > 100*time.Millisecond {
		level = slog.LevelWarn
	}

	logger.Log(ctx, level, "operation completed",
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	)

	return result, nil
}

// Fatal logs an error message and exits the application.
// This should only be used for unrecoverable errors during startup.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// Component returns the component name for this logger.
func (l *Logger) Component() string {
	return l.component
}

// getRequestIDFromContext extracts the request ID from context.
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// getComponentFromContext extracts the component name from context.
func getComponentFromContext(ctx context.Context) string {
	if component, ok := ctx.Value(componentKey).(string); ok {
		return component
	}
	return ""
}

// SetRequestIDInContext adds a request ID to the context.
func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// SetComponentInContext adds a component name to the context.
func SetComponentInContext(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, componentKey, component)
}

// createHandler creates the appropriate slog handler based on configuration.
func createHandler(config Config) (slog.Handler, error) {
	level := parseLevel(config.Level)

	var handler slog.Handler

	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Customize timestamp format
				if a.Key == slog.TimeKey {
					return slog.String(slog.TimeKey, a.Value.Time().UTC().Format(time.RFC3339Nano))
				}
				return a
			},
		})
	case FormatPretty:
		handler = newPrettyHandler(config.Output, level)
	default:
		handler = slog.NewTextHandler(config.Output, &slog.HandlerOptions{
			Level: level,
		})
	}

	// Add global attributes (service, version, hostname)
	attrs := []slog.Attr{
		slog.String("service", config.ServiceName),
		slog.String("version", config.Version),
	}

	if config.Hostname != "" {
		attrs = append(attrs, slog.String("hostname", config.Hostname))
	}

	if config.PID > 0 {
		attrs = append(attrs, slog.Int("pid", config.PID))
	}

	handler = handler.WithAttrs(attrs)

	return handler, nil
}

// parseLevel converts string level to slog.Level.
func parseLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "warning", "WARN", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
