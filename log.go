package xlog

import (
	"os"
)

const (
	EnvLogLevel  = "LOG_LEVEL"
	EnvLogFormat = "LOG_FORMAT"

	JSONFormat    = "json"
	TextFormat    = "text"
	DefaultFormat = "default"
)

var logger *Logger

func init() {
	logger = NewLogger(LogLevel(os.Getenv(EnvLogLevel)), os.Getenv(EnvLogFormat))
}

func SetLogger(l *Logger) {
	logger = l
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Fatal(msg string, args ...any) {
	logger.Fatal(msg, args...)
	os.Exit(1)
}
