package guild

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
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
	log.Trace("--> role.GetAdminRoles")
	defer log.Trace("<-- role.GetAdminRoles")

	filter := bson.M{"guild_id": guildID}
	server := &Guild{}
	err := db.FindOne(GUILD_COLLECTION, filter, server)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Debug("server not found in the database")
	}
	if server.GuildID == "" {
		server = readGuildFromFile(guildID)
	}

	return server.AdminRoles
}

// GetGuildRoles returns the list of roles for a guild.
func GetGuildRoles(s *discordgo.Session, guildID string) []*discordgo.Role {
	log.Trace("--> role.GetGuildRoles")
	defer log.Trace("<-- Gole.getGuildRoles")

	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "error": err, "guildRoles": guildRoles}).Error("failed to get guild roles")
		return nil
	}
	return guildRoles
}

// GetGuildRole returns the role for a guild with the given name.
// If the role is not found, it returns nil.
func GetGuildRole(s *discordgo.Session, guildID string, roleName string) *discordgo.Role {
	log.Trace("--> role.GetGuildRole")
	defer log.Trace("<-- role.GetGuildRole")

	guildRoles := GetGuildRoles(s, guildID)
	for _, role := range guildRoles {
		if role.Name == roleName {
			return role
		}
	}
	log.WithFields(log.Fields{"guildID": guildID, "roleName": roleName}).Debug("role not found")
	return nil
}

// GetMemberRoles returns the list of roles names for a member with the given set of role IDs
func GetMemberRoles(guildRoles []*discordgo.Role, roleIDs []string) []string {
	log.Trace("--> role.GetMemberRoles")
	defer log.Trace("<-- role.GetMemberRoles")

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
	log.Trace("--> role.MemberHasRole")
	defer log.Trace("<-- role.MemberHasRole")

	// Check to see if the member already has the role
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "error": err}).Error("failed to read member")
		return true
	}
	if slices.Contains(member.Roles, role.ID) {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleName": role.Name}).Warn("member already has role")
		return true
	}

	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleName": role.Name, "memberRoles": member.Roles}).Debug("member does not have role")
	return false
}

// AssignRole assigns a role to the member in the guild.
func AssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	log.Trace("--> role.AssignRole")
	defer log.Trace("<-- role.AssignRole")

	guildRoles := GetGuildRoles(s, guildID)
	roleID := ""
	for _, role := range guildRoles {
		if role.Name == roleName {
			roleID = role.ID
			break
		}
	}
	if roleID == "" {
		log.WithFields(log.Fields{"guildID": guildID, "roleName": roleName}).Error("role not found")
		return nil
	}

	err := s.GuildMemberRoleAdd(guildID, memberID, roleID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleID": roleID, "error": err}).Error("failed to assign role")
	}

	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleID": roleID}).Info("assigned role")
	return err
}

// UnAssignRole removes a role to the member in the guild.
func UnAssignRole(s *discordgo.Session, guildID string, memberID string, roleName string) error {
	log.Trace("--> role.UnAssignRole")
	defer log.Trace("<-- role.UnAssignRole")

	guildRoles := GetGuildRoles(s, guildID)
	roleID := ""
	for _, role := range guildRoles {
		if role.Name == roleName {
			roleID = role.ID
			break
		}
	}
	if roleID == "" {
		log.WithFields(log.Fields{"guildID": guildID, "roleName": roleName}).Error("role not found")
		return nil
	}

	err := s.GuildMemberRoleRemove(guildID, memberID, roleID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleID": roleID, "error": err}).Error("failed to unassign role")
	}

	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "roleID": roleID}).Info("unassigned role")
	return err
}

// CheckAdminRole checks if a member has any admin role in the server.
func CheckAdminRole(adminRoles []string, memberRoles []string) bool {
	log.Trace("--> role.CheckAdminRole")
	defer log.Trace("<-- role.CheckAdminRole")

	if len(adminRoles) == 0 {
		log.WithFields(log.Fields{"adminRoles": adminRoles}).Trace("not using admin roles")
		return true
	}

	for _, memberRole := range memberRoles {
		if slices.Contains(adminRoles, memberRole) {
			log.WithFields(log.Fields{"memberRoles": memberRoles, "adminRoles": adminRoles}).Trace("member has admin role")
			return true
		}
	}
	log.WithFields(log.Fields{"memberRoles": memberRoles, "adminRoles": adminRoles}).Trace("member does not have admin role")
	return false
}

// IsAdmin checks if a member is an admin in a guild.
// It returns true if the member has any admin role in the server.
// It returns false if the member does not have any admin role.
func IsAdmin(s *discordgo.Session, guildID string, memberID string) bool {
	log.Trace("--> guild.IsAdmin")
	defer log.Trace("<-- guild.IsAdmin")

	guildRoles := GetGuildRoles(s, guildID)
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "error": err}).Error("failed to get guild member")
		return false
	}
	memberRoles := GetMemberRoles(guildRoles, member.Roles)
	adminRoles := GetAdminRoles(guildID)
	isAdmin := CheckAdminRole(adminRoles, memberRoles)
	log.WithFields(log.Fields{"guildID": guildID, "memberID": memberID, "isAdmin": isAdmin, "adminRoles": adminRoles, "memberRoles": memberRoles}).Debug("isAdmin")

	return isAdmin
}
