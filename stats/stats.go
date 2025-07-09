package stats

import (
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
For each player:
- Keep track of when they last played. Also keep track of their individual stats
  for the past day; beyond that, they can be deleted.
  - Not sure how best to do this, or if it is even necessary. I worry about table row explosion.
  - The server is keeping track of averages per minute, so do we need much beyond that?
    - Just report the game stats, and whether it is a new since the last minute reported?
- No outputs for this right now

For each server:
- Keep track of:
  - Average & total number of unique players per period per individual game over a period
  - Average & total number of members that participate in an individual game over a period
  - Average & total number of times the average player played an individual game in a period
  - Average & total earnings per player per period per individual game
  - Average & total unique outcomes per individual game over a period
  - Average & total number of times each game is played over a period

Time Periods:
- Daily, Monthly, Yearly, All-Time

- When a new minute is calculated, redo the current hour
  - When the new hour reaches 60 minutes, delete the oldest hour and add the newest
- When a new hour is calculated, redo the current day
  - When the new day reaches 1, delete the oldest
- When a new day is calculated, redo the current monthly
  - When the new month reaches the end of the month, delete the oldest month & add the newest
- All-Time is updated on an hourly basis. Keep track of the timestamps, and use that to convert the average to
  an hourly basis. Multiply out by the number of hours in the period, add the new one, and then divide by the number of
  new hours in the period.
*/

// Race game outcomes.
const (
	Win   = "win"   // The member won the race.
	Place = "place" // The member came in second in the race.
	Show  = "show"  // The member came in third in the race.
	Lose  = "lose"  // The member lost the race.
)

// Heist game outcomes.
const (
	Escaped     = "escaped"     // The member successfully escaped the heist.
	Apprehended = "apprehended" // The member was apprehended during the heist.
	Died        = "died"        // The member died during the heist.
)

// Stats represents the statistics for a game over a period of time in a guild.
type Stats struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id" bson:"guild_id"`
	Game    string             `json:"game" bson:"game"`
	Daily   *GameStats         `json:"daily,omitempty" bson:"daily,omitempty"`
	Weekly  *GameStats         `json:"weekly,omitempty" bson:"weekly,omitempty"`
	Monthly *GameStats         `json:"monthly,omitempty" bson:"monthly,omitempty"`
	AllTime *GameStats         `json:"all_time,omitempty" bson:"all_time,omitempty"`
	mutex   sync.Mutex         `json:"-" bson:"-"`
}

// GameStats represents the statistics for a game in a guild over a specific period.
type GameStats struct {
	UniqueMembers      int            `json:"unique_members" bson:"unique_members"`
	TotalMembers       int            `json:"total_members" bson:"total_members"`
	AverageTimesPlayed int            `json:"average_times_played" bson:"average_times_played"`
	Earnings           int            `json:"earnings" bson:"earnings"`
	Outcomes           map[string]int `json:"outcomes" bson:"outcomes"`
	TotalTimesPlayed   int            `json:"total_times_played" bson:"total_times_played"`
}

// NewStats creates a new Stats instance for a specific guild and game.
func NewStats(guildID, game string) *Stats {
	return &Stats{
		GuildID: guildID,
		Game:    game,
		Daily:   NewGameStats(),
		Weekly:  NewGameStats(),
		Monthly: NewGameStats(),
		AllTime: NewGameStats(),
	}
}

// NewGameStats creates a new GameStats instance for a given period of time with default values.
func NewGameStats() *GameStats {
	return &GameStats{
		UniqueMembers:      0,
		TotalMembers:       0,
		AverageTimesPlayed: 0,
		Earnings:           0,
		Outcomes:           make(map[string]int),
		TotalTimesPlayed:   0,
	}
}
