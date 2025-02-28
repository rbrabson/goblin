package race

import (
	"sync"
	"time"

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
	Racers      []*RaceParticpant
	Betters     []*RaceMember
	interaction *discordgo.InteractionCreate
	config      *Config
	mutex       sync.Mutex
}

// RacePartipant is used to track a member who is paricipating in a race.
type RaceParticpant struct {
	Member       *RaceMember // Member who is racing
	LastMove     int64       // Distance moved on the last move
	LastPosition int64       // Initialize to the end of the race track
	Position     int64       // Initialize to the end of the race track
	Speed        float64     // Calculate at end to sort the racers
	Turn         int64       // How many turns it took to move from the starting position to the end of the track
	Prize        int         // The amount of credits earned in the race
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
		Racers:      make([]*RaceParticpant, 0, 10),
		Betters:     make([]*RaceMember, 0, 10),
		interaction: nil,
		config:      config,
		mutex:       sync.Mutex{},
	}
	currentRaces[guildID] = race
	log.WithFields(log.Fields{"guild": guildID}).Info("new race")

	return race
}

func (r *Race) AddRacer(racer *RaceParticpant) {
	log.Trace("--> race.Race.AddRacer")
	defer log.Trace("<-- race.Race.AddRacer")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Racers = append(r.Racers, racer)
	log.WithFields(log.Fields{"guild": r.GuildID, "racer": racer.Member.MemberID}).Info("add racer to current race")
}

func (r *Race) RunRace() {
	log.Trace("--> race.Race.RunRace")
	defer log.Trace("<-- race.Race.RunRace")

	// TODO; implement. This needs to pass back an array of race legs to be displayed that shows where each racer is
	//       located on the track.
}

// End ends the current race.
func (r *Race) End() {
	raceLock.Lock()
	defer raceLock.Unlock()
	delete(currentRaces, r.GuildID)

	config := GetConfig(r.GuildID)
	config.LastRaceEnded = time.Now()
	writeConfig(config)

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

// newRaceParticipant creates a new RaceParticpant for the given member. This is used to
// track the position of the member in the race.
func newRaceParitipcant(member *RaceMember) *RaceParticpant {
	log.Trace("--> race.newRaceParticipant")
	defer log.Trace("<-- race.newRaceParticipant")

	participant := &RaceParticpant{
		Member:       member,
		LastMove:     0,
		LastPosition: 0,
		Position:     0,
		Speed:        0,
		Turn:         0,
		Prize:        0,
	}

	return participant
}
