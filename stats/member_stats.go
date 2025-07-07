package stats

import "time"

type MemberStats struct {
	GuildID     string            `json:"guild_id" bson:"guild_id"`
	MemberID    string            `json:"member_id" bson:"member_id"`
	Game        string            `json:"game" bson:"game"`
	FirstPlayed time.Time         `json:"first_played" bson:"first_played"`
	LastPlayed  time.Time         `json:"last_played" bson:"last_played"`
	LastUpdated time.Time         `json:"last_updated" bson:"last_updated"`
	Last24Hours []MemberGameStats `json:"last_24_hours" bson:"last_24_hours"`
	CurrentHour GameStats         `json:"current_hour" bson:"current_hour"`
}

type MemberGameStats struct {
	TotalEarnings   int            `json:"total_earnings" bson:"total_earnings"`
	TotalGames      int            `json:"total_games" bson:"total_games"`
	AverageEarnings float64        `json:"average_earnings" bson:"average_earnings"`
	AverageGames    float64        `json:"average_games" bson:"average_games"`
	Outcomes        map[string]int `json:"outcomes" bson:"outcomes"`
}

func AddMemberStats(guildID string, memberID string, game string, result string, earnings int) {
	// TODO: figure out what is required, and whether we should create a struct for that
	//       not sure how to tell when to get rid of the previous entries w/out recording every one, which I really
	//       don't want to have a monthly record for every member's every game, but I don't know how to do this unless
	//       I do. But I may have to do that to get the data I want.
	//
	// TODO: Try to find a clever way to only save a days worth of data. Maybe track all transactions daily, then do
	//       a running total of daily for a week, then running totals for a week. I don't think that'll be perfect,
	//       but it will be something.
	//
	// TODO: maybe something like: new_average = (old_average * (n-1) + new_value) / n
	//       it is a bit more complicated w/ the rolling averages, since we only want some of the values to be included
	//       in this case, perhaps multiply the old_average * (n-num_days_since), where "n" is the time period we are
	//                     interested in
	// TODO: figure out how best to update the server stats
}

func AddStats(guildID string, memberID string, game string, result string, outcome string, earnings int) {
	// Keep track of the stats for the individual member in the guild for the game.
	// - In particular, we need to know the first and last time the game was played.
	//   - Used for updating the unique players in the days, weeks, months and all-time stats.
	// - A lot of what is in GameStats is needed here. Keep track of the last 24 hours, plus the current hour.
	//   - Once the current hour is over, remove the oldest hour and add the new one.
	//   - Then update the overall game stats for the guild.
	//   - Can be tricky as we need to run all the current status for every player hourly, and cannot count on a player
	//     playing each hour.
}
