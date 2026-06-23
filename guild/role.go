package guild

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	db *mongo.MongoDB
)

// SetDB sets the database to be used by the role package.
func SetDB(database *mongo.MongoDB) {
	db = database
}

// GuildRoles is the cached Discord role list for a guild.
type GuildRoles struct {
	ID       bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID  string        `json:"guild_id" bson:"guild_id"`
	Roles    []*GuildRole  `json:"roles" bson:"roles"`
	SyncedAt time.Time     `json:"synced_at" bson:"synced_at"`
	Version  int           `json:"version" bson:"version"`
}

// GuildRole is a Discord role cached for slash command display and ID lookups.
type GuildRole struct {
	RoleID   string `json:"role_id" bson:"role_id"`
	RoleName string `json:"role_name" bson:"role_name"`
	Position int    `json:"position" bson:"position"`
}

// Sort keeps roles ordered by Discord position.
func (guildRoles *GuildRoles) Sort() {
	slices.SortFunc(guildRoles.Roles, func(a *GuildRole, b *GuildRole) int {
		return a.Position - b.Position
	})
}

// GetRoleName returns a role name by role ID.
func (guildRoles *GuildRoles) GetRoleName(roleID string) string {
	for _, role := range guildRoles.Roles {
		if role.RoleID == roleID {
			return role.RoleName
		}
	}

	return roleID
}

// GetRoleID returns a role ID by role name.
func (guildRoles *GuildRoles) GetRoleID(roleName string) string {
	for _, role := range guildRoles.Roles {
		if role.RoleName == roleName {
			return role.RoleID
		}
	}

	return ""
}

// GetAdminRoles returns the list of admin role names for a given guild.
// If the guild is not found, it returns nil.
// If there are no admin roles, it returns an empty slice.
func GetAdminRoles(guildID string) []string {
	server := GetGuild(guildID)
	if server == nil {
		return nil
	}

	return server.GetAdminRoles()
}

// GetAdminRoleIDs returns the list of admin role IDs for a given guild.
func GetAdminRoleIDs(guildID string) []string {
	server := GetGuild(guildID)
	if server == nil {
		return nil
	}

	return server.GetAdminRoleIDs()
}

// GetCachedGuildRoles returns cached roles for a guild, sorted by position.
func GetCachedGuildRoles(guildID string) *GuildRoles {
	guildRoles := readGuildRoles(guildID)
	if guildRoles == nil {
		return &GuildRoles{
			GuildID: guildID,
			Roles:   []*GuildRole{},
		}
	}

	guildRoles.Sort()
	return guildRoles
}

// SyncGuildRoles refreshes the guild_roles table from Discord.
func SyncGuildRoles(s *discordgo.Session, guildID string) *GuildRoles {
	roles := GetGuildRoles(s, guildID)
	if roles == nil {
		return GetCachedGuildRoles(guildID)
	}

	guildRoles := &GuildRoles{
		GuildID:  guildID,
		Roles:    make([]*GuildRole, 0, len(roles)),
		SyncedAt: time.Now(),
	}

	for _, role := range roles {
		guildRoles.Roles = append(guildRoles.Roles, &GuildRole{
			RoleID:   role.ID,
			RoleName: role.Name,
			Position: role.Position,
		})
	}

	guildRoles.Sort()

	current := readGuildRoles(guildID)
	if current != nil {
		guildRoles.ID = current.ID
		guildRoles.Version = current.Version + 1
	} else {
		guildRoles.Version = 1
	}

	if err := writeGuildRoles(guildRoles); err != nil {
		slog.Error("failed to sync guild roles", "guildID", guildID, "error", err)
		return current
	}

	return guildRoles
}

// SyncGuildAdminRoles refreshes a guild's role cache and migrates admin roles from role names to role IDs.
// Missing admin roles are pruned by UpdateAdminRolesToIDs.
func SyncGuildAdminRoles(s *discordgo.Session, guildID string) {
	if s == nil {
		slog.Warn("unable to sync guild admin roles: discord session is nil", "guildID", guildID)
		return
	}
	if guildID == "" {
		slog.Warn("unable to sync guild admin roles: guild ID is empty")
		return
	}

	guildRoles := SyncGuildRoles(s, guildID)
	if guildRoles == nil {
		slog.Warn("unable to sync guild admin roles: role cache unavailable", "guildID", guildID)
		return
	}

	server := GetGuild(guildID)
	if server == nil {
		slog.Warn("unable to sync guild admin roles: guild config unavailable", "guildID", guildID)
		return
	}

	server.UpdateAdminRolesToIDs(guildRoles)
}

// SyncAllGuildRoles refreshes the guild_roles table for every guild the bot is running on.
// It also migrates guild admin roles from role names to role IDs.
func SyncAllGuildRoles(s *discordgo.Session) {
	if s == nil || s.State == nil {
		slog.Warn("unable to sync guild roles at startup: discord session state is unavailable")
		return
	}

	guilds := s.State.Guilds
	slog.Info("syncing guild roles at startup", "numGuilds", len(guilds))

	for _, discordGuild := range guilds {
		if discordGuild == nil {
			continue
		}

		SyncGuildAdminRoles(s, discordGuild.ID)
	}

	slog.Info("finished syncing guild roles at startup", "numGuilds", len(guilds))
}

