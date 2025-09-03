package stats

import (
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	millisToDays int = 1000 * 60 * 60 * 24
)

var (
	statsLock = sync.Mutex{}
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

// getPlayerStats retrieves the player statistics for a specific guild, member, and game.
// If the player stats do not exist, it creates a new PlayerStats instance.
func getPlayerStats(guildID string, memberID string, game string) *PlayerStats {
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
		LastPlayed:          time.Time{},
		NumberOfTimesPlayed: 0,
	}
	writePlayerStats(ps)
	return ps
}

// GetUniquePlayers retrieves the number of unique players for a specific guild, game, and date range.
func GetUniquePlayers(guildID string, game string, startDate time.Time, endDate time.Time) (int, error) {
	statsLock.Lock()
	defer statsLock.Unlock()

	slog.Debug("getting unique players",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	// Create aggregation pipeline to get unique players
	pipeline := make(mongo.Pipeline, 0, 3)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild and date range
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "last_played", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild, date range, and game
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
				{Key: "last_played", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		})
	}

	// Stage 2: Group by member_id to get unique players
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$member_id"},
		}},
	})

	// Stage 3: Count the unique players
	pipeline = append(pipeline, bson.D{
		{Key: "$count", Value: "unique_players"},
	})

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get unique players count",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Any("error", err),
		)
		return 0, err
	}

	if len(docs) == 0 {
		slog.Debug("no unique players found",
			slog.String("guild_id", guildID),
			slog.String("game", game),
		)
		return 0, nil
	}

	count := getInt(docs[0]["unique_players"])

	slog.Debug("unique players count retrieved",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Int("unique_players", count),
	)

	return count, nil
}

// GetPlayerRetention finds players who played after a specific date but haven't played
// recently for the duration provided (i.e., players who became inactive)
func GetPlayerRetention(guildID string, game string, after time.Time, inactiveDuration time.Duration) (*PlayerRetention, error) {
	statsLock.Lock()
	defer statsLock.Unlock()

	slog.Debug("calculating player retention",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("after", after),
		slog.Duration("inactive_duration", inactiveDuration),
	)

	today := today()
	endDate := today.AddDate(0, 0, -1)
	inactiveDays := int(inactiveDuration.Hours()/24) + 1 // Convert duration to the number of days
	pipeline := make(mongo.Pipeline, 0, 7)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild and game
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
			}},
		})
	}

	// Stage 2: Group by member_id to get their first and last played dates
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$member_id"},
			{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
			{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			{Key: "total_games", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
		}},
	})

	// Stage 3: Filter players who played after the specified date
	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "last_played", Value: bson.D{
				{Key: "$gte", Value: after},
			}},
		}},
	})

	// Stage 4: Add fields to calculate inactive status
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "days_since_last_played", Value: bson.D{
				{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$subtract", Value: bson.A{endDate, "$last_played"}}},
					millisToDays, // Convert milliseconds to days
				}},
			}},
		}},
	})

	// Stage 5: Categorize players as inactive or active
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "is_active", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$lt", Value: bson.A{"$days_since_last_played", inactiveDays}},
					}},
					{Key: "then", Value: 1},
					{Key: "else", Value: 0},
				}},
			}},
			{Key: "is_inactive", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$gte", Value: bson.A{"$days_since_last_played", inactiveDays}},
					}},
					{Key: "then", Value: 1},
					{Key: "else", Value: 0},
				}},
			}},
		}},
	})

	// Stage 6: Group all players to calculate totals and percentages
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil}, // Group all documents together
			{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "active_players", Value: bson.D{{Key: "$sum", Value: "$is_active"}}},
			{Key: "inactive_players", Value: bson.D{{Key: "$sum", Value: "$is_inactive"}}},
		}},
	})

	// Stage 7: Calculate percentages
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "inactive_percentage", Value: bson.D{
				{Key: "$multiply", Value: bson.A{
					bson.D{{Key: "$divide", Value: bson.A{"$inactive_players", "$total_players"}}},
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
	})

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
		InactivePlayers:    getInt(result["inactive_players"]), // Players who became inactive
		InactivePercentage: getFloat64(result["inactive_percentage"]),
		ActivePlayers:      getInt(result["active_players"]), // Players still active
		ActivePercentage:   getFloat64(result["active_percentage"]),
	}

	slog.Debug("player retention calculated",
		slog.Time("start_date", after),
		slog.Time("end_date", endDate),
		slog.Int("inactive_days", int(inactiveDuration.Hours())/24),
		slog.Int("total_eligible_players", getInt(result["total_players"])),
		slog.Int("aactive_players", retention.ActivePlayers),
		slog.Float64("active_percentage", retention.ActivePercentage),
		slog.Int("inactive_players", retention.InactivePlayers),
		slog.Float64("inactive_percentage", retention.InactivePercentage),
	)

	return retention, nil
}

