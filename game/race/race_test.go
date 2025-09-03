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

func TestCalculateWinnings(t *testing.T) {
	// Create a race
	race := GetRace("123")
	if race == nil {
		t.Error("expected race to be created")
		return
	}

	// Add participants
	member1 := &RaceMember{
		GuildID:  "123",
		MemberID: "456",
	}
	member2 := &RaceMember{
		GuildID:  "123",
		MemberID: "789",
	}
	member3 := &RaceMember{
		GuildID:  "123",
		MemberID: "101",
	}

	racer1 := race.addRaceParticipant(member1)
	racer2 := race.addRaceParticipant(member2)
	racer3 := race.addRaceParticipant(member3)

	// Add a better
	better := &RaceMember{
		GuildID:  "123",
		MemberID: "202",
	}

	// Create a bet on racer1
	raceBetter := getRaceBetter(better, racer1)
	race.addBetter(raceBetter)

	// Create the final race leg with positions
	raceLeg := &RaceLeg{
		ParticipantPositions: []*RaceParticipantPosition{
			{
				RaceParticipant: racer1,
				Position:        0,
				Movement:        60,
				Speed:           10.5, // Fastest (winner)
				Turn:            10,
				Finished:        true,
			},
			{
				RaceParticipant: racer2,
				Position:        0,
				Movement:        60,
				Speed:           12.3, // Second place
				Turn:            10,
				Finished:        true,
			},
			{
				RaceParticipant: racer3,
				Position:        0,
				Movement:        60,
				Speed:           15.7, // Third place
				Turn:            10,
				Finished:        true,
			},
		},
	}

	// Calculate winnings
	calculateWinnings(race, raceLeg)

	// Verify results
	if race.RaceResult.Win == nil {
		t.Error("expected a winner")
	} else if race.RaceResult.Win.Participant != racer1 {
		t.Error("expected racer1 to win")
	}

	if race.RaceResult.Place == nil {
		t.Error("expected a second place")
	} else if race.RaceResult.Place.Participant != racer2 {
		t.Error("expected racer2 to place")
	}

	if race.RaceResult.Show == nil {
		t.Error("expected a third place")
	} else if race.RaceResult.Show.Participant != racer3 {
		t.Error("expected racer3 to show")
	}

	// Verify bet winnings
	if raceBetter.Winnings == 0 {
		t.Error("expected better to win")
	}

	// Cleanup
	ResetRace("123")
	filter := bson.M{"guild_id": "123", "member_id": "456"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "456"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123", "member_id": "789"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "789"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123", "member_id": "101"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "101"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123", "member_id": "202"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "202"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123"}
	if err := db.Delete(RaceConfigCollection, filter); err != nil {
		slog.Error("Error deleting race config",
			slog.String("guildID", "123"),
			slog.Any("error", err),
		)
	}
}

func TestRaceChecks(t *testing.T) {
	// Test raceStartChecks
	err := raceStartChecks("123", "456")
	if err != nil {
		t.Error("expected no error from raceStartChecks for new race, got:", err)
	}

	// Create a race
	race := GetRace("123")
	if race == nil {
		t.Error("expected race to be created")
		return
	}

	// Add a participant
	member1 := &RaceMember{
		GuildID:  "123",
		MemberID: "456",
	}
	race.addRaceParticipant(member1)

	// Test raceJoinChecks
	err = raceJoinChecks(race, "789")
	if err != nil {
		t.Error("expected no error from raceJoinChecks for new member, got:", err)
	}

	// Test raceJoinChecks with existing member (should fail)
	err = raceJoinChecks(race, "456")
	if err == nil {
		t.Error("expected error from raceJoinChecks for existing member")
	}

	// Note: We're not testing raceBetChecks here because it depends on timing
	// In a real test environment, we would need to mock time.Now() to properly test this

	// Cleanup
	ResetRace("123")
	filter := bson.M{"guild_id": "123", "member_id": "456"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "456"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123"}
	if err := db.Delete(RaceConfigCollection, filter); err != nil {
		slog.Error("Error deleting race config",
			slog.String("guildID", "123"),
			slog.Any("error", err),
		)
	}
}

func TestResetRace(t *testing.T) {
	// Create a race
	race := GetRace("123")
	if race == nil {
		t.Error("expected race to be created")
		return
	}

	// Verify the race exists in the currentRaces map
	if currentRaces["123"] == nil {
		t.Error("expected race to be in currentRaces map")
		return
	}

	// Reset the race
	ResetRace("123")

	// Verify the race has been removed from the currentRaces map
	if currentRaces["123"] != nil {
		t.Error("expected race to be removed from currentRaces map")
	}

	// Cleanup
	filter := bson.M{"guild_id": "123"}
	if err := db.Delete(RaceConfigCollection, filter); err != nil {
		slog.Error("Error deleting race config",
			slog.String("guildID", "123"),
			slog.Any("error", err),
		)
	}
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

	racers := GetRaceAvatars("123", "clash")
	if len(racers) < 2 {
		for i, racer := range racers {
			t.Error("racer: ", i, " ", racer)
		}
		t.Error("expected at least 2 racers")
	}

	member1 := &RaceMember{
		GuildID:  "123",
		MemberID: "456",
	}
	member2 := &RaceMember{
		GuildID:  "123",
		MemberID: "789",
	}
	race.addRaceParticipant(member1)
	race.addRaceParticipant(member2)

	race.RunRace(60)

	result := race.RaceResult
	if result.Win == nil {
		t.Error("expected a winner")
	}
	if result.Place == nil {
		t.Error("expected a second place")
	}
	if result.Show != nil {
		// With only 2 participants, there should be no show position
		t.Error("expected no show position with only 2 participants")
	}

	// Verify the race results have been calculated
	if result.Win == nil || result.Place == nil {
		t.Error("expected race results to be calculated")
	}

	// Cleanup
	filter := bson.M{"guild_id": "123", "member_id": "456"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "456"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123", "member_id": "789"}
	if err := db.Delete(RaceMemberCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "789"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123", "theme": "clash"}
	if err := db.DeleteMany(RacerCollection, filter); err != nil {
		slog.Error("Error deleting race member",
			slog.String("guildID", "123"),
			slog.String("memberID", "789"),
			slog.Any("error", err),
		)
	}
	filter = bson.M{"guild_id": "123"}
	if err := db.Delete(RaceConfigCollection, filter); err != nil {
		slog.Error("Error deleting race config",
			slog.String("guildID", "123"),
			slog.Any("error", err),
		)
	}
}
