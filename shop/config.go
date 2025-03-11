package shop

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Config represents the configuration for the shop in a guild.
type Config struct {
	ID          primitive.ObjectID     `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID     string                 `json:"guild_id" bson:"guild_id"`
	ChannelID   string                 `json:"channel_id" bson:"channel_id"`
	MessageID   string                 `json:"message_id" bson:"message_id"`
	Interaction *discordgo.Interaction `json:"interaction" bson:"interaction"`
}

// GetConfig reads the configuration from the database. If the config does not exist,
// then one is created.
func GetConfig(guildID string) (*Config, error) {
	log.Trace("--> shop.GetConfig")
	defer log.Trace("<-- shop.GetConfig")

	config, _ := readConfig(guildID)
	if config == nil {
		config = newConfig(guildID)
	}

	return config, nil
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

	c.ChannelID = channelID
	c.Interaction = nil
	writeConfig(c)
	log.WithFields(log.Fields{"guildID": c.GuildID, "channel": channelID}).Debug("set shop channel")
}

// SetInteraction saves the interaction used to publish the shop items.
func (c *Config) SetInteraction(interaction *discordgo.Interaction) {
	log.Trace("--> shop.Config.SetInteraction")
	defer log.Trace("<-- shop.Config.SetInteraction")

	c.Interaction = interaction
	writeConfig(c)
	log.WithFields(log.Fields{"guildID": c.GuildID, "interaction": interaction.ID}).Debug("set shop interaction")
}
