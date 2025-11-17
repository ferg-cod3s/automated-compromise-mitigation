// Package logging provides custom log formatters.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Color codes for pretty printing
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorBold   = "\033[1m"
)

// prettyHandler implements slog.Handler for human-readable output.
type prettyHandler struct {
	opts  slog.HandlerOptions
	attrs []slog.Attr
	group string
	mu    sync.Mutex
	w     io.Writer
}

// newPrettyHandler creates a new pretty handler.
func newPrettyHandler(w io.Writer, level slog.Level) slog.Handler {
	return &prettyHandler{
		opts: slog.HandlerOptions{
			Level: level,
		},
		w: w,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle handles the Record.
func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format: 2025-11-17 14:32:45.123 INFO  [component] message key=value key2=value2

	buf := make([]byte, 0, 1024)

	// Timestamp
	timestamp := r.Time.Format("2006-01-02 15:04:05.000")
	buf = append(buf, colorGray...)
	buf = append(buf, timestamp...)
	buf = append(buf, colorReset...)
	buf = append(buf, ' ')

	// Level with color
	levelStr, color := getLevelDisplay(r.Level)
	buf = append(buf, color...)
	buf = append(buf, colorBold...)
	buf = append(buf, fmt.Sprintf("%-5s", levelStr)...)
	buf = append(buf, colorReset...)
	buf = append(buf, ' ')

	// Component (if present)
	component := ""
	requestID := ""
	otherAttrs := make([]slog.Attr, 0)

	// Collect attributes
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "component":
			component = a.Value.String()
		case "request_id":
			requestID = a.Value.String()
		default:
			otherAttrs = append(otherAttrs, a)
		}
		return true
	})

	// Add component
	if component != "" {
		buf = append(buf, colorCyan...)
		buf = append(buf, '[')
		buf = append(buf, component...)
		buf = append(buf, ']')
		buf = append(buf, colorReset...)
		buf = append(buf, ' ')
	}

	// Message
	buf = append(buf, r.Message...)

	// Request ID (if present)
	if requestID != "" {
		buf = append(buf, ' ')
		buf = append(buf, colorPurple...)
		buf = append(buf, "request_id="...)
		buf = append(buf, requestID...)
		buf = append(buf, colorReset...)
	}

	// Other attributes
	if len(otherAttrs) > 0 {
		for _, attr := range otherAttrs {
			buf = append(buf, ' ')
			buf = append(buf, formatAttr(attr)...)
		}
	}

	buf = append(buf, '\n')

	_, err := h.w.Write(buf)
	return err
}

// WithAttrs returns a new Handler whose attributes consist of both the receiver's attributes and the arguments.
func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &prettyHandler{
		opts:  h.opts,
		attrs: append(h.attrs[:len(h.attrs):len(h.attrs)], attrs...),
		group: h.group,
		w:     h.w,
	}
	return newHandler
}

// WithGroup returns a new Handler with the given group appended to the receiver's existing groups.
func (h *prettyHandler) WithGroup(name string) slog.Handler {
	newHandler := &prettyHandler{
		opts:  h.opts,
		attrs: h.attrs,
		group: name,
		w:     h.w,
	}
	return newHandler
}

// getLevelDisplay returns the display string and color for a log level.
func getLevelDisplay(level slog.Level) (string, string) {
	switch level {
	case slog.LevelDebug:
		return "DEBUG", colorBlue
	case slog.LevelInfo:
		return "INFO", colorGreen
	case slog.LevelWarn:
		return "WARN", colorYellow
	case slog.LevelError:
		return "ERROR", colorRed
	default:
		return level.String(), colorReset
	}
}

// formatAttr formats a single attribute for display.
func formatAttr(attr slog.Attr) string {
	key := attr.Key
	value := attr.Value

	switch value.Kind() {
	case slog.KindString:
		str := value.String()
		if needsQuoting(str) {
			return fmt.Sprintf("%s=%q", key, str)
		}
		return fmt.Sprintf("%s=%s", key, str)
	case slog.KindInt64:
		return fmt.Sprintf("%s=%d", key, value.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%s=%d", key, value.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%s=%.2f", key, value.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%s=%t", key, value.Bool())
	case slog.KindDuration:
		return fmt.Sprintf("%s=%s", key, value.Duration())
	case slog.KindTime:
		return fmt.Sprintf("%s=%s", key, value.Time().Format(time.RFC3339))
	default:
		return fmt.Sprintf("%s=%v", key, value.Any())
	}
}

// needsQuoting returns true if a string needs to be quoted in output.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	return strings.ContainsAny(s, " \t\n\r\"'")
}
