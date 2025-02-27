package race

import "time"

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
