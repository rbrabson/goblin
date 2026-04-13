package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Initialize sets the logging level. If the LOG_LEVEL environment variable isn't set or the value
// isn't recognized, logging defaults to the `info` level
func Initialize() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("Error loading .env file",
			slog.Any("error", err),
		)
	}
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	SetLevel(level)
}

// SetLevel sets the logging level. If the value isn't recognized, logging defaults to the `info` level
func SetLevel(level string) {
	level = strings.ToLower(level)

	var sslog *slog.Logger
	switch level {
	case "error":
		sslog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	case "warn":
		sslog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "info":
		sslog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "debug":
		sslog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		sslog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	slog.SetDefault(sslog)
}
