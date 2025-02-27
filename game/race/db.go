package race

import (
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CONFIG_COLLECTION = "race_configs"
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
