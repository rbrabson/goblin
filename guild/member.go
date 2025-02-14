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

// GetMember returns a member in the guild (server). If one doesnt' exist, then it is created.
func GetMember(guild *Guild, memberID string) *Member {
	log.Trace("--> guild.GetMember")
	defer log.Trace("<-- guild.GetMember")

	member := guild.Members[memberID]
	if member != nil {
		return member
	}

	member = loadMember(guild, memberID)
	if member != nil {
		member.guildID = guild.ID
		guild.Members[memberID] = member
		return member
	}

	return newMember(guild, memberID)
}

// SetName updates the name of the member and, if it has changed, writes the name to the database.
func (member *Member) SetName(name string) *Member {
	if name != member.Name {
		member.Name = name
		member.write()
	}
	return member
}

// newMember creates a new member in the guild (server).
func newMember(guild *Guild, memberID string) *Member {
	log.Trace("--> guild.newMember")
	defer log.Trace("<-- guild.newMember")

	member := &Member{
		ID:      memberID,
		guildID: guild.ID,
	}
	guild.Members[member.ID] = member
	member.write()
	log.WithFields(log.Fields{"guild": guild.ID, "member": memberID}).Info("created new member")

	return member
}

// loadMember reads the member from the database and returns the value, if it exists, or returns nil if the
// member does not exist in the database
func loadMember(guild *Guild, memberID string) *Member {
	log.Trace("--> guild.loadMember")
	defer log.Trace("<-- guild.loadMember")

	var member Member
	err := db.Read(guild.ID, MEMBER_COLLECTION, memberID, &member)
	if err != nil {
		log.WithFields(log.Fields{"guild": guild.ID, "member": memberID}).Info("member not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": member.guildID, "id": member.ID, "name": member.Name}).Info("load member from the database")
	return &member
}

// write creates or updates the member data in the database being used by the Discord bot.
func (member *Member) write() error {
	log.Trace("--> guild.Member.write")
	defer log.Trace("<-- guild.Member.write")

	db.Write(member.guildID, MEMBER_COLLECTION, member.ID, member)
	log.WithFields(log.Fields{"guild": member.guildID, "id": member.ID, "name": member.Name}).Info("save member to the database")
	return nil
}
