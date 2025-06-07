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

func TestGetBank(t *testing.T) {
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

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
	}
	banks = append(banks, bank)
}

func TestGetAccounts(t *testing.T) {
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

	account := GetAccount(bank.GuildID, "67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
	filter := bson.M{
		"guild_id": bank.GuildID,
	}
	sort := bson.M{
		"member_id": 1,
	}
	accounts = GetAccounts(bank.GuildID, filter, sort, 10)
	if len(accounts) == 0 {
		t.Error("no accounts returned")
	}
}
