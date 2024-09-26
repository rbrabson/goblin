package guild

import (
	log "github.com/sirupsen/logrus"
)

const (
	MEMBER_COLLECTION = "members"
)

// Member is a member of a given guild
type Member struct {
	ID      string `json:"_id" bson:"_id"`
	Name    string `json:"name" bson:"name"`
	guildID string `json:"-" bson:"-"`
}

// NewMember creates a new member in the guild (server).
func NewMember(guild Guild, memberID string, memberName string) *Member {
	log.Trace("--> guild.NewMember")
	defer log.Trace("<-- guild.NewMember")

	member := &Member{
		ID:      memberID,
		Name:    memberName,
		guildID: guild.ID,
	}
	guild.Members[member.ID] = member
	member.Write()
	log.WithFields(log.Fields{"guild": guild.ID, "member": memberID}).Info("created new member")

	return member
}

// GetMember returns a member in the guild (server). If one doesnt' exist, then it is created.
func GetMember(guild Guild, memberID string, memberName string) *Member {
	log.Trace("--> guild.GetMember")
	defer log.Trace("<-- guild.GetMember")

	member := guild.Members[memberID]
	if member == nil {
		member = NewMember(guild, memberID, memberName)
	}

	return member
}

// Write creates or updates the member data in the database being used by the Discord bot.
func (member *Member) Write() error {
	log.Trace("--> guild.Member.write")
	defer log.Trace("<-- guild.Member.Write")

	db.Save(member.guildID, MEMBER_COLLECTION, member.ID, member)
	log.WithFields(log.Fields{"guild": member.guildID, "id": member.ID, "name": member.Name}).Info("save member to the database")
	return nil
}
