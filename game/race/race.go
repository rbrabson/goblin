package race

import (
	"math/rand"
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
	RacePartipants []*RaceMember                // The list of participants who are racing
	Betters        []*RaceBetter                // The list of members who are betting on the outcome of the race
	RaceLegs       []*RaceLeg                   // The list of legs in the race
	RaceResult     *RaceResult                  // The results of the race
	interaction    *discordgo.InteractionCreate // Interaction used in sending message updates
	config         *Config                      // Race configuration (avoids having to read from the database)
	mutex          sync.Mutex                   // Lock used to synchronize access to the race
}

// RaceResults is the final results of the race. This includes the winner, 2nd place, and 3rd place finishers, as
// well as the speed at which they finished.
type RaceResult struct {
	Win       *RaceParticipant // First place in the race
	Place     *RaceParticipant // Second place in the race
	Show      *RaceParticipant // Third place in the race
	WinTime   int              // Speed for the winner in the race
	PlaceTime int              // Speed for the 2nd plalce finisher in the race
	ShowTime  int              // Speed for the 3rd place finisher in the race
}

// RaceLeg is a single leg in a race. This covers the movement for all racers during the given turn.
type RaceLeg struct {
	ParticipantPositions []*RaceParticipantPosition // The results for each member in a given leg of the race
}

// RacePartipant is used to track the movement of a given member during a single leg of a race.
type RaceParticipantPosition struct {
	RaceParticipant *RaceParticipant // Member who is racing
	Position        int              // Position of the member on the track for a given leg of the race
	Movement        int              // Amount of movement for the member on the track for a given leg of the race
	Speed           float64          // Speed at which the member moved during the leg of the race
	Turn            int              // Turn in which the member is racing
	Finished        bool             // The member has crossed the finish line
}

type RaceParticipant struct {
	Member *RaceMember // Member who is racing
	Racer  *Racer      // Racer assigned to the member
}

// RaceBetter is a member who is betting on the outcome of the race.
type RaceBetter struct {
	Member *RaceMember              // Member who is betting on the outcome of the the race
	Racer  *RaceParticipantPosition // Racer on which the member is betting
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
		RacePartipants: make([]*RaceMember, 0, 10),
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
func newRaceBetter(member *RaceMember, racer *RaceParticipantPosition) *RaceBetter {
	log.Trace("--> race.newRaceBetter")
	defer log.Trace("<-- race.newRaceBetter")

	raceBetter := &RaceBetter{
		Member: member,
		Racer:  racer,
	}

	return raceBetter
}

// AddRacer adds a race partipant to the given race.
func (r *Race) AddRacer(raceMember *RaceMember) {
	log.Trace("--> race.Race.AddRacer")
	defer log.Trace("<-- race.Race.AddRacer")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.RacePartipants = append(r.RacePartipants, raceMember)
	log.WithFields(log.Fields{"guild": r.GuildID, "racer": raceMember.MemberID}).Info("add racer to current race")
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

	/*
	   Keep an arry for every member of their position on the track, until every parcipant has ended
	   In this array, keep track of who gets to the end first (i.e., the results of the race). If two
	   members reach the end at the same time, randomize the results, or use some other mechanism to decide
	   which member wins.

	   By using an array, this lets the full race results be calculated, and then displayed later on. Again, keeping
	   a spearation of duties between running the race and displaying the results. This will help in testing the logic
	   as well, as the display logic can be tested independent of the race calculations.
	*/

	/*
		   Basic logic:
		   - while not all racers are done
		       run a race leg
			   record the results for each racer
		   - runLeg
		      go through each racer and update their position
			  then save in the position in the current leg
			  figure out which results need to be saved for each race leg
			  looks like the current position is the only thing that is critical.
			- want to calculate the speed for each racer to display at the end
			  this is how long it takes to reach the end of the race (i.e., the time to complete the race)
			       r.Speed = float64(r.Turn) + float64(r.LastPosition)/float64(r.LastMove)
			  where r.Turn is the number of turns to reach the end, r.LastPosition is the position before
			       reaching the finish line (going fron 60 -> 0, for instance), and LastMove is the
				   movement calcualted that causes the member to cross the finish line.

	*/

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
func newRaceParitipcant(member *RaceMember, racers []*Racer) *RaceParticipant {
	log.Trace("--> race.newRaceParticipant")
	defer log.Trace("<-- race.newRaceParticipant")

	index := rand.Intn(len(racers))
	participant := &RaceParticipant{
		Member: member,
		Racer:  racers[index],
	}
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "racer": participant.Racer.Emoji}).Debug("new race particiapnt")

	return participant
}

// Move updates the race particpant for a given turn.
func (rpp *RaceParticipantPosition) Move(turn int) *RaceParticipantPosition {
	log.Trace("-->race.RaceParticpant.Move")
	defer log.Trace("<-- race.RaceParticpant.Move")

	if rpp.Position <= 0 {
		newPosition := &RaceParticipantPosition{
			RaceParticipant: rpp.RaceParticipant,
			Finished:        true,
			Speed:           rpp.Speed,
		}
		return newPosition
	}

	movement := rpp.RaceParticipant.Racer.calculateMovement(turn)
	speed := float64(rpp.Turn) + float64(rpp.Position)/float64(movement)
	newPosition := &RaceParticipantPosition{
		RaceParticipant: rpp.RaceParticipant,
		Position:        rpp.Position - movement,
		Movement:        rpp.Movement,
		Speed:           speed,
		Turn:            rpp.Turn + 1,
		Finished:        false,
	}

	return newPosition
}
