package blackjack

import (
	"fmt"
	"strings"

	bj "github.com/rbrabson/blackjack"
)

type Symbols map[string]map[string]string

func GetSymbols() *Symbols {
	return &Symbols{
		"Suits": {
			"Diamonds": ":diamonds:",
			"Hearts":   ":hearts:",
			"Clubs":    ":clubs:",
			"Spades":   ":spades:",
		},
		"HiddenCard": {
			"Card": "🂠",
		},
		"Diamonds": {
			"Ace":   "A",
			"Two":   "2",
			"Three": "3",
			"Four":  "4",
			"Five":  "5",
			"Six":   "6",
			"Seven": "7",
			"Eight": "8",
			"Nine":  "9",
			"Ten":   "10",
			"Jack":  "J",
			"Queen": "Q",
			"King":  "K",
		},
		"Clubs": {
			"Ace":   "A",
			"Two":   "2",
			"Three": "3",
			"Four":  "4",
			"Five":  "5",
			"Six":   "6",
			"Seven": "7",
			"Eight": "8",
			"Nine":  "9",
			"Ten":   "10",
			"Jack":  "J",
			"Queen": "Q",
			"King":  "K",
		},
		"Hearts": {
			"Ace":   "A",
			"Two":   "2",
			"Three": "3",
			"Four":  "4",
			"Five":  "5",
			"Six":   "6",
			"Seven": "7",
			"Eight": "8",
			"Nine":  "9",
			"Ten":   "10",
			"Jack":  "J",
			"Queen": "Q",
			"King":  "K",
		},
		"Spades": {
			"Ace":   "A",
			"Two":   "2",
			"Three": "3",
			"Four":  "4",
			"Five":  "5",
			"Six":   "6",
			"Seven": "7",
			"Eight": "8",
			"Nine":  "9",
			"Ten":   "10",
			"Jack":  "J",
			"Queen": "Q",
			"King":  "K",
		},
	}
}

// GetHand returns a string representation of the hand using the provided symbols.
func (s Symbols) GetHand(hand *bj.Hand, hidden bool) string {
	cards := make([]string, 0, len(hand.Cards()))
	var sb strings.Builder
	for idx, card := range hand.Cards() {
		if hidden && idx == 0 {
			cards = append(cards, s["HiddenCard"]["Card"])
		} else {
			suit := s["Suits"][card.Suit.String()]
			rank := s[card.Suit.String()][card.Rank.String()]
			cards = append(cards, rank+suit)
		}
	}
	sb.WriteString(strings.Join(cards, " "))
	switch {
	case hand.IsBlackjack():
		sb.WriteString(" (blackjack)")
	case hand.IsBusted():
		sb.WriteString(" (busted)")
	default:
		sb.WriteString(" (value: ")
		sb.WriteString(fmt.Sprintf("%d", handValue(hand, hidden)))
		sb.WriteString(")")
	}

	return sb.String()
}
