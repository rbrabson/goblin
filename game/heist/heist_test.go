package heist

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

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
			slog.Any("error", err),
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
			slog.Any("error", err),
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

func TestHeistMemberApprehended(t *testing.T) {
	testSetup()
	defer testTeardown()

	member := getHeistMember(GuildId, "apprehended_test")
	member.guildMember.SetName("Apprehended Test", "", "")

	// Initial state check
	if member.Status != FREE {
		t.Errorf("Expected initial status to be Free, got %s", member.Status)
	}

	// Set up heist for the member
	heist, _ := NewHeist(GuildId, OrganizerId)
	member.heist = heist
	defer heist.End()

	// Test apprehension
	member.Apprehended()

	// Verify status changed to Apprehended
	if member.Status != Apprehended {
		t.Errorf("Expected status to be Apprehended, got %s", member.Status)
	}

	// Verify jail time is set
	if member.JailTimer.IsZero() {
		t.Errorf("Expected jail timer to be set")
	}

	// Verify remaining jail time is positive
	if member.RemainingJailTime() <= 0 {
		t.Errorf("Expected remaining jail time to be positive, got %v", member.RemainingJailTime())
	}
}

func TestHeistMemberDied(t *testing.T) {
	testSetup()
	defer testTeardown()

	member := getHeistMember(GuildId, "died_test")
	member.guildMember.SetName("Died Test", "", "")

	// Initial state check
	if member.Status != FREE {
		t.Errorf("Expected initial status to be Free, got %s", member.Status)
	}

	// Set up heist for the member
	heist, _ := NewHeist(GuildId, OrganizerId)
	member.heist = heist
	defer heist.End()

	// Test death
	member.Died()

	// Verify status changed to Dead
	if member.Status != Dead {
		t.Errorf("Expected status to be Dead, got %s", member.Status)
	}

	// Verify death time is set
	if member.DeathTimer.IsZero() {
		t.Errorf("Expected death timer to be set")
	}

	// Verify remaining death time is positive
	if member.RemainingDeathTime() <= 0 {
		t.Errorf("Expected remaining death time to be positive, got %v", member.RemainingDeathTime())
	}
}

func TestHeistMemberEscaped(t *testing.T) {
	testSetup()
	defer testTeardown()

	member := getHeistMember(GuildId, "escaped_test")
	member.guildMember.SetName("Escaped Test", "", "")

	// Set initial criminal level
	initialLevel := member.CriminalLevel

	// Set up heist for the member
	heist, _ := NewHeist(GuildId, OrganizerId)
	member.heist = heist
	defer heist.End()

	// Test escape
	member.Escaped()

	// Verify status is still Free
	if member.Status != FREE {
		t.Errorf("Expected status to be Free, got %s", member.Status)
	}

	// Verify criminal level increased
	if member.CriminalLevel <= initialLevel {
		t.Errorf("Expected criminal level to increase from %d, got %d", initialLevel, member.CriminalLevel)
	}
}

func TestHeistMemberUpdateStatus(t *testing.T) {
	testSetup()
	defer testTeardown()

	member := getHeistMember(GuildId, "status_test")
	member.guildMember.SetName("Status Test", "", "")

	// Set up heist for the member
	heist, _ := NewHeist(GuildId, OrganizerId)
	member.heist = heist
	defer heist.End()

	// Test with member in jail
	member.Apprehended()

	// Set jail time to past
	member.JailTimer = time.Now().Add(-1 * time.Hour)

	// Update status
	member.UpdateStatus()

	// Verify status changed back to FREE
	if member.Status != FREE {
		t.Errorf("Expected status to be Free after jail time expired, got %s", member.Status)
	}

	// Test with dead member
	member.Died()

	// Set death time to past
	member.DeathTimer = time.Now().Add(-1 * time.Hour)

	// Update status
	member.UpdateStatus()

	// Verify status changed back to FREE
	if member.Status != FREE {
		t.Errorf("Expected status to be Free after death time expired, got %s", member.Status)
	}
}

