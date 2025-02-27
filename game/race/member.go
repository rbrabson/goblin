package race

import (
	"github.com/rbrabson/goblin/guild"
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
func GetMember(g *guild.Guild, memberID string) *Member {
	log.Trace("--> race.GetMember")
	defer log.Trace("<-- race.GetMember")

	member, err := getMember(g, memberID)
	if err != nil {
		member = newMember(g, memberID)
	}
	return member
}

// getMember gets a member from the database. If the member doesn't exist, then
// nil is returned.
func getMember(g *guild.Guild, memberID string) (*Member, error) {
	log.Trace("--> race.getMember")
	defer log.Trace("<-- race.getMember")

	member := readMember(g, memberID)
	if member == nil {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

// newMember returns a new race member for the guild. The member is saved to
// the database.
func newMember(guild *guild.Guild, memberID string) *Member {
	log.Trace("--> race.newMember")
	defer log.Trace("<-- race.newMember")

	member := &Member{
		GuildID:     guild.GuildID,
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
	log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID}).Info("new member")

	return member
}

// WinRace is called when the race member won a race.
func (m *Member) WinRace() {
	m.RacesWon++
	writeMember(m)
}

// PlaceInRace is called when the race member places (comes in 2nd) in a race.
func (m *Member) PlaceInRace() {
	m.RacesPlaced++
	writeMember(m)

}

// ShowInRace is called when the race member shows (comes in 3rd) in a race.
func (m *Member) ShowInRace() {
	m.RacesShowed++
	writeMember(m)
}

// LoseRace is called when the race member fails to win, place or show in a race.
func (m *Member) LoseRace() {
	m.RacesLost++
	writeMember(m)
}

// PlaceBet is used to place a bet on a member of a race.
func (m *Member) PlaceBet() {
	// TODO: need to decrement the bank account for the bet.
	m.BetsMade++
	writeMember(m)
}

// WinBet is used when a member wins a bet on a race.
func (m *Member) WinBet() {
	// TODO: need to increment the bank account for the better. Probably need to pass in
	// the bet amount.
	m.BetsWon++
	writeMember(m)
}
