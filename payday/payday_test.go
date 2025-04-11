package payday

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
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

func TestGetPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()

	payday := GetPayday("12345")
	if payday == nil {
		t.Error("payday is nil")
	}
	paydays = append(paydays, payday)
}

func TestNewPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Error("payday is nil")
		return
	}
	paydays = append(paydays, payday)

	if payday.GuildID != "12345" {
		t.Errorf("expected GuildID to be '12345', got '%s'", payday.GuildID)
	}
	if payday.Amount != DEFAULT_PAYDAY_AMOUNT {
		t.Errorf("expected Amount to be '%d', got '%d'", DEFAULT_PAYDAY_AMOUNT, payday.Amount)
	}
	if payday.PaydayFrequency != DEFAULT_PAYDAY_FREQUENCY {
		t.Errorf("expected PaydayFrequency to be '%s', got '%s'", DEFAULT_PAYDAY_FREQUENCY, payday.PaydayFrequency)
	}
}

func TestSetPaydayAmount(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Error("payday is nil")
		return
	}
	paydays = append(paydays, payday)

	newAmount := 10000
	payday.SetPaydayAmount(newAmount)
	payday = readPayday("12345")
	if payday.Amount != newAmount {
		t.Errorf("expected Amount to be '%d', got '%d'", newAmount, payday.Amount)
	}
}

func TestSetPaydayFrequency(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			db.Delete(PAYDAY_COLLECTION, bson.M{"guild_id": payday.GuildID})
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Error("payday is nil")
		return
	}
	paydays = append(paydays, payday)

	newFrequency := 48 * time.Hour
	payday.SetPaydayFrequency(newFrequency)
	payday = readPayday("12345")
	if payday.PaydayFrequency != newFrequency {
		t.Errorf("expected PaydayFrequency to be '%s', got '%s'", newFrequency, payday.PaydayFrequency)
	}
}
