package guild

import (
	"log/slog"
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	GuildId = "12345"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetAdminRoles(t *testing.T) {
	setup()
	defer teardown()

	adminRoles := GetAdminRoles(GuildId)
	if adminRoles == nil {
		t.Error("expected admin roles to be not nil")
		return
	}

	if len(adminRoles) == 0 {
		t.Error("expected roles to be not empty")
		return
	}
}

func TestCheckAdminRoles(t *testing.T) {
	setup()
	defer teardown()

	adminRoles := GetAdminRoles(GuildId)
	if adminRoles == nil {
		t.Error("expected admin roles to be not nil")
		return
	}
	if len(adminRoles) == 0 {
		t.Error("admin roles not found")
		return
	}

	guildRoles := []string{"Role 1", "Role 2", "Admin", "Role 3"}
	if !CheckAdminRole(guildRoles, adminRoles) {
		t.Error("admin roles not found")
		return
	}
}

func setup() {
	type Server struct {
		GuildID    string   `bson:"guild_id"`
		AdminRoles []string `bson:"admin_roles"`
	}
	server := &Server{
		GuildID:    GuildId,
		AdminRoles: []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"},
	}
	if err := db.UpdateOrInsert(GuildCollection, bson.M{"guild_id": GuildId}, server); err != nil {
		slog.Error("Error inserting guild",
			slog.String("guildID", server.GuildID),
			slog.Any("error", err),
		)
	}
}

func teardown() {
	if err := db.Delete(GuildCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("Error deleting guild",
			slog.String("guildID", GuildId),
			slog.Any("error", err),
		)
	}
}

func TestGetMemberRoles(t *testing.T) {
	// Create mock guild roles
	guildRoles := []*discordgo.Role{
		{ID: "role1", Name: "Admin"},
		{ID: "role2", Name: "Moderator"},
		{ID: "role3", Name: "User"},
		{ID: "role4", Name: "Guest"},
	}

	// Create mock member role IDs
	memberRoleIDs := []string{"role1", "role3"}

	// Get member roles
	memberRoles := GetMemberRoles(guildRoles, memberRoleIDs)

	// Verify the result
	if len(memberRoles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(memberRoles))
	}

	// Check that the correct roles were returned
	expectedRoles := []string{"Admin", "User"}
	for i, role := range expectedRoles {
		if i >= len(memberRoles) || memberRoles[i] != role {
			t.Errorf("Expected role %s at position %d, got %s", role, i, memberRoles[i])
		}
	}
}
