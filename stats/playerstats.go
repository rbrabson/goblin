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

// GamesPlayed holds the statistics of games played in a guild.
type GamesPlayed struct {
	TotalGamesPlayed         int     `json:"total_games_played"`
	AverageGamesPlayedPerDay float64 `json:"average_games_played_per_day"`
	TotalNumberOfPlayers     int     `json:"total_number_of_players"`
	AverageGamesPerPlayer    float64 `json:"average_games_per_player"`
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

// GetPlayerRetention finds players who played after a specific date but haven't played
// recently for the duration provided (i.e., players who became inactive)
func GetPlayerRetention(guildID string, game string, afterDate time.Time, inactiveDuration time.Duration) (*PlayerRetention, error) {
	today := today()
	inactiveThreshold := today.Add(-inactiveDuration)

	slog.Debug("Calculating player churn",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("after_date", afterDate),
		slog.Duration("inactive_duration", inactiveDuration),
		slog.Time("inactive_threshold", inactiveThreshold),
	)

	// Pipeline to find players who were active after the date but are now inactive
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild and game
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
			}},
		},
		// Stage 2: Group by member_id to get their first and last played dates
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$member_id"},
				{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
				{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
				{Key: "total_games", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
			}},
		},
		// Stage 3: Filter players who played after the specified date
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "last_played", Value: bson.D{
					{Key: "$gte", Value: afterDate},
				}},
			}},
		},
		// Stage 4: Add fields to calculate churn status
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "days_since_last_played", Value: bson.D{
					{Key: "$divide", Value: bson.A{
						bson.D{{Key: "$subtract", Value: bson.A{today, "$last_played"}}},
						1000 * 60 * 60 * 24, // Convert milliseconds to days
					}},
				}},
				{Key: "inactive_threshold", Value: inactiveThreshold},
				{Key: "after_date", Value: afterDate},
			}},
		},
		// Stage 5: Categorize players as churned or active
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "is_churned", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$lt", Value: bson.A{"$last_played", "$inactive_threshold"}},
						}},
						{Key: "then", Value: 1},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
		// Stage 6: Group all players to calculate totals and percentages
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Group all documents together
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "churned_players", Value: bson.D{{Key: "$sum", Value: "$is_churned"}}},
				{Key: "active_players", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$subtract", Value: bson.A{1, "$is_churned"}},
					}},
				}},
				{Key: "avg_days_since_last_played", Value: bson.D{{Key: "$avg", Value: "$days_since_last_played"}}},
			}},
		},
		// Stage 7: Calculate percentages
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "churned_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$churned_players", "$total_players"}}},
						100,
					}},
				}},
				{Key: "active_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$active_players", "$total_players"}}},
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
		InactivePlayers:    getInt(result["churned_players"]), // Players who became inactive
		InactivePercentage: getFloat64(result["churned_percentage"]),
		ActivePlayers:      getInt(result["active_players"]), // Players still active
		ActivePercentage:   getFloat64(result["active_percentage"]),
	}

	slog.Debug("player retentiono calculated",
		slog.Int("total_eligible_players", getInt(result["total_players"])),
		slog.Int("inactive_players", retention.InactivePlayers),
		slog.Float64("inactive_percentage", retention.InactivePercentage),
		slog.Float64("avg_days_since_last_played", getFloat64(result["avg_days_since_last_played"])),
	)

	return retention, nil
}

// GetGamesPlayed calculates the games played statistics for a specific guild and game.
func GetGamesPlayed(guildID string, game string, startDate time.Time, endDate time.Time) (*GamesPlayed, error) {
	slog.Debug("calculating games played statistics",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	// Pipeline to calculate games played statistics
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild, game, and date range
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
				{Key: "last_played", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		},
		// Stage 2: Group by member_id to get total games per player
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$member_id"},
				{Key: "total_games_by_player", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
				{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
				{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			}},
		},
		// Stage 3: Group all players to calculate overall statistics
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Group all documents together
				{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$total_games_by_player"}}},
				{Key: "total_number_of_players", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "earliest_game", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
				{Key: "latest_game", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
				{Key: "games_per_player", Value: bson.D{{Key: "$push", Value: "$total_games_by_player"}}},
			}},
		},
		// Stage 4: Calculate averages and additional metrics
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "date_range_days", Value: bson.D{
					{Key: "$divide", Value: bson.A{
						bson.D{{Key: "$subtract", Value: bson.A{endDate, startDate}}},
						1000 * 60 * 60 * 24, // Convert milliseconds to days
					}},
				}},
				{Key: "actual_activity_days", Value: bson.D{
					{Key: "$divide", Value: bson.A{
						bson.D{{Key: "$subtract", Value: bson.A{"$latest_game", "$earliest_game"}}},
						1000 * 60 * 60 * 24, // Convert milliseconds to days
					}},
				}},
			}},
		},
		// Stage 5: Calculate final averages
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "average_games_played_per_day", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$gt", Value: bson.A{"$date_range_days", 0}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$divide", Value: bson.A{"$total_games_played", "$date_range_days"}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
				{Key: "average_games_per_player", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$gt", Value: bson.A{"$total_number_of_players", 0}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$divide", Value: bson.A{"$total_games_played", "$total_number_of_players"}},
						}},
						{Key: "else", Value: 0},
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
		return &GamesPlayed{
			TotalGamesPlayed:         0,
			AverageGamesPlayedPerDay: 0,
			TotalNumberOfPlayers:     0,
			AverageGamesPerPlayer:    0,
		}, nil
	}

	result := docs[0]
	gamesPlayed := &GamesPlayed{
		TotalGamesPlayed:         getInt(result["total_games_played"]),
		AverageGamesPlayedPerDay: getFloat64(result["average_games_played_per_day"]),
		TotalNumberOfPlayers:     getInt(result["total_number_of_players"]),
		AverageGamesPerPlayer:    getFloat64(result["average_games_per_player"]),
	}

	slog.Debug("Games played statistics calculated",
		slog.Int("total_games", gamesPlayed.TotalGamesPlayed),
		slog.Int("total_players", gamesPlayed.TotalNumberOfPlayers),
		slog.Float64("avg_games_per_day", gamesPlayed.AverageGamesPlayedPerDay),
		slog.Float64("avg_games_per_player", gamesPlayed.AverageGamesPerPlayer),
		slog.Float64("date_range_days", getFloat64(result["date_range_days"])),
		slog.Float64("actual_activity_days", getFloat64(result["actual_activity_days"])),
	)

	return gamesPlayed, nil
}
