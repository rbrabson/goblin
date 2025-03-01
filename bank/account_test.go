package bank

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
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

func TestDeposit(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			db.Delete(BANK_COLLECTION, bson.M{"guild_id": bank.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
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
	account.SetBalance(0)
	account.Deposit(100)
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestWithdraw(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			db.Delete(BANK_COLLECTION, bson.M{"guild_id": bank.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
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
	account.SetBalance(200)
	account.Withdraw(100)
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestSetBalance(t *testing.T) {
	bank := GetBank("12345")
	account := GetAccount(bank.GuildID, "54321")
	account.SetBalance(500)
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
	writeAccount(account)

	ResetMonthlyBalances()

	account = GetAccount(bank.GuildID, "54321")
	if account.MonthlyBalance != 0 {
		t.Errorf("Expected monthly balance to be 0, got %d", account.MonthlyBalance)
	}
}
