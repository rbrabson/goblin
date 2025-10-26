package blackjack

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	blackjackMemberCollection = "blackjack_members"
	blackjackConfigCollection = "blackjack_configs"
)

// readConfig loads the blackjack configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	var config Config
	filter := bson.M{"guild_id": guildID}
	err := db.FindOne(blackjackConfigCollection, filter, &config)
	if err != nil {
		slog.Debug("blackjack config not found in the database, using default",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read blackjack config from the database",
		slog.String("guildID", guildID),
	)

	return &config
}

// writeConfig creates or updates the blackjack configuration in the database
func writeConfig(config *Config) {
	var filter bson.M
	if config.ID != primitive.NilObjectID {
		filter = bson.M{"_id": config.ID}
	} else {
		filter = bson.M{"guild_id": config.GuildID}
	}
	if err := db.UpdateOrInsert(blackjackConfigCollection, filter, config); err != nil {
		slog.Error("error writing blackjack config to the database",
			slog.String("guildID", config.GuildID),
			slog.Any("error", err),
		)
	}
	slog.Debug("write blackjack config to the database",
		slog.String("guildID", config.GuildID),
	)
}

// readMember loads the blackjack member from the database. If it does not exist then
// a `nil` value is returned.
func readMember(guildID, memberID string) *Member {
	var member Member
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	err := db.FindOne(blackjackMemberCollection, filter, &member)
	if err != nil {
		slog.Debug("blackjack member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read blackjack member from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return &member
}

// writeMember creates or updates the blackjack member in the database
func writeMember(member *Member) {
	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	if err := db.UpdateOrInsert(blackjackMemberCollection, filter, member); err != nil {
		slog.Error("error writing blackjack member to the database",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("error", err),
		)
	}
	slog.Debug("write blackjack member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}
