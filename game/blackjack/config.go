package blackjack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	CONFIG_TABLE_NAME = "config"
)

// Config holds the configuration settings for the blackjack game.
type Config struct {
	MaxPlayers        int           `json:"max_players"`
	Decks             int           `json:"decks"`
	BetAmount         int           `json:"bet_amount"`
	BlackjackPay      float64       `json:"blackjack_pay"`
	DelayBetweenGames time.Duration `json:"delay_between_games"`
	WaitForPlayers    time.Duration `json:"wait_for_players"`
	PlayerTimeout     time.Duration `json:"player_timeout"`
}

// String returns a string representation of the Config struct.
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("Config{")
	sb.WriteString(fmt.Sprintf("MaxPlayers: %d, ", c.MaxPlayers))
	sb.WriteString(fmt.Sprintf("Decks: %d, ", c.Decks))
	sb.WriteString(fmt.Sprintf("BetAmount: %d, ", c.BetAmount))
	sb.WriteString(fmt.Sprintf("BlackjackPay: %.2f%%", c.BlackjackPay))
	sb.WriteString(fmt.Sprintf("DelayBetweenGames: %v, ", c.DelayBetweenGames))
	sb.WriteString(fmt.Sprintf("WaitForPlayers: %v", c.WaitForPlayers))
	sb.WriteString("}")
	return sb.String()
}

// GetConfig retrieves the blackjack configuration, either from a file or defaults.
func GetConfig() *Config {
	config := readConfigFromFile()
	if config == nil {
		config = defaultConfig()
	}
	return config
}

// defaultConfig retrieves the configuration for a given guild ID.
func defaultConfig() *Config {
	return &Config{
		MaxPlayers:        5,
		Decks:             6,
		BetAmount:         50,
		BlackjackPay:      1.5,
		DelayBetweenGames: 10 * time.Second,
		WaitForPlayers:    30 * time.Second,
		PlayerTimeout:     15 * time.Second,
	}
}

// readConfigFromFile reads the configuration from a JSON file and returns a Config instance.
func readConfigFromFile() *Config {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "blackjack", "config", CONFIG_TABLE_NAME+".json")
	bytes, _ := os.ReadFile(configFileName)

	var config Config
	_ = json.Unmarshal(bytes, &config)
	config.DelayBetweenGames *= time.Second
	config.WaitForPlayers *= time.Second
	config.PlayerTimeout *= time.Second

	return &config
}
