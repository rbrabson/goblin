package slots

import "go.mongodb.org/mongo-driver/bson/primitive"

type Member struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID          string             `json:"guild_id" bson:"guild_id"`
	MemberID         string             `json:"member_id" bson:"member_id"`
	CurrentWinStreak int                `json:"current_win_streak" bson:"current_win_streak"`
	LongestWinStreak int                `json:"longest_win_streak" bson:"longest_win_streak"`
	TotalWins        int                `json:"total_wins" bson:"total_wins"`
	TotalLosses      int                `json:"total_losses" bson:"total_losses"`
	TotalBet         int                `json:"total_bet" bson:"total_bet"`
	TotalWinnings    int                `json:"total_winnings" bson:"total_winnings"`
}

func GetMemberKey(guildID, userID string) *Member {
	member := readMember(guildID, userID)
	if member == nil {
		member = NewMember(guildID, userID)
	}
	return member
}

func NewMember(guildID, userID string) *Member {
	member := &Member{
		ID:               primitive.NewObjectID(),
		GuildID:          guildID,
		MemberID:         userID,
		CurrentWinStreak: 0,
		LongestWinStreak: 0,
		TotalWins:        0,
		TotalLosses:      0,
		TotalBet:         0,
		TotalWinnings:    0,
	}
	writeMember(member)

	return member
}

func (m *Member) AddResults(spinResult *SpinResult) {
	m.TotalBet += spinResult.Bet
	if spinResult.Payout > 0 {
		m.TotalWinnings += spinResult.Payout
		m.TotalWins++
		m.CurrentWinStreak++
		if m.CurrentWinStreak > m.LongestWinStreak {
			m.LongestWinStreak = m.CurrentWinStreak
		}
	} else {
		m.TotalLosses++
		m.CurrentWinStreak = 0
	}

	writeMember(m)
}
