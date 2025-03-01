package guild

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	MEMBER_COLLECTION = "guild_members"
)

// readMember reads the member from the database and returns the value, if it exists, or returns nil if the
// member does not exist in the database
func readMember(guildID string, memberID string) *Member {
	log.Trace("--> guild.loadMember")
	defer log.Trace("<-- guild.loadMember")

	filter := bson.M{
		"guild_id":  guildID,
		"member_id": memberID,
	}
	var member Member
	err := db.FindOne(MEMBER_COLLECTION, filter, &member)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID}).Debug("guild member not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "name": member.Name}).Debug("read guild member from the database")
	return &member
}

// writeMember creates or updates the member data in the database.
func writeMember(member *Member) error {
	log.Trace("--> guild.Member.writeMember")
	defer log.Trace("<-- guild.Member.writeMember")

	filter := bson.M{
		"guild_id":  member.GuildID,
		"member_id": member.MemberID,
	}
	err := db.UpdateOrInsert(MEMBER_COLLECTION, filter, member)
	if err != nil {
		log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Error("unable to create or update guild in the database")
		return err
	}

	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "name": member.Name}).Debug("write guild member to the database")
	return nil
}
