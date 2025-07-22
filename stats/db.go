package stats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
