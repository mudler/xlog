package xlog

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const (
	EnvLogLevel  = "LOG_LEVEL"
	EnvLogFormat = "LOG_FORMAT"

	JSONFormat = "json"
)

var logger *slog.Logger

func init() {
	logger = NewLogger(LogLevel(os.Getenv(EnvLogLevel)), os.Getenv(EnvLogFormat))
}

func SetLogger(l *slog.Logger) {
	logger = l
}

func NewLogger(level LogLevel, format string) *slog.Logger {
	var handler slog.Handler

	showCode := level.ToSlogLevel() == slog.LevelDebug

	opts := &slog.HandlerOptions{
		AddSource: showCode,
		Level:     level.ToSlogLevel(),
	}

	switch strings.ToLower(format) {
	case JSONFormat:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func _log(level slog.Level, msg string, args ...any) {
	logger.Log(context.Background(), level, msg, args...)
}

func Info(msg string, args ...any) {
	_log(slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	_log(slog.LevelDebug, msg, args...)
}

func Error(msg string, args ...any) {
	_log(slog.LevelError, msg, args...)
}

func Warn(msg string, args ...any) {
	_log(slog.LevelWarn, msg, args...)
}

func Fatal(msg string, args ...any) {
	_log(slog.LevelError, msg, args...)
	os.Exit(1)
}
