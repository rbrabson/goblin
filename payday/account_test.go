package payday

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/internal/disctime"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		sslog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetAccount(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(PAYDAY_ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Errorf("newPayday() returned nil")
		return
	}
	paydays = append(paydays, payday)

	account := payday.GetAccount("67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
}

func TestNewAccount(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(PAYDAY_ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Errorf("newPayday() returned nil")
		return
	}
	paydays = append(paydays, payday)

	account := newAccount(payday, "67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)

	if account.MemberID != "67890" {
		t.Errorf("expected MemberID to be '67890', got '%s'", account.MemberID)
		return
	}
	if account.GuildID != "12345" {
		t.Errorf("expected GuildID to be '12345', got '%s'", account.GuildID)
		return
	}
}

func TestSetNextPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(PAYDAY_ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Errorf("newPayday() returned nil")
		return
	}
	paydays = append(paydays, payday)

	account := newAccount(payday, "67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)

	nextPayday := disctime.NextMonth(time.Now())
	account.setNextPayday(nextPayday)
	account = readAccount(payday, "67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	if !account.NextPayday.Equal(nextPayday) {
		t.Errorf("expected NextPayday to be '%s', got '%s'", nextPayday, account.NextPayday)
		return
	}
}

func TestGetNextPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			db.Delete(PAYDAY_ACCOUNT_COLLECTION, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Errorf("newPayday() returned nil")
		return
	}
	paydays = append(paydays, payday)

	account := newAccount(payday, "67890")
	if account == nil {
		t.Error("account is nil")
		return
	}
	nextPayday := disctime.NextMonth(time.Now())
	account.setNextPayday(nextPayday)
	account = readAccount(payday, "67890")
	if !account.getNextPayday().Equal(nextPayday) {
		t.Errorf("expected NextPayday to be '%s', got '%s'", nextPayday, account.getNextPayday())
		return
	}

	accounts = append(accounts, account)
}
