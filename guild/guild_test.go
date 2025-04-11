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
			db.Delete(MEMBER_COLLECTION, bson.M{"guild_id": member.GuildID, "member_id": member.MemberID})
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
