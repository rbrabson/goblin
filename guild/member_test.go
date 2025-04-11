package guild

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		sslog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestSetName(t *testing.T) {
	members := make([]*Member, 0, 1)
	defer func() {
		for _, member := range members {
			db.Delete(MEMBER_COLLECTION, bson.M{"guild_id": member.GuildID, "member_id": member.MemberID})
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
