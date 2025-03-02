package race

import (
	"math/rand"
	"sort"
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
	GuildID       string                       // Guild (server) on which the race is taking place
	Racers        []*RaceParticipant           // The list of participants who are racing
	Betters       []*RaceBetter                // The list of members who are betting on the outcome of the race
	RaceLegs      []*RaceLeg                   // The list of legs in the race
	RaceResult    *RaceResult                  // The results of the race
	RaceStartTime time.Time                    // The time at which the race is started
	interaction   *discordgo.InteractionCreate // Interaction used in sending message updates
	config        *Config                      // Race configuration (avoids having to read from the database)
	mutex         sync.Mutex                   // Lock used to synchronize access to the race
}

// RaceResults is the final results of the race. This includes the winner, 2nd place, and 3rd place finishers, as
// well as the speed at which they finished.
type RaceResult struct {
	Win   *RaceParticipantResult // First place in the race
	Place *RaceParticipantResult // Second place in the race
	Show  *RaceParticipantResult // Third place in the race
}

type RaceParticipantResult struct {
	Participant *RaceParticipant // Participant in the race
	RaceTime    float64          // Time at which the participant finished
	Winnings    int              // Amount the participant won
}

// RaceLeg is a single leg in a race. This covers the movement for all racers during the given turn.
type RaceLeg struct {
	ParticipantPositions []*RaceParticipantPosition // The results for each member in a given leg of the race
}

// RacePartipantPosition is used to track the movement of a given member during a single leg of a race.
type RaceParticipantPosition struct {
	RaceParticipant *RaceParticipant // Member who is racing
	Position        int              // Position of the member on the track for a given leg of the race
	Movement        int              // Amount of movement for the member on the track for a given leg of the race
	Speed           float64          // Speed at which the member moved during the leg of the race
	Turn            int              // Turn in which the member is racing
	Finished        bool             // The member has crossed the finish line
}

// RaceParticpant is a member who is racing. This includes the member and the racer assigned to them.
type RaceParticipant struct {
	Member *RaceMember // Member who is racing
	Racer  *Racer      // Racer assigned to the member
	Prize  int         // Amount earned in the race
}

// RaceBetter is a member who is betting on the outcome of the race.
type RaceBetter struct {
	Member *RaceMember      // Member who is betting on the outcome of the the race
	Racer  *RaceParticipant // Racer on which the member is betting
}

// GetRace gets the race for the guild. If a race isn't in progress, then a new one is created.
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
		GuildID:       guildID,
		Racers:        make([]*RaceParticipant, 0, 10),
		Betters:       make([]*RaceBetter, 0, 10),
		RaceStartTime: time.Now(),
		RaceResult:    &RaceResult{},
		interaction:   nil,
		config:        config,
		mutex:         sync.Mutex{},
	}
	currentRaces[guildID] = race
	log.WithFields(log.Fields{"guild": guildID}).Info("new race")

	return race
}

// GetRaceBetter returns a new better for a race.
func GetRaceBetter(member *RaceMember, racer *RaceParticipant) *RaceBetter {
	log.Trace("--> race.GetRaceBetter")
	defer log.Trace("<-- race.GetRaceBetter")

	raceBetter := &RaceBetter{
		Member: member,
		Racer:  racer,
	}

	return raceBetter
}

// AddRacer adds a race partipant to the given race.
func (r *Race) AddRacer(raceParticipant *RaceParticipant) error {
	log.Trace("--> race.Race.AddRacer")
	defer log.Trace("<-- race.Race.AddRacer")

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if err := raceJoinChecks(r, raceParticipant); err != nil {
		return err
	}

	r.Racers = append(r.Racers, raceParticipant)
	log.WithFields(log.Fields{"guild": r.GuildID, "racer": raceParticipant.Member.MemberID}).Info("add racer to current race")

	return nil
}

// Adds a better for the given race.
func (race *Race) AddBetter(better *RaceBetter) error {
	log.Trace("--> race.Race.AddBetter")
	defer log.Trace("<-- race.Race.AddBetter")

	race.mutex.Lock()
	defer race.mutex.Unlock()

	if err := raceBetChecks(race, better); err != nil {
		return err
	}

	race.Betters = append(race.Betters, better)
	log.WithFields(log.Fields{"guild": race.GuildID, "better": better.Member.MemberID}).Info("add better to current race")

	return nil
}

// RunRace runs a race, calculating the results of each leg of the race and the
// ultimate winners of the race.
func (race *Race) RunRace(trackLength int) {
	log.Trace("--> race.Race.RunRace")
	defer log.Trace("<-- race.Race.RunRace")

	race.mutex.Lock()
	defer race.mutex.Unlock()

	// Create the initial starting positions and add them to an initial race leg
	raceLeg := &RaceLeg{
		ParticipantPositions: make([]*RaceParticipantPosition, 0, len(race.Racers)),
	}
	for _, racer := range race.Racers {
		participantPosition := &RaceParticipantPosition{
			RaceParticipant: racer,
			Position:        trackLength,
		}
		raceLeg.ParticipantPositions = append(raceLeg.ParticipantPositions, participantPosition)
	}
	race.RaceLegs = append(race.RaceLegs, raceLeg)
	previousLeg := raceLeg

	// Run the race until all racers cross the finish line
	turn := 0
	stillRacing := true
	for stillRacing {
		turn++

		// Create and add a new race leg
		newRaceLeg := &RaceLeg{
			ParticipantPositions: make([]*RaceParticipantPosition, 0, len(race.Racers)),
		}

		// Run the new race leg
		stillRacing = false
		for _, previousPosition := range previousLeg.ParticipantPositions {
			newPosition := Move(previousPosition, turn)
			newRaceLeg.ParticipantPositions = append(newRaceLeg.ParticipantPositions, newPosition)
			if !newPosition.Finished {
				stillRacing = true
			}
		}

		race.RaceLegs = append(race.RaceLegs, newRaceLeg)
		previousLeg = newRaceLeg
		log.WithFields(log.Fields{"guildID": race.GuildID, "turn": turn}).Trace("run race leg")
	}

	calculateWinnings(race, previousLeg)
}

