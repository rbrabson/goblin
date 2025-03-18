package blackjack

import "go.mongodb.org/mongo-driver/bson/primitive"

// HoldOn17 determines how the dealer plays when they are dealt a 17.
type HoldOn17 int

const (
	S17 HoldOn17 = iota
	H17
)

// Surrender determines if and when the player can surrender their hand.
type Surrender int

const (
	NoSurrender Surrender = iota
	LateSurrender
	EarlySurrender
)

// Config is the configuration for a game of blackjack.
type Config struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NumDecks  int                `json:"num_decks" bson:"num_decks"`
	HoldOn17  HoldOn17           `json:"hold_on_17" bson:"hold_on_17"`
	Surrender Surrender          `json:"surrender" bson:"surrender"`
	NumSplits int                `json:"num_splits" bson:"num_splits"`
	BetAount  int                `json:"bet_amount" bson:"bet_amount"`
}

func GetConfig(guildID string) *Config {
	return nil
}
