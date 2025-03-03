package race

import (
	"errors"

	"github.com/rbrabson/goblin/internal/discmsg"
	"golang.org/x/text/language"
)

var (
	ErrAlreadyBetOnRace  = errors.New("you have already bet on the race")
	ErrAlreadyJoinedRace = errors.New("you have already joined the race")
	ErrBettingHasOpened  = errors.New("betting has opened, so you can't join the race")
	ErrBettingNotOpened  = errors.New("betting has not opened yet")
	ErrConfigNotFound    = errors.New("the configuration file not found")
	ErrMemberNotFound    = errors.New("the member was not found")
	ErrNoRacersFound     = errors.New("no racers were found")
	ErrRaceHasStarted    = errors.New("the race has already started")
	ErrRacerNotFound     = errors.New("the racer was not found")
)

// The maximum number of race members have already joined the race.
type ErrRaceFull struct {
	MaxNumRacersAllowed int
}

// Error returns the error message for ErrRaceFull.
func (e ErrRaceFull) Error() string {
	p := discmsg.GetPrinter(language.AmericanEnglish)
	return p.Sprintf("you can't join the race, as there are already %d entered into the race", e.MaxNumRacersAllowed)
}
