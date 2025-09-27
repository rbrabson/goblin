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

	result := docs[0]
	lastPlayed := getTimeFromPipeline(result["last_date_played"])

	return lastPlayed
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
	firstGameDate := getTimeFromPipeline(result["first_game_date"])

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
	firstGameDate := getTimeFromPipeline(result["day"])

	return firstGameDate
}
