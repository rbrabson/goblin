package heist

import (
	"encoding/json"
	"time"
)

// Configuration data for new heists
type Config struct {
	GuildID      string        `json:"_id" bson:"_id"`
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
}

func NewHeistConfig() *Config {
	return nil
}

func GetHeistConfig() *Config {
	return nil
}

func (config *Config) Write() {

}

// String returns a string representation of the heist configuration
func (config *Config) String() string {
	out, _ := json.Marshal(config)
	return string(out)
}
