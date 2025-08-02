package stats

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	memberStats2 []*MemberStats2
)

func TestPercentageOfInactivePlayersLastMonth(t *testing.T) {
	testSetup2(t)
	defer testTeardown2(t)

	today := today()
	oneMonthAgo := today.AddDate(0, -1, 0) // One month ago

	// Pipeline to find the percentage of players who haven't played in the past month
	pipeline := mongo.Pipeline{
		// Stage 1: Match documents for the specific guild and game
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "guild_id", Value: "test_guild"},
				{Key: "game", Value: "test_game"},
			}},
		},
		// Stage 2: Group by member_id and find their last activity day
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$member_id"},
				{Key: "last_played", Value: bson.D{{Key: "$max", Value: "$last_played"}}},
			}},
		},
		// Stage 3: Add a field to categorize players as active or inactive
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "is_inactive", Value: bson.D{
					{Key: "$cond", Value: bson.D{
						{Key: "if", Value: bson.D{
							{Key: "$lt", Value: bson.A{"$last_played", oneMonthAgo}},
						}},
						{Key: "then", Value: 1},
						{Key: "else", Value: 0},
					}},
				}},
			}},
		},
		// Stage 4: Group all players to calculate totals and percentages
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil}, // Group all documents together
				{Key: "total_players", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "inactive_players", Value: bson.D{{Key: "$sum", Value: "$is_inactive"}}},
				{Key: "active_players", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$subtract", Value: bson.A{1, "$is_inactive"}},
					}},
				}},
			}},
		},
		// Stage 5: Calculate percentages
		bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "inactive_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$inactive_players", "$total_players"}}},
						100,
					}},
				}},
				{Key: "active_percentage", Value: bson.D{
					{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{"$active_players", "$total_players"}}},
						100,
					}},
				}},
			}},
		},
	}

	docs, err := db.Aggregate("member_stats", pipeline)
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) == 0 {
		t.Log("No player data found")
		return
	}

	result := docs[0]
	totalPlayers := getInt(result["total_players"])
	inactivePlayers := getInt(result["inactive_players"])
	activePlayers := getInt(result["active_players"])
	inactivePercentage := getFloat64(result["inactive_percentage"])
	activePercentage := getFloat64(result["active_percentage"])

	t.Logf("Player Activity Summary (Past Month):")
	t.Logf("Total Players: %d", totalPlayers)
	t.Logf("Active Players: %d (%.2f%%)", activePlayers, activePercentage)
	t.Logf("Inactive Players: %d (%.2f%%)", inactivePlayers, inactivePercentage)

	// Assertions
	if totalPlayers != activePlayers+inactivePlayers {
		t.Errorf("Total players (%d) should equal active (%d) + inactive (%d)",
			totalPlayers, activePlayers, inactivePlayers)
	}

	if inactivePercentage < 0 || inactivePercentage > 100 {
		t.Errorf("Inactive percentage should be between 0 and 100, got %.2f", inactivePercentage)
	}

	if activePercentage < 0 || activePercentage > 100 {
		t.Errorf("Active percentage should be between 0 and 100, got %.2f", activePercentage)
	}

	// Check that percentages add up to approximately 100%
	totalPercentage := activePercentage + inactivePercentage
	if totalPercentage < 99.9 || totalPercentage > 100.1 {
		t.Errorf("Total percentage should be approximately 100%%, got %.2f", totalPercentage)
	}

	t.Error("Inactive Players analysis completed.")
}

// Helper functions for type conversion
func getInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func getFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0.0
	}
}

// testSetup initializes the test environment, including database connections and any necessary data.
func testSetup2(t *testing.T) {
	t.Log("Setting up test environment...")
	var ms *MemberStats2
	today := today()

	ms = &MemberStats2{
		GuildID:             "test_guild",
		MemberID:            "test_member_1",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -14, 0),
		LastPlayed:          today.AddDate(0, -8, 0),
		NumberOfTimesPlayed: 5,
	}
	if err := writeMemberStats2(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats2 = append(memberStats2, ms)

	ms = &MemberStats2{
		GuildID:             "test_guild",
		MemberID:            "test_member_2",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -13, 0), // Two days ago
		LastPlayed:          today.AddDate(0, 0, 0),   // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writeMemberStats2(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats2 = append(memberStats2, ms)

	ms = &MemberStats2{
		GuildID:             "test_guild",
		MemberID:            "test_member_3",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -48, 0), // Two days ago
		LastPlayed:          today.AddDate(0, -48, 0), // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writeMemberStats2(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats2 = append(memberStats2, ms)

	ms = &MemberStats2{
		GuildID:             "test_guild",
		MemberID:            "test_member_4",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -96, 0), // Two days ago
		LastPlayed:          today.AddDate(0, -23, 0), // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writeMemberStats2(ms); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	memberStats2 = append(memberStats2, ms)
}

// testTeardown cleans up the test environment, closing database connections and removing test data.
func testTeardown2(t *testing.T) {
	t.Log("Tearing down test environment...")

	// Remove all member_stats from the database
	for _, ms := range memberStats2 {
		err := deleteMemberStats2(ms)
		if err != nil {
			t.Error("Error deleting member stats", "error", err)
		}
	}
	memberStats2 = nil
}
