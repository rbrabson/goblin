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
	Greenhorn CriminalLevel = 0
	Renegade  CriminalLevel = 1
	Veteran   CriminalLevel = 10
	Commander CriminalLevel = 25
	WarChief  CriminalLevel = 50
	Legend    CriminalLevel = 75
	Immortal  CriminalLevel = 100
)

type MemberStatus string

const (
	Escaped     MemberStatus = "Escaped"
	Free        MemberStatus = "Free"
	Dead        MemberStatus = "Dead"
	Apprehended MemberStatus = "Apprehended"
	OOB         MemberStatus = "Out on Bail"
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
		CriminalLevel: Greenhorn,
		Status:        Free,
	}
	writeMember(member)
	slog.Debug("create heist member",
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
	member.Status = Apprehended
	member.JailCounter++
	member.TotalJail++
	member.Spree = 0
	member.CriminalLevel++
	member.BailCost = bailCost

	writeMember(member)
	slog.Debug("heist member apprehended",
		slog.Int("bail", member.BailCost),
		slog.Any("criminalLevel", member.CriminalLevel),
		slog.Time("deathTimer", member.DeathTimer),
		slog.String("guild", member.GuildID),
		slog.Int("jailCounter", member.JailCounter),
		slog.String("member", member.MemberID),
		slog.Duration("sentence", member.Sentence),
		slog.Int("spree", member.Spree),
		slog.Any("status", member.Status),
		slog.Time("timer", member.JailTimer),
		slog.Int("totalDeaths", member.Deaths),
		slog.Int("totalJail", member.TotalJail),
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
	member.Status = Dead

	writeMember(member)

	slog.Debug("heist member died",
		slog.Int("bail", member.BailCost),
		slog.Any("criminalLevel", member.CriminalLevel),
		slog.Time("deathTimer", member.DeathTimer),
		slog.String("guild", member.GuildID),
		slog.Int("jailCounter", member.JailCounter),
		slog.String("member", member.MemberID),
		slog.Duration("sentence", member.Sentence),
		slog.Int("spree", member.Spree),
		slog.Any("status", member.Status),
		slog.Time("timer", member.JailTimer),
		slog.Int("totalDeaths", member.Deaths),
		slog.Int("totalJail", member.TotalJail),
	)
}

// Escaped updates the member when they successfully escape during a heist.
func (member *HeistMember) Escaped() {
	member.Spree++
	writeMember(member)

	slog.Debug("heist member escaped",
		slog.Int("bail", member.BailCost),
		slog.Any("criminalLevel", member.CriminalLevel),
		slog.Time("deathTimer", member.DeathTimer),
		slog.String("guild", member.GuildID),
		slog.Int("jailCounter", member.JailCounter),
		slog.String("member", member.MemberID),
		slog.Duration("sentence", member.Sentence),
		slog.Int("spree", member.Spree),
		slog.Any("status", member.Status),
		slog.Time("timer", member.JailTimer),
		slog.Int("totalDeaths", member.Deaths),
		slog.Int("totalJail", member.TotalJail),
	)
}

// BailedOut updates the status of a heist member to "Out on Bail" and saves the changes to the database.
func (member *HeistMember) BailedOut() {
	member.Status = OOB
	writeMember(member)
}

// UpdateStatus updates the status of the member based on the current time. If the member is in jail
// or dead, then the status is updated to Free when the time has expired.
func (member *HeistMember) UpdateStatus() {
	switch member.Status {
	case Apprehended:
		if member.RemainingJailTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	case OOB:
		if member.RemainingJailTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	case Dead:
		if member.RemainingDeathTime() <= 0 {
			member.ClearJailAndDeathStatus()
		}
	}
}

// ClearJailAndDeathStatus is called when a player is released from jail or rises from the grave.
func (member *HeistMember) ClearJailAndDeathStatus() {
	if member.Status == Dead {
		slog.Debug("heist member risen from the grave",
			slog.Int("bail", member.BailCost),
			slog.Any("criminalLevel", member.CriminalLevel),
			slog.Time("deathTimer", member.DeathTimer),
			slog.String("guild", member.GuildID),
			slog.Int("jailCounter", member.JailCounter),
			slog.String("member", member.MemberID),
			slog.Duration("sentence", member.Sentence),
			slog.Int("spree", member.Spree),
			slog.Any("status", member.Status),
			slog.Time("timer", member.JailTimer),
			slog.Int("totalDeaths", member.Deaths),
			slog.Int("totalJail", member.TotalJail),
		)
	} else if member.Status == Apprehended || member.Status == OOB {
		slog.Debug("heist member released from jail",
			slog.Int("bail", member.BailCost),
			slog.Any("criminalLevel", member.CriminalLevel),
			slog.Time("deathTimer", member.DeathTimer),
			slog.String("guild", member.GuildID),
			slog.Int("jailCounter", member.JailCounter),
			slog.String("member", member.MemberID),
			slog.Duration("sentence", member.Sentence),
			slog.Int("spree", member.Spree),
			slog.Any("status", member.Status),
			slog.Time("timer", member.JailTimer),
			slog.Int("totalDeaths", member.Deaths),
			slog.Int("totalJail", member.TotalJail),
		)
	}

	member.BailCost = 0
	member.DeathTimer = time.Time{}
	member.JailCounter = 0
	member.JailTimer = time.Time{}
	member.Sentence = 0
	member.Spree = 0
	member.Status = Free

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
	case level >= Immortal:
		return "Immortal"
	case level >= Legend:
		return "Legend"
	case level >= WarChief:
		return "WarChief"
	case level >= Commander:
		return "Commander"
	case level >= Veteran:
		return "Veteran"
	case level >= Renegade:
		return "Renegade"
	case level >= Greenhorn:
		return "Greenhorn"
	default:
		return "Unknown"
	}
}

// String returns a string representation of the status of the member of a heist
func (status MemberStatus) String() string {
	return string(status)
}
