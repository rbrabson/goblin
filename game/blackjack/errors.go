package blackjack

import "errors"

var (
	ErrAllPlayersBusted    = errors.New("all players have busted. The dealer wins.")
	ErrCannotDoubleDown    = errors.New("you cannot double down on this hand.")
	ErrCannotSplit         = errors.New("you cannot split this hand.")
	ErrCannotSurrender     = errors.New("you cannot surrender this hand.")
	ErrGameActive          = errors.New("the game has already started.")
	ErrGameFull            = errors.New("the game is already full.")
	ErrGameNotStarted      = errors.New("the blackjack game has not started yet. Please wait for the game to start before joining.")
	ErrGameStarted         = errors.New("an active blackjack game already exists.")
	ErrNoActiveGame        = errors.New("there is no active blackjack game. Please start a new game to play.")
	ErrNotActivePlayer     = errors.New("you are not the active player.")
	ErrPlayerAlreadyInGame = errors.New("you already joined the game.")
	ErrPlayerNotFound      = errors.New("player not found in the game.")
)
