package blackjack

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member represents a member's statistics for the blackjack game.
type Member struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID             string             `json:"guild_id" bson:"guild_id"`
	MemberID            string             `json:"member_id" bson:"member_id"`
	CurrentWinStreak    int                `json:"current_win_streak" bson:"current_win_streak"`
	LongestWinStreak    int                `json:"longest_win_streak" bson:"longest_win_streak"`
	CurrentLosingStreak int                `json:"current_losing_streak" bson:"current_losing_streak"`
	LongestLosingStreak int                `json:"longest_losing_streak" bson:"longest_losing_streak"`
	TotalWins           int                `json:"total_wins" bson:"total_wins"`
	TotalLosses         int                `json:"total_losses" bson:"total_losses"`
	TotalBet            int                `json:"total_bet" bson:"total_bet"`
	TotalWinnings       int                `json:"total_winnings" bson:"total_winnings"`
	MaxWin              int                `json:"max_win" bson:"max_win"`
	LastPlayed          time.Time          `json:"last_played" bson:"last_played"`
}

func (m *Member) String() string {
	return "Member{" +
		"ID: " + m.ID.Hex() +
		", GuildID: " + m.GuildID +
		", MemberID: " + m.MemberID +
		", CurrentWinStreak: " + string(rune(m.CurrentWinStreak)) +
		", LongestWinStreak: " + string(rune(m.LongestWinStreak)) +
		", CurrentLosingStreak: " + string(rune(m.CurrentLosingStreak)) +
		", LongestLosingStreak: " + string(rune(m.LongestLosingStreak)) +
		", TotalWins: " + string(rune(m.TotalWins)) +
		", TotalLosses: " + string(rune(m.TotalLosses)) +
		", TotalBet: " + string(rune(m.TotalBet)) +
		", TotalWinnings: " + string(rune(m.TotalWinnings)) +
		", MaxWin: " + string(rune(m.MaxWin)) +
		", LastPlayed: " + m.LastPlayed.String() +
		"}"
}

// GetMember retrieves the member statistics for a specific guild and user.
// If the member does not exist, a new member is created and returned.
func GetMember(guildID, userID string) *Member {
	member := readMember(guildID, userID)
	if member == nil {
		member = newMember(guildID, userID)
	}
	return member
}

// newMember creates a new Member instance with default values and writes it to the database.
func newMember(guildID, userID string) *Member {
	member := &Member{
		ID:       primitive.NewObjectID(),
		GuildID:  guildID,
		MemberID: userID,
	}
	writeMember(member)

	return member
}
