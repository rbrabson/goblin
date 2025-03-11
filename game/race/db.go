package race

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RACE_CONFIG_COLLECTION = "race_configs"
	RACE_MEMBER_COLLECTION = "race_members"
	RACER_COLLECTION       = "race_avatars"
)

// readConfig loads the race configuration from the database. If it does not exist then
// a `nil` value is returned.
func readConfig(guildID string) *Config {
	log.Trace("--> race.readConfig")
	defer log.Trace("<-- race.readConfig")

	filter := bson.M{"guild_id": guildID}
	var config Config
	err := db.FindOne(RACE_CONFIG_COLLECTION, filter, &config)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "error": err}).Debug("race configuration not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": guildID}).Debug("read race configuration from the database")

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
	err := db.UpdateOrInsert(RACE_CONFIG_COLLECTION, filter, config)
	if err != nil {
		log.WithFields(log.Fields{"guild": config.GuildID, "error": err}).Error("failed to write the race configuration to the database")
	}
}

// readConfig loads the race member from the database. If it does not exist then
// a `nil` value is returned.
func readRaceMember(guildID string, memberID string) *RaceMember {
	log.Trace("--> race.readRaceMember")
	defer log.Trace("<-- race.readRaceMember")

	filter := bson.D{{Key: "guild_id", Value: guildID}, {Key: "member_id", Value: memberID}}
	var member RaceMember
	err := db.FindOne(RACE_MEMBER_COLLECTION, filter, &member)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID, "member": memberID, "error": err}).Debug("race member not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": guildID, "member": memberID}).Debug("read race member from the database")

	return &member
}

// Write creates or updates the race member in the database
func writeRaceMember(member *RaceMember) {
	log.Trace("--> race.writeRaceMember")
	defer log.Trace("<-- race.writeRaceMember")

	var filter bson.M
	if member.ID != primitive.NilObjectID {
		filter = bson.M{"_id": member.ID}
	} else {
		filter = bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}
	}
	db.UpdateOrInsert(RACE_MEMBER_COLLECTION, filter, member)
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID}).Debug("write race member to the database")
}

// readAllRaces loads the racers that may be used in racers that match the filter criteria.
func readAllRacers(filter bson.D) ([]*RaceAvatar, error) {
	log.Trace("--> race.readAllRacers")
	defer log.Trace("<-- race.readAllRacers")

	var racers []*RaceAvatar
	sort := bson.D{{Key: "crew_size", Value: 1}}
	err := db.FindMany(RACER_COLLECTION, filter, &racers, sort, 0)
	if err != nil || len(racers) == 0 {
		log.WithError(err).Warn("unable to read racers")
		if err != nil {
			return nil, err
		}
		return nil, ErrNoRacersFound
	}

	log.WithField("racers", racers).Info("load racers")

	return racers, nil
}

// writeRacer creates or updates the target in the database.
func writeRacer(racer *RaceAvatar) {
	log.Trace("--> race.Target.writeRacer")
	defer log.Trace("<-- race.Target.writeRacer")

	var filter bson.D
	if racer.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: racer.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: racer.GuildID}, {Key: "theme", Value: racer.Theme}, {Key: "emoji", Value: racer.Emoji}, {Key: "movement_speed", Value: racer.MovementSpeed}}
	}

	db.UpdateOrInsert(RACER_COLLECTION, filter, racer)
	log.WithFields(log.Fields{"guild": racer.GuildID, "theme": racer.Theme}).Debug("create or update race avatar")
}
