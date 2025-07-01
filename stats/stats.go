package stats

import "go.mongodb.org/mongo-driver/bson/primitive"

// Period represents a time interval or duration such as daily, weekly, monthly, or all-time.
type Period string

const (
	Daily   Period = "daily"
	Weekly  Period = "weekly"
	Monthly Period = "monthly"
	AllTime Period = "all_time"
)

// Stats represents statistical data related to a guild, including unique member counts and associated identifiers.
type Stats struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	UniqueMembers UniqueMembers      `json:"unique_members" bson:"unique_members"`
}

// UniqueMembers represents the count of unique members within specific time frames.
type UniqueMembers struct {
	// TODO: need to track how much the average player actually plays various games
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
