package stats

import (
	"sync"
	"time"
)

// 365 days a year
// - keep track of the total only for a period
// - keep track of the next day's total
// - when the day completes, combine the two
//   -- (previous total) / (period - 1)) + current day's total
// - for average, divide the total by the period

// Stats
// - keep track of the previous 365 days + the current day
// - Take the last <period> days from the set
//   -- Total is the sum
//   -- Average is the Total / Period
// - Every day, replace the oldest day (if there are at least 365)
//   -- If less than 365, add to the set

// I think that is all that is needed. Can also use this to keep track of the
//    max values over a period, etc.

// Unique stats may be different
// - Need to know how many times a day they play, what days they play, etc.

type MemberStats struct {
	GuildID            string             `json:"guild_id" bson:"guild_id"`
	MemberID           string             `json:"member_id" bson:"member_id"`
	Game               string             `json:"game" bson:"game"`
	FirstPlayed        time.Time          `json:"first_played" bson:"first_played"`
	LastPlayed         time.Time          `json:"last_played" bson:"last_played"`
	LastUpdated        time.Time          `json:"last_updated" bson:"last_updated"`
	HourlyMemberStats  HourlyMemberStats  `json:"hourly_stats" bson:"hourly_stats"`
	DailyMemberStats   DailyMemberStats   `json:"daily_stats" bson:"daily_stats"`
	WeeklyMemberStats  WeeklyMemberStats  `json:"weekly_stats" bson:"weekly_stats"`
	MonthlyMemberStats MonthlyMemberStats `json:"monthly_stats" bson:"monthly_stats"`
	AllTimeMemberStats AllTimeMemberStats `json:"all_time_stats" bson:"all_time_stats"`
	mutex              sync.Mutex         `json:"-" bson:"-"`
}

type HourlyMemberStats struct {
	CurrentDay  []MemberGameStats `json:"last_24_hours" bson:"last_24_hours"`
	CurrentHour MemberGameStats   `json:"current_hour" bson:"current_hour"`
}

type DailyMemberStats struct {
	CurrentMonth []MemberGameStats `json:"current_month" bson:"current_month"`
	CurrentDay   MemberGameStats   `json:"current_day" bson:"current_day"`
}

type WeeklyMemberStats struct {
	CurrentMonth []MemberGameStats `json:"current_month" bson:"current_month"`
	CurrentWeek  MemberGameStats   `json:"current_week" bson:"current_week"`
}

type MonthlyMemberStats struct {
	CurrentYear  []MemberGameStats `json:"current_year" bson:"current_year"`
	CurrentMonth MemberGameStats   `json:"current_month" bson:"current_month"`
}

type AllTimeMemberStats struct {
	AllTime GameStats `json:"all_time" bson:"all_time"`
}

type MemberGameStats struct {
	TotalEarnings   int            `json:"total_earnings" bson:"total_earnings"`
	TotalGames      int            `json:"total_games" bson:"total_games"`
	AverageEarnings float64        `json:"average_earnings" bson:"average_earnings"`
	AverageGames    float64        `json:"average_games" bson:"average_games"`
	Outcomes        map[string]int `json:"outcomes" bson:"outcomes"`
}

func NewMemberStats(guildID string, memberID string, game string) *MemberStats {
	return &MemberStats{
		GuildID:     guildID,
		MemberID:    memberID,
		Game:        game,
		FirstPlayed: time.Now(),
		LastPlayed:  time.Time{},
		LastUpdated: time.Time{},
		HourlyMemberStats: HourlyMemberStats{
			CurrentDay:  []MemberGameStats{},
			CurrentHour: MemberGameStats{},
		},
		DailyMemberStats: DailyMemberStats{
			CurrentMonth: []MemberGameStats{},
			CurrentDay:   MemberGameStats{},
		},
		WeeklyMemberStats: WeeklyMemberStats{
			CurrentMonth: []MemberGameStats{},
			CurrentWeek:  MemberGameStats{},
		},
		MonthlyMemberStats: MonthlyMemberStats{
			CurrentYear:  []MemberGameStats{},
			CurrentMonth: MemberGameStats{},
		},
		AllTimeMemberStats: AllTimeMemberStats{
			AllTime: GameStats{},
		},
	}

	// TODO: write this to the database
}

func GetMemberStats(guildID string, memberID string, game string) *MemberStats {
	// TODO: check the database for the status first
	return NewMemberStats(guildID, memberID, game)
}

func (ms *MemberStats) AddGameStats(outcome string, earnings int) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	const daysInWeek = 7.0

	// Update last played time
	now := time.Now()
	durationSinceLastUpdate := now.Sub(ms.LastUpdated)

	ms.HourlyMemberStats.CurrentHour.AverageEarnings = calcNewAverage(ms.HourlyMemberStats.CurrentHour.AverageEarnings, float64(earnings), daysInWeek, durationSinceLastUpdate)
	ms.HourlyMemberStats.CurrentHour.TotalEarnings += earnings
	ms.HourlyMemberStats.CurrentHour.AverageGames = calcNewAverage(ms.HourlyMemberStats.CurrentHour.AverageGames, 1.0, daysInWeek, durationSinceLastUpdate)
	ms.HourlyMemberStats.CurrentHour.TotalGames++
	ms.HourlyMemberStats.CurrentHour.Outcomes[outcome]++ // TODO: not sure if this will work, or if I need to add to the map first

	ms.LastPlayed = time.Now()
	ms.LastUpdated = time.Now()
	// TODO: update the database with the new stats
}

func calcNewAverage(previousAverage float64, newValue float64, period float64, durationSinceLastUpdate time.Duration) float64 {
	// TODO: test this to see if it works correctly
	oldDailyTotal := period * previousAverage
	oldDailyAverage := oldDailyTotal / period

	daysSinceLastUpdate := period - (durationSinceLastUpdate.Hours() / 24)
	newPreviousTotal := oldDailyAverage * daysSinceLastUpdate
	newDailyTotal := newPreviousTotal + newValue
	newDailyAverage := newDailyTotal / period

	return newDailyAverage
}
