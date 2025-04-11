package guild

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	GUILD_COLLECTION  = "guilds"
	MEMBER_COLLECTION = "guild_members"
)

// readMember reads the member from the database and returns the value, if it exists, or returns nil if the
// member does not exist in the database
func readMember(guildID string, memberID string) *Member {
	filter := bson.M{
		"guild_id":  guildID,
		"member_id": memberID,
	}
	var member Member
	err := db.FindOne(MEMBER_COLLECTION, filter, &member)
	if err != nil {
		slog.Debug("guild member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	return &member
}

// writeMember creates or updates the member data in the database.
func writeMember(member *Member) error {
	filter := bson.M{
		"guild_id":  member.GuildID,
		"member_id": member.MemberID,
	}
	err := db.UpdateOrInsert(MEMBER_COLLECTION, filter, member)
	if err != nil {
		slog.Error("unable to create or update guild in the database",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("write guild member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
		slog.String("name", member.Name),
	)
	return nil
}

// readGuild gets the guild from the database and returns the value, if it exists, or returns nil if the
func readGuild(guildID string) *Guild {
	filter := bson.M{"guild_id": guildID}
	var guild Guild
	err := db.FindOne(GUILD_COLLECTION, filter, &guild)
	if err != nil {
		slog.Debug("guild not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	return &guild
}

// writeGuild creates or updates the guild data in the database being used by the Discord bot.
func writeGuild(guild *Guild) error {
	filter := bson.M{"guild_id": guild.GuildID}
	err := db.UpdateOrInsert(GUILD_COLLECTION, filter, guild)
	if err != nil {
		slog.Error("unable to save guild to the database",
			slog.String("guildID", guild.GuildID),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("save guild to the database",
		slog.String("guildID", guild.GuildID),
	)

	return nil
}
