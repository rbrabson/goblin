package blackjack

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	bj "github.com/rbrabson/blackjack"
)

const (
	// symbolsFile = "symbols"
	symbolsFile = "symbols"
)

type Symbols map[string]map[string]string

func GetSymbols() Symbols {
	symbols := readSymbolsFromFile()
	if symbols == nil {
		symbols = getDefaultSymbols()
	}
	return symbols
}

// readSymbolsFromFile reads the symbols from the configuration file. If the file cannot be read or unmarshaled, nil is returned.
func readSymbolsFromFile() Symbols {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "blackjack", "symbols", symbolsFile+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read the symbols file",
			slog.String("file", symbolsFile),
			slog.Any("error", err),
		)
		return nil
	}

	var symbols Symbols
	err = json.Unmarshal(bytes, &symbols)
	if err != nil {
		slog.Error("failed to unmarshal the symbols file",
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err))
		return nil
	}

	slog.Info("loaded symbols from file",
		slog.String("symbolsFile", symbolsFile),
	)

	return symbols
}

// getDefaultSymbols returns the default symbols for the game.
func getDefaultSymbols() Symbols {
	return Symbols{
		"Cards": {
			"Multiple": "",
			"Back":     "🂠",
		},
		"Diamonds": {
			"Ace":   "A:diamonds:",
			"Two":   "2:diamonds:",
			"Three": "3:diamonds:",
			"Four":  "4:diamonds:",
			"Five":  "5:diamonds:",
			"Six":   "6:diamonds:",
			"Seven": "7:diamonds:",
			"Eight": "8:diamonds:",
			"Nine":  "9:diamonds:",
			"Ten":   "10:diamonds:",
			"Jack":  "J:diamonds:",
			"Queen": "Q:diamonds:",
			"King":  "K:diamonds:",
		},
		"Clubs": {
			"Ace":   "A:clubs:",
			"Two":   "2:clubs:",
			"Three": "3:clubs:",
			"Four":  "4:clubs:",
			"Five":  "5:clubs:",
			"Six":   "6:clubs:",
			"Seven": "7:clubs:",
			"Eight": "8:clubs:",
			"Nine":  "9:clubs:",
			"Ten":   "10:clubs:",
			"Jack":  "J:clubs:",
			"Queen": "Q:clubs:",
			"King":  "K:clubs:",
		},
		"Hearts": {
			"Ace":   "A:hearts:",
			"Two":   "2:hearts:",
			"Three": "3:hearts:",
			"Four":  "4:hearts:",
			"Five":  "5:hearts:",
			"Six":   "6:hearts:",
			"Seven": "7:hearts:",
			"Eight": "8:hearts:",
			"Nine":  "9:hearts:",
			"Ten":   "10:hearts:",
			"Jack":  "J:hearts:",
			"Queen": "Q:hearts:",
			"King":  "K:hearts:",
		},
		"Spades": {
			"Ace":   "A:spades:",
			"Two":   "2:spades:",
			"Three": "3:spades:",
			"Four":  "4:spades:",
			"Five":  "5:spades:",
			"Six":   "6:spades:",
			"Seven": "7:spades:",
			"Eight": "8:spades:",
			"Nine":  "9:spades:",
			"Ten":   "10:spades:",
			"Jack":  "J:spades:",
			"Queen": "Q:spades:",
			"King":  "K:spades:",
		},
	}
}

// GetHand returns a string representation of the hand using the provided symbols.
func (s Symbols) GetHand(hand *bj.Hand, hidden bool) string {
	cards := make([]string, 0, len(hand.Cards()))
	var sb strings.Builder
	for idx, card := range hand.Cards() {
		if hidden && idx == 0 {
			cards = append(cards, s["Cards"]["Back"])
		} else {
			card := s[card.Suit.String()][card.Rank.String()]
			cards = append(cards, card)
		}
	}
	sb.WriteString(strings.Join(cards, ""))
	sb.WriteString(fmt.Sprintf(" (value: %s)", GetHandValue(hand, hidden)))

	return sb.String()
}

// GetHandWithoutValue returns a string representation of the hand using the provided symbols.
func (s Symbols) GetHandWithoutValue(hand *bj.Hand, hidden bool) string {
	cards := make([]string, 0, len(hand.Cards()))
	var sb strings.Builder
	for idx, card := range hand.Cards() {
		if hidden && idx == 0 {
			cards = append(cards, s["Cards"]["Back"])
		} else {
			card := s[card.Suit.String()][card.Rank.String()]
			cards = append(cards, card)
		}
	}
	sb.WriteString(strings.Join(cards, ""))

	return sb.String()
}

// GetHandValue returns a string representation of the hand value using the provided symbols.
func GetHandValue(hand *bj.Hand, hidden bool) string {
	switch {
	case hand.IsBlackjack():
		return (" (blackjack)")
	case hand.IsBusted():
		return fmt.Sprintf("%d, busted", handValue(hand, hidden))
	case hand.IsSurrendered():
		return fmt.Sprintf("%d, surrendered", handValue(hand, hidden))
	default:
		return fmt.Sprintf("%d", handValue(hand, hidden))
	}
}
