package race

import (
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Config represents the configuration for the race game.
type Config struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID          string             `json:"guild_id" bson:"guild_id"`
	BetAmount        int                `json:"bet_amount" bson:"bet_amount"`
	Currency         string             `json:"currency" bson:"currency"`
	MaxNumRacers     int                `json:"max_num_racers" bson:"max_num_racers"`
	MaxPrizeAmount   int                `json:"max_prize_amount" bson:"max_prize_amount"`
	MinNumRacers     int                `json:"min_num_racers" bson:"min_num_racers"`
	MinPrizeAmount   int                `json:"min_price_amount" bson:"min_price_amount"`
	Theme            string             `json:"theme" bson:"theme"`
	WaitBetweenRaces time.Duration      `json:"wait_beween_races" bson:"wait_between_races"`
	WaitForBets      time.Duration      `json:"wait_for_bets" bson:"wait_for_bets"`
	WaitToStart      time.Duration      `json:"wait_to_start" bson:"wait_to_start"`
	StartingLine     string             `json:"starting_line" bson:"starting_line"`
	EndingLine       string             `json:"ending_line" bson:"ending_line"`
}

// GetConfig gets the race configuration for the guild. If the configuration does not
// exist, then a new one is created.
func GetConfig(guildID string) *Config {
	config, err := getConfig(guildID)
	if err != nil {
		config = newConfig(guildID)
	}
	return config
}

// getConfig reads the race configuration from the database. If the configuration
// does not exist, then an error is returned.
func getConfig(guildID string) (*Config, error) {
	log.Trace("--> race.getConfig")
	defer log.Trace("<-- race.getConfig")

	config := readConfig(guildID)
	if config == nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}

// newConfig creates a new race configuration for the guild. The configuration is saved to
// the database.
func newConfig(guildID string) *Config {
	log.Trace("--> race.newConfig")
	defer log.Trace("<-- race.newConfig")

	config := &Config{
		GuildID:          guildID,
		Theme:            "clash",
		BetAmount:        100,
		Currency:         "credit",
		StartingLine:     "🏁",
		MaxNumRacers:     10,
		MaxPrizeAmount:   1250,
		MinNumRacers:     2,
		MinPrizeAmount:   750,
		WaitBetweenRaces: time.Minute * 5,
		WaitForBets:      time.Minute * 1,
		WaitToStart:      time.Second * 30,
	}

	writeConfig(config)
	log.WithFields(log.Fields{"guild": guildID}).Info("race configuration created")

	return config
}
