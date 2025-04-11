package heist

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

const (
	GUILD_ID     = "12345"
	ORGANIZER_ID = "12345"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
	guild.SetDB(db)
	bank.SetDB(db)
}

func TestNewHeist(t *testing.T) {
	testSetup()
	defer testTeardown()

	organizer := guild.GetMember(GUILD_ID, ORGANIZER_ID)
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GUILD_ID, ORGANIZER_ID)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}

	if heist == nil {
		t.Errorf("Expected heist, got nil")
		return
	}
	defer heist.End()

	if heist.GuildID != GUILD_ID {
		t.Errorf("Expected %s, got %s", GUILD_ID, heist.GuildID)
	}
	if heist.Organizer.MemberID != ORGANIZER_ID {
		t.Errorf("Expected %s, got %s", ORGANIZER_ID, heist.Organizer.MemberID)
	}
	if len(heist.Crew) != 1 {
		t.Errorf("Expected 1, got %d", len(heist.Crew))
	}
	if heist.StartTime.IsZero() {
		t.Errorf("Expected non-zero start time")
	}
}

func TestHeistChecks(t *testing.T) {
	testSetup()
	defer testTeardown()

	organizer := guild.GetMember(GUILD_ID, ORGANIZER_ID)
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GUILD_ID, ORGANIZER_ID)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	err = heistChecks(heist, heist.Organizer)
	if err == nil {
		t.Errorf("Expected non-nil, got nil error")
		return
	}
	defer heist.End()

	member := getHeistMember(GUILD_ID, "abcdef")
	member.guildMember.SetName("Crew Member 1", "", "")
	err = heistChecks(heist, member)
	if err != nil {
		t.Errorf("Got %s", err.Error())
		return
	}
	heist.AddCrewMember(member)
	err = heistChecks(heist, member)
	if err == nil {
		t.Errorf("Expected non-nil, got nil error")
		return
	}
}

func TestStartHeist(t *testing.T) {
	testSetup()
	defer testTeardown()

	organizer := guild.GetMember(GUILD_ID, ORGANIZER_ID).SetName("Organizer", "", "")
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GUILD_ID, ORGANIZER_ID)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	if heist == nil {
		t.Errorf("Expected heist, got nil")
		return
	}
	defer heist.End()

	member := getHeistMember(GUILD_ID, "abcdef")
	member.guildMember.SetName("Crew Member 1", "", "")
	heist.AddCrewMember(member)

	res, err := heist.Start()
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	if len(res.AllResults) != 2 {
		t.Errorf("Expected 2, got %d", len(res.AllResults))
	}
}

func testSetup() {}

func testTeardown() {
	db.DeleteMany(guild.MEMBER_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(bank.ACCOUNT_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(bank.BANK_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(CONFIG_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(HEIST_MEMBER_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(TARGET_COLLECTION, bson.M{"guild_id": GUILD_ID})
	db.DeleteMany(THEME_COLLECTION, bson.M{"guild_id": GUILD_ID})
	delete(alertTimes, GUILD_ID)
}
