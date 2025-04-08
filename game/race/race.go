package race

import (
	"math/rand/v2"
	"sort"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var (
	lastRaceTimes = make(map[string]time.Time)
	currentRaces  = make(map[string]*Race)
	raceLock      = sync.Mutex{}
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
	RaceStartTime time.Time                    // The time at which the race is started (first created)
	raceAvatars   []*RaceAvatar                // The avatars of the racers
	interaction   *discordgo.InteractionCreate // Interaction used in sending message updates
	config        *Config                      // Race configuration (avoids having to read from the database)
	mutex         sync.Mutex                   // Lock used to synchronize access to the race
}

// RaceParticpant is a member who is racing. This includes the member and the racer assigned to them.
type RaceParticipant struct {
	Member *RaceMember // Member who is racing
	Racer  *RaceAvatar // Racer assigned to the member
}

// RaceBetter is a member who is betting on the outcome of the race.
type RaceBetter struct {
	Member   *RaceMember      // Member who is betting on the outcome of the the race
	Racer    *RaceParticipant // Racer on which the member is betting
	Winnings int              // Amount won by the better
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

// GetRace gets the race for the guild. If a race isn't in progress, then a new one is created.
func GetRace(guildID string) *Race {
	race := currentRaces[guildID]
	if race == nil {
		race = newRace(guildID)
	}

	return race
}

// newRace creates a new race for the guild.
func newRace(guildID string) *Race {
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
	race.raceAvatars = GetRaceAvatars(race.GuildID, race.config.Theme)

	currentRaces[guildID] = race
	log.WithFields(log.Fields{"guild": guildID}).Info("new race")

	return race
}

// addRaceParticiapnt returns a new race participant for a member in the race. The race
// participant is added to the race.
func (r *Race) addRaceParticipant(member *RaceMember) *RaceParticipant {
	participant := &RaceParticipant{
		Member: member,
		Racer:  getRaceAvatar(r),
	}
	member.TotalRaces++
	r.mutex.Lock()
	r.Racers = append(r.Racers, participant)
	defer r.mutex.Unlock()

	return participant
}

// getRaceParticipant returns a racer for a given race. If the member isn't in the race, then
// nil is returned.
func (r *Race) getRaceParticipant(memberID string) *RaceParticipant {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, racer := range r.Racers {
		if racer.Member.MemberID == memberID {
			return racer
		}
	}
	return nil
}

// getRaceBetter returns a new better for a race.
func getRaceBetter(member *RaceMember, racer *RaceParticipant) *RaceBetter {
	raceBetter := &RaceBetter{
		Member: member,
		Racer:  racer,
	}

	return raceBetter
}

// addBetter adds a better for the given race.
func (race *Race) addBetter(better *RaceBetter) error {
	race.mutex.Lock()
	defer race.mutex.Unlock()

	race.Betters = append(race.Betters, better)
	log.WithFields(log.Fields{"guild": race.GuildID, "better": better.Member.MemberID}).Debug("add better to current race")

	return nil
}

// RunRace runs a race, calculating the results of each leg of the race and the
// ultimate winners of the race.
func (race *Race) RunRace(trackLength int) {
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
	log.WithFields(log.Fields{"guildID": race.GuildID, "numRacers": len(race.Racers), "trackLength": trackLength}).Debug("starting race")
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
		log.WithFields(log.Fields{"guildID": race.GuildID, "turn": turn}).Trace("completed race leg")
	}

	calculateWinnings(race, previousLeg)
}

// End ends the current race.
func (r *Race) End() {
	raceLock.Lock()
	defer raceLock.Unlock()

	// The race runs if there are 2 or more racers. If that is the case, then reset the time the last
	// successful race ran.
	race := currentRaces[r.GuildID]
	if race != nil && len(race.Racers) >= r.config.MinNumRacers {
		lastRaceTimes[r.GuildID] = time.Now()
	}

	delete(currentRaces, r.GuildID)

	if r.RaceResult != nil && len(r.Racers) >= r.config.MinNumRacers {
		log.WithFields(log.Fields{"guild": r.GuildID, "racers": len(r.Racers)}).Info("processing race results")
		for _, racer := range r.Racers {
			switch {
			case r.RaceResult.Win != nil && racer.Member.MemberID == r.RaceResult.Win.Participant.Member.MemberID:
				racer.Member.WinRace(r.RaceResult.Win.Winnings)
			case r.RaceResult.Place != nil && racer.Member.MemberID == r.RaceResult.Place.Participant.Member.MemberID:
				racer.Member.PlaceInRace(r.RaceResult.Place.Winnings)
			case r.RaceResult.Show != nil && racer.Member.MemberID == r.RaceResult.Show.Participant.Member.MemberID:
				racer.Member.ShowInRace(r.RaceResult.Show.Winnings)
			default:
				racer.Member.LoseRace()
			}
		}

		log.WithFields(log.Fields{"guild": r.GuildID, "betters": len(r.Betters)}).Info("processing race bets")
		for _, better := range r.Betters {
			if better.Winnings != 0 {
				better.Member.WinBet(better.Winnings)
			} else {
				better.Member.LoseBet()
			}
		}
	}

	log.WithFields(log.Fields{"guild": r.GuildID}).Info("end race")
}

// ResetRace resets a hung race for a given guild.
func ResetRace(guildID string) {
	log.Trace("---> race.ResetRace")
	defer log.Trace("<-- race.ResetRace")

	delete(currentRaces, guildID)
	delete(lastRaceTimes, guildID)
	log.WithFields(log.Fields{"guild": guildID}).Info("reset race")
}

// getRaceAvatar returns a  random race avatar to be used by a race participant.
func getRaceAvatar(race *Race) *RaceAvatar {
	if len(race.raceAvatars) == 0 {
		race.raceAvatars = GetRaceAvatars(race.GuildID, race.config.Theme)
	}

	index := len(race.raceAvatars) - 1
	avatar := race.raceAvatars[index]
	race.raceAvatars[index] = nil
	race.raceAvatars = race.raceAvatars[:index]
	return avatar
}

// Move returns the new race position for a particpant based on the previous position and the current turn.
func Move(previousPosition *RaceParticipantPosition, turn int) *RaceParticipantPosition {
	// Already done with the race
	if previousPosition.Position <= 0 {
		newPosition := &RaceParticipantPosition{
			RaceParticipant: previousPosition.RaceParticipant,
			Finished:        true,
			Speed:           previousPosition.Speed,
		}
		log.WithFields(log.Fields{"guildID": previousPosition.RaceParticipant.Member.GuildID, "memberID": previousPosition.RaceParticipant.Member.MemberID}).Trace("racer already finished")
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

// raceStartChecks checks to see if a race can be started.
func raceStartChecks(guildID string, memberID string) error {
	config := GetConfig(guildID)

	race := currentRaces[guildID]
	if race != nil {
		log.WithFields(log.Fields{"guild_id": guildID}).Debug("race already in progress")
		return ErrRaceAlreadyInProgress
	}

	lastRaceTime := lastRaceTimes[guildID]
	if time.Since(lastRaceTime) < config.WaitBetweenRaces {
		timeSinceLastRace := time.Since(lastRaceTime)
		timeUntilRaceCanStart := config.WaitBetweenRaces - timeSinceLastRace
		log.WithFields(log.Fields{"guild_id": guildID, "timeUntilRaceCanStart": timeUntilRaceCanStart}).Debug("racers are resting")
		return ErrRacersAreResting{timeUntilRaceCanStart}
	}
	delete(lastRaceTimes, guildID)

	return nil
}

// raceJoinChecks checks to see if a racer is able to join the race.
func raceJoinChecks(race *Race, memberID string) error {
	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart + race.config.WaitForBets)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("race has started")
		return ErrRaceHasStarted
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("betting has opened")
		return ErrBettingHasOpened
	}

	if len(race.Racers) >= race.config.MaxNumRacers {
		log.WithFields(log.Fields{"guild_id": race.GuildID, "maxNumRacers": race.config.MaxNumRacers, "numRacers": len(race.Racers)}).Warn("too many racers already joined")
		return ErrRaceAlreadyFull
	}

	for _, r := range race.Racers {
		if r.Member.MemberID == memberID {
			return ErrAlreadyJoinedRace
		}
	}

	return nil
}

// raceBetChecks checks to see if a better is able to place a bet on the current race.
func raceBetChecks(race *Race, memberID string) error {
	if time.Now().Before(race.RaceStartTime.Add(race.config.WaitToStart)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("betting has opened")
		return ErrBettingNotOpened
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart + race.config.WaitForBets)) {
		log.WithFields(log.Fields{"guild_id": race.GuildID}).Warn("race has started")
		return ErrRaceHasStarted
	}

	for _, b := range race.Betters {
		if b.Member.MemberID == memberID {
			return ErrAlreadyBetOnRace
		}
	}

	return nil
}

