package heist

import (
	"encoding/json"
	"time"

	"github.com/rbrabson/dgame/guild"
)

const (
	GAME_ID           = "heist"
	CONFIG_COLLECTION = "config"
)

var (
	configs = make(map[string]*Config)
)

// Configuration data for new heists
type Config struct {
	ID           string        `json:"_id" bson:"_id"`
	AlertTime    time.Time     `json:"alert_time" bson:"alert_time"`
	BailBase     int           `json:"bail_base" bson:"bail_base"`
	CrewOutput   string        `json:"crew_output" bson:"crew_output"`
	DeathTimer   time.Duration `json:"death_timer" bson:"death_timer"`
	HeistCost    int           `json:"heist_cost" bson:"heist_cost"`
	PoliceAlert  time.Duration `json:"police_alert" bson:"police_alert"`
	SentenceBase time.Duration `json:"sentence_base" bson:"sentence_base"`
	Theme        string        `json:"theme" bson:"theme"`
	Targets      string        `json:"targets" bson:"targets"`
	WaitTime     time.Duration `json:"wait_time" bson:"wait_time"`
	guildID      string        `json:"-" bson:"-"`
}

// NewConfig creates a new default configuration for the specified guild.
func NewConfig(guild *guild.Guild) *Config {
	config := &Config{
		ID:      GAME_ID,
		guildID: guild.ID,
	}
	configs[config.guildID] = config
	config.Write()
	return config
}

// GetConfig retrieves the heist configuration for the specified guild. If
// the configuration does not exist, it is created and added to the database.
func GetConfig(guild *guild.Guild) *Config {
	config := configs[guild.ID]
	if config != nil {
		return config
	}
	config = LoadConfig(guild)
	if config != nil {
		configs[config.guildID] = config
		return config
	}
	return NewConfig(guild)
}

// LoadConfig loads the heist configuration from the database. If it does not exist then
// a `nil` value is returned.
func LoadConfig(guild *guild.Guild) *Config {
	var config Config
	db.Read(guild.ID, CONFIG_COLLECTION, guild.ID, &config)
	return &config
}

// Write stores the configuration in the database.
func (config *Config) Write() {
	db.Write(config.guildID, CONFIG_COLLECTION, config.guildID, config)
}

// String returns a string representation of the heist configuration
func (config *Config) String() string {
	out, _ := json.Marshal(config)
	return string(out)
}
