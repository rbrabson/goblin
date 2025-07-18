package race

import (
	"log/slog"
	"math/rand/v2"
	"sort"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
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
	raceAvatars   []*Avatar                    // The avatars of the racers
	interaction   *discordgo.InteractionCreate // Interaction used in sending message updates
	config        *Config                      // Race configuration (avoids having to read from the database)
	mutex         sync.Mutex                   // Lock used to synchronize access to the race
}

// RaceParticipant is a member who is racing. This includes the member and the racer assigned to them.
type RaceParticipant struct {
	Member *RaceMember // Member who is racing
	Racer  *Avatar     // Racer assigned to the member
}

// RaceBetter is a member who is betting on the outcome of the race.
type RaceBetter struct {
	Member   *RaceMember      // Member who is betting on the outcome of the the race
	Racer    *RaceParticipant // Racer on which the member is betting
	Winnings int              // Amount won by the better
}

// RaceResult is the final results of the race. This includes the winner, 2nd place, and 3rd place finishers, as
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

// RaceParticipantPosition is used to track the movement of a given member during a single leg of a race.
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
	slog.Info("new race",
		slog.String("guildID", guildID),
	)

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
func (r *Race) addBetter(better *RaceBetter) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Betters = append(r.Betters, better)
	slog.Debug("add better to current race",
		slog.String("guildID", r.GuildID),
		slog.String("memberID", better.Member.MemberID),
	)

	return nil
}

// RunRace runs a race, calculating the results of each leg of the race and the
// ultimate winners of the race.
func (r *Race) RunRace(trackLength int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Create the initial starting positions and add them to an initial race leg
	raceLeg := &RaceLeg{
		ParticipantPositions: make([]*RaceParticipantPosition, 0, len(r.Racers)),
	}
	for _, racer := range r.Racers {
		participantPosition := &RaceParticipantPosition{
			RaceParticipant: racer,
			Position:        trackLength,
		}
		raceLeg.ParticipantPositions = append(raceLeg.ParticipantPositions, participantPosition)
	}
	r.RaceLegs = append(r.RaceLegs, raceLeg)
	previousLeg := raceLeg

	// Run the race until all racers cross the finish line
	slog.Debug("starting race",
		slog.String("guildID", r.GuildID),
		slog.Int("numRacers", len(r.Racers)),
		slog.Int("trackLength", trackLength),
	)
	turn := 0
	stillRacing := true
	for stillRacing {
		turn++

		// Create and add a new race leg
		newRaceLeg := &RaceLeg{
			ParticipantPositions: make([]*RaceParticipantPosition, 0, len(r.Racers)),
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

		r.RaceLegs = append(r.RaceLegs, newRaceLeg)
		previousLeg = newRaceLeg
	}

	calculateWinnings(r, previousLeg)
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
		slog.Info("processing race results",
			slog.String("guildID", r.GuildID),
			slog.Int("numRacers", len(r.Racers)),
		)
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

		slog.Info("processing race bets",
			slog.String("guildID", r.GuildID),
			slog.Int("numBetters", len(r.Betters)),
		)
		// Pay the winning bets
		for _, better := range r.Betters {
			if better.Winnings != 0 {
				better.Member.WinBet(better.Winnings)
			} else {
				better.Member.LoseBet()
			}
		}
	}

	slog.Info("end race",
		slog.String("guildID", r.GuildID),
	)
}

// ResetRace resets a hung race for a given guild.
func ResetRace(guildID string) {
	delete(currentRaces, guildID)
	delete(lastRaceTimes, guildID)
	slog.Info("reset race",
		slog.String("guildID", guildID),
	)
}

// getRaceAvatar returns a  random race avatar to be used by a race participant.
func getRaceAvatar(race *Race) *Avatar {
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
		slog.Debug("race already in progress",
			slog.String("guildID", guildID),
		)
		return ErrRaceAlreadyInProgress
	}

	lastRaceTime := lastRaceTimes[guildID]
	if time.Since(lastRaceTime) < config.WaitBetweenRaces {
		timeSinceLastRace := time.Since(lastRaceTime)
		timeUntilRaceCanStart := config.WaitBetweenRaces - timeSinceLastRace
		slog.Debug("racers are resting",
			slog.String("guildID", guildID),
			slog.Duration("timeUntilRaceCanStart", timeUntilRaceCanStart),
		)
		return ErrRacersAreResting{timeUntilRaceCanStart}
	}
	delete(lastRaceTimes, guildID)

	return nil
}

// raceJoinChecks checks to see if a racer is able to join the race.
func raceJoinChecks(race *Race, memberID string) error {
	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart + race.config.WaitForBets)) {
		slog.Warn("race has started",
			slog.String("guildID", race.GuildID),
		)
		return ErrRaceHasStarted
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart)) {
		slog.Warn("betting has opened",
			slog.String("guildID", race.GuildID),
		)
		return ErrBettingHasOpened
	}

	if len(race.Racers) >= race.config.MaxNumRacers {
		slog.Warn("too many racers already joined",
			slog.String("guildID", race.GuildID),
			slog.Int("maxNumRacers", race.config.MaxNumRacers),
			slog.Int("numRacers", len(race.Racers)),
		)
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
		slog.Warn("betting has opened",
			slog.String("guildID", race.GuildID),
		)
		return ErrBettingNotOpened
	}

	if time.Now().After(race.RaceStartTime.Add(race.config.WaitToStart + race.config.WaitForBets)) {
		slog.Warn("race has started",
			slog.String("guildID", race.GuildID),
		)
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
	prize := r.IntN(race.config.MaxPrizeAmount-race.config.MinPrizeAmount) + race.config.MinPrizeAmount
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
