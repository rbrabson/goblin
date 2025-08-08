package stats

import (
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	serverStatsLock = &sync.Mutex{}
)

// ServerStats represents the statistics for a specific game in a guild on a specific day.
type ServerStats struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	Game        string             `json:"game" bson:"game"`
	Day         time.Time          `json:"day" bson:"day"`
	Players     int                `json:"players" bson:"players"`
	GamesPlayed int                `json:"games_played" bson:"games_played"`
}

// ServerStatsGamesPlayed represents the number of players and games played for a specific game in a guild.
type ServerGamesPlayed struct {
	Players                     int     `json:"players" bson:"players"`
	AverageGamesPerPlayer       float64 `json:"average_games_per_player" bson:"average_games_per_player"`
	GamesPlayed                 int     `json:"games_played" bson:"games_played"`
	AverageGamesPlayedPerDay    float64 `json:"average_games_played_per_day" bson:"average_games_played_per_day"`
	AverageGamesPerPlayerPerDay float64 `json:"average_games_per_player_per_day" bson:"average_games_per_player_per_day"`
}

// GetServerStats retrieves the server statistics for a specific game in a guild on a specific day.
func GetServerStats(guildID string, game string, day time.Time) *ServerStats {
	serverStatsLock.Lock()
	defer serverStatsLock.Unlock()

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

// GetGamesPlayedForGame retrieves the aggregated games played statistics for all games in a guild.
func GetGamesPlayedForGame(guildID string, game string, startDate time.Time, endDate time.Time) (*ServerGamesPlayed, error) {
	serverStatsLock.Lock()
	defer serverStatsLock.Unlock()

	slog.Debug("calculating server games played statistics for all games",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	// Pipeline to aggregate server statistics across all games
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild and date range
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
		// Stage 3: Calculate additional averages
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "date_range_days", Value: bson.D{
					{Key: "$divide", Value: bson.A{
						bson.D{{Key: "$subtract", Value: bson.A{endDate, startDate}}},
						1000 * 60 * 60 * 24, // Convert milliseconds to days
					}},
				}},
			}},
		},
		// Stage 4: Calculate final averages
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
							{Key: "$gt", Value: bson.A{"$total_players", 0}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$divide", Value: bson.A{"$total_games_played", "$total_players"}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
				{Key: "average_games_per_player_per_day", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$and", Value: bson.A{
								bson.D{{Key: "$gt", Value: bson.A{"$total_players", 0}}},
								bson.D{{Key: "$gt", Value: bson.A{"$date_range_days", 0}}},
							}},
						}},
						{Key: "then", Value: bson.D{
							{Key: "$divide", Value: bson.A{
								bson.D{{Key: "$divide", Value: bson.A{"$total_games_played", "$total_players"}}},
								"$date_range_days",
							}},
						}},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
	}

	docs, err := db.Aggregate(ServerStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &ServerGamesPlayed{
			Players:                     0,
			AverageGamesPerPlayer:       0,
			GamesPlayed:                 0,
			AverageGamesPlayedPerDay:    0,
			AverageGamesPerPlayerPerDay: 0,
		}, nil
	}

	result := docs[0]
	gamesPlayed := &ServerGamesPlayed{
		Players:                     getInt(result["total_players"]),
		AverageGamesPerPlayer:       getFloat64(result["average_games_per_player"]),
		GamesPlayed:                 getInt(result["total_games_played"]),
		AverageGamesPlayedPerDay:    getFloat64(result["average_games_played_per_day"]),
		AverageGamesPerPlayerPerDay: getFloat64(result["average_games_per_player_per_day"]),
	}

	slog.Debug("server games played statistics calculated for all games",
		slog.Int("total_players", gamesPlayed.Players),
		slog.Int("total_games", gamesPlayed.GamesPlayed),
		slog.Float64("avg_games_per_player", gamesPlayed.AverageGamesPerPlayer),
		slog.Float64("avg_games_per_day", gamesPlayed.AverageGamesPlayedPerDay),
		slog.Float64("avg_games_per_player_per_day", gamesPlayed.AverageGamesPerPlayerPerDay),
		slog.Int("unique_days", getInt(result["unique_days"])),
		slog.Float64("date_range_days", getFloat64(result["date_range_days"])),
	)

	return gamesPlayed, nil
}
