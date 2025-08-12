package stats

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GameStats represents the statistics for a specific game in a guild on a specific day.
type GameStats struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	Game          string             `json:"game" bson:"game"`
	Day           time.Time          `json:"day" bson:"day"`
	UniquePlayers int                `json:"unique_players" bson:"unique_players"`
	TotalPlayers  int                `json:"total_players" bson:"total_players"`
	GamesPlayed   int                `json:"games_played" bson:"games_played"`
}

// GamesPlayed represents the statistics for games played in a guild on a specific day.
type GamesPlayed struct {
	NumberOfDays                int
	UniquePlayers               int
	UniquePlayersPerDay         float64
	TotalPlayers                int
	TotalPlayersPerDay          float64
	TotalGamesPlayed            int
	AverageGamesPerPlayer       float64
	AverageGamesPerDay          float64
	AverageGamesPerPlayerPerDay float64
	AveragePlayersPerGame       float64
}

// getGameStats retrieves the game statistics for a specific game in a guild on a specific day.
func getGameStats(guildID string, game string, day time.Time) *GameStats {
	gs, err := readGameStats(guildID, game, day)
	if err != nil || gs == nil {
		gs = newGameStats(guildID, game, day)
	}
	return gs
}

// newGameStats creates a new GameStats instance for a specific game in a guild on a specific day.
func newGameStats(guildID string, game string, day time.Time) *GameStats {
	return &GameStats{
		GuildID: guildID,
		Game:    game,
		Day:     day,
	}
}

// UpdateGameStats updates the game statistics for a specific game in a guild.
func UpdateGameStats(guildID string, game string, memberIDs []string) {
	statsLock.Lock()
	defer statsLock.Unlock()

	today := today()

	var newUniquePlayersForGame, newUniquePlayersForAllGames int
	for _, memberID := range memberIDs {
		ps := getPlayerStats(guildID, memberID, game)
		if ps.LastPlayed.Before(today) {
			newUniquePlayersForGame++
			slog.Debug("first game played by member for today",
				slog.String("guild_id", ps.GuildID),
				slog.String("member_id", ps.MemberID),
				slog.String("game", ps.Game),
				slog.Time("day", today),
			)
		}
		lastDatePlayed := getLastDatePlayed(guildID, memberID)
		if lastDatePlayed.Before(today) {
			newUniquePlayersForAllGames++
			slog.Debug("first time playing any game by member for today",
				slog.String("guild_id", ps.GuildID),
				slog.String("member_id", ps.MemberID),
				slog.Time("day", today),
			)
		}

		ps.LastPlayed = today
		ps.NumberOfTimesPlayed++
		writePlayerStats(ps)
		slog.Debug("player stats updated",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.String("game", game),
			slog.Time("last_played", ps.LastPlayed),
			slog.Int("number_of_times_played", ps.NumberOfTimesPlayed),
		)
	}

	gs := getGameStats(guildID, game, today)
	gs.UniquePlayers += newUniquePlayersForGame
	gs.TotalPlayers += len(memberIDs)
	gs.GamesPlayed++
	writeGameStats(gs)
	slog.Debug("server stats for game updated",
		slog.String("guild_id", gs.GuildID),
		slog.String("game", gs.Game),
		slog.Time("day", gs.Day),
		slog.Int("games_played", gs.GamesPlayed),
		slog.Int("new_unique_players_for_game", newUniquePlayersForGame),
		slog.Int("unique_players", gs.UniquePlayers),
		slog.Int("new_total_players_for_game", len(memberIDs)),
		slog.Int("total_players", gs.TotalPlayers),
	)

	// Update unique players for all games
	gsAll := getGameStats(guildID, "all", today)
	gsAll.UniquePlayers += newUniquePlayersForAllGames
	gsAll.TotalPlayers += len(memberIDs)
	gsAll.GamesPlayed++
	writeGameStats(gsAll)
	slog.Debug("server stats for all games updated",
		slog.String("guild_id", gsAll.GuildID),
		slog.String("game", gsAll.Game),
		slog.Time("day", gsAll.Day),
		slog.Int("games_played", gsAll.GamesPlayed),
		slog.Int("new_unique_players_for_all_games", newUniquePlayersForAllGames),
		slog.Int("unique_players", gsAll.UniquePlayers),
		slog.Int("new_total_players_for_all_games", len(memberIDs)),
		slog.Int("total_players", gsAll.TotalPlayers),
	)
}

