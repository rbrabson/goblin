package heist

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rbrabson/goblin/guild"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type MemberStatus string

const (
	ESCAPED     = "Escaped"
	FREE        = "Free"
	DEAD        = "Dead"
	APPREHENDED = "Apprehended"
	OOB         = "Out on Bail"
	UNKNOWN     = "Unknown"
)

// HeistMember is the status of a member who has participated in at least one heist
type HeistMember struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID       string             `json:"guild_id" bson:"guild_id"`
	MemberID      string             `json:"member_id" bson:"member_id"`
	BailCost      int                `json:"bail_cost" bson:"bail_cost"`
	CriminalLevel CriminalLevel      `json:"criminal_level" bson:"criminal_level"`
	Deaths        int                `json:"deaths" bson:"deaths"`
	DeathTimer    time.Time          `json:"death_timer" bson:"death_timer"`
	JailCounter   int                `json:"jail_counter" bson:"jail_counter"`
	JailTimer     time.Time          `json:"jail_timer" bson:"jail_timer"`
	Sentence      time.Duration      `json:"sentence" bson:"sentence"`
	Spree         int                `json:"spree" bson:"spree"`
	Status        MemberStatus       `json:"status" bson:"status"`
	TotalJail     int                `json:"total_jail" bson:"total_jail"`
	heist         *Heist             `json:"-" bson:"-"`
	guildMember   *guild.Member      `json:"-" bson:"-"`
}

// getHeistMember gets a member for heists. If the member does not exist, then nil is returned.
func getHeistMember(guildID string, memberID string) *HeistMember {
	member := readMember(guildID, memberID)
	if member == nil {
		member = newHeistMember(guildID, memberID)
	}
	member.guildMember = guild.GetMember(guildID, memberID)

	return member
}

// newHeistMember creates a new member for heists. It is called when guild member
// first plans or joins a heist.
func newHeistMember(guildID string, memberID string) *HeistMember {
	member := &HeistMember{
		GuildID:       guildID,
		MemberID:      memberID,
		CriminalLevel: GREENHORN,
		Status:        FREE,
	}
	writeMember(member)
	sslog.Debug("create heist member",
		slog.String("guildID", member.GuildID),
		slog.String("memberID", member.MemberID),
	)

	return member
}

// Apprehended updates the member when they are caught during a heist.
func (member *HeistMember) Apprehended() {
	bailCost := member.heist.config.BailBase
	if member.Status == OOB {
		bailCost *= 3
	}
	member.Sentence = time.Duration(int64(member.heist.config.SentenceBase) * int64(member.JailCounter+1))
	member.JailTimer = time.Now().Add(member.Sentence)
	member.Status = APPREHENDED
	member.JailCounter++
	member.TotalJail++
	member.Spree = 0
	member.CriminalLevel++
	member.BailCost = bailCost

	writeMember(member)
	sslog.Debug("heist member apprehended",
		"bail", member.BailCost,
		"criminalLevel", member.CriminalLevel,
		"deathTimer", member.DeathTimer,
		"guild", member.GuildID,
		"jailCounter", member.JailCounter,
		"member", member.MemberID,
		"sentence", member.Sentence,
		"spree", member.Spree,
		"status", member.Status,
		"timer", member.JailTimer,
		"totalDeaths", member.Deaths,
		"totalJail", member.TotalJail,
	)
}

// Died updates the member when they die during a heist.
func (member *HeistMember) Died() {
	member.BailCost = 0
	member.CriminalLevel = 0
	member.Deaths++
	member.DeathTimer = time.Now().Add(member.heist.config.DeathTimer)
	member.JailCounter = 0
	member.JailTimer = time.Time{}
	member.Sentence = 0
	member.Spree = 0
	member.Status = DEAD

	writeMember(member)

	sslog.Debug("heist member died",
		"bail", member.BailCost,
		"criminalLevel", member.CriminalLevel,
		"deathTimer", member.DeathTimer,
		"guild", member.GuildID,
		"jailCounter", member.JailCounter,
		"member", member.MemberID,
		"sentence", member.Sentence,
		"spree", member.Spree,
		"status", member.Status,
		"timer", member.JailTimer,
		"totalDeaths", member.Deaths,
		"totalJail", member.TotalJail,
	)
}