// getAggregatePlayerStats retrieves aggregated player stats for a specific member and game
func getAggregatePlayerStats(guildID string, memberID string, game string) (*PlayerStats, error) {
	slog.Debug("getting aggregated player stats",
		slog.String("guild_id", guildID),
		slog.String("member_id", memberID),
		slog.String("game", game),
	)

	pipeline := make(mongo.Pipeline, 0, 4)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild and member
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "member_id", Value: memberID},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild, member, and game
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "member_id", Value: memberID},
				{Key: "game", Value: game},
			}},
		})
	}

	// Stage 2: Group by member_id and aggregate stats across all matching games
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$member_id"},
			{Key: "guild_id", Value: bson.D{{Key: "$first", Value: "$guild_id"}}},
			{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
			{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
			{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			{Key: "games", Value: bson.D{{Key: "$addToSet", Value: "$game"}}}, // Track which games they played
		}},
	})

	// Stage 3: Project fields for the result
	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "member_id", Value: "$_id"},
			{Key: "guild_id", Value: 1},
			{Key: "game", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$eq", Value: bson.A{bson.D{{Key: "$size", Value: "$games"}}, 1}},
					}},
					{Key: "then", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$games", 0}}}},
					{Key: "else", Value: game}, // Use the requested game parameter
				}},
			}},
			{Key: "first_played", Value: 1},
			{Key: "last_played", Value: 1},
			{Key: "number_of_times_played", Value: "$total_games_played"},
		}},
	})

	// Stage 4: Limit to 1 result (should only be one member anyway)
	pipeline = append(pipeline, bson.D{
		{Key: "$limit", Value: 1},
	})

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get aggregated player stats",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.String("game", game),
			slog.Any("error", err),
		)
		return nil, err
	}

	if len(docs) == 0 {
		slog.Debug("no player stats found",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.String("game", game),
		)
		return nil, nil // No stats found
	}

	doc := docs[0]
	ps := &PlayerStats{
		GuildID:             getString(doc["guild_id"]),
		MemberID:            getString(doc["member_id"]),
		Game:                getString(doc["game"]),
		FirstPlayed:         getTimeFromPipeline(doc["first_played"]),
		LastPlayed:          getTimeFromPipeline(doc["last_played"]),
		NumberOfTimesPlayed: getInt(doc["number_of_times_played"]),
	}

	slog.Debug("aggregated player stats retrieved",
		slog.String("guild_id", ps.GuildID),
		slog.String("member_id", ps.MemberID),
		slog.String("game", ps.Game),
		slog.Int("total_games_played", ps.NumberOfTimesPlayed),
		slog.Time("first_played", ps.FirstPlayed),
		slog.Time("last_played", ps.LastPlayed),
	)

	return ps, nil
}

// getPlayerStatsForMostActiveMembers returns the most active players using aggregation pipeline
func getPlayerStatsForMostActiveMembers(guildID string, game string) []*PlayerStats {
	slog.Debug("getting most active members",
		slog.String("guild_id", guildID),
		slog.String("game", game),
	)

	pipeline := make(mongo.Pipeline, 0, 5)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild and game
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
			}},
		})
	}

	// Stage 2: Group by member_id and sum number_of_times_played across all games
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$member_id"},
			{Key: "guild_id", Value: bson.D{{Key: "$first", Value: "$guild_id"}}},
			{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$number_of_times_played"}}},
			{Key: "first_played", Value: bson.D{{Key: "$min", Value: "$first_played"}}},
			{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			{Key: "games", Value: bson.D{{Key: "$addToSet", Value: "$game"}}}, // Track which games they played
		}},
	})

	// Stage 3: Sort by total_games_played (descending) and _id (ascending for tie-breaking)
	pipeline = append(pipeline, bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "total_games_played", Value: -1},
			{Key: "_id", Value: 1},
		}},
	})

	// Stage 4: Limit to top 10 players
	pipeline = append(pipeline, bson.D{
		{Key: "$limit", Value: 10},
	})

	// Stage 5: Project fields for the result
	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "member_id", Value: "$_id"},
			{Key: "guild_id", Value: 1},
			{Key: "game", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$eq", Value: bson.A{bson.D{{Key: "$size", Value: "$games"}}, 1}},
					}},
					{Key: "then", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$games", 0}}}},
					{Key: "else", Value: "all"},
				}},
			}},
			{Key: "first_played", Value: 1},
			{Key: "last_played", Value: 1},
			{Key: "number_of_times_played", Value: "$total_games_played"},
		}},
	})

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get most active members",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Any("error", err),
		)
		return []*PlayerStats{}
	}

	playerStats := make([]*PlayerStats, 0, len(docs))
	for _, doc := range docs {
		ps := &PlayerStats{
			GuildID:             getString(doc["guild_id"]),
			MemberID:            getString(doc["member_id"]),
			Game:                getString(doc["game"]),
			FirstPlayed:         getTimeFromPipeline(doc["first_played"]),
			LastPlayed:          getTimeFromPipeline(doc["last_played"]),
			NumberOfTimesPlayed: getInt(doc["number_of_times_played"]),
		}
		playerStats = append(playerStats, ps)
	}

	slog.Debug("most active members retrieved",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Int("count", len(playerStats)),
	)

	return playerStats
}

// Helper functions for type conversion
func getString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// getTimeFromPipeline retrieves a time.Time from a BSON DateTime or a time.Time.
func getTimeFromPipeline(value interface{}) time.Time {
	var t time.Time
	switch v := value.(type) {
	case primitive.DateTime:
		t = v.Time()
	case time.Time:
		t = v
	default:
		slog.Error("unknown type for time conversion",
			slog.Any("value", value),
		)
		t = time.Time{}
	}

	return t.UTC()
}
