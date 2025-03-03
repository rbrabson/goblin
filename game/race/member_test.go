package race

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
	bank.SetDB(db)
	guild.SetDB(db)
}

func TestGetRaceMember(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.RacesWon = 1
	writeRaceMember(member)

	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}
	if member.RacesWon != 1 {
		t.Error("expected RacesWon to be 1")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestRaceMemberWinRace(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.WinRace(100)
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.RacesWon != 1 {
		t.Error("expected RacesWon to be 1")
	}
	if member.TotalEarnings != 100 {
		t.Error("expected TotalEarnings to be 100")
	}
	if member.RacesWon != 1 {
		t.Error("expected RacesWon to be 1")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberPlacedInRace(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.PlaceInRace(100)
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.RacesPlaced != 1 {
		t.Error("expected RacesWon to be 1")
	}
	if member.TotalEarnings != 100 {
		t.Error("expected TotalEarnings to be 100")
	}
	if member.RacesWon != 0 {
		t.Error("expected RacesWon to be 0")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberHowedInRace(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.ShowInRace(100)
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.RacesShowed != 1 {
		t.Error("expected RacesShowed to be 1")
	}
	if member.TotalEarnings != 100 {
		t.Error("expected TotalEarnings to be 100")
	}
	if member.RacesWon != 0 {
		t.Error("expected RacesWon to be 0")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberLoseRace(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.LoseRace()
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.RacesWon != 0 {
		t.Error("expected RacesWon to be 0")
	}
	if member.TotalEarnings != 0 {
		t.Error("expected TotalEarnings to be 0")
	}
	if member.RacesLost != 1 {
		t.Error("expected RacesLost to be 1")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceBetOnRace(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.PlaceBet(100)
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.BetsMade != 1 {
		t.Error("expected BetsMade to be 1")
	}
	if member.TotalEarnings != -100 {
		t.Error("expected TotalEarnings to be -100")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceWinBet(t *testing.T) {
	member := GetRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	member.WinBet(100)
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	if member.BetsWon != 1 {
		t.Error("expected BetsMade to be 1")
	}
	if member.BetsEarnings != 100 {
		t.Error("expected BetsEarnings to be 100")
	}
	if member.TotalEarnings != 100 {
		t.Error("expected TotalEarnings to be 100")
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RACE_MEMBER_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.ACCOUNT_COLLECTION, filter)
	if err != nil {
		t.Error(err)
	}
}