func TestHeistMemberClearJailAndDeathStatus(t *testing.T) {
	testSetup()
	defer testTeardown()

	member := getHeistMember(GuildId, "clear_status_test")
	member.guildMember.SetName("Clear Status Test", "", "")

	// Set up heist for the member
	heist, _ := NewHeist(GuildId, OrganizerId)
	member.heist = heist
	defer heist.End()

	// Set member to jail
	member.Apprehended()

	// Clear status
	member.ClearJailAndDeathStatus()

	// Verify status is FREE
	if member.Status != FREE {
		t.Errorf("Expected status to be Free after clearing, got %s", member.Status)
	}

	// Verify jail time is cleared
	if !member.JailTimer.IsZero() {
		t.Errorf("Expected jail timer to be cleared")
	}

	// Set member to dead
	member.Died()

	// Clear status
	member.ClearJailAndDeathStatus()

	// Verify status is FREE
	if member.Status != FREE {
		t.Errorf("Expected status to be Free after clearing, got %s", member.Status)
	}

	// Verify death time is cleared
	if !member.DeathTimer.IsZero() {
		t.Errorf("Expected death timer to be cleared")
	}
}

func testSetup() {
	// Create a default target for testing
	target := &Target{
		GuildID:  GuildId,
		Theme:    HeistDefaultTheme,
		Name:     "Test Target",
		CrewSize: 5,
		Success:  0.5,
		Vault:    1000,
		VaultMax: 1000,
		IsAtMax:  true,
	}
	writeTarget(target)

	// Create a default theme for testing
	theme := &Theme{
		GuildID: GuildId,
		Name:    "Test Theme",
		EscapedMessages: []*HeistMessage{
			{
				Message:     "{} escaped with the loot!",
				BonusAmount: 0,
				Result:      Escaped,
			},
		},
		ApprehendedMessages: []*HeistMessage{
			{
				Message:     "{} was caught by the police!",
				BonusAmount: 0,
				Result:      Apprehended,
			},
		},
		DiedMessages: []*HeistMessage{
			{
				Message:     "{} died during the heist!",
				BonusAmount: 0,
				Result:      Dead,
			},
		},
		Jail:     "Jail",
		OOB:      "Out on Bail",
		Police:   "Police",
		Bail:     "Bail",
		Crew:     "Crew",
		Sentence: "Sentence",
		Heist:    "Heist",
		Vault:    "Vault",
	}
	writeTheme(theme)
}

func TestGetTargets(t *testing.T) {
	testSetup()
	defer testTeardown()

	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil {
		t.Errorf("Expected targets, got nil")
		return
	}

	// Verify that our test target is in the list
	found := false
	for _, target := range targets {
		if target.GuildID == GuildId && target.Name == "Test Target" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find test target in targets list")
	}
}

func TestStealFromVault(t *testing.T) {
	testSetup()
	defer testTeardown()

	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil || len(targets) == 0 {
		t.Errorf("Expected targets, got nil or empty")
		return
	}

	// Find our test target
	var target *Target
	for _, tgt := range targets {
		if tgt.GuildID == GuildId && tgt.Name == "Test Target" {
			target = tgt
			break
		}
	}

	if target == nil {
		t.Errorf("Could not find test target")
		return
	}

	// Initial vault amount
	initialVault := target.Vault

	// Steal half the vault
	stealAmount := initialVault / 2
	target.StealFromValut(stealAmount)

	// Verify vault amount decreased
	if target.Vault != initialVault - stealAmount {
		t.Errorf("Expected vault to be %d, got %d", initialVault - stealAmount, target.Vault)
	}

	// Verify IsAtMax is false
	if target.IsAtMax {
		t.Errorf("Expected IsAtMax to be false after stealing")
	}

	// Steal more than what's in the vault
	target.StealFromValut(target.Vault + 100)

	// Verify vault is 0
	if target.Vault != 0 {
		t.Errorf("Expected vault to be 0 after stealing more than available, got %d", target.Vault)
	}
}

func TestGetTarget(t *testing.T) {
	testSetup()
	defer testTeardown()

	// Create multiple targets with different crew sizes
	targets := []*Target{
		{
			GuildID:  GuildId,
			Theme:    HeistDefaultTheme,
			Name:     "Small Target",
			CrewSize: 3,
			Success:  0.7,
			Vault:    500,
			VaultMax: 500,
		},
		{
			GuildID:  GuildId,
			Theme:    HeistDefaultTheme,
			Name:     "Medium Target",
			CrewSize: 5,
			Success:  0.5,
			Vault:    1000,
			VaultMax: 1000,
		},
		{
			GuildID:  GuildId,
			Theme:    HeistDefaultTheme,
			Name:     "Large Target",
			CrewSize: 10,
			Success:  0.3,
			Vault:    2000,
			VaultMax: 2000,
		},
	}

	// Test with crew size smaller than smallest target
	target := getTarget(targets, 2)
	if target.Name != "Small Target" {
		t.Errorf("Expected Small Target for crew size 2, got %s", target.Name)
	}

	// Test with crew size between targets
	target = getTarget(targets, 4)
	if target.Name != "Medium Target" {
		t.Errorf("Expected Medium Target for crew size 4, got %s", target.Name)
	}

	// Test with crew size larger than largest target
	target = getTarget(targets, 15)
	if target.Name != "Large Target" {
		t.Errorf("Expected Large Target for crew size 15, got %s", target.Name)
	}
}

