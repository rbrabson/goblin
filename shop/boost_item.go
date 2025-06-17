package shop

import "go.mongodb.org/mongo-driver/bson/primitive"

// Game is a type used to define the game to which an operation, boost, or effect applies.
type Game string

const (
	RaceGame  Game = "race"  // RaceGame specifies that an operation, boost, or effect applies to the race game.
	HeistGame Game = "heist" // HeistGame specifies that an operation, boost, or effect applies to the heist game.
)

// Effect is a type used to specify the aspect or attribute that an operation, change, or effect influences.
type Effect string

const (
	Speed       Effect = "speed"  // Speed modifies the speed of the race avatar.
	EscapeBonus Effect = "escape" // EscapeBonus modifies the likelihood that a crew member escapes during a heist.
)

// BoostItem is a structure defining a boost that may be purchased and which affects the behavior of a game.
type BoostItem struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	Game   Game               `json:"game" bson:"game"`
	Effect Effect             `json:"effect" bson:"effect"`
}