// calculateWinngins calculates the earnings for the racers that wins, places and shows.
func calculateWinnings(race *Race, lastLeg *RaceLeg) {
	source := rand.NewPCG(rand.Uint64(), rand.Uint64())
	r := rand.New(source)
	// sort the participants in the final race leg
	sort.Slice(lastLeg.ParticipantPositions, func(i, j int) bool {
		if lastLeg.ParticipantPositions[i].Speed == lastLeg.ParticipantPositions[j].Speed {
			return r.IntN(2) == 0
		}
		return lastLeg.ParticipantPositions[i].Speed < lastLeg.ParticipantPositions[j].Speed
	})

	// Calculate the winners of the race and save in the results
	prize := r.IntN(int(race.config.MaxPrizeAmount-race.config.MinPrizeAmount)) + race.config.MinPrizeAmount
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
		racePosition := lastLeg.ParticipantPositions[2]
		race.RaceResult.Show = &RaceParticipantResult{
			Participant: racePosition.RaceParticipant,
			RaceTime:    racePosition.Speed,
			Winnings:    int(float64(prize) * 0.50),
		}
	}

	// Pay the winning bets
	if race.RaceResult.Win != nil {
		winner := race.RaceResult.Win.Participant
		winningBet := race.config.BetAmount * len(race.Racers)
		for _, better := range race.Betters {
			if better.Racer == winner {
				better.Winnings = winningBet
			}
		}
	}
}
