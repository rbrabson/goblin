package stats

import (
	"sync"
	"time"
)

var (
	memberStatsLock = &sync.Mutex{}
)

// MemberStats represents the statistics of a member in a specific game within a guild.
type MemberStats struct {
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	Game        string             `json:"game" bson:"game"`
	FirstPlayed time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed  time.Time          `json:"last_played" bson:"last_played"`
	DailyStats  []*MembeDailyStats `json:"daily_stats" bson:"daily_stats"`
	mutex       sync.Mutex         `json:"-" bson:"-"`
}

// MembeDailyStats represents the daily statistics of a member in a specific game.
type MembeDailyStats struct {
	Earnings   int            `json:"total_earnings" bson:"total_earnings"`
	TotalGames int            `json:"total_games" bson:"total_games"`
	Results    map[string]int `json:"results" bson:"results"`
	Day        time.Time      `json:"day" bson:"day"`
}

// NewMemberStats initializes a new MemberStats instance for a specific guild and member.
func (m *MemberStats) NewMemberStats(guildID, memberID, game string) {
	m.GuildID = guildID
	m.MemberID = memberID
	m.Game = game
	m.FirstPlayed = time.Now()
	m.LastPlayed = time.Time{}
	m.DailyStats = []*MembeDailyStats{}
}

// NewMemberDailyStats initializes a new MemberGameStats instance.
func NewMemberDailyStats() *MembeDailyStats {
	return &MembeDailyStats{
		Earnings:   0,
		TotalGames: 0,
		Results:    map[string]int{},
	}
}

// AddOutcome adds the outcome of a game played by the member, updating their statistics accordingly.
func (m *MemberStats) AddOutcome(result string, earnings int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.LastPlayed = time.Now()
	if len(m.DailyStats) == 0 {
		m.DailyStats = append(m.DailyStats, NewMemberDailyStats())
	}

	mgs := m.DailyStats[len(m.DailyStats)-1]
	mgs.AddOutcome(result, earnings)
}

// AddOutcome adds the outcome of a game played by the member, updating their daily statistics accordingly.
func (mgs *MembeDailyStats) AddOutcome(result string, earnings int) {
	mgs.Earnings += earnings
	mgs.TotalGames++

	if r, ok := mgs.Results[result]; ok {
		mgs.Results[result] = r + 1
		return
	}

	mgs.Results[result] = 1
}

func updateDailyStats() {
	// This function should be called once a day to update the daily stats for all members.
	// It should iterate through all MemberStats, check if they have played in the past year,
	// and manage their daily stats accordingly.

	memberStatsLock.Lock()
	defer memberStatsLock.Unlock()

	// If any guild member hasn't playyed in the past year, delete their daily stats.
}

func daysInYear(date time.Time) int {
	year := date.Year()
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return 366 // Leap year
	}
	return 365 // Non-leap year
}