// Escaped updates the member when they successfully escape during a heist.
func (member *HeistMember) Escaped() {
	member.Spree++
	writeMember(member)

	sslog.Debug("heist member escaped",
		"bail", member.BailCost,
		"criminalLevel", member.CriminalLevel,
		"deathTimer", member.DeathTimer,
		"guild", member.GuildID,
		"jailCounter", member.JailCounter,
		"member", member.MemberID,
		"sentence", member.Sentence,
		"spree", member.Spree,
		"status", member.Status,
		"timer", member.JailTimer,
		"totalDeaths", member.Deaths,
		"totalJail", member.TotalJail,
	)
}

// UpdateStatus updates the status of the member based on the current time. If the member is in jail
// or dead, then the status is updated to FREE when the time has expired.
func (member *HeistMember) UpdateStatus() {
	switch member.Status {
	case APPREHENDED:
		if member.RemainingJailTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	case OOB:
		if member.RemainingJailTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	case DEAD:
		if member.RemainingDeathTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	}
}

// ClearJailAndDeathStatus is called when a player is released from jail or rises from the grave.
func (member *HeistMember) ClearJailAndDeathStatus() {
	if member.Status == DEAD {
		sslog.Debug("heist member risen from the grave",
			"guildID", member.GuildID,
			"memberID", member.MemberID,
			"bail", member.BailCost,
			"deathTimer", member.DeathTimer,
			"jailCounter", member.JailCounter,
			"jailTimer", member.JailTimer,
			"sentence", member.Sentence,
			"spree", member.Spree,
			"status", member.Status,
		)
	} else if member.Status == APPREHENDED || member.Status == OOB {
		sslog.Debug("heist member released from jail",
			"guildID", member.GuildID,
			"memberID", member.MemberID,
			"bail", member.BailCost,
			"deathTimer", member.DeathTimer,
			"jailCounter", member.JailCounter,
			"jailTimer", member.JailTimer,
			"sentence", member.Sentence,
			"spree", member.Spree,
			"status", member.Status,
		)
	}

	member.BailCost = 0
	member.DeathTimer = time.Time{}
	member.JailCounter = 0
	member.JailTimer = time.Time{}
	member.Sentence = 0
	member.Spree = 0
	member.Status = FREE

	writeMember(member)
}

// RemainingJailTime returns the amount of time remaining on the player's sentence has been served
func (member *HeistMember) RemainingJailTime() time.Duration {
	if member.JailTimer.Before(time.Now()) {
		return 0
	}
	return time.Until(member.JailTimer)
}

// RemainingDeathTime returns the amount of time before the member can be resurected.
func (member *HeistMember) RemainingDeathTime() time.Duration {
	if member.DeathTimer.Before(time.Now()) {
		return 0
	}
	return time.Until(member.DeathTimer)
}

// String returns a string representation of the HeistMember.
func (member *HeistMember) String() string {
	return fmt.Sprintf("HeistMember{ID=%s, GuildID=%s, MemberID=%s, BailCost=%d, CriminalLevel=%s, Deaths=%d, DeathTimer=%s, JailCounter=%d, JailTimer=%s, Sentence=%s, Spree=%d, Status=%s, TotalJail=%d}",
		member.ID.Hex(),
		member.GuildID,
		member.MemberID,
		member.BailCost,
		member.CriminalLevel,
		member.Deaths,
		member.DeathTimer,
		member.JailCounter,
		member.JailTimer,
		member.Sentence,
		member.Spree,
		member.Status,
		member.TotalJail,
	)
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
	return string(status)
}
