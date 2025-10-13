package payday

import (
	"log/slog"
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
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
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
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
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
	if payday.Amount != DefaultPaydayAmount {
		t.Errorf("expected Amount to be '%d', got '%d'", DefaultPaydayAmount, payday.Amount)
	}
	if payday.PaydayFrequency != DefaultPaydayFrequency {
		t.Errorf("expected PaydayFrequency to be '%s', got '%s'", DefaultPaydayFrequency, payday.PaydayFrequency)
	}
}

func TestSetPaydayAmount(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
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
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
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

func TestPaydayString(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	payday := readPaydayFromFile("12345")
	if payday == nil {
		t.Error("payday is nil")
		return
	}
	paydays = append(paydays, payday)

	// Test the String method
	str := payday.String()
	expected := "Payday{ID=" + payday.ID.Hex() + ", GuildID=12345, Amount=" +
		"5000, PaydayFrequency=23h0m0s}"
	if str != expected {
		t.Errorf("expected String() to return '%s', got '%s'", expected, str)
	}
}

func TestGetDefaultPayday(t *testing.T) {
	paydays := make([]*Payday, 0, 1)
	defer func() {
		for _, payday := range paydays {
			if err := db.Delete(PaydayCollection, bson.M{"guild_id": payday.GuildID}); err != nil {
				slog.Error("Error deleting payday",
					slog.String("guildID", payday.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()

	guildID := "54321"
	payday := getDefaultPayday(guildID)
	if payday == nil {
		t.Error("payday is nil")
		return
	}
	paydays = append(paydays, payday)

	if payday.GuildID != guildID {
		t.Errorf("expected GuildID to be '%s', got '%s'", guildID, payday.GuildID)
	}
	if payday.Amount != DefaultPaydayAmount {
		t.Errorf("expected Amount to be '%d', got '%d'", DefaultPaydayAmount, payday.Amount)
	}
	if payday.PaydayFrequency != DefaultPaydayFrequency {
		t.Errorf("expected PaydayFrequency to be '%s', got '%s'", DefaultPaydayFrequency, payday.PaydayFrequency)
	}
}
