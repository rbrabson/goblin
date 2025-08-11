package stats

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	PlayerStatsCollection = "player_stats"
	ServerStatsCollection = "server_stats"
	GameStatsCollection   = "game_stats"
)

// readMemberStats retrieves the member statistics for a specific member in a guild for a specific game.
func readPlayerStats(guildID string, memberID string, game string) (*PlayerStats, error) {
	var ps PlayerStats
	filter := bson.M{"guild_id": guildID, "member_id": memberID, "game": game}
	err := db.FindOne(PlayerStatsCollection, filter, &ps)
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

// writePlayerStats updates or inserts the player statistics for a specific member in a guild.
func writePlayerStats(ps *PlayerStats) error {
	var filter bson.M
	if ps.ID != primitive.NilObjectID {
		filter = bson.M{"_id": ps.ID}
	} else {
		filter = bson.M{"guild_id": ps.GuildID, "member_id": ps.MemberID, "game": ps.Game}
	}

	err := db.UpdateOrInsert(PlayerStatsCollection, filter, ps)
	if err != nil {
		slog.Info("Writing player stats", "collection", PlayerStatsCollection, "PlayerStats", ps, "filter", filter, "error", err)
		return err
	}
	return nil
}

// deletePlayerStats removes the player statistics for a specific member in a guild.
func deletePlayerStats(ps *PlayerStats) error {
	var filter bson.M
	if ps.ID != primitive.NilObjectID {
		filter = bson.M{"_id": ps.ID}
	} else {
		filter = bson.M{"guild_id": ps.GuildID, "member_id": ps.MemberID, "game": ps.Game}
	}
	err := db.DeleteMany(PlayerStatsCollection, filter)
	if err != nil {
		return err
	}
	return nil
}

// Get all the matching player stats for the guild.
func readMultiplePlayerStats(guildID string, filter interface{}, sortBy interface{}, limit int64) []*PlayerStats {
	var playerStats []*PlayerStats
	err := db.FindMany(PlayerStatsCollection, filter, &playerStats, sortBy, limit)
	if err != nil {
		slog.Error("unable to read player stats from the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read player stats from the database",
		slog.String("guildID", guildID),
		slog.Int("count", len(playerStats)),
	)

	return playerStats
}

// readGameStats retrieves the game statistics for a specific game in a guild.
func readGameStats(guildID string, game string, day time.Time) (*GameStats, error) {
	var gs GameStats
	filter := bson.M{"guild_id": guildID, "game": game, "day": day}
	err := db.FindOne(GameStatsCollection, filter, &gs)
	if err != nil {
		return nil, err
	}
	return &gs, nil
}

// writeGameStats updates or inserts the game statistics a guild.
func writeGameStats(gs *GameStats) error {
	var filter bson.M
	if gs.ID != primitive.NilObjectID {
		filter = bson.M{"_id": gs.ID}
	} else {
		filter = bson.M{"guild_id": gs.GuildID, "game": gs.Game, "day": gs.Day}
	}

	err := db.UpdateOrInsert(GameStatsCollection, filter, gs)
	if err != nil {
		slog.Info("writing game stats", "collection", GameStatsCollection, "GameStats", gs, "filter", filter, "error", err)
		return err
	}
	return nil
}

// deleteGameStats removes the game statistics for a specific game in a guild.
func deleteGameStats(gs *GameStats) error {
	var filter bson.M
	if gs.ID != primitive.NilObjectID {
		filter = bson.M{"_id": gs.ID}
	} else {
		filter = bson.M{"guild_id": gs.GuildID, "game": gs.Game, "day": gs.Day}
	}
	err := db.DeleteMany(GameStatsCollection, filter)
	if err != nil {
		return err
	}
	return nil
}

// getLastDatePlayed retrieves the last date a member played a game in a guild.
func getLastDatePlayed(guildID string, memberID string) time.Time {
	// Use aggregation pipeline to find the maximum last_played date for the member
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild and member
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: guildID},
				{Key: "member_id", Value: memberID},
			}},
		},
		// Stage 2: Group all documents and find the maximum last_played date
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "last_date_played", Value: bson.D{
					{Key: "$max", Value: "$last_played"},
				}},
			}},
		},
	}

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get last date played",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.Any("error", err),
		)
		return time.Time{}
	}

	if len(docs) == 0 {
		slog.Debug("no game data found for member",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
		)
		return time.Time{}
	}

	var t time.Time
	result := docs[0]
	if lastDatePlayed, ok := result["last_date_played"].(primitive.DateTime); ok {
		t = lastDatePlayed.Time().UTC()
		slog.Debug("found last date played",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.Time("last_date_played", t),
		)
	} else {
		slog.Warn("unexpected data type for last_date_played",
			slog.String("guild_id", guildID),
			slog.String("member_id", memberID),
			slog.Any("value", result["last_date_played"]),
		)
		t = time.Time{}
	}

	return t
}

