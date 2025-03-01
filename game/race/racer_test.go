package race

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
	log.SetLevel(log.DebugLevel)
}

func TestGetRacers(t *testing.T) {
	racers := GetRacers("123", "clash")
	if len(racers) == 0 {
		t.Error("expected racers to be created")
		return
	}

	racers = GetRacers("123", "clash")
	if (len(racers)) == 0 {
		t.Error("expected racers to be found")
	}

	filter := bson.M{"guild_id": "123", "theme": "clash"}
	err := db.Delete(RACER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCalculateMovement(t *testing.T) {
	racers := GetRacers("123", "clash")
	if len(racers) == 0 {
		t.Error("expected racers to be created")
		return
	}
	racer := racers[0]

	movement := racer.calculateMovement(1)
	log.Info("movement: ", movement)

	movement = racer.calculateMovement(2)
	log.Info("movement: ", movement)

	movement = racer.calculateMovement(3)
	log.Info("movement: ", movement)

	filter := bson.M{"guild_id": "123", "theme": "clash"}
	err := db.Delete(RACER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
		return
	}
}
