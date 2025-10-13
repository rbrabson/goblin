package leaderboard

import (
	"log"
	"log/slog"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = mongo.NewDatabase()
}

func TestNewLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)
}

func TestGetLeaderboards(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the leaderboards
	lbs := getLeaderboards()
	if lbs == nil {
		t.Errorf("GetLeaderboards() returned nil")
		return
	}
}

func TestGetLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the leaderboard
	lb = getLeaderboard("12345")
	if lb == nil {
		t.Errorf("GetLeaderboard() returned nil")
		return
	}
}

func TestSetChannel(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	// Test SetChannel
	// Create a new leaderboard
	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Set the channel
	lb.setChannel("54321")
	lb = getLeaderboard(lb.GuildID)
	if lb.ChannelID != "54321" {
		t.Errorf("SetChannel() failed")
		return
	}
}

func TestGetCurrentLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}
	banks = append(banks, b)

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the monthly leaderboard
	accounts := lb.getCurrentLeaderboard()
	bankAccounts = append(bankAccounts, accounts...)
	if accounts == nil {
		t.Errorf("GetCurrentLeaderboard() returned nil")
		return
	}
	if len(accounts) != 1 {
		t.Errorf("GetCurrentLeaderboard() returned an empty array")
		return
	}
}

func TestGetMonthlyLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the monthly leaderboard
	accounts := lb.getMonthlyLeaderboard()
	bankAccounts = append(bankAccounts, accounts...)
	if accounts == nil {
		t.Errorf("GetMonthlyLeaderboard() returned nil")
		return
	}
	if len(accounts) != 1 {
		t.Errorf("GetMonthlyLeaderboard() returned an empty array")
		return
	}
}

func TestGetLifetimeLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}
	banks = append(banks, b)

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the Lifetime leaderboard
	accounts := lb.getLifetimeLeaderboard()
	bankAccounts = append(bankAccounts, accounts...)
	if accounts == nil {
		t.Errorf("GetLifetimeLeaderboard() returned nil")
		return
	}
	if len(accounts) != 1 {
		t.Errorf("GetLifetimeLeaderboard() returned an empty array")
		return
	}
}

func TestGetCurrentRanking(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}
	banks = append(banks, b)

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the player ranking
	rank := getCurrentRanking(lb, bankAccount)
	if rank != 1 {
		t.Errorf("GetCurrentLeaderboard() returned an empty array")
		return
	}
}

func TestGetMonthlyRanking(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}
	banks = append(banks, b)

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the player ranking
	rank := getMonthlyRanking(lb, bankAccount)
	if rank != 1 {
		t.Errorf("GetMonthlyLeaderboard() returned an empty array")
		return
	}
}

