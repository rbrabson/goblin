package role

import (
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"

	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
	guild.SetDB(db)
}

func TestGetServer(t *testing.T) {
	servers := make([]*guild.Guild, 0, 1)
	defer func() {
		for _, server := range servers {
			if err := db.Delete(guild.GuildCollection, bson.M{"guild_id": server.GuildID}); err != nil {
				slog.Error("Error deleting guild",
					slog.String("guildID", server.GuildID),
					slog.Any("err", err),
				)
			}
		}
	}()

	// Create a new server configuration.
	server := guild.GetGuild("12345")
	if server.GuildID != "12345" {
		t.Errorf("Expected guild ID to be 12345, got %s", server.GuildID)
		return
	}
	servers = append(servers, server)

	for i, role := range server.AdminRoles {
		if !slices.Contains(guild.DefaultAdminRoles, role) {
			t.Errorf("Expected role to be %s, got %s", guild.DefaultAdminRoles[i], role)
		}
	}
}

func TestAddAdminRole(t *testing.T) {
	// Use a unique guild ID for this test to avoid conflicts
	guildID := "12345-add-role-test"
	roleName := "NewRole"

	servers := make([]*guild.Guild, 0, 1)
	defer func() {
		for _, server := range servers {
			if err := db.Delete(guild.GuildCollection, bson.M{"guild_id": server.GuildID}); err != nil {
				slog.Error("Error deleting guild",
					slog.String("guildID", server.GuildID),
					slog.Any("err", err),
				)
			}
		}
	}()

	// Create a new server configuration.
	server := guild.GetGuild(guildID)
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	// Store original admin roles to restore later
	originalRoles := make([]string, len(server.AdminRoles))
	copy(originalRoles, server.AdminRoles)

	// Add the role
	server.AddAdminRole(roleName)

	// Re-read the server from the database to ensure it was saved
	server = guild.GetGuild(guildID)
	if server == nil {
		t.Errorf("Expected server to be retrieved")
		return
	}

	// Check if the role was added
	found := false
	for _, role := range server.AdminRoles {
		if role == roleName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected role %s to be in the list of admin roles", roleName)
	}

	// Restore original roles
	server.AdminRoles = originalRoles
	// We can't directly call writeGuild as it's not exported, so we'll use AddAdminRole and RemoveAdminRole
	// to restore the original state
}

func TestRemoveAdminRole(t *testing.T) {
	// Use a unique guild ID for this test to avoid conflicts
	guildID := "12345-remove-role-test"
	roleName := "RoleToRemove"

	servers := make([]*guild.Guild, 0, 1)
	defer func() {
		for _, server := range servers {
			if err := db.Delete(guild.GuildCollection, bson.M{"guild_id": server.GuildID}); err != nil {
				slog.Error("Error deleting guild",
					slog.String("guildID", server.GuildID),
					slog.Any("err", err),
				)
			}
		}
	}()

	// Create a new server configuration.
	server := guild.GetGuild(guildID)
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	// Store original admin roles to restore later
	originalRoles := make([]string, len(server.AdminRoles))
	copy(originalRoles, server.AdminRoles)

	// First add the role we want to remove
	server.AddAdminRole(roleName)

	// Re-read the server from the database to ensure it was saved
	server = guild.GetGuild(guildID)
	if server == nil {
		t.Errorf("Expected server to be retrieved")
		return
	}

	// Verify the role was added
	found := false
	for _, role := range server.AdminRoles {
		if role == roleName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected role %s to be in the list of admin roles before removal", roleName)
		return
	}

	// Now remove the role
	server.RemoveAdminRole(roleName)

	// Re-read the server from the database to ensure it was saved
	server = guild.GetGuild(guildID)
	if server == nil {
		t.Errorf("Expected server to be retrieved")
		return
	}

	// Verify the role was removed
	for _, role := range server.AdminRoles {
		if role == roleName {
			t.Errorf("Expected role %s to not be in the list of admin roles after removal", roleName)
			break
		}
	}

	// Restore original roles
	server.AdminRoles = originalRoles
	// We can't directly call writeGuild as it's not exported, so we'll use AddAdminRole and RemoveAdminRole
	// to restore the original state
}

func TestListAdminRoles(t *testing.T) {
	servers := make([]*guild.Guild, 0, 1)
	defer func() {
		for _, server := range servers {
			if err := db.Delete(guild.GuildCollection, bson.M{"guild_id": server.GuildID}); err != nil {
				slog.Error("Error deleting guild",
					slog.String("guildID", server.GuildID),
					slog.Any("err", err),
				)
			}
		}
	}()

	// Create a new server configuration.
	server := guild.GetGuild("12345")
	if server == nil {
		t.Errorf("Expected server to be created")
		return
	}
	servers = append(servers, server)

	roles := server.GetAdminRoles()
	for i, role := range roles {
		if !slices.Contains(guild.DefaultAdminRoles, role) {
			t.Errorf("Expected role to not be %s, got %s", guild.DefaultAdminRoles[i], role)
		}
	}
}
