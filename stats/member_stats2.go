package stats

import (
	"sync"
	"time"
)

type MemberStats2 struct {
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	Game        string             `json:"game" bson:"game"`
	FirstPlayed time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed  time.Time          `json:"last_played" bson:"last_played"`
	DailyStats  []MemberGameStats2 `json:"daily_stats" bson:"daily_stats"`
	mutex       sync.Mutex         `json:"-" bson:"-"`
}

type MemberGameStats2 struct {
	Earnings   int             `json:"total_earnings" bson:"total_earnings"`
	TotalGames int             `json:"total_games" bson:"total_games"`
	Outcomes   []MemberOutcome `json:"outcomes" bson:"outcomes"`
}

type MemberOutcome struct {
	Outcome       string    `json:"outcome" bson:"outcome"`
	Total         int       `json:"total" bson:"total"`
	CurrentStreak int       `json:"current_streak" bson:"current_streak"`
	BestStreak    int       `json:"best_streak" bson:"best_streak"`
	LastOccurence time.Time `json:"last_occurrence" bson:"last_occurrence"`
}
