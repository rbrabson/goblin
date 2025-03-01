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

func TestGetBank(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			db.Delete(BANK_COLLECTION, bson.M{"guild_id": bank.GuildID})
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
