package race

import (
	"fmt"

	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
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
func GetRaceMember(guildID string, memberID string) *RaceMember {
	log.Trace("--> race.GetRaceMember")
	defer log.Trace("<-- race.GetRaceMember")

	member := readRaceMember(guildID, memberID)
	if member == nil {
		member = newRaceMember(guildID, memberID)
	}
	member.guildMember = guild.GetMember(guildID, memberID)
	return member
}

// newRaceMember returns a new race member for the guild. The member is saved to
// the database.
func newRaceMember(guildID string, memberID string) *RaceMember {
	log.Trace("--> race.newRaceMember")
	defer log.Trace("<-- race.newRaceMember")

	member := &RaceMember{
		GuildID:  guildID,
		MemberID: memberID,
	}

	writeRaceMember(member)
	log.WithFields(log.Fields{"guild": guildID, "member": memberID}).Info("new race member")

	return member
}

// WinRace is called when the race member won a race.
func (m *RaceMember) WinRace(amount int) {
	log.Trace("--> race.Member.WinRace")
	log.Trace("<-- race.Member.WinRace")

	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	bankAccount.Deposit(amount)

	m.RacesWon++
	m.TotalEarnings += amount
	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "winnings": amount}).Info("won race")
}

// PlaceInRace is called when the race member places (comes in 2nd) in a race.
func (m *RaceMember) PlaceInRace(amount int) {
	log.Trace("--> race.Member.PlaceInRace")
	log.Trace("<-- race.Member.PlaceInRace")

	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	bankAccount.Deposit(amount)

	m.RacesPlaced++
	m.TotalEarnings += amount
	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "winnings": amount}).Info("placed in race")
}

// ShowInRace is called when the race member shows (comes in 3rd) in a race.
func (m *RaceMember) ShowInRace(amount int) {
	log.Trace("--> race.Member.ShowInRace")
	log.Trace("<-- race.Member.ShowInRace")

	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	bankAccount.Deposit(amount)

	m.RacesShowed++
	m.TotalEarnings += amount
	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "winnings": amount}).Info("showed in race")
}

// LoseRace is called when the race member fails to win, place or show in a race.
func (m *RaceMember) LoseRace() {
	log.Trace("--> race.Member.LoseRace")
	log.Trace("<-- race.Member.LoseRace")

	m.RacesLost++
	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("lost race")
}

// PlaceBet is used to place a bet on a member of a race.
func (m *RaceMember) PlaceBet(betAmount int) error {
	log.Trace("-->race.Member.PlaceBet")
	defer log.Trace("<-- race.Member.PlaceBet")

	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	err := bankAccount.Withdraw(betAmount)
	if err != nil {
		return err
	}

	m.BetsMade++
	m.TotalEarnings -= betAmount

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "betAmount": betAmount}).Info("placed bet")

	return nil
}

// WinBet is used when a member wins a bet on a race.
func (m *RaceMember) WinBet(winnings int) {
	log.Trace("--> race.Member.WinBet")
	defer log.Trace("<-- race.Member.WinBet")

	bankAccount := bank.GetAccount(m.GuildID, m.MemberID)
	bankAccount.Deposit(winnings)

	m.BetsWon++
	m.BetsEarnings += winnings
	m.TotalEarnings += winnings
	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "winnings": winnings}).Info("win bet")
}

// WinBet is used when a member wins a bet on a race.
func (m *RaceMember) LoseBet() {
	log.Trace("--> race.Member.LoseBet")
	defer log.Trace("<-- race.Member.LoseBet")

	writeRaceMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("lose bet")
}

func (m *RaceMember) String() string {
	return fmt.Sprintf("RaceMember{GuildID: %s, MemberID: %s, RacesLost: %d, RacesPlaced: %d, RacesShowed: %d, RacesWon: %d, TotalRaces: %d, BetsEarnings: %d, BetsMade: %d, BetsWon: %d, TotalEarnings: %d}",
		m.GuildID, m.MemberID, m.RacesLost, m.RacesPlaced, m.RacesShowed, m.RacesWon, m.TotalRaces, m.BetsEarnings, m.BetsMade, m.BetsWon, m.TotalEarnings)
}
