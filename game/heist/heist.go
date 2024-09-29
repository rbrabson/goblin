package heist

import (
	"encoding/json"
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

// Heist is a heist that is being planned, is in progress, or has completed
type Heist struct {
	PlannerID   *string                      `json:"planner_id" bson:"planner_id"`
	CrewIDs     []*string                    `json:"crew_ids" bson:"crew_ids"`
	State       HeistState                   `json:"state" bson:"state"`
	StartTime   time.Time                    `json:"start_time" bson:"start_time"`
	interaction *discordgo.InteractionCreate `json:"-" bson:"-"`
	guildID     string                       `json:"-" bson:"-"`
}

// HeistResult are the results of a heist
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

// String returns a string representation of the heist
func (heist *Heist) String() string {
	out, _ := json.Marshal(heist)
	return string(out)
}

// String returns a string representation of the resuilt of a heist
func (result *HeistResult) String() string {
	out, _ := json.Marshal(result)
	return string(out)
}

// String returns a string representation of the resuilt for a single member of a heist
func (result *HeistMemberResult) String() string {
	out, _ := json.Marshal(result)
	return string(out)
}
