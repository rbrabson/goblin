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

func TestSetName(t *testing.T) {
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

	member := GetMember(guild.GuildID, "67890").SetName("userName", "displayName", "")
	if member == nil {
		t.Errorf("GetMember() member not found or created")
		return
	}
	members = append(members, member)

	member.SetName("userName", "", "")
	member = GetMember(guild.GuildID, member.MemberID)
	if member == nil || member.Name != "userName" {
		t.Errorf("SetName() member name not set")
		return
	}
	if member.Name != "userName" {
		t.Errorf("SetName() member name not set")
	}

	member.SetName("userName", "displayName", "")
	member = GetMember(guild.GuildID, member.MemberID)
	if member == nil {
		t.Errorf("SetName() member not found")
		return
	}
	if member.Name != "displayName" {
		t.Errorf("SetName() member name not set")
	}
}