// getFirstGameDate retrieves the earliest date a game was played by any member in a guild.
func getFirstGameDate(guildID string, game string) time.Time {
	var pipeline mongo.Pipeline
	if game == "" || game == "all" {
		pipeline = mongo.Pipeline{
			// Stage 1: Match documents for the specific guild for all games
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "guild_id", Value: guildID},
				}},
			},
			// Stage 2: Group all documents and find the minimum first_played date
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "first_game_date", Value: bson.D{
						{Key: "$min", Value: "$first_played"},
					}},
				}},
			},
		}
	} else {
		// Use aggregation pipeline to find the minimum first_played date
		pipeline = mongo.Pipeline{
			// Stage 1: Match documents for the specific guild and game
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "guild_id", Value: guildID},
					{Key: "game", Value: game},
				}},
			},
			// Stage 2: Group all documents and find the minimum first_played date
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "first_game_date", Value: bson.D{
						{Key: "$min", Value: "$first_played"},
					}},
				}},
			},
		}
	}

	docs, err := db.Aggregate(PlayerStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get first game date",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Any("error", err),
		)
		return today().AddDate(-1, 0, 0) // Default to 1 years ago if no data found
	}

	if len(docs) == 0 {
		slog.Debug("no game data found",
			slog.String("guild_id", guildID),
			slog.String("game", game),
		)
		return today().AddDate(-1, 0, 0) // Default to 1 years ago if no data found
	}

	result := docs[0]
	if firstGameDate, ok := result["first_game_date"].(primitive.DateTime); ok {
		t := firstGameDate.Time().UTC()
		slog.Debug("found first game date",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Time("first_game_date", t),
		)
		return t
	}
	slog.Warn("unexpected data type for first_game_date",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Any("value", result["first_game_date"]),
	)

	firstGameDate := today().AddDate(-1, 0, 0)
	slog.Debug("defaulting to 1 year ago for first game date",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("default_first_game_date", firstGameDate),
	)
	return firstGameDate
}

// getFirstGameDate retrieves the earliest date a game was played by any member in a guild.
func getFirstServerGameDate(guildID string, game string) time.Time {
	var pipeline mongo.Pipeline
	if game == "" || game == "all" {
		pipeline = mongo.Pipeline{
			// Stage 1: Match documents for the specific guild for all games
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "guild_id", Value: guildID},
				}},
			},
			// Stage 2: Group all documents and find the minimum first_played date
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "day", Value: bson.D{
						{Key: "$min", Value: "$day"},
					}},
				}},
			},
		}
	} else {
		// Use aggregation pipeline to find the minimum first_played date
		pipeline = mongo.Pipeline{
			// Stage 1: Match documents for the specific guild and game
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "guild_id", Value: guildID},
					{Key: "game", Value: game},
				}},
			},
			// Stage 2: Group all documents and find the minimum first_played date
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "day", Value: bson.D{
						{Key: "$min", Value: "$day"},
					}},
				}},
			},
		}
	}

	docs, err := db.Aggregate(GameStatsCollection, pipeline)
	if err != nil {
		slog.Error("failed to get first game date",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Any("error", err),
		)
		return today().AddDate(-1, 0, 0) // Default to 1 years ago if no data found
	}

	if len(docs) == 0 {
		slog.Debug("no game data found",
			slog.String("guild_id", guildID),
			slog.String("game", game),
		)
		return today().AddDate(-1, 0, 0) // Default to 1 years ago if no data found
	}

	result := docs[0]
	if firstGameDate, ok := result["day"].(primitive.DateTime); ok {
		t := firstGameDate.Time().UTC()
		slog.Debug("found first game date",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Time("day", t),
		)
		return t
	}
	slog.Warn("unexpected data type for first_game_date",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Any("value", result["day"]),
	)

	firstGameDate := today().AddDate(-1, 0, 0)
	slog.Debug("defaulting to 1 year ago for first game date",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Time("day", firstGameDate),
	)
	return firstGameDate
}
