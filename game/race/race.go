package race

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/guild"
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
	Racers      []*Member
	Betters     []*Member
	interaction *discordgo.InteractionCreate
	config      *Config
	mutex       sync.Mutex
}

// GetRace gets the race for the guild. If a race isn't in progress, then a new one
// is created.
func GetRace(g *guild.Guild) *Race {
	log.Trace("--> race.GetRace")
	defer log.Trace("<-- race.GetRace")

	raceLock.Lock()
	defer raceLock.Unlock()
	race := currentRaces[g.GuildID]
	if race == nil {
		race = newRace(g)
	}
	return race
}

// newRace creates a new race for the guild.
func newRace(g *guild.Guild) *Race {
	log.Trace("--> race.newRace")
	defer log.Trace("<-- race.newRace")

	config := GetConfig(g)
	race := &Race{
		GuildID:     g.GuildID,
		Racers:      make([]*Member, 0, 10),
		Betters:     make([]*Member, 0, 10),
		interaction: nil,
		config:      config,
		mutex:       sync.Mutex{},
	}
	currentRaces[g.GuildID] = race
	log.WithFields(log.Fields{"guild": g.GuildID}).Info("new race")

	return race
}

// End ends the current race.
func (r *Race) End() {
	raceLock.Lock()
	defer raceLock.Unlock()

	delete(currentRaces, r.GuildID)
	log.WithFields(log.Fields{"guild": r.GuildID}).Info("end race")
}
