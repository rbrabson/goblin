package slots

import (
	"time"

	"github.com/rbrabson/goblin/stats"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	cooldown = 5 * time.Second
)

// Member represents a member's statistics for the slots game.
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
	LastPlayed       time.Time          `json:"last_played" bson:"last_played"`
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

// IsInCooldown checks if the member is in cooldown. If not, it updates the LastPlayed time and returns false.
// If the member is in cooldown, it returns true.
func (m *Member) IsInCooldown(config *Config) bool {
	if time.Since(m.LastPlayed) < time.Duration(config.Cooldown) {
		return false
	}
	m.LastPlayed = time.Now()
	writeMember(m)
	return true
}

// GetCooldownRemaining returns the remaining cooldown time for the member.
func (m *Member) GetCooldownRemaining(config *Config) time.Duration {
	remaining := cooldown - time.Since(m.LastPlayed)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AddResults updates the member's statistics based on the results of a spin.
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

	memberIDs := []string{m.MemberID}
	stats.UpdateGameStats(m.GuildID, "slots", memberIDs)
}
