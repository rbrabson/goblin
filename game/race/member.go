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
func getMember(guild *guild.Guild, memberID string) (*Member, error) {
	log.Trace("--> race.getMember")
	defer log.Trace("<-- race.getMember")

	// TODO: readMember
	return nil, nil
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

	// TODO: writeMember
	log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID}).Info("new member")

	return member
}

func (m *Member) WinRace() {
	m.RacesWon++
	// TODO: writeMember
}

func (m *Member) PlaceInRace() {
	m.RacesPlaced++
	// TODO: writeMember

}

func (m *Member) ShowInRace() {
	m.RacesShowed++
	// TODO: writeMember
}

func (m *Member) LoseRace() {
	m.RacesLost++
	// TODO: writeMember
}

func (m *Member) PlaceBet() {
	m.BetsMade++
	// TODO: writeMember
}

func (m *Member) WinBet() {
	m.BetsWon++
	// TODO: writeMember
}
