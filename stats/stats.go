package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	HoursPerDay   = 24.0
	HoursPerWeek  = HoursPerDay * 7
	HoursPerMonth = HoursPerDay * 30.436875 // 30.436875 is the average number of days in a month
)

type Stats struct {
	ID                   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID              string             `json:"guild_id" bson:"guild_id"`
	Game                 string             `json:"game" bson:"game"`
	Period               string             `json:"period" bson:"period"`
	AverageUniquePlayers float64            `json:"average_unique_players" bson:"average_unique_players"`
	AverageEarnings      float64            `json:"average_earnings" bson:"average_earnings"`
	AverageGamesPlayed   float64            `json:"average_games_played" bson:"average_games_played"`
	FirstUpdated         time.Time          `json:"first_updated" bson:"first_updated"`
	NumPeriods           int                `json:"num_periods" bson:"num_periods"`
}

// GetStats retrieves the statistics for a specific guild and game for a given period.
// If the stats do not exist, it creates a new entry for that guild and game.
func GetStats(guildID, game, period string) *Stats {
	// TODO: Read from the database & return the stats if found.
	//       If not found, create a new entry for that guild and game.
	stats := newStats(guildID, game, period)

	return stats
}

// newStats creates a new Stats entry with the given guildID, game, and period.
func newStats(guildID, game, period string) *Stats {
	stats := &Stats{
		GuildID:              guildID,
		Game:                 game,
		Period:               period,
		AverageUniquePlayers: 0,
		AverageEarnings:      0,
		AverageGamesPlayed:   0,
		FirstUpdated:         today(),
		NumPeriods:           0,
	}

	return stats
}

// Update updates the statistics with the given unique players, earnings, and games played.
func (s *Stats) Update(uniquePlayers, earnings, gamesPlayed int) {
	s.updateUniquePlayers(uniquePlayers)
	s.updateEarnings(earnings)
	s.updateGamesPlayed(gamesPlayed)
	s.NumPeriods++

	// TODO: write to the database
}

// updateUniquePlayers updates the average unique players based on the current period.
func (s *Stats) updateUniquePlayers(count int) {
	oldTotal := s.AverageUniquePlayers * float64(s.NumPeriods)
	s.AverageUniquePlayers = (oldTotal + float64(count)) / float64((s.NumPeriods + 1))
}

// updateEarnings updates the average earnings based on the current period.
func (s *Stats) updateEarnings(earnings int) {
	oldTotal := s.AverageEarnings * float64(s.NumPeriods)
	s.AverageEarnings = (oldTotal + float64(earnings)) / float64(s.NumPeriods+1)
}

// updateGamesPlayed updates the average games played based on the current period.
func (s *Stats) updateGamesPlayed(count int) {
	oldTotal := s.AverageGamesPlayed * float64(s.NumPeriods)
	s.AverageGamesPlayed = (oldTotal + float64(count)) / float64(s.NumPeriods+1)
}
