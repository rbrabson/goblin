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
		log.WithFields(log.Fields{"guildID": i.GuildID, "error": err}).Error("Error getting channel, error")
		return nil
	}
	if channel == nil {
		log.Error("channel is nil")
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
	mute := int64(discordgo.PermissionSendMessages)
	allowFlagsNoSend := c.everyonePermissions.Allow & mute
	denyFlagsNoSend := c.everyonePermissions.Deny ^ mute
	log.WithFields(log.Fields{"guildID": c.i.GuildID, "currentAllow": c.everyonePermissions.Allow, "currentDeny": c.everyonePermissions.Deny, "allow": allowFlagsNoSend, "deny": denyFlagsNoSend, "mute": mute}).Debug("muting the channel")
	err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, c.everyonePermissions.Type, allowFlagsNoSend, denyFlagsNoSend)
	if err != nil {
		log.WithFields(log.Fields{"guildID": c.i.GuildID, "channelID": c.i.ChannelID, "error": err}).Warning("failed to mute the channel")
	} else {
		log.WithFields(log.Fields{"guildID": c.i.GuildID, "channelID": c.i.ChannelID}).Debug("muted the channel")
	}
}

// UnmuteChannel resets the permissions for `@everyone` to what they were before the channel was muted.
func (c *Mute) UnmuteChannel() {
	if c.everyonePermissions.ID != "" {
		err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, c.everyonePermissions.Type, c.everyonePermissions.Allow, c.everyonePermissions.Deny)
		if err != nil {
			log.WithFields(log.Fields{"guildID": c.i.GuildID, "channelID": c.i.ChannelID, "error": err}).Warning("failed to unmute the channel")
			return
		}
		log.WithFields(log.Fields{"guildID": c.i.GuildID, "channelID": c.i.ChannelID}).Debug("reset the channel permissions")
	} else {
		log.WithFields(log.Fields{"guildID": c.i.GuildID, "channelID": c.i.ChannelID}).Error("permissions unknown; not possible to un-mute the channel")
	}
}