func TestResetVaultsToMaximumValue(t *testing.T) {
	testSetup()
	defer testTeardown()

	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil || len(targets) == 0 {
		t.Errorf("Expected targets, got nil or empty")
		return
	}

	// Find our test target
	var target *Target
	for _, tgt := range targets {
		if tgt.GuildID == GuildId && tgt.Name == "Test Target" {
			target = tgt
			break
		}
	}

	if target == nil {
		t.Errorf("Could not find test target")
		return
	}

	// Steal from vault to reduce it
	target.StealFromValut(500)

	// Verify vault is reduced
	if target.Vault >= target.VaultMax {
		t.Errorf("Expected vault to be less than max after stealing")
		return
	}

	// Reset vaults
	ResetVaultsToMaximumValue(GuildId)

	// Get updated target
	targets = GetTargets(GuildId, HeistDefaultTheme)
	for _, tgt := range targets {
		if tgt.GuildID == GuildId && tgt.Name == "Test Target" {
			target = tgt
			break
		}
	}

	// Verify vault is reset to max
	if target.Vault != target.VaultMax {
		t.Errorf("Expected vault to be reset to max value %d, got %d", target.VaultMax, target.Vault)
	}

	// Verify IsAtMax is true
	if !target.IsAtMax {
		t.Errorf("Expected IsAtMax to be true after reset")
	}
}

func TestGetThemeNames(t *testing.T) {
	testSetup()
	defer testTeardown()

	themeNames, err := GetThemeNames(GuildId)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err.Error())
		return
	}

	if themeNames == nil {
		t.Errorf("Expected theme names, got nil")
		return
	}

	// Verify that our test theme is in the list
	found := false
	for _, name := range themeNames {
		if name == "Test Theme" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find test theme in theme names list")
	}
}

func TestGetThemes(t *testing.T) {
	testSetup()
	defer testTeardown()

	themes := GetThemes(GuildId)
	if themes == nil {
		t.Errorf("Expected themes, got nil")
		return
	}

	// Verify that our test theme is in the list
	found := false
	for _, theme := range themes {
		if theme.GuildID == GuildId && theme.Name == "Test Theme" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find test theme in themes list")
	}
}

func TestGetTheme(t *testing.T) {
	testSetup()
	defer testTeardown()

	// Set up config to use our test theme
	config := GetConfig(GuildId)
	config.Theme = "Test Theme"
	writeConfig(config)

	// Get the theme
	theme := GetTheme(GuildId)
	if theme == nil {
		t.Errorf("Expected theme, got nil")
		return
	}

	// Verify it's our test theme
	if theme.Name != "Test Theme" {
		t.Errorf("Expected Test Theme, got %s", theme.Name)
	}

	// Verify theme properties
	if len(theme.EscapedMessages) != 1 {
		t.Errorf("Expected 1 escaped message, got %d", len(theme.EscapedMessages))
	}

	if len(theme.ApprehendedMessages) != 1 {
		t.Errorf("Expected 1 apprehended message, got %d", len(theme.ApprehendedMessages))
	}

	if len(theme.DiedMessages) != 1 {
		t.Errorf("Expected 1 died message, got %d", len(theme.DiedMessages))
	}

	if theme.Jail != "Jail" {
		t.Errorf("Expected Jail, got %s", theme.Jail)
	}
}

func TestHeistEnd(t *testing.T) {
	testSetup()
	defer testTeardown()

	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}

	// Add a crew member
	member := getHeistMember(GuildId, "end_test")
	member.guildMember.SetName("End Test", "", "")
	if err := heist.AddCrewMember(member); err != nil {
		t.Errorf("Failed to add crew member: %s", err.Error())
		return
	}

	// End the heist
	heist.End()

	// Verify the heist is removed from the active heists
	if GetHeist(GuildId) != nil {
		t.Errorf("Expected heist to be removed from active heists")
	}
}

