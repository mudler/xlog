package xlog

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

type Logger struct {
	logger               *slog.Logger
	level                LogLevel
	debuggingInformation bool
}

func NewLogger(level LogLevel, format string) *Logger {
	var handler slog.Handler

	debuggingInformation := level.ToSlogLevel() == slog.LevelDebug

	opts := &slog.HandlerOptions{
		AddSource: debuggingInformation,
		Level:     level.ToSlogLevel(),
	}

	switch strings.ToLower(format) {
	case JSONFormat:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case TextFormat:
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = newColorTextHandler(os.Stdout, opts)
	}

	return &Logger{
		logger:               slog.New(handler),
		level:                level,
		debuggingInformation: debuggingInformation,
	}
}

func (l *Logger) _log(level slog.Level, msg string, args ...any) {
	if l.debuggingInformation {
		_, f, line, _ := runtime.Caller(2)
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
