package guild

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/rbrabson/goblin/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DEFAULT_ADMIN_ROLES = []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"}
)

var (
	sslog = logger.GetLogger()
)

// Guild is the configuration for a guild (guild).
type Guild struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	AdminRoles []string           `json:"admin_roles" bson:"admin_roles"`
}

// GetAllGuilds returns all guilds in the database.
func GetAllGuilds() []*Guild {
	guilds := make([]*Guild, 0)
	err := db.FindMany(GUILD_COLLECTION, bson.M{}, &guilds, bson.M{}, 0)
	if err != nil {
		sslog.Error("failed to get all guilds")
		return nil
	}

	sslog.Debug("all guilds",
		slog.Int("numGuilds", len(guilds)),
	)
	return guilds
}

// GetGuild returns the guild configuration for a given guild (guild).
func GetGuild(guildID string) *Guild {
	guild := readGuild(guildID)
	if guild == nil {
		guild = readGuildFromFile(guildID)
	}

	return guild
}

// readGuildFromFile creates a new guild configuration for a given guild (guild).
func readGuildFromFile(guildID string) *Guild {

	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "guild", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		sslog.Error("failed to read default guild config",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultGuild(guildID)
	}

	guild := &Guild{}
	err = json.Unmarshal(bytes, guild)
	if err != nil {
		sslog.Error("failed to unmarshal default guild config",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
			slog.String("config", string(bytes)),
		)
		return getDefaultGuild(guildID)
	}
	guild.GuildID = guildID

	writeGuild(guild)
	sslog.Info("create new guild",
		slog.String("guildID", guild.GuildID),
		slog.String("file", configFileName),
	)

	return guild
}

func getDefaultGuild(guildID string) *Guild {
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
	if slices.Contains(guild.AdminRoles, roleName) {
		sslog.Warn("role already exists",
			slog.String("guildID", guild.GuildID),
			slog.String("roleName", roleName),
			"adminRoles", guild.AdminRoles,
		)
		return
	}

	guild.AdminRoles = append(guild.AdminRoles, roleName)
	writeGuild(guild)
	sslog.Info("adde role",
		slog.String("guildID", guild.GuildID),
		slog.String("roleName", roleName),
	)

}

// RemoveAdminRole removes a role from the list of admin roles for the guild.
func (guild *Guild) RemoveAdminRole(roleName string) {
	for i, role := range guild.AdminRoles {
		if role == roleName {
			guild.AdminRoles = append(guild.AdminRoles[:i], guild.AdminRoles[i+1:]...)
			writeGuild(guild)
			sslog.Info("removed admin role",
				slog.String("guildID", guild.GuildID),
				slog.String("roleName", roleName),
			)
			return
		}
	}
	sslog.Warn("role not found",
		slog.String("guildID", guild.GuildID),
		slog.String("roleName", roleName),
	)

}

// GetAdminRoles returns the list of admin roles for the guild.
func (guild *Guild) GetAdminRoles() []string {
	return guild.AdminRoles
}

// String returns a string representation of the guild.
func (guild *Guild) String() string {
	return fmt.Sprintf("Guild{guildID = %s, adminRoles = %v}", guild.GuildID, guild.AdminRoles)
}
