package stats

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PlayerStats holds the statistics of a player in a game.
type PlayerStats struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	MemberID            string             `json:"member_id" bson:"member_id"`
	Game                string             `json:"game" bson:"game"`
	FirstPlayed         time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed          time.Time          `json:"last_played" bson:"last_played"`
	NumberOfTimesPlayed int                `json:"number_of_times_played" bson:"number_of_times_played"`
}

// PlayerRetention holds the retention statistics of players in a game.
type PlayerRetention struct {
	InactivePlayers    int     `json:"inactive_players"`
	InactivePercentage float64 `json:"inactive_percentage"`
	ActivePlayers      int     `json:"active_players"`
	ActivePercentage   float64 `json:"active_percentage"`
}

// GetPlayerStats retrieves the player statistics for a specific guild, member, and game.
// If the player stats do not exist, it creates a new PlayerStats instance.
func GetPlayerStats(guildID string, memberID string, game string) *PlayerStats {
	ps, _ := readPlayerStats(guildID, memberID, game)
	if ps == nil {
		ps = newPlayerStats(guildID, memberID, game)
	}

	return ps
}

// newPlayerStats creates a new PlayerStats instance with the current date as FirstPlayed and LastPlayed.
func newPlayerStats(guildID string, memberID string, game string) *PlayerStats {
	today := today()
	ps := &PlayerStats{
		GuildID:             guildID,
		MemberID:            memberID,
		Game:                game,
		FirstPlayed:         today,
		LastPlayed:          today,
		NumberOfTimesPlayed: 0,
	}
	writePlayerStats(ps)
	return ps
}

// GamePlayed updates the PlayerStats when a game is played.
func (ps *PlayerStats) GamePlayed() {
	ps.LastPlayed = today()
	ps.NumberOfTimesPlayed++
	writePlayerStats(ps)
}

// GetPlayerActivity retrieves the activity statistics of players in a game for a specific guild.
func GetPlayerActivity(guildID string, game string, checkAfter time.Time, duration time.Duration) (*PlayerRetention, error) {
	pastDate := today().Add(-duration)

	slog.Debug("Calculating player activity",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Duration("duration", duration),
		slog.Time("past_date", pastDate),
	)

	// Pipeline to find the percentage of players who played longer than the duration
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for players who started on or after the start date
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
				{Key: "last_played", Value: bson.D{
					{Key: "$gte", Value: checkAfter},
				}},
			}},
		},
		// Stage 2: Group by member_id to get their first and last played dates
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$member_id"},
				{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
				{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			}},
		},
		// Stage 3: Add fields to calculate if player is retained
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "played_duration", Value: bson.D{
					{Key: "$subtract", Value: bson.A{"$last_played", "$first_played"}},
				}},
				{Key: "past_date", Value: pastDate},
			}},
		},
		// Stage 4: Categorize players as retained or not retained
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "is_retained", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$and", Value: bson.A{
								// Player must have played past the cutoff date
								bson.D{{Key: "$gte", Value: bson.A{"$last_played", "$past_date"}}},
							}},
						}},
						{Key: "then", Value: 1},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
		// Stage 5: Group all players to calculate totals and percentages
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Group all documents together
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "retained_players", Value: bson.D{{Key: "$sum", Value: "$is_retained"}}},
				{Key: "not_retained_players", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$subtract", Value: bson.A{1, "$is_retained"}},
					}},
				}},
			}},
		},
		// Stage 6: Calculate percentages
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "retained_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$retained_players", "$total_players"}}},
						100,
					}},
				}},
				{Key: "not_retained_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$not_retained_players", "$total_players"}}},
						100,
					}},
				}},
			}},
		},
	}

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &PlayerRetention{
			InactivePlayers:    0,
			InactivePercentage: 0,
			ActivePlayers:      0,
			ActivePercentage:   0,
		}, nil
	}

	result := docs[0]
	retention := &PlayerRetention{
		InactivePlayers:    getInt(result["not_retained_players"]),
		InactivePercentage: getFloat64(result["not_retained_percentage"]),
		ActivePlayers:      getInt(result["retained_players"]),
		ActivePercentage:   getFloat64(result["retained_percentage"]),
	}

	return retention, nil
}
