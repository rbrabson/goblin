package heist

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	GAME_ID = "heist"
)

const (
	BAIL_BASE           = 250
	CREW_OUTPUT         = "None"
	DEATH_TIMER         = time.Duration(45 * time.Second)
	HEIST_COST          = 1500
	POLICE_ALERT        = time.Duration(60 * time.Second)
	SENTENCE_BASE       = time.Duration(45 * time.Second)
	WAIT_TIME           = time.Duration(60 * time.Second)
	HEIST_DEFAULT_THEME = "clash"
)

// Configuration data for new heists
type Config struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID      string             `json:"guild_id" bson:"guild_id"`
	Theme        string             `json:"theme" bson:"theme"`
	BailBase     int                `json:"bail_base" bson:"bail_base"`
	CrewOutput   string             `json:"crew_output" bson:"crew_output"`
	DeathTimer   time.Duration      `json:"death_timer" bson:"death_timer"`
	HeistCost    int                `json:"heist_cost" bson:"heist_cost"`
	PoliceAlert  time.Duration      `json:"police_alert" bson:"police_alert"`
	SentenceBase time.Duration      `json:"sentence_base" bson:"sentence_base"`
	Targets      string             `json:"targets" bson:"targets"`
	WaitTime     time.Duration      `json:"wait_time" bson:"wait_time"`
}

// GetConfig retrieves the heist configuration for the specified guild. If
// the configuration does not exist, nil is returned.
func GetConfig(guildID string) *Config {
	config := readConfig(guildID)
	if config == nil {
		config = NewConfig(guildID)
	}
	return config
}

// NewConfig creates a new default configuration for the specified guild.
func NewConfig(guildID string) *Config {
	config := &Config{
		GuildID:      guildID,
		BailBase:     BAIL_BASE,
		CrewOutput:   CREW_OUTPUT,
		DeathTimer:   DEATH_TIMER,
		HeistCost:    HEIST_COST,
		PoliceAlert:  POLICE_ALERT,
		SentenceBase: SENTENCE_BASE,
		Targets:      HEIST_DEFAULT_THEME,
		Theme:        HEIST_DEFAULT_THEME,
		WaitTime:     WAIT_TIME,
	}
	writeConfig(config)

	return config
}

// String returns a string representation of the heist configuration
func (config *Config) String() string {
	out, _ := json.Marshal(config)
	return string(out)
}
