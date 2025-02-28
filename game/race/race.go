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
	GuildID        string                       // Guild (server) on which the race is taking place
	RacePartipants []*RaceParticpant            // The list of participants who are racing
	Betters        []*RaceBetter                // The list of members who are betting on the outcome of the race
	interaction    *discordgo.InteractionCreate // Interaction used in sending message updates
	config         *Config                      // Race configuration (avoids having to read from the database)
	mutex          sync.Mutex                   // Lock used to synchronize access to the race
}

// RacePartipant is used to track a member who is paricipating in a race.
type RaceParticpant struct {
	Member          *RaceMember // Member who is racing
	LastMove        int         // Distance moved on the last move
	CurrentPosition int         // Current position on the race track
	LastPosition    int         // Initialize to the end of the race track
	Position        int         // Initialize to the end of the race track
	Speed           float64     // Calculate at end to sort the racers
	Turn            int         // How many turns it took to move from the starting position to the end of the track
	Prize           int         // The amount of credits earned in the race
	racer           *Racer      // The racer assigned to the member
}

// RaceBetter is a member who is betting on the outcome of the race.
type RaceBetter struct {
	Member *RaceMember     // Member who is betting on the outcome of the the race
	Racer  *RaceParticpant // Racer on which the member is betting
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
		GuildID:        guildID,
		RacePartipants: make([]*RaceParticpant, 0, 10),
		Betters:        make([]*RaceBetter, 0, 10),
		interaction:    nil,
		config:         config,
		mutex:          sync.Mutex{},
	}
	currentRaces[guildID] = race
	log.WithFields(log.Fields{"guild": guildID}).Info("new race")

	return race
}

// newRaceBetter returns a new better for a race.
func newRaceBetter(member *RaceMember, racer *RaceParticpant) *RaceBetter {
	log.Trace("--> race.newRaceBetter")
	defer log.Trace("<-- race.newRaceBetter")

	raceBetter := &RaceBetter{
		Member: member,
		Racer:  racer,
	}

	return raceBetter
}

// AddRacer adds a race partipant to the given race.
func (r *Race) AddRacer(raceParicipant *RaceParticpant) {
	log.Trace("--> race.Race.AddRacer")
	defer log.Trace("<-- race.Race.AddRacer")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.RacePartipants = append(r.RacePartipants, raceParicipant)
	log.WithFields(log.Fields{"guild": r.GuildID, "racer": raceParicipant.Member.MemberID}).Info("add racer to current race")
}

// Adds a better for the given race.
func (r *Race) AddBetter(better *RaceBetter) {
	log.Trace("--> race.Race.AddBetter")
	defer log.Trace("<-- race.Race.AddBetter")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Betters = append(r.Betters, better)
	log.WithFields(log.Fields{"guild": r.GuildID, "better": better.Member.MemberID}).Info("add better to current race")
}

func (r *Race) RunRace() {
	log.Trace("--> race.Race.RunRace")
	defer log.Trace("<-- race.Race.RunRace")

	r.mutex.Lock()
	defer r.mutex.Unlock()

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

// Move updates the race particpant for a given turn.
func (rp *RaceParticpant) Move() bool {
	log.Trace("-->race.RaceParticpant.Move")
	defer log.Trace("<-- race.RaceParticpant.Move")

	if rp.Position > 0 {
		rp.LastPosition = rp.Position
		rp.LastMove = rp.racer.calculateMovement(rp.Turn)
		rp.Position -= rp.LastMove
		rp.Turn++
		if rp.Position <= 0 {
			rp.Speed = float64(rp.Turn) + float64(rp.LastPosition)/float64(rp.LastMove)
		}
		return true
	}

	return false
}
