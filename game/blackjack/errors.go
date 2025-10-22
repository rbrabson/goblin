package blackjack

import "errors"

var (
	ErrGameActive          = errors.New("you cannot join an active game")
	ErrPlayerAlreadyInGame = errors.New("you already joined the game")
)
