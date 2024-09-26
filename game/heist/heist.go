package heist

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/dgame/database"
)

var (
	db database.Client
)

type HeistState int

const (
	PLANNED HeistState = iota
	STARTED
	COMPLETED
)

// Configuration data for new heists
type HeistConfig struct {
	GuildID      string        `json:"_id" bson:"_id"`
	AlertTime    time.Time     `json:"alert_time" bson:"alert_time"`
	BailBase     int64         `json:"bail_base" bson:"bail_base"`
	CrewOutput   string        `json:"crew_output" bson:"crew_output"`
	DeathTimer   time.Duration `json:"death_timer" bson:"death_timer"`
	HeistCost    int64         `json:"heist_cost" bson:"heist_cost"`
	PoliceAlert  time.Duration `json:"police_alert" bson:"police_alert"`
	SentenceBase time.Duration `json:"sentence_base" bson:"sentence_base"`
	Theme        string        `json:"theme" bson:"theme"`
	Targets      string        `json:"targets" bson:"targets"`
	WaitTime     time.Duration `json:"wait_time" bson:"wait_time"`
}

// Heist is a heist that is being planned, is in progress, or has completed
type Heist struct {
	PlannerID   *string                      `json:"planner_id" bson:"planner_id"`
	CrewIDs     []*string                    `json:"crew_ids" bson:"crew_ids"`
	State       HeistState                   `json:"state" bson:"state"`
	StartTime   time.Time                    `json:"start_time" bson:"start_time"`
	interaction *discordgo.InteractionCreate `json:"-" bson:"-"`
	guildID     string                       `json:"-" bson:"-"`
}

type HeistResult struct {
	Escaped       int
	Apprehended   int
	Dead          int
	memberResults []*HeistMemberResult
	survivingCrew []*HeistMemberResult
	Target        *Target
}

// HeistMemberResult is the result for a single member of the heist
type HeistMemberResult struct {
	player        *Member
	status        string
	message       string
	stolenCredits int
	bonusCredits  int
}

// Init initializes the heist game to use the provided database for reading and writing data
func Init(database database.Client) {
	db = database
}

// String returns a string representation of the state of a heist
func (state HeistState) String() string {
	switch state {
	case PLANNED:
		return "Planned"
	case STARTED:
		return "Started"
	case COMPLETED:
		return "Completed"
	default:
		return "Unknown"
	}
}
