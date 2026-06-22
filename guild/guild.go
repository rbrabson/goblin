package guild

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	defaultGuildTheme = "clash"
)

var (
	DefaultAdminRoles = []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"}
	guildLock         = sync.Mutex{}
)

// Guild is the configuration for a guild (server).
type Guild struct {
	ID         bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string        `json:"guild_id" bson:"guild_id"`
	AdminRoles []string      `json:"admin_roles" bson:"admin_roles"`
}

// GetAllGuilds returns all guilds in the database.
func GetAllGuilds() []*Guild {
	guildLock.Lock()
	defer guildLock.Unlock()

	guilds := make([]*Guild, 0)
	err := db.FindMany(GuildCollection, bson.M{}, &guilds, bson.M{}, 0)
	if err != nil {
		slog.Error("failed to get all guilds", "error", err)
		return nil
	}

	slog.Debug("all guilds",
		slog.Int("numGuilds", len(guilds)),
	)
	return guilds
}

// GetGuild returns the guild configuration for a given guild (server).
func GetGuild(guildID string) *Guild {
	guildLock.Lock()
	defer guildLock.Unlock()

	guild := readGuild(guildID)
	if guild == nil {
		guild = readGuildFromFile(guildID)
	}

	return guild
}

// readGuildFromFile creates a new guild configuration for a given guild (server).
func readGuildFromFile(guildID string) *Guild {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	guildTheme := os.Getenv("DISCORD_GUILD_THEME")
	if guildTheme == "" {
		guildTheme = defaultGuildTheme
	}
	configFileName := filepath.Join(configDir, "guild", "config", guildTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default guild config", "guildID", guildID, "file", configFileName, "error", err)
		return getDefaultGuild(guildID)
	}

	guild := &Guild{}
	err = json.Unmarshal(bytes, guild)
	if err != nil {
		slog.Error("failed to unmarshal default guild config", "guildID", guildID, "file", configFileName, "error", err, "config", string(bytes))
		return getDefaultGuild(guildID)
	}
	guild.GuildID = guildID

	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config", "guildID", guildID, "file", configFileName, "error", err)
	}
	slog.Info("create new guild", "guildID", guild.GuildID, "file", configFileName)

	return guild
}

// getDefaultGuild creates a new guild configuration for a given guild (server).
func getDefaultGuild(guildID string) *Guild {
	guild := &Guild{
		GuildID: guildID,
	}
	guild.AdminRoles = make([]string, len(DefaultAdminRoles))
	copy(guild.AdminRoles, DefaultAdminRoles)
	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config", "guildID", guildID, "file", guild.GuildID, "error", err)
	}

	return guild
}

// AddAdminRole adds a role to the list of admin roles for the guild.
func (guild *Guild) AddAdminRole(roleID string) {
	guildLock.Lock()
	defer guildLock.Unlock()

	guild.refresh()

	if slices.Contains(guild.AdminRoles, roleID) {
		slog.Warn("role already exists", "guildID", guild.GuildID, "roleID", roleID, "adminRoles", guild.AdminRoles)
		return
	}

	guild.AdminRoles = append(guild.AdminRoles, roleID)
	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config", "guildID", guild.GuildID, "file", guild.GuildID, "error", err)
	}
	slog.Info("added role", "guildID", guild.GuildID, "roleID", roleID)

}

// RemoveAdminRole removes a role from the list of admin roles for the guild.
func (guild *Guild) RemoveAdminRole(roleID string) {
	guildLock.Lock()
	defer guildLock.Unlock()

	guild.refresh()

	for i, role := range guild.AdminRoles {
		if role == roleID {
			guild.AdminRoles = append(guild.AdminRoles[:i], guild.AdminRoles[i+1:]...)
			if err := writeGuild(guild); err != nil {
				slog.Error("failed to write default guild config", "guildID", guild.GuildID, "file", guild.GuildID, "error", err)
			}
			slog.Info("removed admin role", "guildID", guild.GuildID, "roleID", roleID)
			return
		}
	}
	slog.Warn("role not found", "guildID", guild.GuildID, "roleID", roleID)
}

// refresh refreshes the guild configuration.
func (guild *Guild) refresh() {
	currentGuild := readGuild(guild.GuildID)
	if currentGuild == nil {
		return
	}
	guild.AdminRoles = currentGuild.AdminRoles
}

// GetAdminRoles returns the list of admin role names for the guild.
func (guild *Guild) GetAdminRoles() []string {
	guildRoles := GetCachedGuildRoles(guild.GuildID)
	adminRoles := make([]string, 0, len(guild.AdminRoles))

	for _, role := range guildRoles.Roles {
		if slices.Contains(guild.AdminRoles, role.RoleID) {
			adminRoles = append(adminRoles, role.RoleName)
		}
	}

	return adminRoles
}

// GetAdminRoleIDs returns the list of admin role IDs for the guild.
func (guild *Guild) GetAdminRoleIDs() []string {
	adminRoles := make([]string, len(guild.AdminRoles))
	copy(adminRoles, guild.AdminRoles)
	return adminRoles
}

// UpdateAdminRolesToIDs updates admin roles from role names to role IDs.
func (guild *Guild) UpdateAdminRolesToIDs(guildRoles *GuildRoles) {
	guildLock.Lock()
	defer guildLock.Unlock()

	if guildRoles == nil {
		slog.Warn("unable to update admin roles to IDs: guild roles are nil", "guildID", guild.GuildID)
		return
	}

	guild.refresh()

	roleIDs := make(map[string]struct{}, len(guildRoles.Roles))
	roleIDsByName := make(map[string]string, len(guildRoles.Roles))
	for _, role := range guildRoles.Roles {
		roleIDs[role.RoleID] = struct{}{}
		roleIDsByName[role.RoleName] = role.RoleID
	}

	updatedAdminRoles := make([]string, 0, len(guild.AdminRoles))
	changed := false

	for _, adminRole := range guild.AdminRoles {
		if _, ok := roleIDs[adminRole]; ok {
			if !slices.Contains(updatedAdminRoles, adminRole) {
				updatedAdminRoles = append(updatedAdminRoles, adminRole)
			} else {
				changed = true
			}
			continue
		}

		roleID, ok := roleIDsByName[adminRole]
		if !ok {
			slog.Warn("unable to convert admin role to role ID; pruning missing role",
				"guildID", guild.GuildID,
				"adminRole", adminRole,
			)
			changed = true
			continue
		}

		changed = true
		if !slices.Contains(updatedAdminRoles, roleID) {
			updatedAdminRoles = append(updatedAdminRoles, roleID)
		}
	}

	if !changed {
		return
	}

	guild.AdminRoles = updatedAdminRoles
	if err := writeGuild(guild); err != nil {
		slog.Error("failed to migrate admin roles to role IDs", "guildID", guild.GuildID, "error", err)
		return
	}

	slog.Info("migrated admin roles to role IDs", "guildID", guild.GuildID, "adminRoles", guild.AdminRoles)
}

// String returns a string representation of the guild.
func (guild *Guild) String() string {
	return fmt.Sprintf("Guild{guildID = %s, adminRoles = %v}", guild.GuildID, guild.AdminRoles)
}
