package slots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/rbrabson/goblin/discord"
)

const (
	CONFIG_TABLE_NAME = "config"
)

// Config represents the configuration for the slots game.
type Config struct {
	Cooldown time.Duration `json:"cooldown"`
}

// GetConfig retrieves the configuration for the slots game.
func GetConfig() *Config {
	return newConfig()
}

// newConfig creates a new Config instance by reading from the configuration file.
func newConfig() *Config {
	return readConfigFromFile()
}

// readConfigFromFile reads the configuration from a JSON file and returns a Config instance.
func readConfigFromFile() *Config {
	configFileName := filepath.Join(discord.DISCORD_CONFIG_DIR, "slots", "config", CONFIG_TABLE_NAME+".json")
	bytes, _ := os.ReadFile(configFileName)

	var config Config
	_ = json.Unmarshal(bytes, &config)
	config.Cooldown *= time.Second

	return &config
}
