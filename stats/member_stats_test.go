package stats

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

var (
	memberStats []*MemberStats
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
	guild.SetDB(db)
	bank.SetDB(db)
}

func TestMemberStatsDailyAverage(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	today := today()

	// Get the stats for the past week
	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "guild_id", Value: "test_guild"},
			{Key: "game", Value: "test_game"},
			{Key: "day", Value: bson.D{
				{Key: "$gte", Value: today.AddDate(0, 0, -7)},
				{Key: "$lt", Value: today},
			}},
		}},
	}

	// Can match on an "id" of "$day" as well to get daily values.
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$guild_id"},
			{Key: "earnings", Value: bson.D{{Key: "$avg", Value: "$total_earnings"}}},
			{Key: "games_played", Value: bson.D{{Key: "$avg", Value: "$total_played"}}},
		}},
	}

	// Sort by _id (guild_id) in ascending order
	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}}

	// Create the aggregation pipeline and run it.
	pipeline := mongodb.Pipeline{matchStage, groupStage, sortStage}
	docs, err := db.Aggregate("member_stats", pipeline)
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range docs {
		t.Log("Document", "doc", doc)
	}
	t.Error("Completed testMemberStats")
}

func TestMemberStatsWeeklyAverage(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	today := today()

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "guild_id", Value: "test_guild"},
			{Key: "game", Value: "test_game"},
			{Key: "day", Value: bson.D{
				{Key: "$gte", Value: today.AddDate(0, 0, -7)},
				{Key: "$lt", Value: today},
			}},
		}},
	}

	// Can match on an "id" of "$day" as well to get daily values.
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$guild_id"},
			{Key: "earnings", Value: bson.D{{Key: "$avg", Value: "$total_earnings"}}},
			{Key: "games_played", Value: bson.D{{Key: "$avg", Value: "$total_played"}}},
		}},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "month", Value: bson.D{{Key: "$month", Value: "$day"}}},
			{Key: "day", Value: bson.D{{Key: "$dateToString", Value: bson.D{
				{Key: "format", Value: "%Y-%m-%d"},
				{Key: "date", Value: "$day"},
			}}}},
			{Key: "earnings", Value: "$earnings"},
			{Key: "games_played", Value: "$games_played"},
		}},
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}}

	// Create the aggregation pipeline and run it.
	pipeline := mongodb.Pipeline{matchStage, groupStage, projectStage, sortStage}
	docs, err := db.Aggregate("member_stats", pipeline)
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range docs {
		t.Log("Document", "doc", doc)
	}
	t.Error("Completed testMemberStats")
}

func TestAggregate(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)

	endDate := today()
	startDate := endDate.AddDate(0, 0, -4) // 4 days ago
	numDays := endDate.Sub(startDate).Hours() / 24

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "guild_id", Value: "test_guild"},
			{Key: "game", Value: "test_game"},
			{Key: "day", Value: bson.D{
				{Key: "$gte", Value: startDate},
				{Key: "$lt", Value: endDate},
			}},
		}},
	}

	// Can match on an "id" of "$day" as well to get daily values.
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "guild_id", Value: "$guild_id"},
				{Key: "member_id", Value: "$member_id"},
			}},
			{Key: "member_id", Value: bson.D{{Key: "$sum", Value: 1}}},
		}},
	}

	regroupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id.member_id"},
			{Key: "guild_id", Value: bson.D{{Key: "$first", Value: "$_id.guild_id"}}},
			{Key: "member_id", Value: bson.D{{Key: "$first", Value: "$_id.member_id"}}},
			{Key: "number_of_times_played", Value: bson.D{{Key: "$first", Value: "$member_id"}}},
		}},
	}

	// Create the aggregation pipeline and run it.
	pipeline := mongodb.Pipeline{matchStage, groupStage, regroupStage}
	docs, err := db.Aggregate("member_stats", pipeline)
	if err != nil {
		t.Fatal(err)
	}
	var totalPerDay int
	for _, doc := range docs {
		switch v := doc["number_of_times_played"].(type) {
		case int32:
			totalPerDay += int(v)
		default:
			t.Errorf("Unexpected type for number_of_times_played: %T", v)
		}
		t.Logf("Document: guild_id=%s, member_id=%s, number_of_times_played=%d", doc["guild_id"], doc["member_id"], doc["number_of_times_played"])
	}
	t.Error("Completed testMemberStats")

	t.Logf("Unique number of players over %v days is %d", numDays, len(docs))
	t.Logf("Average number of players per day over time period is %f", float64(len(docs))/numDays)
	t.Logf("Total number of days all players played is %d", totalPerDay)
	t.Logf("Average number of days each player played is %f", float64(totalPerDay)/float64(len(docs)))
}

func testSetup(t *testing.T) {
	var ms *MemberStats
	today := today()

	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member",
		Game:        "test_game",
		Earnings:    4000,
		TotalPlayed: 80,
		Day:         today, // Today
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats = append(memberStats, ms)

	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member",
		Game:        "test_game",
		Earnings:    1000,
		TotalPlayed: 20,
		Day:         today.AddDate(0, 0, -1), // Yesterday,
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats = append(memberStats, ms)

	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member",
		Game:        "test_game",
		Earnings:    2500,
		TotalPlayed: 180,
		Day:         today.AddDate(0, 0, -2), // Two days ago
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats = append(memberStats, ms)

	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member_2",
		Game:        "test_game",
		Earnings:    500,
		TotalPlayed: 10,
		Day:         today.AddDate(0, 0, -2), // Two days ago
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}

	memberStats = append(memberStats, ms)
	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member_2",
		Game:        "test_game",
		Earnings:    10000,
		TotalPlayed: 60,
		Day:         today.AddDate(0, 0, -3), // Three days ago
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats = append(memberStats, ms)

	ms = &MemberStats{
		GuildID:     "test_guild",
		MemberID:    "test_member_3",
		Game:        "test_game",
		Earnings:    10000,
		TotalPlayed: 60,
		Day:         today.AddDate(0, 0, -4),
	}
	if err := writeMemberStats(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats = append(memberStats, ms)
}

func testTeardown(t *testing.T) {
	// Remove all member_stats from the database
	for _, ms := range memberStats {
		err := db.Delete(MemberStatsCollection, ms)
		if err != nil {
			t.Error("Error deleting member stats", "error", err)
		}
	}
	memberStats = nil
}
