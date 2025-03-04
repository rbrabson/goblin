package channel

import (
	"github.com/rbrabson/goblin/guild"
	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

// Mute is used for muting and unmuting a channel on a server
type Mute struct {
	channel             *discordgo.Channel
	everyoneID          string
	everyonePermissions discordgo.PermissionOverwrite
	s                   *discordgo.Session
	i                   *discordgo.InteractionCreate
}

// NewChannelMute creates a channelMute for the given session and interaction.
func NewChannelMute(s *discordgo.Session, i *discordgo.InteractionCreate) *Mute {
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Error("Error getting channel, error:", err)
		return nil
	}
	if channel == nil {
		log.Error("Channel is nil")
		return nil
	}

	c := Mute{
		s:       s,
		i:       i,
		channel: channel,
	}

	guildRoles := guild.GetGuildRoles(s, channel.GuildID)
	for _, guildlRole := range guildRoles {
		if guildlRole.Name == "@everyone" {
			c.everyoneID = guildlRole.ID
		}
	}

	for _, p := range channel.PermissionOverwrites {
		if p.ID == c.everyoneID {
			c.everyonePermissions = *p
			break
		}
	}

	return &c
}

// MuteChannel sets the channel so that `@everyone`	 can't send messages to the channel.
func (c *Mute) MuteChannel() {
	err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)
	if err != nil {
		log.Warning("Failed to mute the channel, error:", err)
	}
}

// UnmuteChannel resets the permissions for `@everyone` to what they were before the channel was muted.
func (c *Mute) UnmuteChannel() {
	if c.everyonePermissions.ID != "" {
		allow := int64(discordgo.PermissionSendMessages)
		err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, c.everyonePermissions.Type, allow, c.everyonePermissions.Deny)
		if err != nil {
			log.Warning("Failed to unmute the channel, error:", err)
			return
		}
		log.WithFields(log.Fields{"channelID": c.i.ChannelID}).Info("reset the mute permissions for the channel")
	} else {
		log.WithFields(log.Fields{"channelID": c.i.ChannelID}).Error("permissions unknown; not possible to un-mute the channel")
	}
}
