package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameOutcome represents the outcome of a game.
type GameOutcome string

// Race game outcome.
const (
	Win   GameOutcome = "win"
	Place GameOutcome = "place"
	Show  GameOutcome = "show"
	Lose  GameOutcome = "lose"
)

// Heist game outcome.
const (
	Escaped  GameOutcome = "escaped"
	Captured GameOutcome = "captured"
	Died     GameOutcome = "died"
)

// MemberStats represents statistical data for a member including games played and earnings over various time periods.
type MemberStats struct {
	ID          primitive.ObjectID           `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string                       `json:"guild_id" bson:"guild_id"`
	MemberID    string                       `json:"member_id" bson:"member_id"`
	Game        string                       `json:"game" bson:"game"`
	GamesPlayed GamesPlayed                  `json:"games_played" bson:"games_played"`
	Earnings    GameEarnings                 `json:"earnings" bson:"earnings"`
	Results     map[GameOutcome]GameOutcomes `json:"results" bson:"results"`
	Streaks     map[GameOutcome]Streak       `json:"streaks" bson:"streaks"`
	Created     time.Time                    `json:"created" bson:"created"`
	Updated     time.Time                    `json:"updated" bson:"updated"`
}

// GamesPlayed represents the count and average of games played by a user over different time periods, such as daily or all-time.
type GamesPlayed struct {
	Daily   GamesPlayedStats `json:"daily" bson:"daily"`
	Weekly  GamesPlayedStats `json:"weekly" bson:"weekly"`
	Monthly GamesPlayedStats `json:"monthly" bson:"monthly"`
	Total   int              `json:"total" bson:"total"`
}

type GamesPlayedStats struct {
	Minimum int `json:"minimum" bson:"minimum"`
	Maximum int `json:"maximum" bson:"maximum"`
	Current int `json:"current" bson:"current"`
	Average int `json:"average" bson:"average"`
}

// GameEarnings represents the earnings statistics of a game over various time periods such as daily, weekly, and monthly.
type GameEarnings struct {
	Daily   EarningsStats `json:"daily" bson:"daily"`
	Weekly  EarningsStats `json:"weekly" bson:"weekly"`
	Monthly EarningsStats `json:"monthly" bson:"monthly"`
	Total   int           `json:"total" bson:"total"`
}

// EarningsStats represents statistical data for earnings, including the average and total earnings over a specific period.
type EarningsStats struct {
	Minimum int `json:"minimum" bson:"minimum"`
	Maximum int `json:"maximum" bson:"maximum"`
	Current int `json:"current" bson:"current"`
	Average int `json:"average" bson:"average"`
}

// Streak represents a user's streak data, including the current streak and the longest streak achieved.
type Streak struct {
	Current int `json:"current" bson:"current"`
	Longest int `json:"longest" bson:"longest"`
}

type GameOutcomes struct {
	Daily   GameOutcomesStats `json:"daily" bson:"daily"`
	Weekly  GameOutcomesStats `json:"weekly" bson:"weekly"`
	Monthly GameOutcomesStats `json:"monthly" bson:"monthly"`
	Total   int               `json:"total" bson:"total"`
}

// GameOutcomesStats represents the result of a game along with statistical data such as average and total scores.
type GameOutcomesStats struct {
	Minimum int `json:"minimum" bson:"minimum"`
	Maximum int `json:"maximum" bson:"maximum"`
	Current int `json:"current" bson:"current"`
	Average int `json:"average" bson:"average"`
	Total   int `json:"total" bson:"total"`
}

func AverageGamesPlayedByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func TotalGamesPlayedByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func AverageEarningsByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func TotalEarningsByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func LongestStreakByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func CurrentStreakByMember(guildID string, memberID string, game string, period Period) int {
	return 0
}

func AddMemberStats(guildID string, memberID string, game string, result string, earnings int) {
	// TODO: figure out what is required, and whether we should create a struct for that
	//       not sure how to tell when to get rid of the previous entries w/out recording every one, which I really
	//       don't want to have a monthly record for every member's every game, but I don't know how to do this unless
	//       I do. But I may have to do that to get the data I want.
	//
	// TODO: Try to find a clever way to only save a days worth of data. Maybe track all transactions daily, then do
	//       a running total of daily for a week, then running totals for a week. I don't think that'll be perfect,
	//       but it will be something.
	//
	// TODO: maybe something like: new_average = (old_average * (n-1) + new_value) / n
	//       it is a bit more complicated w/ the rolling averages, since we only want some of the values to be included
	//       in this case, perhaps multiply the old_average * (n-num_days_since), where "n" is the time period we are
	//                     interested in
}

// String returns the string representation of the GameOutcome value.
func (gr GameOutcome) String() string {
	return string(gr)
}
