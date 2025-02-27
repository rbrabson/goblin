package race

type RaceMember struct {
	ID          string `json:"_id" bson:"_id"`
	GuildID     string `json:"guild_id" bson:"guild_id"`
	MemberID    string `json:"member_id" bson:"member_id"`
	NumRaces    int    `json:"num_races" bson:"num_races"`
	Win         int    `json:"win" bson:"win"`
	Place       int    `json:"place" bson:"place"`
	Show        int    `json:"show" bson:"show"`
	Loses       int    `json:"loses" bson:"loses"`
	Earnings    int    `json:"earnings" bson:"earnings"`
	BetsPlaced  int    `json:"bets_placed" bson:"bets_placed"`
	BetsWon     int    `json:"bets_won" bson:"bets_won"`
	BetEarnings int    `json:"bet_earnings" bson:"bet_earnings"`
	TotalBets   int    `json:"total_bets" bson:"total_bets"`
}
