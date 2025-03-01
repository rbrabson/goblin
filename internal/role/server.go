package role

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DEFAULT_ADMIN_ROLES = []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"}
)

// Server is the configuration for a guild (server).
type Server struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	AdminRoles []string           `json:"admin_roles" bson:"admin_roles"`
}

// GetServer returns the server configuration for a given guild (server).
func GetServer(guildID string) *Server {
	server := readServer(guildID)
	if server == nil {
		server = newServer(guildID)
	}

	log.Tracef("Server: %v", server)
	return server
}

// newServer creates a new server configuration for a given guild (server).
func newServer(guildID string) *Server {
	server := &Server{
		GuildID: guildID,
	}
	server.AdminRoles = make([]string, len(DEFAULT_ADMIN_ROLES))
	copy(server.AdminRoles, DEFAULT_ADMIN_ROLES)
	writeServer(server)

	return server
}

// AddAdminRole adds a role to the list of admin roles for the server.
func (server *Server) AddAdminRole(roleName string) {
	log.Trace("--> server.Server.AddAdminRole")
	defer log.Trace("<-- server.Server.AddAdminRole")

	if slices.Contains(server.AdminRoles, roleName) {
		log.WithFields(log.Fields{"roleName": roleName, "adminRoles": server.AdminRoles}).Warn("role already exists")
		return
	}

	server.AdminRoles = append(server.AdminRoles, roleName)
	writeServer(server)
	log.WithFields(log.Fields{"roleName": roleName, "adminRoles": server.AdminRoles}).Info("added admin role")
}

// RemoveAdminRole removes a role from the list of admin roles for the server.
func (server *Server) RemoveAdminRole(roleName string) {
	log.Trace("--> server.Server.RemoveAdminRole")
	defer log.Trace("<-- server.Server.RemoveAdminRole")

	for i, role := range server.AdminRoles {
		if role == roleName {
			server.AdminRoles = append(server.AdminRoles[:i], server.AdminRoles[i+1:]...)
			writeServer(server)
			log.WithFields(log.Fields{"roleName": roleName, "adminRoles": server.AdminRoles}).Info("removed admin role")
			return
		}
	}
	log.WithFields(log.Fields{"roleName": roleName, "adminRoles": server.AdminRoles}).Warn("role not found")
}

// GetAdminRoles returns the list of admin roles for the server.
func (server *Server) GetAdminRoles() []string {
	log.Trace("--> server.Server.GetAdminRoles")
	defer log.Trace("<-- server.Server.GetAdminRoles")

	return server.AdminRoles
}

// String returns a string representation of the server.
func (server *Server) String() string {
	return fmt.Sprintf("Server{guildID = %s, adminRoles = %v}", server.GuildID, server.AdminRoles)
}
