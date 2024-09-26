package discord

import (
	"os"
	"strconv"
)

// getenvBool returns a boolean to indicate whether the environment variable
// is set to "true"
func getenvBool(key string) bool {
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