func TestHeistCancel(t *testing.T) {
	testSetup()
	defer testTeardown()

	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}

	// Add a crew member
	member := getHeistMember(GuildId, "cancel_test")
	member.guildMember.SetName("Cancel Test", "", "")
	if err := heist.AddCrewMember(member); err != nil {
		t.Errorf("Failed to add crew member: %s", err.Error())
		return
	}

	// Cancel the heist
	heist.Cancel()

	// Verify the heist is removed from the active heists
	if GetHeist(GuildId) != nil {
		t.Errorf("Expected heist to be removed from active heists")
	}
}

func TestCalculateSuccessRate(t *testing.T) {
	testSetup()
	defer testTeardown()

	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	defer heist.End()

	// Add crew members
	for i := 0; i < 3; i++ {
		member := getHeistMember(GuildId, fmt.Sprintf("success_test_%d", i))
		member.guildMember.SetName(fmt.Sprintf("Success Test %d", i), "", "")
		if err := heist.AddCrewMember(member); err != nil {
			t.Errorf("Failed to add crew member: %s", err.Error())
			return
		}
	}

	// Get a target
	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil || len(targets) == 0 {
		t.Errorf("Expected targets, got nil or empty")
		return
	}

	// Calculate success rate
	successRate := calculateSuccessRate(heist, targets[0])

	// Verify success rate is within expected range
	if successRate < 0 || successRate > 100 {
		t.Errorf("Expected success rate between 0 and 100, got %d", successRate)
	}
}

func TestCalculateBonusRate(t *testing.T) {
	testSetup()
	defer testTeardown()

	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	defer heist.End()

	// Add crew members
	for i := 0; i < 3; i++ {
		member := getHeistMember(GuildId, fmt.Sprintf("bonus_test_%d", i))
		member.guildMember.SetName(fmt.Sprintf("Bonus Test %d", i), "", "")
		if err := heist.AddCrewMember(member); err != nil {
			t.Errorf("Failed to add crew member: %s", err.Error())
			return
		}
	}

	// Get a target
	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil || len(targets) == 0 {
		t.Errorf("Expected targets, got nil or empty")
		return
	}

	// Calculate bonus rate
	bonusRate := calculateBonusRate(heist, targets[0])

	// Verify bonus rate is within expected range
	if bonusRate < 0 || bonusRate > 100 {
		t.Errorf("Expected bonus rate between 0 and 100, got %d", bonusRate)
	}
}

func TestCalculateCredits(t *testing.T) {
	testSetup()
	defer testTeardown()

	heist, err := NewHeist(GuildId, OrganizerId)
	if err != nil {
		t.Errorf("Expected nil, got %s", err.Error())
		return
	}
	defer heist.End()

	// Add crew members
	for i := 0; i < 3; i++ {
		member := getHeistMember(GuildId, fmt.Sprintf("credits_test_%d", i))
		member.guildMember.SetName(fmt.Sprintf("Credits Test %d", i), "", "")
		if err := heist.AddCrewMember(member); err != nil {
			t.Errorf("Failed to add crew member: %s", err.Error())
			return
		}
	}

	// Get a target
	targets := GetTargets(GuildId, HeistDefaultTheme)
	if targets == nil || len(targets) == 0 {
		t.Errorf("Expected targets, got nil or empty")
		return
	}

	// Create a heist result
	result := &HeistResult{
		Target:     targets[0],
		TotalStolen: 1000,
		AllResults: make([]*HeistMemberResult, 0, len(heist.Crew)),
		heist:      heist,
	}

	// Add member results
	for _, member := range heist.Crew {
		memberResult := &HeistMemberResult{
			Player:  member,
			Status:  string(Escaped),
			Message: "{} escaped with the loot!",
			heist:   heist,
		}
		result.AllResults = append(result.AllResults, memberResult)
	}

	// Calculate credits
	calculateCredits(result)

	// Verify credits are calculated
	totalCredits := 0
	for _, memberResult := range result.AllResults {
		totalCredits += memberResult.StolenCredits + memberResult.BonusCredits
	}

	// Verify total credits don't exceed vault amount
	if totalCredits > result.TotalStolen {
		t.Errorf("Expected total credits (%d) to not exceed total stolen amount (%d)", totalCredits, result.TotalStolen)
	}
}

func testTeardown() {
	if err := db.DeleteMany(guild.MemberCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all members",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(bank.AccountCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all account",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(bank.BankCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all banks",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(ConfigCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all configs",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(HeistMemberCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all heist members",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(TargetCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all targets",
			slog.Any("error", err),
		)
	}
	if err := db.DeleteMany(ThemeCollection, bson.M{"guild_id": GuildId}); err != nil {
		slog.Error("error deleting all themes",
			slog.Any("error", err),
		)
	}
	delete(alertTimes, GuildId)
}
