package stats

import (
	"testing"
)

var (
	playerStats []*PlayerStats
)

func TestInactivePlayersLastMonth(t *testing.T) {
	testSetup2(t)
	defer testTeardown2(t)

	today := today()
	startDate := today.AddDate(-10, 0, 0)
	endDate := today.AddDate(0, -1, 0)
	duration := today.Sub(endDate)
	retention, err := GetPlayerRetention("test_guild", "test_game", startDate, duration)
	if err != nil {
		t.Error("Error calculating retention", "error", err)
		return
	}
	if retention == nil {
		t.Error("Retention data is nil")
		return
	}

	t.Logf("Player Activity Summary (Past Month):")
	t.Logf("Total Players: %d", retention.ActivePlayers+retention.InactivePlayers)
	t.Logf("Active Players: %d (%.2f%%)", retention.ActivePlayers, retention.ActivePercentage)
	t.Logf("Inactive Players: %d (%.2f%%)", retention.InactivePlayers, retention.InactivePercentage)

	t.Errorf("Inactive Players analysis completed.")
}

func TestPlayersWhoQuitAfterOneMonth(t *testing.T) {
	testSetup2(t)
	defer testTeardown2(t)

	today := today()
	endDate := today.AddDate(0, -1, 0)
	duration := today.Sub(endDate)
	retention, err := GetPlayerRetentionDuration("test_guild", "test_game", duration)
	if err != nil {
		t.Error("Error calculating retention", "error", err)
		return
	}
	if retention == nil {
		t.Error("Retention data is nil")
		return
	}

	t.Logf("Player Activity Summary (Past Month):")
	t.Logf("Total Players: %d", retention.ActivePlayers+retention.InactivePlayers)
	t.Logf("Players Still Playing: %d (%.2f%%)", retention.ActivePlayers, retention.ActivePercentage)
	t.Logf("Players No Longer Playing: %d (%.2f%%)", retention.InactivePlayers, retention.InactivePercentage)

	t.Errorf("Inactive Players analysis completed.")
}

// testSetup initializes the test environment, including database connections and any necessary data.
func testSetup2(t *testing.T) {
	t.Log("Setting up test environment...")
	var ps *PlayerStats
	today := today()

	ps = &PlayerStats{
		GuildID:             "test_guild",
		MemberID:            "test_member_1",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -14, 0),
		LastPlayed:          today.AddDate(0, -8, 0),
		NumberOfTimesPlayed: 5,
	}
	if err := writePlayerStats(ps); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	playerStats = append(playerStats, ps)

	ps = &PlayerStats{
		GuildID:             "test_guild",
		MemberID:            "test_member_2",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -13, 0), // Two days ago
		LastPlayed:          today.AddDate(0, 0, 0),   // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writePlayerStats(ps); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	playerStats = append(playerStats, ps)

	ps = &PlayerStats{
		GuildID:             "test_guild",
		MemberID:            "test_member_3",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -48, 0), // Two days ago
		LastPlayed:          today.AddDate(0, -48, 0), // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writePlayerStats(ps); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	playerStats = append(playerStats, ps)

	ps = &PlayerStats{
		GuildID:             "test_guild",
		MemberID:            "test_member_4",
		Game:                "test_game",
		FirstPlayed:         today.AddDate(0, -96, 0), // Two days ago
		LastPlayed:          today.AddDate(0, -23, 0), // One day ago
		NumberOfTimesPlayed: 5,
	}
	if err := writePlayerStats(ps); err != nil {
		t.Error("Error writing member stats", "error", err)
		return
	}
	playerStats = append(playerStats, ps)
}

// testTeardown cleans up the test environment, closing database connections and removing test data.
func testTeardown2(t *testing.T) {
	t.Log("Tearing down test environment...")

	// Remove all player_stats from the database
	for _, ps := range playerStats {
		err := deletePlayerStats(ps)
		if err != nil {
			t.Error("Error deleting member stats", "error", err)
		}
	}
	playerStats = nil
}
