package race

import (
	"time"

	"github.com/rbrabson/goblin/guild"
)

type Config struct {
	ID               string        `json:"_id" bson:"_id"`
	GuildID          string        `json:"guild_id" bson:"guild_id"`
	BetAmount        int64         `json:"bet_amount" bson:"bet_amount"`
	Currency         string        `json:"currency" bson:"currency"`
	MaxNumRacers     int           `json:"max_num_racers" bson:"max_num_racers"`
	MaxPrizeAmount   int           `json:"max_prize_amount" bson:"max_prize_amount"`
	MinNumRacers     int           `json:"min_num_racers" bson:"min_num_racers"`
	MinPriceAmount   int           `json:"min_price_amount" bson:"min_price_amount"`
	Theme            string        `json:"theme" bson:"theme"`
	WaitBetweenRaces time.Duration `json:"wait_beween_races" bson:"wait_between_races"`
	WaitForBets      time.Duration `json:"wait_for_bets" bson:"wait_for_bets"`
	WaitToStart      time.Duration `json:"wait_to_start" bson:"wait_to_start"`
	StartingLine     string        `json:"starting_line" bson:"starting_line"`
	EndingLine       string        `json:"ending_line" bson:"ending_line"`
}

func GetConfig(g *guild.Guild) *Config {
	config, err := getConfig(g)
	if err != nil {
		config = newConfig(g)
	}
	return config
}

func getConfig(guild *guild.Guild) (*Config, error) {
	// TODO: readConfig
	return nil, nil
}

func newConfig(guild *guild.Guild) *Config {
	config := &Config{
		GuildID:          guild.GuildID,
		Theme:            "clash",
		BetAmount:        100,
		Currency:         "credit",
		MaxNumRacers:     10,
		MaxPrizeAmount:   1250,
		MinNumRacers:     2,
		MinPriceAmount:   750,
		WaitBetweenRaces: time.Minute * 5,
		WaitForBets:      time.Minute * 1,
		WaitToStart:      time.Second * 30,
	}
	return config
}
