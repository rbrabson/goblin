package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	sslog *slog.Logger
)

// init initializes the logger by loading environment variables from a .env file and setting the logging level
func init() {
	godotenv.Load(".env")
	initializeLogger()
}

// initializeLogger sets the logging level. If the LOG_LEVEL environment variable isn't set or the value
// isn't recognized, logging defaults to the `debug` level
func initializeLogger() {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
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
}

// GetLogger returns the global slog.Logger instance
func GetLogger() *slog.Logger {
	if sslog == nil {
		// Default to debug level if logger is not initialized
		sslog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return sslog
}
