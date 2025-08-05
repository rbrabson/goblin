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

func getFirstGameDate(guildID string, game string) time.Time {
	// Use aggregation pipeline to find the minimum first_played date
	pipeline := mongo.Pipeline{
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
		slog.Debug("found first game date",
			slog.String("guild_id", guildID),
			slog.String("game", game),
			slog.Time("first_game_date", firstGameDate.Time()),
		)
		return firstGameDate.Time()
	}

	slog.Warn("unexpected data type for first_game_date",
		slog.String("guild_id", guildID),
		slog.String("game", game),
		slog.Any("value", result["first_game_date"]),
	)
	return today().AddDate(-1, 0, 0) // Default to 1 years ago if no data found
}
