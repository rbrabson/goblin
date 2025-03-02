package guild

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	GUILD_ID = "12345"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
}

func TestGetAdminRoles(t *testing.T) {
	setup()
	defer teardown()

	adminRoles := GetAdminRoles(GUILD_ID)
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

	adminRoles := GetAdminRoles(GUILD_ID)
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
		GuildID:    GUILD_ID,
		AdminRoles: []string{"Admin", "Admins", "Administrator", "Mod", "Mods", "Moderator"},
	}
	db.UpdateOrInsert(GUILD_COLLECTION, bson.M{"guild_id": GUILD_ID}, server)
}

func teardown() {
	db.Delete(GUILD_COLLECTION, bson.M{"guild_id": GUILD_ID})
}
