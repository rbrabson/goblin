package race

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/guild"
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

func GetRace(g *guild.Guild) *Race {
	raceLock.Lock()
	defer raceLock.Unlock()
	race := currentRaces[g.GuildID]
	if race == nil {
		race = newRace(g)
	}
	return race
}

func newRace(g *guild.Guild) *Race {
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

	return race
}

// End ends the current race.
func (r *Race) End() {
	raceLock.Lock()
	defer raceLock.Unlock()
	delete(currentRaces, r.GuildID)
}
