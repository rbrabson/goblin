package guild

import (
	"github.com/rbrabson/dgame/database"
	log "github.com/sirupsen/logrus"
)

var (
	guilds = make(map[string]*Guild)
	db     database.Client
)

// A Guild is a Discord server that a user may be a member of
type Guild struct {
	ID      string
	Members map[string]*Member
}

// Init initializes the guild package with the database connection to be used for
// retrieving and storing guilds and guild members.
func Init(dbase database.Client) {
	db = dbase
}

// GetGuild returns a guild (server) with the given guild ID. If one is not already
// known, then it is created.
func GetGuild(guildID string) *Guild {
	log.Trace("--> guild.GetGuild")
	defer log.Trace("<-- guild.GetGuild")

	guild := guilds[guildID]
	if guild == nil {
		guild = new(guildID)
	}

	return guild
}

// new creates a new guild (server) with the given guild ID.
func new(guildID string) *Guild {
	log.Trace("--> guild.new")
	defer log.Trace("<-- guild.new")

	guild := &Guild{
		ID:      guildID,
		Members: make(map[string]*Member),
	}
	guilds[guild.ID] = guild
	log.WithFields(log.Fields{"guild": guild.ID}).Info("create new guild")

	return guild
}
