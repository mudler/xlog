package xlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// Logger wraps slog.Logger with level-aware logging and optional log deduplication.
type Logger struct {
	logger               *slog.Logger
	level                LogLevel
	debuggingInformation bool
}

// LoggerOption configures NewLogger behavior.
type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	dedup *bool // nil = auto-detect terminal, true = force on, false = force off
}

// WithDedup forces log deduplication on, regardless of whether output is a terminal.
func WithDedup() LoggerOption {
	return func(c *loggerConfig) {
		v := true
		c.dedup = &v
	}
}

// WithoutDedup forces log deduplication off, even when output is a terminal.
func WithoutDedup() LoggerOption {
	return func(c *loggerConfig) {
		v := false
		c.dedup = &v
	}
}

// NewLogger creates a new Logger with the given level and format.
// By default, consecutive identical log lines are automatically deduplicated
// when output is a terminal. Use WithDedup() or WithoutDedup() to override.
func NewLogger(level LogLevel, format string, opts ...LoggerOption) *Logger {
	var cfg loggerConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level.ToSlogLevel(),
	}

	handler := NewHandler(format, os.Stdout, handlerOpts)

	enableDedup := false
	if cfg.dedup != nil {
		enableDedup = *cfg.dedup
	} else {
		enableDedup = isTerminalWriter(os.Stdout)
	}

	if enableDedup {
		handler = NewDeduplicatingHandler(handler, os.Stdout)
	}

	return &Logger{
		logger:               slog.New(handler),
		level:                level,
		debuggingInformation: level.ToSlogLevel() == slog.LevelDebug,
	}
}

// NewHandler creates an slog.Handler for the given format and options.
// This allows callers to build custom handler chains (e.g., wrapping with middleware).
func NewHandler(format string, w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	switch strings.ToLower(format) {
	case JSONFormat:
		return slog.NewJSONHandler(w, opts)
	case TextFormat:
		return slog.NewTextHandler(w, opts)
	default:
		return newColorTextHandler(w, opts)
	}
}

// NewLoggerWithHandler creates a Logger using a pre-built slog.Handler.
// No automatic deduplication is applied — the caller controls the handler chain.
func NewLoggerWithHandler(handler slog.Handler, level LogLevel) *Logger {
	return &Logger{
		logger:               slog.New(handler),
		level:                level,
		debuggingInformation: level.ToSlogLevel() == slog.LevelDebug,
	}
}

func (l *Logger) _log(level slog.Level, msg string, args ...any) {
	if l.debuggingInformation {
		_, f, line, _ := runtime.Caller(3)
		group := slog.Group(
			"caller",
			slog.Attr{
				Key:   "file",
				Value: slog.AnyValue(f),
			},
			slog.Attr{
				Key:   "L",
				Value: slog.AnyValue(line),
			},
		)
		args = append(args, group)
	}
	l.logger.Log(context.Background(), level, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l._log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l._log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l._log(slog.LevelError, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l._log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l._log(slog.LevelError, msg, args...)
	os.Exit(1)
}

// isTerminalWriter checks if the given writer is connected to a terminal.
func isTerminalWriter(w io.Writer) bool {
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
