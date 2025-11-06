package guild

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultGuildTheme = "clash"
)

var (
	DefaultAdminRoles = []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"}
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
	err := db.FindMany(GuildCollection, bson.M{}, &guilds, bson.M{}, 0)
	if err != nil {
		slog.Error("failed to get all guilds")
		return nil
	}

	slog.Debug("all guilds",
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
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	guildTheme := os.Getenv("DISCORD_GUILD_THEME")
	if guildTheme == "" {
		guildTheme = defaultGuildTheme
	}
	configFileName := filepath.Join(configDir, "guild", "config", guildTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default guild config",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultGuild(guildID)
	}

	guild := &Guild{}
	err = json.Unmarshal(bytes, guild)
	if err != nil {
		slog.Error("failed to unmarshal default guild config",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
			slog.String("config", string(bytes)),
		)
		return getDefaultGuild(guildID)
	}
	guild.GuildID = guildID

	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
	}
	slog.Info("create new guild",
		slog.String("guildID", guild.GuildID),
		slog.String("file", configFileName),
	)

	return guild
}

func getDefaultGuild(guildID string) *Guild {
	guild := &Guild{
		GuildID: guildID,
	}
	guild.AdminRoles = make([]string, len(DefaultAdminRoles))
	copy(guild.AdminRoles, DefaultAdminRoles)
	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config",
			slog.String("guildID", guildID),
			slog.String("file", guild.GuildID),
			slog.Any("error", err),
		)
	}

	return guild
}

// AddAdminRole adds a role to the list of admin roles for the guild.
func (guild *Guild) AddAdminRole(roleName string) {
	if slices.Contains(guild.AdminRoles, roleName) {
		slog.Warn("role already exists",
			slog.String("guildID", guild.GuildID),
			slog.String("roleName", roleName),
			slog.Any("adminRoles", guild.AdminRoles),
		)
		return
	}

	guild.AdminRoles = append(guild.AdminRoles, roleName)
	if err := writeGuild(guild); err != nil {
		slog.Error("failed to write default guild config",
			slog.String("guildID", guild.GuildID),
			slog.String("file", guild.GuildID),
			slog.Any("error", err),
		)
	}
	slog.Info("adde role",
		slog.String("guildID", guild.GuildID),
		slog.String("roleName", roleName),
	)

}

// RemoveAdminRole removes a role from the list of admin roles for the guild.
func (guild *Guild) RemoveAdminRole(roleName string) {
	for i, role := range guild.AdminRoles {
		if role == roleName {
			guild.AdminRoles = append(guild.AdminRoles[:i], guild.AdminRoles[i+1:]...)
			if err := writeGuild(guild); err != nil {
				slog.Error("failed to write default guild config",
					slog.String("guildID", guild.GuildID),
					slog.String("file", guild.GuildID),
					slog.Any("error", err),
				)
			}
			slog.Info("removed admin role",
				slog.String("guildID", guild.GuildID),
				slog.String("roleName", roleName),
			)
			return
		}
	}
	slog.Warn("role not found",
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
