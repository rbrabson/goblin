package race

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultBabyDragonBuffPercent = 85
)

var (
	track = strings.Repeat("•   ", 20)
)

// Config represents the configuration for the race game.
type Config struct {
	ID                    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID               string             `json:"guild_id" bson:"guild_id"`
	BetAmount             int                `json:"bet_amount" bson:"bet_amount"`
	Currency              string             `json:"currency" bson:"currency"`
	MaxPrizeAmount        int                `json:"max_prize_amount" bson:"max_prize_amount"`
	MaxNumRacers          int                `json:"max_num_racers" bson:"max_num_racers"`
	MinNumRacers          int                `json:"min_num_racers" bson:"min_num_racers"`
	MinPrizeAmount        int                `json:"min_price_amount" bson:"min_price_amount"`
	Theme                 string             `json:"theme" bson:"theme"`
	WaitBetweenRaces      time.Duration      `json:"wait_beween_races" bson:"wait_between_races"`
	WaitForBets           time.Duration      `json:"wait_for_bets" bson:"wait_for_bets"`
	WaitToStart           time.Duration      `json:"wait_to_start" bson:"wait_to_start"`
	StartingLine          string             `json:"starting_line" bson:"starting_line"`
	Track                 string             `json:"track" bson:"track"`
	EndingLine            string             `json:"ending_line" bson:"ending_line"`
	BabyDragonBuffPercent int                `json:"babydragon_buff_percent" bson:"babydragon_buff_percent"`
}

// GetConfig gets the race configuration for the guild. If the configuration does not
// exist, then a new one is created.
func GetConfig(guildID string) *Config {
	config := readConfig(guildID)
	if config == nil {
		config = readConfigFromFile(guildID)
	}
	if config.BabyDragonBuffPercent == 0 {
		config.BabyDragonBuffPercent = defaultBabyDragonBuffPercent
		writeConfig(config)
		slog.Debug("set baby dragon buff percent",
			slog.String("guildID", guildID),
			slog.Int("babydragon_buff_percent", config.BabyDragonBuffPercent),
		)
	}
	return config
}

// readConfigFromFile gets a new configuration for the guild. If the oconfiguration cannot be
// read from the configuration file or decdoded, then a default configuration is
// returned.
func readConfigFromFile(guildID string) *Config {
	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "race", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default race config",
			slog.String("guildID", guildID),
			slog.String("theme", configTheme),
			slog.Any("error", err),
		)
		// If the configuration file does not exist, then create a new configuration.
		return getDefauiltConfig(guildID)
	}

	config := &Config{}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		slog.Error("failed to unmarshal default race config",
			slog.String("guildID", guildID),
			slog.String("theme", configTheme),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefauiltConfig(guildID)
	}
	config.GuildID = guildID

	writeConfig(config)
	slog.Info("create new race config",
		slog.String("guildID", guildID),
		slog.String("theme", configTheme),
	)

	return config
}

// getDefauiltConfig creates a new race configuration for the guild. The configuration is saved to
// the database.
func getDefauiltConfig(guildID string) *Config {
	config := &Config{
		GuildID:               guildID,
		Theme:                 "clash",
		BetAmount:             100,
		Currency:              "credit",
		StartingLine:          ":checkered_flag:",
		EndingLine:            "<:gems:312346463453708289>",
		Track:                 track,
		MaxPrizeAmount:        1250,
		MaxNumRacers:          10,
		MinNumRacers:          2,
		MinPrizeAmount:        750,
		WaitBetweenRaces:      time.Second * 60,
		WaitForBets:           time.Second * 30,
		WaitToStart:           time.Second * 45,
		BabyDragonBuffPercent: defaultBabyDragonBuffPercent,
	}

	writeConfig(config)
	slog.Info("race configuration created",
		slog.String("guildID", guildID),
		slog.String("theme", config.Theme),
	)

	return config
}
