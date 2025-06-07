package bank

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestDeposit(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
	if err := account.SetBalance(0); err != nil {
		slog.Error("error setting balance to 0",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if err := account.Deposit(100); err != nil {
		slog.Error("error depositing to 100",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestWithdraw(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
	if err := account.SetBalance(200); err != nil {
		slog.Error("error setting balance to 200",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if err := account.Withdraw(100); err != nil {
		slog.Error("error withdrawing to 100",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestSetBalance(t *testing.T) {
	bank := GetBank("12345")
	account := GetAccount(bank.GuildID, "54321")
	if err := account.SetBalance(500); err != nil {
		slog.Error("error setting balance to 500",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	account = GetAccount(bank.GuildID, "54321")
	if account.CurrentBalance != 500 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestResetMonthlyBalances(t *testing.T) {
	bank := GetBank("12345")
	account := GetAccount(bank.GuildID, "54321")
	account.CurrentBalance = 500
	account.MonthlyBalance = 750
	if err := writeAccount(account); err != nil {
		slog.Error("error writing account",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	ResetMonthlyBalances()

	account = GetAccount(bank.GuildID, "54321")
	if account.MonthlyBalance != 0 {
		t.Errorf("Expected monthly balance to be 0, got %d", account.MonthlyBalance)
	}
}
