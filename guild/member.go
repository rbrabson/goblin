package guild

import (
	"cmp"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member is a member of a given guild
type Member struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID  string             `json:"guild_id" bson:"guild_id"`
	MemberID string             `json:"member_id" bson:"member_id"`
	Name     string             `json:"name" bson:"name"`
}

// GetMember returns a member in the guild (server). If one doesnt' exist, then one is created with a blank name.
func GetMember(guildID string, memberID string) *Member {
	member := readMember(guildID, memberID)

	if member == nil {
		member = newMember(guildID, memberID)
	}

	return member
}

// SetName updates the name of the member as known on this guild (server).
func (member *Member) SetName(userName string, displayName string) *Member {
	name := cmp.Or(displayName, userName)
	name = strings.Trim(name, "# ")

	if member.Name != name {
		member.Name = name
		writeMember(member)
		log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "name": member.Name}).Debug("set member name")
	}

	return member
}

// newMember creates a new member in the guild (server).
func newMember(guildID string, memberID string) *Member {
	member := &Member{
		MemberID: memberID,
		GuildID:  guildID,
	}
	writeMember(member)
	log.WithFields(log.Fields{"guild": guildID, "member": memberID}).Info("created new member")

	return member
}

// String returns a string representation of the Member.
func (member *Member) String() string {
	return fmt.Sprintf("Member{ID=%s, GuildID=%s, MemberID=%s, Name=%s}",
		member.ID.Hex(),
		member.GuildID,
		member.MemberID,
		member.Name)
}
