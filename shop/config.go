package shop

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Config represents the configuration for the shop in a guild.
type Config struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID        string             `json:"guild_id" bson:"guild_id"`
	ChannelID      string             `json:"channel_id" bson:"channel_id"`
	MessageID      string             `json:"message_id" bson:"message_id"`
	ModChannelID   string             `json:"mod_channel_id" bson:"mod_channel_id"`
	NotificationID string             `json:"notification_id" bson:"notification_id"`
}

// GetConfig reads the configuration from the database. If the config does not exist,
// then one is created.
func GetConfig(guildID string) *Config {
	config, _ := readConfig(guildID)
	if config == nil {
		config = newConfig(guildID)
	}

	return config
}

// newConfig creates a new configuration for the given guild ID and writes it to the database.
func newConfig(guildID string) *Config {
	config := &Config{
		ID:      primitive.NewObjectID(),
		GuildID: guildID,
	}
	writeConfig(config)

	return config
}

// SetChannel sets the channel to which to publish the shop items
func (c *Config) SetChannel(channelID string) {
	if c.ChannelID != channelID {
		c.ChannelID = channelID
		c.MessageID = ""
		writeConfig(c)
		sslog.Debug("set shop channel",
			slog.String("guildID", c.GuildID),
			slog.String("channel", channelID),
		)
	}
}

// SetModChannel sets the channel to which to publish the shop purchases and expirations.
func (c *Config) SetModChannel(channelID string) {
	if c.ModChannelID != channelID {
		c.ModChannelID = channelID
		writeConfig(c)
		sslog.Debug("set shop mod channel",
			slog.String("guildID", c.GuildID),
			slog.String("channel", channelID),
		)
	}
}

// SetModChannel sets the channel to which to notify a user (e.g., ModMail) about an action to take to complete a member's purchase.
func (c *Config) SetNotificationID(id string) {
	if c.NotificationID != id {
		c.NotificationID = id
		writeConfig(c)
		sslog.Debug("set shop notification ID",
			slog.String("guildID", c.GuildID),
			slog.String("member", id),
		)
	}
}

// SetMessageID saves the interaction used to publish the shop items.
func (c *Config) SetMessageID(messageID string) {
	if c.MessageID != messageID {
		c.MessageID = messageID
		writeConfig(c)
		sslog.Debug("set shop message ID",
			slog.String("guildID", c.GuildID),
			slog.String("messageID", messageID),
		)
	}
}

// String returns a string representation of the config.
func (c *Config) String() string {
	return "Config{" +
		"ID: " + c.ID.Hex() +
		", GuildID: " + c.GuildID +
		", ChannelID: " + c.ChannelID +
		", MessageID: " + c.MessageID +
		", ModChannelID: " + c.ModChannelID +
		", NotificationID: " + c.NotificationID +
		"}"
}
