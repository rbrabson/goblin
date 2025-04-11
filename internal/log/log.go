package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Initialize sets the logging level. If the LOG_LEVEL environment variable isn't set or the value
// isn't recognized, logging defaults to the `debug` level
func Initialize() {
	godotenv.Load(".env")
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var sslog *slog.Logger
	switch level {
	case "error":
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	case "warn":
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "info":
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "debug":
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	slog.SetDefault(sslog)
}
