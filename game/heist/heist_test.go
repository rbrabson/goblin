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
	GuildId     = "12345"
	OrganizerId = "12345"
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

	organizer := guild.GetMember(GuildId, OrganizerId)
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}

	if heist == nil {
		t.Errorf("Expected heist, got nil")
		return
	}
	defer heist.End()

	if heist.GuildID != GuildId {
		t.Errorf("Expected %s, got %s", GuildId, heist.GuildID)
	}
	if heist.Organizer.MemberID != OrganizerId {
		t.Errorf("Expected %s, got %s", OrganizerId, heist.Organizer.MemberID)
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

	organizer := guild.GetMember(GuildId, OrganizerId)
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GuildId, OrganizerId)
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

	member := getHeistMember(GuildId, "abcdef")
	member.guildMember.SetName("Crew Member 1", "", "")
	err = heistChecks(heist, member)
	if err != nil {
		t.Errorf("Got %s", err.Error())
		return
	}
	if err := heist.AddCrewMember(member); err != nil {
		slog.Error("error adding crew member to heist",
			slog.String("guildID", heist.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("err", err),
		)
	}
	err = heistChecks(heist, member)
	if err == nil {
		t.Errorf("Expected non-nil, got nil error")
		return
	}
}

func TestStartHeist(t *testing.T) {
	testSetup()
	defer testTeardown()

	organizer := guild.GetMember(GuildId, OrganizerId).SetName("Organizer", "", "")
	if organizer == nil {
		t.Errorf("Expected organizer, got nil")
		return
	}
	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	if heist == nil {
		t.Errorf("Expected heist, got nil")
		return
	}
	defer heist.End()

	member := getHeistMember(GuildId, "abcdef")
	member.guildMember.SetName("Crew Member 1", "", "")
	if err := heist.AddCrewMember(member); err != nil {
		slog.Error("error adding crew member to heist",
			slog.String("guildID", heist.GuildID),
			slog.String("memberID", member.MemberID),
			slog.Any("err", err),
		)
	}

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
	if err := db.DeleteMany(guild.MemberCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all members",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(bank.AccountCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all account",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(bank.BankCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all banks",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(ConfigCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all configs",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(HeistMemberCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all heist members",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(TargetCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all targets",
			slog.Any("err", err),
		)
	}
	if err := db.DeleteMany(ThemeCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all themes",
			slog.Any("err", err),
		)
	}
	delete(alertTimes, GuildId)
}
