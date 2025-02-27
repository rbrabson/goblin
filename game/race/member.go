package race

import (
	"github.com/rbrabson/goblin/bank"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member represents a member of a guild that is assigned a racer
type Member struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	GuildID     string             `json:"guild_id" bson:"guild_id"`
	MemberID    string             `json:"member_id" bson:"member_id"`
	RacesLost   int                `json:"races_lost" bson:"races_lost"`
	RacesPlaced int                `json:"races_placed" bson:"races_placed"`
	RacesShowed int                `json:"races_showed" bson:"races_showd"`
	RacesWon    int                `json:"races_won" bson:"races_won"`
	TotalRaces  int                `json:"total_races" bson:"total_races"`
	BetEarnings int                `json:"bet_earnings" bson:"bet_earnings"`
	BetsMade    int                `json:"bets_made" bson:"bets_made"`
	BetsWon     int                `json:"bets_won" bson:"bets_won"`
	TotalBets   int                `json:"total_bets" bson:"total_bets"`
	Earnings    int                `json:"earnings" bson:"earnings"`
	racer       *Racer             `json:"-" bson:"-"`
}

// GetMember gets a race member. THe member is created if it doesn't exist.
func GetMember(guildID string, memberID string) *Member {
	log.Trace("--> race.GetMember")
	defer log.Trace("<-- race.GetMember")

	member, err := getMember(guildID, memberID)
	if err != nil {
		member = newMember(guildID, memberID)
	}
	return member
}

// getMember gets a member from the database. If the member doesn't exist, then
// nil is returned.
func getMember(guildID string, memberID string) (*Member, error) {
	log.Trace("--> race.getMember")
	defer log.Trace("<-- race.getMember")

	member := readMember(guildID, memberID)
	if member == nil {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

// newMember returns a new race member for the guild. The member is saved to
// the database.
func newMember(guildID string, memberID string) *Member {
	log.Trace("--> race.newMember")
	defer log.Trace("<-- race.newMember")

	member := &Member{
		GuildID:     guildID,
		MemberID:    memberID,
		RacesWon:    0,
		RacesPlaced: 0,
		RacesShowed: 0,
		RacesLost:   0,
		TotalRaces:  0,
		BetsMade:    0,
		BetsWon:     0,
		BetEarnings: 0,
		TotalBets:   0,
		Earnings:    0,
	}

	writeMember(member)
	log.WithFields(log.Fields{"guild": guildID, "member": memberID}).Info("new member")

	return member
}

// WinRace is called when the race member won a race.
func (m *Member) WinRace() {
	log.Trace("--> race.Member.WinRace")
	log.Trace("<-- race.Member.WinRace")

	m.RacesWon++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("won race")
}

// PlaceInRace is called when the race member places (comes in 2nd) in a race.
func (m *Member) PlaceInRace() {
	log.Trace("--> race.Member.PlaceInRace")
	log.Trace("<-- race.Member.PlaceInRace")

	m.RacesPlaced++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("placed in race")

}

// ShowInRace is called when the race member shows (comes in 3rd) in a race.
func (m *Member) ShowInRace() {
	log.Trace("--> race.Member.ShowInRace")
	log.Trace("<-- race.Member.ShowInRace")

	m.RacesShowed++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("sbowed in race")
}

// LoseRace is called when the race member fails to win, place or show in a race.
func (m *Member) LoseRace() {
	log.Trace("--> race.Member.LoseRace")
	log.Trace("<-- race.Member.LoseRace")

	m.RacesLost++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID}).Info("lost race")
}

// PlaceBet is used to place a bet on a member of a race.
func (m *Member) PlaceBet(betAmount int) error {
	log.Trace("-->race.Member.PlaceBet")
	defer log.Trace("<-- race.Member.PlaceBet")

	b := bank.GetBank(m.GuildID)
	bankAccount := b.GetAccount(m.MemberID)
	err := bankAccount.Withdraw(betAmount)
	if err != nil {
		return err
	}

	m.BetsMade++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "betAmount": betAmount}).Info("placed bet")

	return nil
}

// WinBet is used when a member wins a bet on a race.
func (m *Member) WinBet(winnings int) {
	log.Trace("--> race.Member.WinBet")
	defer log.Trace("<-- race.Member.WinBet")

	b := bank.GetBank(m.GuildID)
	bankAccount := b.GetAccount(m.MemberID)
	bankAccount.Deposit(winnings)

	m.BetsWon++
	writeMember(m)

	log.WithFields(log.Fields{"guild": m.GuildID, "member": m.MemberID, "winnings": winnings}).Info("won bet")
}
