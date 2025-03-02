package race

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
}

func TestGetRace(t *testing.T) {
	race := GetRace("123")
	if race == nil {
		t.Error("expected race to be created")
		return
	}
	savedRace := currentRaces["123"]
	if savedRace == nil {
		t.Error("expected race to be found")
	}

	racers := GetRacers("123", "clash")
	if len(racers) < 2 {
		for i, racer := range racers {
			t.Error("racer: ", i, " ", racer)
		}
		t.Error("expected at least 2 racers")
	}

	racer1 := &RaceParticipant{
		Member: &RaceMember{
			GuildID:  "123",
			MemberID: "456",
		},
		Racer: racers[0],
	}

	racer2 := &RaceParticipant{
		Member: &RaceMember{
			GuildID:  "123",
			MemberID: "789",
		},
		Racer: racers[1],
	}
	race.addRacer(racer1)
	race.addRacer(racer2)

	race.RunRace(60)

	result := race.RaceResult
	if result.Win != nil {

	}
	if result.Place != nil {

	}
	if result.Show != nil {

	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	db.Delete(RACE_MEMBER_COLLECTION, filter)
	filter = bson.M{"guild_id": "123", "member_id": "789"}
	db.Delete(RACE_MEMBER_COLLECTION, filter)
	filter = bson.M{"guild_id": "123", "theme": "clash"}
	db.DeleteMany(RACER_COLLECTION, filter)
	filter = bson.M{"guild_id": "123"}
	db.Delete(RACE_CONFIG_COLLECTION, filter)
}
