package role

import (
	"log"
	"slices"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"

	discrole "github.com/rbrabson/goblin/internal/role"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
	discrole.SetDB(db)
}

func TestGetServer(t *testing.T) {
	servers := make([]*discrole.Server, 0, 1)
	defer func() {
		for _, server := range servers {
			db.Delete(discrole.SERVER_COLLECTION, bson.M{"guild_id": server.GuildID})
		}
	}()

	// Create a new server configuration.
	server := discrole.GetServer("12345")
	if server.GuildID != "12345" {
		t.Errorf("Expected guild ID to be 12345, got %s", server.GuildID)
		return
	}
	servers = append(servers, server)

	for i, role := range server.AdminRoles {
		if !slices.Contains(discrole.DEFAULT_ADMIN_ROLES, role) {
			t.Errorf("Expected role to be %s, got %s", discrole.DEFAULT_ADMIN_ROLES[i], role)
		}
	}
}

func TestAddAdminRole(t *testing.T) {
	servers := make([]*discrole.Server, 0, 1)
	defer func() {
		for _, server := range servers {
			db.Delete(discrole.SERVER_COLLECTION, bson.M{"guild_id": server.GuildID})
		}
	}()

	// Create a new server configuration.
	server := discrole.GetServer("12345")
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	server.AddAdminRole("NewRole")
	server = discrole.GetServer(server.GuildID)
	if server == nil {
		t.Errorf("Expected server to be retrieved")
		return
	}

	if !slices.Contains(server.AdminRoles, "NewRole") { // This test will fail.
		t.Errorf("Expected role %s to be in the list of admin roles", "NewRole")
	}
}

func TestRemoveAdminRole(t *testing.T) {
	servers := make([]*discrole.Server, 0, 1)
	defer func() {
		for _, server := range servers {
			db.Delete(discrole.SERVER_COLLECTION, bson.M{"guild_id": server.GuildID})
		}
	}()

	// Create a new server configuration.
	server := discrole.GetServer("12345")
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	server.RemoveAdminRole("NewRole")
	server = discrole.GetServer(server.GuildID)
	if server == nil {
		t.Errorf("Expected server to be retrieved")
		return
	}
	if slices.Contains(server.AdminRoles, "NewRole") { // This test will fail.
		t.Errorf("Expected role %s to not be in the list of admin roles", "NewRole")
	}
}

func TestListAdminRoles(t *testing.T) {
	servers := make([]*discrole.Server, 0, 1)
	defer func() {
		for _, server := range servers {
			db.Delete(discrole.SERVER_COLLECTION, bson.M{"guild_id": server.GuildID})
		}
	}()

	// Create a new server configuration.
	server := discrole.GetServer("12345")
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	roles := server.GetAdminRoles()
	for i, role := range roles {
		if !slices.Contains(discrole.DEFAULT_ADMIN_ROLES, role) {
			t.Errorf("Expected role to not be %s, got %s", discrole.DEFAULT_ADMIN_ROLES[i], role)
		}
	}
}
