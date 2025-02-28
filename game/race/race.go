package race

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var (
	currentRaces = make(map[string]*Race)
	raceLock     = sync.Mutex{}
)

// Race represents a race that is currently in progress.
// It contains a list of racers who are particpaing in the race as well as
// betters on the outcome of the race.
type Race struct {
	GuildID     string
	Racers      []*RaceMember
	Betters     []*RaceMember
	interaction *discordgo.InteractionCreate
	config      *Config
	mutex       sync.Mutex
}

// GetRace gets the race for the guild. If a race isn't in progress, then a new one
// is created.
func GetRace(guildID string) *Race {
	log.Trace("--> race.GetRace")
	defer log.Trace("<-- race.GetRace")

	raceLock.Lock()
	defer raceLock.Unlock()
	race := currentRaces[guildID]
	if race == nil {
		race = newRace(guildID)
	}
	return race
}

// newRace creates a new race for the guild.
func newRace(guildID string) *Race {
	log.Trace("--> race.newRace")
	defer log.Trace("<-- race.newRace")

	config := GetConfig(guildID)
	race := &Race{
		GuildID:     guildID,
		Racers:      make([]*RaceMember, 0, 10),
		Betters:     make([]*RaceMember, 0, 10),
		interaction: nil,
		config:      config,
		mutex:       sync.Mutex{},
	}
	currentRaces[guildID] = race
	log.WithFields(log.Fields{"guild": guildID}).Info("new race")

	return race
}

// End ends the current race.
func (r *Race) End() {
	raceLock.Lock()
	defer raceLock.Unlock()

	delete(currentRaces, r.GuildID)
	log.WithFields(log.Fields{"guild": r.GuildID}).Info("end race")
}

// ResetRace resets a hung race for a given guild.
func ResetRace(guildID string) {
	log.Trace("---> race.ResetRace")
	defer log.Trace("<-- race.ResetRace")

	raceLock.Lock()
	defer raceLock.Unlock()

	delete(currentRaces, guildID)
	log.WithFields(log.Fields{"guild": guildID}).Info("reset race")
}
