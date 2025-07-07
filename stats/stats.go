package stats

import "go.mongodb.org/mongo-driver/bson/primitive"

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

// Period represents a time interval or duration such as daily, weekly, monthly, or all-time.
type Period string

const (
	Daily   Period = "daily"
	Weekly  Period = "weekly"
	Monthly Period = "monthly"
	AllTime Period = "allTime"
)

// Stats represents statistical data related to a guild, including unique member counts and associated identifiers.
type Stats struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	UniqueMembers UniqueMembers      `json:"unique_members" bson:"unique_members"`
}

// UniqueMembers represents the count of unique members within specific time frames.
type UniqueMembers struct {
	// TODO: need to track how much the average player actually plays each games
	Daily   int `json:"daily" bson:"daily"`
	Weekly  int `json:"weekly" bson:"weekly"`
	Monthly int `json:"monthly" bson:"monthly"`
	AllTime int `json:"all_time" bson:"all_time"`
}

func AverageGamesPlayed(guildID string, game string, period Period) int {
	return 0
}

func TotalGamesPlayed(guildID string, game string, period Period) int {
	return 0
}

func AverageNumMembersPlayed(guildID string, game string, period Period) int {
	return 0
}

func TotalNumMembersPlayed(guildID string, game string, period Period) int {
	return 0
}

func AverageUniqueMembersPlayed(guildID string, game string, period Period) int {
	return 0
}

func TotalUniqueMembersPlayed(guildID string, game string, period Period) int {
	return 0
}

func AverageWinnings(guildID string, game string, period Period) int {
	return 0
}

func TotalWinnings(guildID string, game string, period Period) int {
	return 0
}

// String returns the string representation of the Period.
func (p Period) String() string {
	return string(p)
}
