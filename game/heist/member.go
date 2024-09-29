package heist

import (
	"encoding/json"
	"time"

	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	MEMBER = "heist_member"
)

type CriminalLevel int

const (
	GREENHORN CriminalLevel = 0
	RENEGADE  CriminalLevel = 1
	VETERAN   CriminalLevel = 10
	COMMANDER CriminalLevel = 25
	WAR_CHIEF CriminalLevel = 50
	LEGEND    CriminalLevel = 75
	IMMORTAL  CriminalLevel = 100
)

type MemberStatus int

const (
	FREE MemberStatus = iota
	DEAD
	APPREHENDED
	OOB
)

var (
	members = make(map[string]map[string]*Member)
)

// Member is the status of a member who has participated in at least one heist
type Member struct {
	ID            string        `json:"_id" bson:"_id"`
	BailCost      int           `json:"bail_cost" bson:"bail_cost"`
	CriminalLevel CriminalLevel `json:"criminal_level" bson:"criminal_level"`
	DeathTimer    time.Time     `json:"death_timer" bson:"death_timer"`
	Deaths        int           `json:"deaths" bson:"deaths"`
	JailCounter   int           `json:"jail_counter" bson:"jail_counter"`
	Sentence      time.Duration `json:"sentence" bson:"sentence"`
	Spree         int           `json:"spree" bson:"spree"`
	Status        MemberStatus  `json:"status" bson:"status"`
	JailTimer     time.Time     `json:"time_served" bson:"time_served"`
	TotalJail     int           `json:"total_jail" bson:"total_jail"`
	guildID       string        `json:"-" bson:"-"`
}

// NewMember creates a new member for heists. It is called when guild member
// first plans or joins a heist.
func NewMember(guild *guild.Guild, guildMember *guild.Member) *Member {
	log.Trace("--> heist.NewMember")
	defer log.Trace("<-- heist.NewMember")

	member := &Member{
		ID:            guildMember.ID,
		CriminalLevel: GREENHORN,
		Status:        FREE,
		guildID:       guild.ID,
	}

	member.Write()
	log.WithFields(log.Fields{"guild": guild.ID, "member": member.ID}).Debug("create heist member")

	return member
}

// GetMember gets a member for heists. If the member does not exist, then
// it is created.
func GetMember(guild *guild.Guild, guildMember *guild.Member) *Member {
	log.Trace("--> heist.GetMember")
	defer log.Trace("<-- heist.GetMember")

	guildMembers := members[guild.ID]
	if guildMembers == nil {
		members[guild.ID] = make(map[string]*Member)
		log.WithField("guild", guild.ID).Debug("creating heist members for the guild")
	}

	member := guildMembers[guildMember.ID]
	if member == nil {
		member := NewMember(guild, guildMember)
		return member
	}

	return member
}

// Release frees the member from jail and removes the jail and death times. It is used when a
// member is released from jail or is revived after dying in a heist.
func (member *Member) Release() {
	log.Trace("--> heist.Member.Release")
	log.Trace("<-- heist.Member.Release")

	member.Status = FREE
	member.DeathTimer = time.Time{}
	member.BailCost = 0
	member.Sentence = 0
	member.JailTimer = time.Time{}
	log.WithFields(log.Fields{"guild": member.guildID, "member": member.ID}).Info("released from jail")

	member.Write()
}

// Clear resets the status of the member and releases them from jail, so they can participate in heists.
func (member *Member) Clear() {
	log.Trace("--> heist.Member.Clear")
	log.Trace("<-- heist.Member.Clear")

	member.CriminalLevel = GREENHORN
	member.JailCounter = 0
	log.WithFields(log.Fields{"guild": member.guildID, "member": member.ID}).Info("status cleared")

	member.Release()
}

// RemainingJailTime returns the amount of time remaining on the player's sentence has been served
func (member *Member) RemainingJailTime() time.Duration {
	log.Trace("--> heist.Member.RemainingJailTime")
	log.Trace("<-- heist.Member.RemainingJailTime")

	if member.JailTimer.IsZero() || member.JailTimer.After(time.Now()) {
		return 0
	}
	return time.Until(member.JailTimer)
}

// RemainingDeathTime returns the amount of time before the member can be resurected.
func (member *Member) RemainingDeathTime() time.Duration {
	log.Trace("--> heist.Member.RemainingDeathTime")
	log.Trace("<-- heist.Member.RemainingDeathTime")

	if member.DeathTimer.IsZero() || member.DeathTimer.After(time.Now()) {
		return 0
	}
	return time.Until(member.DeathTimer)
}

// Write creates or updates the heist member in the database
func (member *Member) Write() {
	log.Trace("--> heist.Member.Write")
	defer log.Trace("<-- heist.Member.Write")

	db.Save(member.guildID, MEMBER, member.ID, member)
	log.WithFields(log.Fields{"guild": member.guildID, "member": member.ID}).Debug("write member to the database")
}

// String returns a string representation of the member of the heist
func (member *Member) String() string {
	out, _ := json.Marshal(member)
	return string(out)
}

// String returns the string representation for a criminal level
func (level CriminalLevel) String() string {
	switch {
	case level >= IMMORTAL:
		return "Immortal"
	case level >= LEGEND:
		return "Legend"
	case level >= WAR_CHIEF:
		return "WarChief"
	case level >= COMMANDER:
		return "Commander"
	case level >= VETERAN:
		return "Veteran"
	case level >= RENEGADE:
		return "Renegade"
	case level >= GREENHORN:
		return "Greenhorn"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the status of the member of a heist
func (status MemberStatus) String() string {
	switch status {
	case FREE:
		return "Free"
	case DEAD:
		return "Dead"
	case APPREHENDED:
		return "Apprehende"
	case OOB:
		return "Out on Bail"
	default:
		return "Unknownn"
	}
}
