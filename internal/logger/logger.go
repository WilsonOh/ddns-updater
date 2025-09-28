package logger

import (
	"log/slog"
	"os"
	"strings"
)

func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Init() {
	opts := &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
