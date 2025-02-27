package race

import (
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CONFIG_COLLECTION = "race_configs"
	MEMBER_COLLECTION = "race_members"
)

// readConfig loads the race configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guild *guild.Guild) *Config {
	log.Trace("--> race.readConfig")
	defer log.Trace("<-- race.readConfig")

	filter := bson.M{"guild_id": guild.GuildID}
	var config Config
	err := db.FindOne(CONFIG_COLLECTION, filter, &config)
	if err != nil {
		log.WithFields(log.Fields{"guild": guild.GuildID, "error": err}).Debug("race configuration not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": guild.GuildID}).Debug("read race configuration from the database")

	return &config
}

// writeConfig stores the race configuration in the database.
func writeConfig(config *Config) {
	log.Trace("--> race.writeConfig")
	defer log.Trace("<-- race.writeConfig")

	var filter bson.M
	if config.ID != primitive.NilObjectID {
		filter = bson.M{"_id": config.ID}
	} else {
		filter = bson.M{"guild_id": config.GuildID}
	}
	err := db.UpdateOrInsert(CONFIG_COLLECTION, filter, config)
	if err != nil {
		log.WithFields(log.Fields{"guild": config.GuildID, "error": err}).Error("failed to write the race configuration to the database")
	}
}

// readConfig loads the race member from the database. If it does not exist then
// a `nil` value is returned.
func readMember(guild *guild.Guild, memberID string) *Member {
	log.Trace("--> race.readMember")
	defer log.Trace("<-- race.readMember")

	filter := bson.M{"guild_id": guild.GuildID, "member_id": memberID}
	var member Member
	err := db.FindOne(MEMBER_COLLECTION, filter, &member)
	if err != nil {
		log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID, "error": err}).Debug("race member not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": guild.GuildID, "member": memberID}).Debug("read race member from the database")

	return &member
}

// Write creates or updates the race member in the database
func writeMember(member *Member) {
	log.Trace("--> race.writeMember")
	defer log.Trace("<-- race.writeMember")

	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	db.UpdateOrInsert(MEMBER_COLLECTION, filter, member)
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Debug("write race member to the database")
}
