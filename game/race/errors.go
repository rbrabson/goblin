package race

import (
	"errors"
	"time"

	"github.com/rbrabson/goblin/internal/format"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	ErrAlreadyBetOnRace      = errors.New("you have already bet on the race")
	ErrAlreadyJoinedRace     = errors.New("you have already joined the race")
	ErrBettingHasOpened      = errors.New("betting has opened, so you can't join the race")
	ErrBettingNotOpened      = errors.New("betting has not opened yet")
	ErrNoRacersFound         = errors.New("no racers were found")
	ErrRaceAlreadyInProgress = errors.New("you can't start a new race as one is already in progress")
	ErrRaceHasStarted        = errors.New("the race has already started")
	ErrRaceAlreadyFull       = errors.New("the race is already full")
)

// ErrRaceFull is the maximum number of race members have already joined the race.
type ErrRaceFull struct {
	MaxNumRacersAllowed int
}

// Error returns the error message for ErrRaceFull.
func (e ErrRaceFull) Error() string {
	p := message.NewPrinter(language.AmericanEnglish)
	return p.Sprintf("you can't join the race, as there are already %d entered into the race", e.MaxNumRacersAllowed)
}

// ErrRacersAreResting is the racers are resting, so the user should try again in a certain amount of time.
type ErrRacersAreResting struct {
	waitTime time.Duration
}

// Error returns the error message for ErrRacersAreResting.
func (e ErrRacersAreResting) Error() string {
	p := message.NewPrinter(language.AmericanEnglish)
	return p.Sprintf("The racers are resting. Try again in %s!", format.Duration(e.waitTime))
}
