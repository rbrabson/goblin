package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MemberStatsCollection = "member_stats"
)

func countUniqueMembers(guildID string, start time.Time, end time.Time) (int, error) {
	filter := bson.M{
		"guild_id": guildID,
		"last_played": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	return db.DistinctCount(MemberStatsCollection, filter, "member_id")
}

// readMemberStats retrieves the member statistics for a specific member in a guild for a specific game.
func readMemberStats(guildID string, memberID string, game string) (*MemberStats, error) {
	var ms MemberStats
	filter := bson.M{"guild_id": guildID, "member_id": memberID, "game": game}
	err := db.FindOne(MemberStatsCollection, filter, &ms)
	if err != nil {
		return nil, err
	}
	return &ms, nil
}

// writeMemberStats updates or inserts the member statistics for a specific member in a guild.
func writeMemberStats(ms *MemberStats) error {
	var filter bson.M
	if ms.ID != primitive.NilObjectID {
		filter = bson.M{"_id": ms.ID}
	} else {
		filter = bson.M{"guild_id": ms.GuildID, "member_id": ms.MemberID, "game": ms.Game, "day": ms.Day}
	}

	err := db.UpdateOrInsert(MemberStatsCollection, filter, ms)
	if err != nil {
		return err
	}
	return nil
}

// deleteMemberStats removes the member statistics for a specific member in a guild.
func deleteMemberStats(ms *MemberStats) error {
	filter := bson.M{"_id": ms.ID}
	err := db.DeleteMany(MemberStatsCollection, filter)
	if err != nil {
		return err
	}
	return nil
}
