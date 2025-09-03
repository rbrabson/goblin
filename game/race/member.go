package race

import (
	"fmt"
	"log/slog"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RaceMember represents a member of a guild that is assigned a racer
type RaceMember struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	MemberID      string             `json:"member_id" bson:"member_id"`
	RacesLost     int                `json:"races_lost" bson:"races_lost"`
	RacesPlaced   int                `json:"races_placed" bson:"races_placed"`
	RacesShowed   int                `json:"races_showed" bson:"races_showed"`
	RacesWon      int                `json:"races_won" bson:"races_won"`
	TotalRaces    int                `json:"total_races" bson:"total_races"`
	BetsEarnings  int                `json:"bets_earnings" bson:"bets_earnings"`
	BetsMade      int                `json:"bets_made" bson:"bets_made"`
	BetsWon       int                `json:"bets_won" bson:"bets_won"`
	TotalEarnings int                `json:"total_earnings" bson:"total_earnings"`
	guildMember   *guild.Member      `json:"-" bson:"-"`
}

// GetRaceMember gets a race member. THe member is created if it doesn't exist.
func GetRaceMember(guildID string, guildMember *guild.Member) *RaceMember {
	member := readRaceMember(guildID, guildMember.MemberID)
	if member == nil {
		member = newRaceMember(guildID, guildMember.MemberID)
	}
	member.guildMember = guild.GetMember(guildID, guildMember.MemberID)
	return member
}

// newRaceMember returns a new race member for the guild. The member is saved to
// the database.
func newRaceMember(guildID string, memberID string) *RaceMember {
	member := &RaceMember{
		GuildID:  guildID,
		MemberID: memberID,
	}

	writeRaceMember(member)
	slog.Info("new race member",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

	return member
}

// WinRace is called when the race member won a race.
func (m *RaceMember) WinRace(amount int) {
	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	if err := bankAccount.Deposit(amount); err != nil {
		slog.Error("error depositing race win amount",
			slog.String("guildID", m.GuildID),
			slog.String("memberID", m.MemberID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
	}

	m.RacesWon++
	m.TotalEarnings += amount
	writeRaceMember(m)

	slog.Info("won race",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
		slog.Int("winnings", amount),
	)
}

// PlaceInRace is called when the race member places (comes in 2nd) in a race.
func (m *RaceMember) PlaceInRace(amount int) {
	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	if err := bankAccount.Deposit(amount); err != nil {
		slog.Error("error depositing race place amount",
			slog.String("guildID", m.GuildID),
			slog.String("memberID", m.MemberID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
	}

	m.RacesPlaced++
	m.TotalEarnings += amount
	writeRaceMember(m)

	slog.Info("placed in race",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
		slog.Int("winnings", amount),
	)
}

// ShowInRace is called when the race member shows (comes in 3rd) in a race.
func (m *RaceMember) ShowInRace(amount int) {
	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	if err := bankAccount.Deposit(amount); err != nil {
		slog.Error("error depositing race show amount",
			slog.String("guildID", m.GuildID),
			slog.String("memberID", m.MemberID),
			slog.Int("amount", amount),
			slog.Any("error", err),
		)
	}

	m.RacesShowed++
	m.TotalEarnings += amount
	writeRaceMember(m)

	slog.Info("showed in race",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
		slog.Int("winnings", amount),
	)
}

// LoseRace is called when the race member fails to win, place or show in a race.
func (m *RaceMember) LoseRace() {
	m.RacesLost++
	writeRaceMember(m)

	slog.Info("lost race",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
	)
}

// PlaceBet is used to place a bet on a member of a race.
func (m *RaceMember) PlaceBet(betAmount int) error {
	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	err := bankAccount.Withdraw(betAmount)
	if err != nil {
		return err
	}

	m.BetsMade++
	m.TotalEarnings -= betAmount

	slog.Info("placed bet")
	slog.Info("placed bet",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
		slog.Int("betAmount", betAmount),
	)

	return nil
}

// WinBet is used when a member wins a bet on a race.
func (m *RaceMember) WinBet(winnings int) {
	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	if err := bankAccount.Deposit(winnings); err != nil {
		slog.Error("error depositing race win bet amount",
			slog.String("guildID", m.GuildID),
			slog.String("memberID", m.MemberID),
			slog.Int("amount", winnings),
			slog.Any("error", err),
		)
	}

	m.BetsWon++
	m.BetsEarnings += winnings
	m.TotalEarnings += winnings
	writeRaceMember(m)

	slog.Info("won bet",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
		slog.Int("winnings", winnings),
	)
}

// LoseBet is used when a member loses a bet on a race.
func (m *RaceMember) LoseBet() {
	writeRaceMember(m)

	slog.Info("lost bet",
		slog.String("guildID", m.GuildID),
		slog.String("memberID", m.MemberID),
	)
}

func (m *RaceMember) String() string {
	return fmt.Sprintf("RaceMember{GuildID: %s, MemberID: %s, RacesLost: %d, RacesPlaced: %d, RacesShowed: %d, RacesWon: %d, TotalRaces: %d, BetsEarnings: %d, BetsMade: %d, BetsWon: %d, TotalEarnings: %d}",
		m.GuildID, m.MemberID, m.RacesLost, m.RacesPlaced, m.RacesShowed, m.RacesWon, m.TotalRaces, m.BetsEarnings, m.BetsMade, m.BetsWon, m.TotalEarnings)
}
