package heist

import (
	"errors"
	"fmt"
	"time"

	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/format"
	"golang.org/x/text/language"
)

var (
	ErrConfigNotFound      = errors.New("configuration file not found")
	ErrHeistInProgress     = errors.New("heist is already in progress")
	ErrAlreadyJoinedHieist = errors.New("you have already joined the heist")
	ErrNoHeist             = errors.New("heist not found")
	ErrNotAllowed          = errors.New("user is not allowed to perform command")
	ErrThemeNotFound       = errors.New("theme not found")
)

// ErrNotEnoughMembers is returned when there are not enough members to start a heist.
type ErrNotEnoughMembers struct {
	Theme Theme
}

// Error returns the error message for ErrNotEnoughMembers.
func (e ErrNotEnoughMembers) Error() string {
	p := discmsg.GetPrinter(language.AmericanEnglish)
	return p.Sprintf("You tried to rally a %s, but no one wanted to follow you. The %s has been cancelled.", e.Theme.Crew, e.Theme.Heist)
}

// ErrNoTargets is returned when there are no targets available to hit.
type ErrNoTargets struct{}

// Error returns the error message for ErrNoTargets.
func (e ErrNoTargets) Error() string {
	return "Oh no! There are no targets!"
}

// ErrNotEnoughCredits is returned when a user does not have enough credits to participate in a heist.
type ErrNotEnoughCredits struct {
	CreditsNeeded int
}

// Error returns the error message for ErrNotEnoughCredits.
func (e ErrNotEnoughCredits) Error() string {
	p := discmsg.GetPrinter(language.AmericanEnglish)
	return p.Sprintf("You do not have enough credits to cover the cost of entry. You need %d credits to participate",
		e.CreditsNeeded,
	)
}

// ErrPoliceOnAlert is returned when the police are on high alert after the last target.
type ErrPoliceOnAlert struct {
	Police        string
	RemainingTime time.Duration
}

// Error returns the error message for ErrPoliceOnAlert.
func (e ErrPoliceOnAlert) Error() string {
	p := discmsg.GetPrinter(language.AmericanEnglish)
	return p.Sprintf("The %s are on high alert after the last target. We should wait for things to cool off before hitting another target. Time remaining: %s.",
		e.Police,
		format.Duration(e.RemainingTime),
	)
}

// ErrInJail is returned when a user is in jail.
type ErrInJail struct {
	Jail          string
	Sentence      string
	RemainingTime time.Duration
	Bail          string
	BailCost      int
}

// Error returns the error message for ErrInJail.
func (e ErrInJail) Error() string {
	remainingTime := format.Duration(e.RemainingTime)
	return fmt.Sprintf("You are in %s. You can wait out your remaining %s of %s, or pay %d credits to be released on %s.",
		e.Jail, e.Sentence, remainingTime, e.BailCost, e.Bail,
	)
}

// ErrDead is returned when a user is dead.
type ErrDead struct {
	RemainingTime time.Duration
}

// Error returns the error message for ErrDead.
func (e ErrDead) Error() string {
	return fmt.Sprintf("You are dead. You will revive in %s", format.Duration(e.RemainingTime))
}
