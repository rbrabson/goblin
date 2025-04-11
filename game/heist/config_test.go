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
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestGetConfig(t *testing.T) {
	guild := guild.GetGuild("12345")
	configs := make([]*Config, 0, 1)
	defer func() {
		for _, config := range configs {
			db.Delete(CONFIG_COLLECTION, bson.M{"guild_id": config.GuildID})
		}
	}()

	config := GetConfig(guild.GuildID)
	if config == nil {
		t.Errorf("Expected config, got nil")
		return
	}
	configs = append(configs, config)

	if config.Theme != HEIST_DEFAULT_THEME {
		t.Errorf("Expected %s, got %s", HEIST_DEFAULT_THEME, config.Theme)
	}
	if config.BailBase != BAIL_BASE {
		t.Errorf("Expected %d, got %d", BAIL_BASE, config.BailBase)
	}
	if config.CrewOutput != CREW_OUTPUT {
		t.Errorf("Expected %s, got %s", CREW_OUTPUT, config.CrewOutput)
	}
	if config.DeathTimer != DEATH_TIMER {
		t.Errorf("Expected %d, got %d", DEATH_TIMER, config.DeathTimer)
	}
	if config.HeistCost != HEIST_COST {
		t.Errorf("Expected %d, got %d", HEIST_COST, config.HeistCost)
	}
	if config.PoliceAlert != POLICE_ALERT {
		t.Errorf("Expected %d, got %d", POLICE_ALERT, config.PoliceAlert)
	}
	if config.SentenceBase != SENTENCE_BASE {
		t.Errorf("Expected %d, got %d", SENTENCE_BASE, config.SentenceBase)
	}
	if config.WaitTime != WAIT_TIME {
		t.Errorf("Expected %d, got %d", WAIT_TIME, config.WaitTime)
	}
	if config.Targets != HEIST_DEFAULT_THEME {
		t.Errorf("Expected empty string, got %s", config.Targets)
	}
}
