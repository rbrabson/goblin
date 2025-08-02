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
func readPlayerStats(guildID string, memberID string, game string) (*PlayerStats, error) {
	var ps PlayerStats
	filter := bson.M{"guild_id": guildID, "member_id": memberID, "game": game}
	err := db.FindOne(MemberStatsCollection, filter, &ps)
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

	err := db.UpdateOrInsert(MemberStatsCollection, filter, ps)
	if err != nil {
		return err
	}
	return nil
}

// deletePlayerStats removes the player statistics for a specific member in a guild.
func deletePlayerStats(ps *PlayerStats) error {
	filter := bson.M{"_id": ps.ID}
	err := db.DeleteMany(MemberStatsCollection, filter)
	if err != nil {
		return err
	}
	return nil
}
