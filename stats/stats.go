package stats

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	memberStatsLock = &sync.Mutex{}
)

// MemberStats represents the daily statistics of a member of a guild in a specific game.
type MemberStats struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	Game        string             `json:"game" bson:"game"`
	Earnings    int                `json:"total_earnings" bson:"total_earnings"`
	TotalPlayed int                `json:"total_played" bson:"total_played"`
	Day         time.Time          `json:"day" bson:"day"`
}

// getMemberStats retrieves the statistics for a specific member in a guild for a specific game.
// If the stats do not exist, it creates a new entry for that member.
func getMemberStats(guildID, memberID, game string) *MemberStats {
	// Read from the database & return the member stats if found.

	return newMemberStats(guildID, memberID, game)
}

// newMemberStats creates a new MemberStats entry for a member in a guild for a specific game.
func newMemberStats(guildID, memberID, game string) *MemberStats {
	ms := &MemberStats{
		ID:          primitive.NewObjectID(),
		GuildID:     guildID,
		MemberID:    memberID,
		Game:        game,
		Earnings:    0,
		TotalPlayed: 0,
		Day:         getToday(),
	}
	// Write the new member stats to the database.
	return ms
}

// Update updates the earnings and total games played for a member's stats.
func UpdateMemberStats(guildID, memberID, game string, earnings int) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()

	ms := getMemberStats(guildID, memberID, game)
	ms.Earnings += earnings
	ms.TotalPlayed++
	// Write the updated stats back to the database.
}

func GetTotalUniquePlayers(guildID, game string, startDate, endDate time.Time) (int, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to count unique members
	// for the specified guild and game within the given date range.
	// Implementation is not provided here.
	return 0, nil
}

func GetAverageUniquePlayers(guildID, game string, startDate, endDate time.Time) (float64, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to calculate the average
	// unique players per day for the specified guild and game within the
	// given date range.
	// Implementation is not provided here.
	return 0.0, nil
}

func GetTotalEarningsPerPlayer(guildID, game string, startDate, endDate time.Time) (int, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to sum the earnings of all
	// members for the specified guild and game within the given date range.
	// Implementation is not provided here.
	return 0, nil
}

func GetAverageEarningsPerPlayer(guildID, game string, startDate, endDate time.Time) (float64, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to calculate the average earnings
	// per player for the specified guild and game within the given date range.
	// Implementation is not provided here.
	return 0.0, nil
}

func GetTotalGamesPlayed(guildID, game string, startDate, endDate time.Time) (int, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to sum the total games played
	// by all members for the specified guild and game within the given date range.
	// Implementation is not provided here.
	return 0, nil
}

func GetAverageGamesPlayed(guildID, game string, startDate, endDate time.Time) (float64, error) {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()
	// This function should query the database to calculate the average number of
	// games played per day for the specified guild and game within the given date range.
	// Implementation is not provided here.
	return 0.0, nil
}

// pruneMemberStats is a daily routine that prunes old member stats that are older than one year.
func pruneMemberStats() {
	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()

	lastYear := getToday().AddDate(-1, 0, 0)
	db.DeleteMany(MemberStatsCollection, bson.M{"day": bson.M{"$lt": lastYear}})
}

// getToday returns the current date with the time set to midnight.
func getToday() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}
