package guild

import (
	"log/slog"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	db *mongo.MongoDB
)

// Sets the database to be used by the role package.
func SetDB(database *mongo.MongoDB) {
	db = database
}

// GetAdminRoles returns the list of admin roles for a given guild.
// If the guild is not found, it returns nil.
// If there are no admin roles, it returns an empty slice.
func GetAdminRoles(guildID string) []string {
	filter := bson.M{"guild_id": guildID}
	server := &Guild{}
	err := db.FindOne(GUILD_COLLECTION, filter, server)
	if err != nil {
		slog.Debug("server not found in the database",
			slog.String("guildID", guildID),
		)
		return nil
	}
	if server.GuildID == "" {
		server = readGuildFromFile(guildID)
	}

	return server.AdminRoles
}

// GetGuildRoles returns the list of roles for a guild.
func GetGuildRoles(s *discordgo.Session, guildID string) []*discordgo.Role {
	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		slog.Error("failed to get guild roles",
			slog.String("guildID", guildID),
			slog.Any("guildRoles", guildRoles),
			slog.Any("error", err),
		)
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
	slog.Debug("role not found",
		slog.String("guildID", guildID),
		slog.String("roleName", roleName),
	)
	return nil
}

// GetMemberRoles returns the list of roles names for a member with the given set of role IDs
func GetMemberRoles(guildRoles []*discordgo.Role, roleIDs []string) []string {
	roleNames := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		for _, role := range guildRoles {
			if role.ID == roleID {
				roleNames = append(roleNames, role.Name)
			}
		}
	}
	return roleNames
}

// MemberHasRole returns a boolean indicating whether a member has a specific role in the guild.
// It returns true if the member has the role, false otherwise.
func MemberHasRole(s *discordgo.Session, guildID string, memberID string, role *discordgo.Role) bool {
	// Check to see if the member already has the role
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		slog.Error("failed to get member",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return true
	}
	if slices.Contains(member.Roles, role.ID) {
		slog.Warn("member already has role",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.String("roleName", role.Name),
			slog.Any("memberRoles", member.Roles),
		)
		return true
	}

	slog.Debug("member does not have role",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
		slog.String("roleName", role.Name),
		slog.Any("memberRoles", member.Roles),
	)
	return false
}

// AssignRole assigns a role to the member in the guild.
func AssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	guildRoles := GetGuildRoles(s, guildID)
	roleID := ""
	for _, role := range guildRoles {
		if role.Name == roleName {
			roleID = role.ID
			break
		}
	}
	if roleID == "" {
		slog.Error("role not found",
			slog.String("guildID", guildID),
			slog.String("roleName", roleName),
			slog.Any("guildRoles", guildRoles),
		)
		return nil
	}

	err := s.GuildMemberRoleAdd(guildID, memberID, roleID)
	if err != nil {
		slog.Error("failed to assign role",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.String("roleID", roleID),
			slog.Any("error", err),
		)
	}

	slog.Info("assigned role",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
		slog.String("roleID", roleID),
	)
	return err
}

// UnAssignRole removes a role to the member in the guild.
func UnAssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	guildRoles := GetGuildRoles(s, guildID)
	roleID := ""
	for _, role := range guildRoles {
		if role.Name == roleName {
			roleID = role.ID
			break
		}
	}
	if roleID == "" {
		slog.Error("role not found",
			slog.String("guildID", guildID),
			slog.String("roleName", roleName),
			slog.Any("guildRoles", guildRoles),
		)
		return nil
	}

	err := s.GuildMemberRoleRemove(guildID, memberID, roleID)
	if err != nil {
		slog.Error("failed to unassign role",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.String("roleID", roleID),
			slog.Any("error", err),
		)
	}

	slog.Info("unassigned role",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
		slog.String("roleID", roleID),
	)
	return err
}

// CheckAdminRole checks if a member has any admin role in the server.
func CheckAdminRole(adminRoles []string, memberRoles []string) bool {
	if len(adminRoles) == 0 {
		return true
	}

	for _, memberRole := range memberRoles {
		if slices.Contains(adminRoles, memberRole) {
			return true
		}
	}
	return false
}

// IsAdmin checks if a member is an admin in a guild.
// It returns true if the member has any admin role in the server.
// It returns false if the member does not have any admin role.
func IsAdmin(s *discordgo.Session, guildID string, memberID string) bool {
	guildRoles := GetGuildRoles(s, guildID)
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		slog.Error("failed to get guild member",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return false
	}
	memberRoles := GetMemberRoles(guildRoles, member.Roles)
	adminRoles := GetAdminRoles(guildID)
	isAdmin := CheckAdminRole(adminRoles, memberRoles)
	slog.Debug("isAdmin",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
		slog.Bool("isAdmin", isAdmin),
		slog.Any("adminRoles", adminRoles),
		slog.Any("memberRoles", memberRoles),
	)

	return isAdmin
}
