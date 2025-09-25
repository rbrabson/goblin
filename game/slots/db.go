package slots

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SlotsMemberCollection = "slots_members"
)

// readMember loads the slots member from the database. If it does not exist then
// a `nil` value is returned.
func readMember(guildID string, memberID string) *Member {
	var member Member
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	err := db.FindOne(SlotsMemberCollection, filter, &member)
	if err != nil {
		slog.Debug("slots member not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read slots member from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return &member
}

// Write creates or updates the slots member in the database
func writeMember(member *Member) {
	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	if err := db.UpdateOrInsert(SlotsMemberCollection, filter, member); err != nil {
		slog.Error("error writing slots member to the database",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("error", err),
		)
	}
	slog.Debug("write slots member to the database",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)
}
