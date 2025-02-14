package discord

import (
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
		return false
	}
	return v
}
