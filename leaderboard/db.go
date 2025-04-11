package leaderboard

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	LEADERBOARD_COLLECTION = "leaderboards"
)

// readLeaderboard reads the leaderboard from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database
func readLeaderboard(guildID string) *Leaderboard {
	filter := bson.M{"guild_id": guildID}
	var lb Leaderboard
	err := db.FindOne(LEADERBOARD_COLLECTION, filter, &lb)
	if err != nil {
		sslog.Debug("leaderboard not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}

	return &lb
}

// writeBank creates or updates the bank for a guild in the database being used by the Discord bot.
func writeLeaderboard(lb *Leaderboard) error {
	filter := bson.M{"guild_id": lb.GuildID}

	err := db.UpdateOrInsert(LEADERBOARD_COLLECTION, filter, lb)
	if err != nil {
		sslog.Error("unable to save leaderboard to the database",
			slog.String("guildID", lb.GuildID),
			slog.Any("error", err),
		)
		return err
	}
	sslog.Debug("save leaderboard to the database",
		slog.String("guildID", lb.GuildID),
	)

	return nil
}
