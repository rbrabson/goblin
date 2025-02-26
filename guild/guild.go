package guild

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// A Guild is a Discord server that a user may be a member of
type Guild struct {
	GuildID string
}

// GetGuild returns a guild (server) with the given guild ID. If one doesn't exist, then it is created.
func GetGuild(guildID string) *Guild {
	log.Trace("--> guild.GetGuild")
	defer log.Trace("<-- guild.GetGuild")

	guild := &Guild{
		GuildID: guildID,
	}
	log.WithFields(log.Fields{"guild": guild.GuildID}).Debug("get guild")
	return guild
}

// GetMember returns a member in the guild (server). If one doesnt' exist, then one is created with a blank name.
func (g *Guild) GetMember(memberID string) *Member {
	log.Trace("--> guild.GetMember")
	defer log.Trace("<-- guild.GetMember")

	return getMember(g, memberID)
}

// String returns a string representation of the Guild.
func (g *Guild) String() string {
	return fmt.Sprintf("Guild{GuildID=%s}", g.GuildID)
}
