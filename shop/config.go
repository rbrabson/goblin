package shop

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Config represents the configuration for the shop in a guild.
type Config struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID      string             `json:"guild_id" bson:"guild_id"`
	ChannelID    string             `json:"channel_id" bson:"channel_id"`
	MessageID    string             `json:"message_id" bson:"message_id"`
	ModChannelID string             `json:"mod_channel_id" bson:"mod_channel_id"`
}

// GetConfig reads the configuration from the database. If the config does not exist,
// then one is created.
func GetConfig(guildID string) *Config {
	log.Trace("--> shop.GetConfig")
	defer log.Trace("<-- shop.GetConfig")

	config, _ := readConfig(guildID)
	if config == nil {
		config = newConfig(guildID)
	}

	return config
}

// newConfig creates a new configuration for the given guild ID and writes it to the database.
func newConfig(guildID string) *Config {
	log.Trace("--> shop.newConfig")
	defer log.Trace("<-- shop.newConfig")

	config := &Config{
		ID:      primitive.NewObjectID(),
		GuildID: guildID,
	}
	writeConfig(config)

	return config
}

// SetChannel sets the channel to which to publish the shop items
func (c *Config) SetChannel(channelID string) {
	log.Trace("--> shop.Config.SetChannel")
	defer log.Trace("<-- shop.Config.SetChannel")

	if c.ChannelID != channelID {
		c.ChannelID = channelID
		c.MessageID = ""
		writeConfig(c)
		log.WithFields(log.Fields{"guildID": c.GuildID, "channel": channelID}).Debug("set shop channel")
	}
}

// SetModChannel sets the channel to which to publish the shop purchases and expirations.
func (c *Config) SetModChannel(channelID string) {
	log.Trace("--> shop.Config.SetModChannel")
	defer log.Trace("<-- shop.Config.SetModChannel")

	if c.ModChannelID != channelID {
		c.ModChannelID = channelID
		writeConfig(c)
		log.WithFields(log.Fields{"guildID": c.GuildID, "channel": channelID}).Debug("set shop mod channel")
	}
}

// SetMessageID saves the interaction used to publish the shop items.
func (c *Config) SetMessageID(messageID string) {
	log.Trace("--> shop.Config.SetMessageID")
	defer log.Trace("<-- shop.Config.SetMessageID")

	if c.MessageID != messageID {
		c.MessageID = messageID
		writeConfig(c)
		log.WithFields(log.Fields{"guildID": c.GuildID, "messageID": messageID}).Debug("set shop message ID")
	}
}
