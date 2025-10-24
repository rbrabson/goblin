package blackjack

import (
	bj "github.com/rbrabson/blackjack"
)

type Symbols map[string]map[string]string

func GetSymbols() *Symbols {
	return &Symbols{
		"cards": {
			"S": "♠",
			"H": "♥",
			"D": "♦",
			"C": "♣",
		},
		"actions": {
			"hit":        "🃏",
			"stand":      "✋",
			"doubleDown": "💰",
			"split":      "🔀",
		},
		"hand": {
			"hidden":    "🂠",
			"busted":    "💥",
			"blackjack": "🃏",
		},
	}
}

func (s Symbols) GetHand(hand *bj.Hand, hidden bool) string {
	if hidden {
		return hand.StringHidden()
	}
	return hand.String()
}
