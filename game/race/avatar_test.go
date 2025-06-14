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

func TestGetRacers(t *testing.T) {
	racers := GetRaceAvatars("123", "clash")
	if len(racers) == 0 {
		t.Error("expected racers to be created")
		return
	}

	racers = GetRaceAvatars("123", "clash")
	if (len(racers)) == 0 {
		t.Error("expected racers to be found")
	}

	filter := bson.M{"guild_id": "123", "theme": "clash"}
	err := db.Delete(RacerCollection, filter)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCalculateMovement(t *testing.T) {
	racers := GetRaceAvatars("123", "clash")
	if len(racers) == 0 {
		t.Error("expected racers to be created")
		return
	}
	racer := racers[0]

	movement := racer.calculateMovement(1)
	slog.Info("movement", slog.Int("movement", movement))

	movement = racer.calculateMovement(2)
	slog.Info("movement", slog.Int("movement", movement))

	movement = racer.calculateMovement(3)
	slog.Info("movement", slog.Int("movement", movement))

	filter := bson.M{"guild_id": "123", "theme": "clash"}
	err := db.Delete(RacerCollection, filter)
	if err != nil {
		t.Error(err)
		return
	}
}
