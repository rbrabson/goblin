package role

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	SERVER_COLLECTION = "servers"
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
	server := &Server{}
	err := db.FindOne(SERVER_COLLECTION, filter, server)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Debug("server not found in the database")
	}
	if server.GuildID == "" {
		server = newServer(guildID)
	}

	return server.AdminRoles
}

// GetGuildRoles returns the list of roles for a guild.
func GetGuildRoles(s *discordgo.Session, guildID string) []*discordgo.Role {
	log.Trace("--> role.GetGuildRoles")
	defer log.Trace("<-- Gole.getGuildRoles")

	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": guildID, "error": err}).Error("failed to get guild roles")
		return nil
	}
	return guildRoles
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

// CheckAdminRole checks if a member has any admin role in the server.
func CheckAdminRole(adminRoles []string, memberRoles []string) bool {
	log.Trace("--> role.CheckAdminRole")
	defer log.Trace("<-- role.CheckAdminRole")

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
	log.Trace("--> role.IsAdmin")
	defer log.Trace("<-- role.IsAdmin")

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
