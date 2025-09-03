package heist

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BailBase          = 250
	CrewOutput        = "None"
	DeathTimer        = 45 * time.Second
	HeistCost         = 1000
	PoliceAlert       = 60 * time.Second
	SentenceBase      = 45 * time.Second
	WaitTime          = 60 * time.Second
	HeistDefaultTheme = "clash"
)

// Config is the configuration data for new heists
type Config struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID         string             `json:"guild_id" bson:"guild_id"`
	Theme           string             `json:"theme" bson:"theme"`
	BailBase        int                `json:"bail_base" bson:"bail_base"`
	BoostPercentage float64            `json:"boost_percentage" bson:"boost_percentage"`
	BoostEnabled    bool               `json:"boost_enabled" bson:"boost_enabled"`
	CrewOutput      string             `json:"crew_output" bson:"crew_output"`
	DeathTimer      time.Duration      `json:"death_timer" bson:"death_timer"`
	HeistCost       int                `json:"heist_cost" bson:"heist_cost"`
	PoliceAlert     time.Duration      `json:"police_alert" bson:"police_alert"`
	SentenceBase    time.Duration      `json:"sentence_base" bson:"sentence_base"`
	Targets         string             `json:"targets" bson:"targets"`
	WaitTime        time.Duration      `json:"wait_time" bson:"wait_time"`
}

// GetConfig retrieves the heist configuration for the specified guild. If
// the configuration does not exist, nil is returned.
func GetConfig(guildID string) *Config {
	config := readConfig(guildID)
	if config == nil {
		config = readConfigFromFile(guildID)
	}
	return config
}

// readConfigFromFile creates a new default configuration for the specified guild.
// If the default configuration file cannot be read or decoded, then a default
// configuration is created.
func readConfigFromFile(guildID string) *Config {
	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "heist", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default heist config",
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultConfig(guildID)
	}

	config := &Config{}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		slog.Error("failed to unmarshal default heist config",
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultConfig(guildID)
	}
	config.GuildID = guildID

	writeConfig(config)
	slog.Info("create new heist config",
		slog.String("guildID", config.GuildID),
	)

	return config
}

// NewConfig creates a new default configuration for the specified guild.
func getDefaultConfig(guildID string) *Config {
	config := &Config{
		GuildID:         guildID,
		BailBase:        BailBase,
		BoostPercentage: 0,
		BoostEnabled:    false,
		CrewOutput:      CrewOutput,
		DeathTimer:      DeathTimer,
		HeistCost:       HeistCost,
		PoliceAlert:     PoliceAlert,
		SentenceBase:    SentenceBase,
		Targets:         HeistDefaultTheme,
		Theme:           HeistDefaultTheme,
		WaitTime:        WaitTime,
	}
	writeConfig(config)

	return config
}

// String returns a string representation of the heist configuration
func (config *Config) String() string {
	out, _ := json.Marshal(config)
	return string(out)
}
