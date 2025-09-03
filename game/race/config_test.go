package race

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetConfig(t *testing.T) {
	// Create
	config := GetConfig("123")
	if config == nil {
		t.Error("expected config to be created")
		return
	}

	config.BetAmount = 1000
	writeConfig(config)

	config = readConfig("123")
	if config.BetAmount != 1000 {
		t.Error("expected BetAmount to be 1000")
	}

	filter := bson.M{"guild_id": "123"}
	err := db.Delete(RaceConfigCollection, filter)
	if err != nil {
		t.Error(err)
		return
	}
}