// End ends the current race.
func (r *Race) End(config *Config) {
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

// newRaceParticipant creates a new RaceParticpant for the given member. This is used to
// track the position of the member in the race.
func newRaceParticipcant(member *RaceMember, racers []*Racer) *RaceParticipant {
	log.Trace("--> race.newRaceParticipcant")
	defer log.Trace("<-- race.newRaceParticipcant")

	index := rand.Intn(len(racers))
	participant := &RaceParticipant{
		Member: member,
		Racer:  racers[index],
	}
	log.WithFields(log.Fields{"guild": member.GuildID, "member": member.MemberID, "racer": participant.Racer.Emoji}).Debug("new race particiapnt")

	return participant
}

// Move returns the new race position for a particpant based on the previous position and the current turn.
func Move(previousPosition *RaceParticipantPosition, turn int) *RaceParticipantPosition {
	log.Trace("-->race.RaceParticpant.Move")
	defer log.Trace("<-- race.RaceParticpant.Move")

	// Already done with the race
	if previousPosition.Position <= 0 {
		newPosition := &RaceParticipantPosition{
			RaceParticipant: previousPosition.RaceParticipant,
			Finished:        true,
			Speed:           previousPosition.Speed,
		}
		return newPosition
	}

	movement := previousPosition.RaceParticipant.Racer.calculateMovement(turn)
	newPosition := &RaceParticipantPosition{
		RaceParticipant: previousPosition.RaceParticipant,
		Position:        previousPosition.Position - movement,
		Movement:        previousPosition.Movement,
		Turn:            previousPosition.Turn + 1,
		Finished:        false,
	}
	newPosition.Speed = float64(newPosition.Turn)
	if newPosition.Position <= 0 {
		newPosition.Speed += float64(previousPosition.Position) / float64(movement)
	}

	return newPosition
}

// Check to see if a new racer can join a race.
func raceJoinChecks(race *Race, racer *RaceParticipant) error {
	log.Trace("--> race.raceChecks")
	defer log.Trace("<-- race.raceChecks")

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitForBets)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("race has started")
		return ErrRaceHasStarted
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("betting has opened")
		return ErrBettingHasOpened
	}

	if race.config.MaxNumRacers == len(race.Racers) {
		log.WithFields(log.Fields{"guild_id": race.GuildID, "maxRacers": race.config.MaxNumRacers}).Warn("max racers already entered")
		return ErrRaceFull{MaxNumRacersAllowed: race.config.MaxNumRacers}
	}

	for _, r := range race.Racers {
		if r.Member.MemberID == racer.Member.MemberID {
			return ErrAlreadyJoinedRace
		}
	}

	return nil
}

func raceBetChecks(race *Race, better *RaceBetter) error {
	log.Trace("--> race.raceChecks")
	defer log.Trace("<-- race.raceChecks")

	if time.Now().Before(race.RaceStartTime.Add(race.config.WaitToStart)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("betting has opened")
		return ErrBettingNotOpened
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitForBets)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("race has started")
		return ErrRaceHasStarted
	}

	for _, b := range race.Betters {
		if b.Member.MemberID == better.Member.MemberID {
			return ErrAlreadyBetOnRace
		}
	}

	return nil
}

// calculateWinngins calculates the earnings for the racers that wins, places and shows.
func calculateWinnings(race *Race, lastLeg *RaceLeg) {
	log.Trace("--> race.calculateWinnings")
	defer log.Trace("<-- race.calculateWinnings")

	// sort the participants in the final race leg
	sort.Slice(lastLeg.ParticipantPositions, func(i, j int) bool {
		if lastLeg.ParticipantPositions[i].Speed == lastLeg.ParticipantPositions[j].Speed {
			return rand.Intn(2) == 0
		}
		return lastLeg.ParticipantPositions[i].Speed < lastLeg.ParticipantPositions[j].Speed
	})

	// Calculate the winners of the race and save in the results
	prize := rand.Intn(int(race.config.MaxPrizeAmount-race.config.MinPrizeAmount)) + race.config.MinPrizeAmount
	prize *= len(race.Racers)

	// Assign the purse for the winner
	if len(lastLeg.ParticipantPositions) > 0 {
		racePosition := lastLeg.ParticipantPositions[0]
		race.RaceResult.Win = &RaceParticipantResult{
			Participant: racePosition.RaceParticipant,
			RaceTime:    racePosition.Speed,
			Winnings:    prize,
		}
	}

	// Assign the purse for the second place finisher
	if len(lastLeg.ParticipantPositions) > 1 {
		racePosition := lastLeg.ParticipantPositions[1]
		race.RaceResult.Place = &RaceParticipantResult{
			Participant: racePosition.RaceParticipant,
			RaceTime:    racePosition.Speed,
			Winnings:    int(float64(prize) * 0.75),
		}
	}

	// Assign the purse for the third place finisher
	if len(lastLeg.ParticipantPositions) > 2 {
		racePosition := lastLeg.ParticipantPositions[1]
		race.RaceResult.Show = &RaceParticipantResult{
			Participant: racePosition.RaceParticipant,
			RaceTime:    racePosition.Speed,
			Winnings:    int(float64(prize) * 0.50),
		}
	}
}
