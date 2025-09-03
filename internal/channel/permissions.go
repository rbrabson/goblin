package channel

import (
	"log/slog"

	"github.com/rbrabson/goblin/guild"

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
		slog.Error("error getting channel to mute",
			slog.String("guildID", i.GuildID),
			slog.String("channelID", i.ChannelID),
			slog.Any("error", err),
		)
		return nil
	}
	if channel == nil {
		slog.Error("channel to mute is is nil",
			slog.String("guildID", i.GuildID),
			slog.String("channelID", i.ChannelID),
		)
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
	if c == nil {
		slog.Error("channelMute is nil")
		return
	}
	mute := int64(discordgo.PermissionSendMessages)
	allowFlagsNoSend := c.everyonePermissions.Allow & mute
	denyFlagsNoSend := c.everyonePermissions.Deny ^ mute
	slog.Debug("muting the channel",
		slog.String("guildID", c.i.GuildID),
		slog.String("channelID", c.i.ChannelID),
		slog.Int64("currentAllow", c.everyonePermissions.Allow),
		slog.Int64("currentDeny", c.everyonePermissions.Deny),
		slog.Int64("allow", allowFlagsNoSend),
		slog.Int64("deny", denyFlagsNoSend),
		slog.Int64("mute", mute),
	)
	err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, c.everyonePermissions.Type, allowFlagsNoSend, denyFlagsNoSend)
	if err != nil {
		slog.Warn("failed to mute the channel",
			slog.String("guildID", c.i.GuildID),
			slog.String("channelID", c.i.ChannelID),
			slog.Any("error", err),
		)
	} else {
		slog.Debug("muted the channel",
			slog.String("guildID", c.i.GuildID),
			slog.String("channelID", c.i.ChannelID),
		)
	}
}

// UnmuteChannel resets the permissions for `@everyone` to what they were before the channel was muted.
func (c *Mute) UnmuteChannel() {
	if c == nil {
		slog.Error("channelMute is nil")
		return
	}
	if c.everyoneID != "" {
		err := c.s.ChannelPermissionSet(c.i.ChannelID, c.everyoneID, c.everyonePermissions.Type, c.everyonePermissions.Allow, c.everyonePermissions.Deny)
		if err != nil {
			slog.Warn("failed to unmute the channel",
				slog.String("guildID", c.i.GuildID),
				slog.String("channelID", c.i.ChannelID),
				slog.Any("error", err),
			)
			return
		}
		slog.Debug("reset the channel permissions",
			slog.String("guildID", c.i.GuildID),
			slog.String("channelID", c.i.ChannelID),
		)
	} else {
		slog.Error("permissions unknown; not possible to un-mute the channel",
			slog.String("guildID", c.i.GuildID),
			slog.String("channelID", c.i.ChannelID),
		)
	}
}