func TestGetLifetimeRanking(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	banks := make([]*bank.Bank, 0, 1)
	defer func() {
		for _, b := range banks {
			if err := db.Delete(bank.BankCollection, bson.M{"guild_id": b.GuildID}); err != nil {
				slog.Error("Error deleting bank",
					slog.String("guildID", b.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	bankAccounts := make([]*bank.Account, 0, 1)
	defer func() {
		for _, account := range bankAccounts {
			if err := db.Delete(bank.AccountCollection, bson.M{"guild_id": account.GuildID, "account_id": account.MemberID}); err != nil {
				slog.Error("Error deleting bank account",
					slog.String("guildID", account.GuildID),
					slog.String("accountID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank.SetDB(db)
	b := bank.GetBank("12345")
	if b == nil {
		t.Errorf("NewBank() returned nil")
		return
	}
	banks = append(banks, b)

	bankAccount := bank.GetAccount(b.GuildID, "54321")
	if bankAccount == nil {
		t.Errorf("GetAccount() returned nil")
		return
	}
	bankAccounts = append(bankAccounts, bankAccount)

	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Get the player ranking
	rank := getLifetimeRanking(lb, bankAccount)
	if rank != 1 {
		t.Errorf("GetLifetimeLeaderboard() returned an empty array")
		return
	}
}

func TestReadLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	// Create a new leaderboard
	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Read the leaderboard from the database
	readLb := readLeaderboard("12345")
	if readLb == nil {
		t.Errorf("readLeaderboard() returned nil")
		return
	}

	// Verify the leaderboard was read correctly
	if readLb.GuildID != "12345" {
		t.Errorf("readLeaderboard() returned incorrect GuildID: got %s, want %s", readLb.GuildID, "12345")
		return
	}
}

func TestWriteLeaderboard(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	// Create a new leaderboard
	lb := &Leaderboard{
		GuildID:   "12345",
		ChannelID: "54321",
	}
	leaderboards = append(leaderboards, lb)

	// Write the leaderboard to the database
	err := writeLeaderboard(lb)
	if err != nil {
		t.Errorf("writeLeaderboard() returned error: %v", err)
		return
	}

	// Read the leaderboard from the database to verify it was written correctly
	readLb := readLeaderboard("12345")
	if readLb == nil {
		t.Errorf("readLeaderboard() returned nil after writeLeaderboard()")
		return
	}

	// Verify the leaderboard was written correctly
	if readLb.GuildID != "12345" {
		t.Errorf("writeLeaderboard() wrote incorrect GuildID: got %s, want %s", readLb.GuildID, "12345")
		return
	}
	if readLb.ChannelID != "54321" {
		t.Errorf("writeLeaderboard() wrote incorrect ChannelID: got %s, want %s", readLb.ChannelID, "54321")
		return
	}
}

func TestPlugin(t *testing.T) {
	// Create a test plugin
	testPlugin := &Plugin{}

	// Test GetName
	if testPlugin.GetName() != PluginName {
		t.Errorf("GetName() returned incorrect name: got %s, want %s", testPlugin.GetName(), PluginName)
	}

	// Test Status
	if testPlugin.Status() != discord.RUNNING {
		t.Errorf("Status() returned incorrect status: got %v, want %v", testPlugin.Status(), discord.RUNNING)
	}

	// Test GetCommands
	commands := testPlugin.GetCommands()
	if len(commands) != len(adminCommands)+len(memberCommands) {
		t.Errorf("GetCommands() returned incorrect number of commands: got %d, want %d",
			len(commands), len(adminCommands)+len(memberCommands))
	}

	// Test GetCommandHandlers
	handlers := testPlugin.GetCommandHandlers()
	if len(handlers) != len(commandHandlers) {
		t.Errorf("GetCommandHandlers() returned incorrect number of handlers: got %d, want %d",
			len(handlers), len(commandHandlers))
	}

	// Test GetComponentHandlers
	componentHandlers := testPlugin.GetComponentHandlers()
	if componentHandlers != nil {
		t.Errorf("GetComponentHandlers() returned non-nil: got %v, want nil", componentHandlers)
	}

	// Test GetHelp
	help := testPlugin.GetHelp()
	if len(help) == 0 {
		t.Errorf("GetHelp() returned empty help")
	}

	// Test GetAdminHelp
	adminHelp := testPlugin.GetAdminHelp()
	if len(adminHelp) == 0 {
		t.Errorf("GetAdminHelp() returned empty help")
	}

	// Test Stop
	testPlugin.Stop()
	if status != discord.STOPPED {
		t.Errorf("Stop() did not set status to STOPPED: got %v, want %v", status, discord.STOPPED)
	}

	// Reset status for other tests
	status = discord.RUNNING
}

func TestPluginInitialize(t *testing.T) {
	// Create a test plugin
	testPlugin := &Plugin{}

	// Create a mock bot and database
	mockBot := &discord.Bot{}
	mockDB := db

	// Test Initialize
	testPlugin.Initialize(mockBot, mockDB)

	// Verify that bot and db were set
	if bot != mockBot {
		t.Errorf("Initialize() did not set bot correctly")
	}

	// Note: We can't easily test the goroutine that's started in Initialize
	// without modifying the code to make it testable
}

// TestLeaderboardString tests the String method of the Leaderboard struct
func TestLeaderboardString(t *testing.T) {
	leaderboards := make([]*Leaderboard, 0, 1)
	defer func() {
		for _, leaderboard := range leaderboards {
			if err := db.Delete(LeaderboardCollection, bson.M{"guild_id": leaderboard.GuildID}); err != nil {
				slog.Error("Error deleting leaderboard",
					slog.String("guildID", leaderboard.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	// Create a new leaderboard
	lb := newLeaderboard("12345")
	if lb == nil {
		t.Errorf("NewLeaderboard() returned nil")
		return
	}
	leaderboards = append(leaderboards, lb)

	// Test the String method
	str := lb.String()
	if str == "" {
		t.Errorf("String() returned empty string")
		return
	}

	// Verify the string contains the expected information
	if !strings.Contains(str, "12345") {
		t.Errorf("String() does not contain GuildID: %s", str)
		return
	}
}
