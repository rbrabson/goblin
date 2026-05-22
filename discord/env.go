package discord

import (
	"log/slog"
	"os"
	"strconv"
)

// GetenvBool returns a boolean to indicate whether the environment variable
// is set to "true"
func GetenvBool(key string) bool {
	s := os.Getenv(key)
	if s == "" {
		return false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		slog.Warn("invalid boolean environment variable",
			slog.String("key", key),
			slog.String("value", s),
			slog.Any("error", err),
		)
		return false
	}
	return v
}
