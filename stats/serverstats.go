package stats

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ServerStats represents the statistics for a specific game in a guild on a specific day.
type ServerStats struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	Game        string             `json:"game" bson:"game"`
	Day         time.Time          `json:"day" bson:"day"`
	Players     int                `json:"players" bson:"players"`
	GamesPlayed int                `json:"games_played" bson:"games_played"`
}

// ServerStatsGamesPlayed represents the number of players and games played for a specific game in a guild.
type ServerGamesPlayed struct {
	Players     int `json:"players" bson:"players"`
	GamesPlayed int `json:"games_played" bson:"games_played"`
}

// GetServerStats retrieves the server statistics for a specific game in a guild on a specific day.
func GetServerStats(guildID string, game string, day time.Time) *ServerStats {
	statsLock.Lock()
	defer statsLock.Unlock()

	ss, _ := readServerStats(guildID, game, day)
	if ss == nil {
		ss = newServerStats(guildID, game, day)
	}

	return ss
}

// NewServerStats creates a new ServerStats instance with the provided guild ID, game, and day.
func newServerStats(guildID string, game string, day time.Time) *ServerStats {
	return &ServerStats{
		GuildID:     guildID,
		Game:        game,
		Day:         day,
		Players:     0,
		GamesPlayed: 0,
	}
}

// GetServerGamesPlayed retrieves the aggregated games played statistics for a specific game in a guild.
func GetServerGamesPlayed(guildID string, game string, startDate time.Time, endDate time.Time) (*ServerGamesPlayed, error) {
	slog.Debug("Calculating server games played statistics",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	// Pipeline to aggregate server statistics
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild, game, and date range
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
				{Key: "day", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		},
		// Stage 2: Group all documents and sum the statistics
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: "$players"}}},
				{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$games_played"}}},
				{Key: "unique_days", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "avg_players_per_day", Value: bson.D{{Key: "$avg", Value: "$players"}}},
				{Key: "avg_games_per_day", Value: bson.D{{Key: "$avg", Value: "$games_played"}}},
			}},
		},
	}

	docs, err := db.Aggregate(ServerStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &ServerGamesPlayed{
			Players:     0,
			GamesPlayed: 0,
		}, nil
	}

	result := docs[0]
	gamesPlayed := &ServerGamesPlayed{
		Players:     getInt(result["total_players"]),
		GamesPlayed: getInt(result["total_games_played"]),
	}

	slog.Debug("Server games played statistics calculated",
		slog.Int("total_players", gamesPlayed.Players),
		slog.Int("total_games", gamesPlayed.GamesPlayed),
		slog.Int("unique_days", getInt(result["unique_days"])),
		slog.Float64("avg_players_per_day", getFloat64(result["avg_players_per_day"])),
		slog.Float64("avg_games_per_day", getFloat64(result["avg_games_per_day"])),
	)

	return gamesPlayed, nil
}

// GetServerGamesPlayed retrieves the aggregated games played statistics for a specific game in a guild.
func GetServerGamesPlayedAllGames(guildID string, startDate time.Time, endDate time.Time) (*ServerGamesPlayed, error) {
	slog.Debug("Calculating server games played statistics",
		slog.String("guild_id", guildID),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	// Pipeline to aggregate server statistics
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild, game, and date range
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "day", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		},
		// Stage 2: Group all documents and sum the statistics
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: "$players"}}},
				{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$games_played"}}},
				{Key: "unique_days", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "avg_players_per_day", Value: bson.D{{Key: "$avg", Value: "$players"}}},
				{Key: "avg_games_per_day", Value: bson.D{{Key: "$avg", Value: "$games_played"}}},
			}},
		},
	}

	docs, err := db.Aggregate(ServerStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &ServerGamesPlayed{
			Players:     0,
			GamesPlayed: 0,
		}, nil
	}

	result := docs[0]
	gamesPlayed := &ServerGamesPlayed{
		Players:     getInt(result["total_players"]),
		GamesPlayed: getInt(result["total_games_played"]),
	}

	slog.Debug("Server games played statistics calculated",
		slog.Int("total_players", gamesPlayed.Players),
		slog.Int("total_games", gamesPlayed.GamesPlayed),
		slog.Int("unique_days", getInt(result["unique_days"])),
		slog.Float64("avg_players_per_day", getFloat64(result["avg_players_per_day"])),
		slog.Float64("avg_games_per_day", getFloat64(result["avg_games_per_day"])),
	)

	return gamesPlayed, nil
}