// GetGamesPlayed retrieves the aggregated games played statistics from the game_stats table
func GetGamesPlayed(guildID string, game string, startDate time.Time, endDate time.Time) (*GamesPlayed, error) {
	slog.Debug("calculating games played statistics from game_stats",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	pipeline := make(mongo.Pipeline, 0, 4)

	if game == "" || game == "all" {
		// Stage 1: Match documents for the specific guild and date range
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "day", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		})
	} else {
		// Stage 1: Match documents for the specific guild, game, and date range
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "game", Value: game},
				{Key: "day", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		})
	}
	// Stage 2: Group all documents and sum the statistics
	pipeline = append(pipeline, bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total_unique_players", Value: bson.D{{Key: "$sum", Value: "$unique_players"}}},
			{Key: "total_players", Value: bson.D{{Key: "$sum", Value: "$total_players"}}},
			{Key: "total_games_played", Value: bson.D{{Key: "$sum", Value: "$games_played"}}},
			{Key: "number_of_days", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "avg_unique_players_per_day", Value: bson.D{{Key: "$avg", Value: "$unique_players"}}},
			{Key: "avg_total_players_per_day", Value: bson.D{{Key: "$avg", Value: "$total_players"}}},
			{Key: "avg_games_per_day", Value: bson.D{{Key: "$avg", Value: "$games_played"}}},
		}},
	})
	// Stage 3: Calculate date range in days
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "date_range_days", Value: bson.D{
				{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$subtract", Value: bson.A{endDate, startDate}}},
					1000 * 60 * 60 * 24, // Convert milliseconds to days
				}},
			}},
		}},
	})
	// Stage 4: Calculate final averages
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "unique_players_per_day", Value: "$avg_unique_players_per_day"},
			{Key: "total_players_per_day", Value: "$avg_total_players_per_day"},
			{Key: "average_games_per_day", Value: "$avg_games_per_day"},
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
						{Key: "$gt", Value: bson.A{"$total_unique_players", 0}},
					}},
					{Key: "then", Value: bson.D{
						{Key: "$divide", Value: bson.A{"$total_games_played", "$total_unique_players"}},
					}},
					{Key: "else", Value: 0},
				}},
			}},
			{Key: "average_games_per_player_per_day", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$and", Value: bson.A{
							bson.D{{Key: "$gt", Value: bson.A{"$total_unique_players", 0}}},
							bson.D{{Key: "$gt", Value: bson.A{"$date_range_days", 0}}},
						}},
					}},
					{Key: "then", Value: bson.D{
						{Key: "$divide", Value: bson.A{
							bson.D{{Key: "$divide", Value: bson.A{"$total_games_played", "$total_unique_players"}}},
							"$date_range_days",
						}},
					}},
					{Key: "else", Value: 0},
				}},
			}},
			{Key: "average_players_per_game", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$gt", Value: bson.A{"$total_games_played", 0}},
					}},
					{Key: "then", Value: bson.D{
						{Key: "$divide", Value: bson.A{"$total_players", "$total_games_played"}},
					}},
					{Key: "else", Value: 0},
				}},
			}},
		}},
	})

	docs, err := db.Aggregate(GameStatsCollection, pipeline)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &GamesPlayed{}, nil
	}

	result := docs[0]
	gamesPlayed := &GamesPlayed{
		NumberOfDays:                getInt(result["number_of_days"]),
		UniquePlayers:               getInt(result["total_unique_players"]),
		UniquePlayersPerDay:         getFloat64(result["unique_players_per_day"]),
		TotalPlayers:                getInt(result["total_players"]),
		TotalPlayersPerDay:          getFloat64(result["total_players_per_day"]),
		TotalGamesPlayed:            getInt(result["total_games_played"]),
		AverageGamesPerPlayer:       getFloat64(result["average_games_per_player"]),
		AverageGamesPerDay:          getFloat64(result["average_games_per_day"]),
		AverageGamesPerPlayerPerDay: getFloat64(result["average_games_per_player_per_day"]),
		AveragePlayersPerGame:       getFloat64(result["average_players_per_game"]),
	}

	slog.Debug("games played statistics calculated from game_stats",
		slog.Int("number_of_days", gamesPlayed.NumberOfDays),
		slog.Int("unique_players", gamesPlayed.UniquePlayers),
		slog.Float64("unique_players_per_day", gamesPlayed.UniquePlayersPerDay),
		slog.Int("total_players", gamesPlayed.TotalPlayers),
		slog.Float64("total_players_per_day", gamesPlayed.TotalPlayersPerDay),
		slog.Int("total_games_played", gamesPlayed.TotalGamesPlayed),
		slog.Float64("avg_games_per_player", gamesPlayed.AverageGamesPerPlayer),
		slog.Float64("avg_games_per_day", gamesPlayed.AverageGamesPerDay),
		slog.Float64("avg_games_per_player_per_day", gamesPlayed.AverageGamesPerPlayerPerDay),
		slog.Float64("avg_players_per_game", gamesPlayed.AveragePlayersPerGame),
		slog.Float64("date_range_days", getFloat64(result["date_range_days"])),
	)

	return gamesPlayed, nil
}
