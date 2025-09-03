package race

import (
	"log/slog"
	"os"
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
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
	bank.SetDB(db)
	guild.SetDB(db)
}

func TestGetRaceMember(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestRaceMemberWinRace(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberPlacedInRace(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberHowedInRace(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceMemberLoseRace(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceBetOnRace(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	if err := member.PlaceBet(100); err != nil {
		slog.Error("error placing bet",
			slog.String("guildID", member.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Int("amount", 100),
			slog.Any("error", err),
		)
	}
	raceMember := readRaceMember("123", "456")
	if raceMember == nil {
		t.Error("expected member to be found")
		return
	}

	if member.BetsMade != 1 {
		t.Error("expected BetsMade to be 1, got ", member.BetsMade)
	}
	if member.TotalEarnings != -100 {
		t.Error("expected TotalEarnings to be 0, got ", member.TotalEarnings)
	}

	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceWinBet(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
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
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}

func TestRaceLoseBet(t *testing.T) {
	guildMember := guild.GetMember("123", "456").SetName("TestUser", "TestNick", "TestGlobalName")
	member := GetRaceMember("123", guildMember)
	if member == nil {
		t.Error("expected member to be created")
		return
	}

	// Place a bet first
	if err := member.PlaceBet(100); err != nil {
		t.Error("error placing bet:", err)
		return
	}

	// Initial state after placing bet
	if member.BetsMade != 1 {
		t.Error("expected BetsMade to be 1, got", member.BetsMade)
	}
	if member.TotalEarnings != -100 {
		t.Error("expected TotalEarnings to be -100, got", member.TotalEarnings)
	}

	// Lose the bet
	member.LoseBet()

	// Verify state after losing bet
	member = readRaceMember("123", "456")
	if member == nil {
		t.Error("expected member to be found")
		return
	}

	// BetsMade should still be 1
	if member.BetsMade != 1 {
		t.Error("expected BetsMade to be 1, got", member.BetsMade)
	}

	// BetsWon should be 0
	if member.BetsWon != 0 {
		t.Error("expected BetsWon to be 0, got", member.BetsWon)
	}

	// TotalEarnings should still be -100 (no change from losing)
	if member.TotalEarnings != -100 {
		t.Error("expected TotalEarnings to be -100, got", member.TotalEarnings)
	}

	// Cleanup
	filter := bson.M{"guild_id": "123", "member_id": "456"}
	err := db.Delete(RaceMemberCollection, filter)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(bank.AccountCollection, filter)
	if err != nil {
		t.Error(err)
	}
}
