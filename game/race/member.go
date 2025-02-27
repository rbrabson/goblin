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
	NumRaces    int                `json:"num_races" bson:"num_races"`
	Win         int                `json:"win" bson:"win"`
	Place       int                `json:"place" bson:"place"`
	Show        int                `json:"show" bson:"show"`
	Loses       int                `json:"loses" bson:"loses"`
	Earnings    int                `json:"earnings" bson:"earnings"`
	BetsPlaced  int                `json:"bets_placed" bson:"bets_placed"`
	BetsWon     int                `json:"bets_won" bson:"bets_won"`
	BetEarnings int                `json:"bet_earnings" bson:"bet_earnings"`
	TotalBets   int                `json:"total_bets" bson:"total_bets"`
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
		NumRaces:    0,
		Win:         0,
		Place:       0,
		Show:        0,
		Loses:       0,
		Earnings:    0,
		BetsPlaced:  0,
		BetsWon:     0,
		BetEarnings: 0,
		TotalBets:   0,
	}

	// TODO: writeMember
	log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID}).Info("new member")

	return member
}

func (m *Member) WinRace() {
}

func (m *Member) PlaceInRace() {

}

func (m *Member) ShowInRace() {

}

func (m *Member) LoseRace() {
}

func (m *Member) PlaceBet() {
}

func (m *Member) WinBet() {
}
