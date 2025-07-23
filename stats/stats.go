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
	LastUpdated          time.Time          `json:"last_updated" bson:"last_updated"`
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
		LastUpdated:          today(),
	}

	return stats
}

// Update updates the statistics with the given unique players, earnings, and games played.
func (s *Stats) Update(uniquePlayers, earnings, gamesPlayed int) {
	s.LastUpdated = today()
	s.updateUniquePlayers(uniquePlayers)
	s.updateEarnings(earnings)
	s.updateGamesPlayed(gamesPlayed)

	// TODO: write to the database
}

// updateUniquePlayers updates the average unique players based on the current period.
func (s *Stats) updateUniquePlayers(count int) {
	periods := periodsSince(s.FirstUpdated, s.LastUpdated, s.Period)

	oldTotal := s.AverageUniquePlayers * periods
	s.AverageUniquePlayers = (oldTotal + float64(count)) / (periods + 1)
}

// updateEarnings updates the average earnings based on the current period.
func (s *Stats) updateEarnings(earnings int) {
	periods := periodsSince(s.FirstUpdated, s.LastUpdated, s.Period)

	oldTotal := s.AverageEarnings * periods
	s.AverageEarnings = (oldTotal + float64(earnings)) / (periods + 1)
}

// updateGamesPlayed updates the average games played based on the current period.
func (s *Stats) updateGamesPlayed(count int) {
	periods := periodsSince(s.FirstUpdated, s.LastUpdated, s.Period)

	oldTotal := s.AverageGamesPlayed * periods
	s.AverageGamesPlayed = (oldTotal + float64(count)) / (periods + 1)
}

// / periodsSince calculates the number of periods (days, weeks, or months) since the first update.
func periodsSince(firstUpdate, lastUpdate time.Time, period string) float64 {
	switch period {
	case "daily":
		return lastUpdate.Sub(firstUpdate).Hours() / HoursPerDay
	case "weekly":
		return lastUpdate.Sub(firstUpdate).Hours() / HoursPerWeek
	case "monthly":
		return lastUpdate.Sub(firstUpdate).Hours() / HoursPerMonth
	default:
		return 0
	}
}
