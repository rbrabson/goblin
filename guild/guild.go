package guild

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DEFAULT_ADMIN_ROLES = []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"}
)

// Guild is the configuration for a guild (guild).
type Guild struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	AdminRoles []string           `json:"admin_roles" bson:"admin_roles"`
}

// GetGuild returns the guild configuration for a given guild (guild).
func GetGuild(guildID string) *Guild {
	guild := readGuild(guildID)
	if guild == nil {
		guild = newGuild(guildID)
	}

	log.Tracef("Guild: %v", guild)
	return guild
}

// newGuild creates a new guild configuration for a given guild (guild).
func newGuild(guildID string) *Guild {
	guild := &Guild{
		GuildID: guildID,
	}
	guild.AdminRoles = make([]string, len(DEFAULT_ADMIN_ROLES))
	copy(guild.AdminRoles, DEFAULT_ADMIN_ROLES)
	writeGuild(guild)

	return guild
}

// AddAdminRole adds a role to the list of admin roles for the guild.
func (guild *Guild) AddAdminRole(roleName string) {
	log.Trace("--> guild.Guild.AddAdminRole")
	defer log.Trace("<-- guild.Guild.AddAdminRole")

	if slices.Contains(guild.AdminRoles, roleName) {
		log.WithFields(log.Fields{"roleName": roleName, "adminRoles": guild.AdminRoles}).Warn("role already exists")
		return
	}

	guild.AdminRoles = append(guild.AdminRoles, roleName)
	writeGuild(guild)
	log.WithFields(log.Fields{"roleName": roleName, "adminRoles": guild.AdminRoles}).Info("added admin role")
}

// RemoveAdminRole removes a role from the list of admin roles for the guild.
func (guild *Guild) RemoveAdminRole(roleName string) {
	log.Trace("--> guild.Guild.RemoveAdminRole")
	defer log.Trace("<-- guild.Guild.RemoveAdminRole")

	for i, role := range guild.AdminRoles {
		if role == roleName {
			guild.AdminRoles = append(guild.AdminRoles[:i], guild.AdminRoles[i+1:]...)
			writeGuild(guild)
			log.WithFields(log.Fields{"roleName": roleName, "adminRoles": guild.AdminRoles}).Info("removed admin role")
			return
		}
	}
	log.WithFields(log.Fields{"roleName": roleName, "adminRoles": guild.AdminRoles}).Warn("role not found")
}

// GetAdminRoles returns the list of admin roles for the guild.
func (guild *Guild) GetAdminRoles() []string {
	log.Trace("--> guild.Guild.GetAdminRoles")
	defer log.Trace("<-- guild.Guild.GetAdminRoles")

	return guild.AdminRoles
}

// String returns a string representation of the guild.
func (guild *Guild) String() string {
	return fmt.Sprintf("Guild{guildID = %s, adminRoles = %v}", guild.GuildID, guild.AdminRoles)
}
