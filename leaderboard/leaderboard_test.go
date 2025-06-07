package leaderboard

import (
	"log"
	"log/slog"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/database/mongo"
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
					slog.Any("err", err),
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