// GetGuildRoles returns the list of roles for a guild.
func GetGuildRoles(s *discordgo.Session, guildID string) []*discordgo.Role {
	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		slog.Error("failed to get guild roles", "guildID", guildID, "guildRoles", guildRoles, "error", err)
		return nil
	}
	return guildRoles
}

// GetGuildRole returns the role for a guild with the given name.
// If the role is not found, it returns nil.
func GetGuildRole(s *discordgo.Session, guildID string, roleName string) *discordgo.Role {
	guildRoles := GetGuildRoles(s, guildID)
	for _, role := range guildRoles {
		if role.Name == roleName {
			return role
		}
	}
	slog.Debug("role not found", "guildID", guildID, "roleName", roleName)
	return nil
}

// GetMemberRoles returns the list of role names for a member with the given set of role IDs
func GetMemberRoles(guildRoles []*discordgo.Role, roleIDs []string) []string {
	roleNamesByID := make(map[string]string, len(guildRoles))
	for _, role := range guildRoles {
		roleNamesByID[role.ID] = role.Name
	}

	roleNames := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleName, ok := roleNamesByID[roleID]; ok {
			roleNames = append(roleNames, roleName)
		}
	}
	return roleNames
}

// MemberHasRole returns a boolean indicating whether a member has a specific role in the guild.
// It returns true if the member has the role, false otherwise.
func MemberHasRole(s *discordgo.Session, guildID string, memberID string, role *discordgo.Role) bool {
	if role == nil {
		slog.Error("role is nil", "guildID", guildID, "memberID", memberID)
		return false
	}

	// Check to see if the member already has the role
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		slog.Error("failed to get member", "guildID", guildID, "memberID", memberID, "error", err)
		return false
	}
	if slices.Contains(member.Roles, role.ID) {
		slog.Warn("member already has role", "guildID", guildID, "memberID", memberID, "roleName", role.Name, "memberRoles", member.Roles)
		return true
	}

	slog.Debug("member does not have role", "guildID", guildID, "memberID", memberID, "roleName", role.Name, "memberRoles", member.Roles)
	return false
}

// AssignRole assigns a role to the member in the guild.
// AssignRole assigns a role to the member in the guild.
func AssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	guildRoles := SyncGuildRoles(s, guildID)
	roleID := guildRoles.GetRoleID(roleName)
	if roleID == "" {
		err := fmt.Errorf("role %q not found", roleName)
		slog.Error("role not found", "guildID", guildID, "roleName", roleName, "guildRoles", guildRoles)
		return err
	}

	err := s.GuildMemberRoleAdd(guildID, memberID, roleID)
	if err != nil {
		slog.Error("failed to assign role", "guildID", guildID, "memberID", memberID, "roleID", roleID, "error", err)
	}

	slog.Info("assigned role", "guildID", guildID, "memberID", memberID, "roleID", roleID)
	return err
}

// UnAssignRole removes a role to the member in the guild.
func UnAssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	guildRoles := SyncGuildRoles(s, guildID)
	roleID := guildRoles.GetRoleID(roleName)
	if roleID == "" {
		err := fmt.Errorf("role %q not found", roleName)
		slog.Error("role not found", "guildID", guildID, "roleName", roleName, "guildRoles", guildRoles)
		return err
	}

	err := s.GuildMemberRoleRemove(guildID, memberID, roleID)
	if err != nil {
		slog.Error("failed to unassign role", "guildID", guildID, "memberID", memberID, "roleID", roleID, "error", err)
	}

	slog.Info("unassigned role", "guildID", guildID, "memberID", memberID, "roleID", roleID)
	return err
}

// CheckAdminRole checks if a member has any admin role in the server.
func CheckAdminRole(adminRoleIDs []string, memberRoleIDs []string) bool {
	if len(adminRoleIDs) == 0 {
		return true
	}

	for _, memberRoleID := range memberRoleIDs {
		if slices.Contains(adminRoleIDs, memberRoleID) {
			return true
		}
	}
	return false
}

// IsAdmin checks if a member is an admin in a guild.
// It returns true if the member has any admin role in the server.
// It returns false if the member does not have any admin role.
func IsAdmin(s *discordgo.Session, guildID string, memberID string) bool {
	guildRoles := SyncGuildRoles(s, guildID)

	server := GetGuild(guildID)
	if server != nil {
		server.UpdateAdminRolesToIDs(guildRoles)
	}

	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		slog.Error("failed to get guild member", "guildID", guildID, "memberID", memberID, "error", err)
		return false
	}
	adminRoleIDs := GetAdminRoleIDs(guildID)
	isAdmin := CheckAdminRole(adminRoleIDs, member.Roles)
	slog.Debug("isAdmin", "guildID", guildID, "memberID", memberID, "isAdmin", isAdmin, "adminRoleIDs", adminRoleIDs, "memberRoleIDs", member.Roles)

	return isAdmin
}
