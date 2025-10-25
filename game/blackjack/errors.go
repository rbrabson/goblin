package blackjack

import "errors"

var (
	ErrGameActive          = errors.New("the game has already started")
	ErrPlayerAlreadyInGame = errors.New("you already joined the game")
	ErrGameFull            = errors.New("the game is already full")
	ErrPlayerNotFound      = errors.New("player not found in the game")
)
