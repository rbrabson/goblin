package heist

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/guild"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../../.env_test")
	if err != nil {
		slog.Error("Error loading .env_test file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetConfig(t *testing.T) {
	g := guild.GetGuild("12345")
	configs := make([]*Config, 0, 1)
	defer func() {
		for _, config := range configs {
			if err := db.Delete(ConfigCollection, bson.M{"guild_id": config.GuildID}); err != nil {
				slog.Error("error deleting config",
					slog.Any("error", err),
				)
			}
		}
	}()

	config := GetConfig(g.GuildID)
	if config == nil {
		t.Errorf("Expected config, got nil")
		return
	}
	configs = append(configs, config)

	if config.Theme != HeistDefaultTheme {
		t.Errorf("Expected %s, got %s", HeistDefaultTheme, config.Theme)
	}
	if config.BailBase != BailBase {
		t.Errorf("Expected %d, got %d", BailBase, config.BailBase)
	}
	if config.CrewOutput != CrewOutput {
		t.Errorf("Expected %s, got %s", CrewOutput, config.CrewOutput)
	}
	if config.DeathTimer != DeathTimer {
		t.Errorf("Expected %d, got %d", DeathTimer, config.DeathTimer)
	}
	if config.HeistCost != HeistCost {
		t.Errorf("Expected %d, got %d", HeistCost, config.HeistCost)
	}
	if config.PoliceAlert != PoliceAlert {
		t.Errorf("Expected %d, got %d", PoliceAlert, config.PoliceAlert)
	}
	if config.SentenceBase != SentenceBase {
		t.Errorf("Expected %d, got %d", SentenceBase, config.SentenceBase)
	}
	if config.WaitTime != WaitTime {
		t.Errorf("Expected %d, got %d", WaitTime, config.WaitTime)
	}
	if config.Targets != HeistDefaultTheme {
		t.Errorf("Expected empty string, got %s", config.Targets)
	}
}
