package guild

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetGuild(t *testing.T) {
	guild := GetGuild("12345")
	if guild == nil {
		t.Errorf("GetGuild() guild not found or created")
		return
	}
}

func TestGetMember(t *testing.T) {
	members := make([]*Member, 0, 1)
	defer func() {
		for _, member := range members {
			if err := db.Delete(MemberCollection, bson.M{"guild_id": member.GuildID, "member_id": member.MemberID}); err != nil {
				slog.Error("Error deleting guild member",
					slog.String("guildID", member.GuildID),
					slog.String("memberID", member.MemberID),
					slog.Any("err", err),
				)
			}
		}
	}()

	guild := GetGuild("12345")
	if guild == nil {
		t.Errorf("GetGuild() guild not found or created")
		return
	}

	member := GetMember(guild.GuildID, "67890")
	if member == nil {
		t.Errorf("GetMember() member not found or created")
		return
	}
	members = append(members, member)
}

func TestAddAndRemoveAdminRole(t *testing.T) {
	// Setup
	guildID := "12345"
	testRole := "TestAdminRole"

	// Get the guild
	guild := GetGuild(guildID)
	if guild == nil {
		t.Errorf("GetGuild() guild not found or created")
		return
	}

	// Store original admin roles to restore later
	originalRoles := make([]string, len(guild.AdminRoles))
	copy(originalRoles, guild.AdminRoles)

	// Cleanup function to restore original state
	defer func() {
		guild = GetGuild(guildID)
		guild.AdminRoles = originalRoles
		if err := writeGuild(guild); err != nil {
			slog.Error("Error restoring guild admin roles",
				slog.String("guildID", guildID),
				slog.Any("err", err),
			)
		}
	}()

	// Test AddAdminRole
	guild.AddAdminRole(testRole)

	// Verify role was added
	guild = GetGuild(guildID) // Re-read from database to ensure it was saved
	found := false
	for _, role := range guild.AdminRoles {
		if role == testRole {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AddAdminRole() failed to add role %s", testRole)
	}

	// Test RemoveAdminRole
	guild.RemoveAdminRole(testRole)

	// Verify role was removed
	guild = GetGuild(guildID) // Re-read from database to ensure it was saved
	for _, role := range guild.AdminRoles {
		if role == testRole {
			t.Errorf("RemoveAdminRole() failed to remove role %s", testRole)
			break
		}
	}
}

func TestGetAllGuilds(t *testing.T) {
	// Get all guilds
	guilds := GetAllGuilds()

	// Verify we got at least one guild (the one created in previous tests)
	if len(guilds) < 1 {
		t.Errorf("GetAllGuilds() returned no guilds")
		return
	}

	// Verify the test guild is in the list
	found := false
	for _, guild := range guilds {
		if guild.GuildID == "12345" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("GetAllGuilds() did not return the test guild")
	}
}
