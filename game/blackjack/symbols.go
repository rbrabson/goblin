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
	symbolsFile = "pt"
)

type Symbols map[string]map[string]string

func GetSymbols() *Symbols {
	symbols := readSymbolsFromFile()
	if symbols == nil {
		symbols = getDefaultSymbols()
	}
	return symbols
}

// readSymbolsFromFile reads the symbols from the configuration file. If the file cannot be read or unmarshaled, nil is returned.
func readSymbolsFromFile() *Symbols {
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

	return &symbols
}

// getDefaultSymbols returns the default symbols for the game.
func getDefaultSymbols() *Symbols {
	return &Symbols{
		"Suits": {
			"Diamonds": ":diamonds:",
			"Hearts":   ":hearts:",
			"Clubs":    ":clubs:",
			"Spades":   ":spades:",
		},
		"Cards": {
			"Multiple": "<:multiple_cards:1431696869970723840>",
			"Back":     "<:backside_of_card:1431696868949360751>",
		},
		"Diamonds": {
			"Ace":   "<:ace_of_diamonds:1431696865648316427>",
			"Two":   "<:two_of_diamonds:1431696928386715689>",
			"Three": "<:three_of_diamonds:1431696924603318544>",
			"Four":  "<:four_of_diamonds:1431696883402670154>",
			"Five":  "<:five_of_diamonds:1431696879569211402>",
			"Six":   "<:six_of_diamonds:1431696912842490018>",
			"Seven": "<:seven_of_diamonds:1431696908392468530>",
			"Eight": "<:eight_of_diamonds:1431696874087252130>",
			"Nine":  "<:nine_of_diamonds:1431696898883981453>",
			"Ten":   "<:ten_of_diamonds:1431696918534295572>",
			"Jack":  "<:jack_of_diamonds:1431696889484541982>",
			"Queen": "<:queen_of_diamonds:1431696903036207305>",
			"King":  "<:king_of_diamonds:1431696894517842052>",
		},
		"Clubs": {
			"Ace":   "<:ace_of_clubs:1431696862737596418>",
			"Two":   "<:two_of_clubs:1431696927472357446>",
			"Three": "<:three_of_clubs:1431696922976063619>",
			"Four":  "<:four_of_clubs:1431696882832380065>",
			"Five":  "<:five_of_clubs:1431696878696923319>",
			"Six":   "<:six_of_clubs:1431696911920009276>",
			"Seven": "<:seven_of_clubs:1431696906437918851>",
			"Eight": "<:eight_of_clubs:1431696872992407702>",
			"Nine":  "<:nine_of_clubs:1431696898007371796>",
			"Ten":   "<:ten_of_clubs:1431696916802174996>",
			"Jack":  "<:jack_of_clubs:1431696888402415928>",
			"Queen": "<:queen_of_clubs:1431696902176637088>",
			"King":  "<:king_of_clubs:1431696892982464582>",
		},
		"Hearts": {
			"Ace":   "<:ace_of_hearts:1431696866625458267>",
			"Two":   "<:two_of_hearts:1431696929288355960>",
			"Three": "<:three_of_hearts:1431696925496967329>",
			"Four":  "<:four_of_hearts:1431696884791115849>",
			"Five":  "<:five_of_hearts:1431696880319860746>",
			"Six":   "<:six_of_hearts:1431696914394382446>",
			"Seven": "<:seven_of_hearts:1431696909877383369>",
			"Eight": "<:eight_of_hearts:1431696875408461915>",
			"Nine":  "<:nine_of_hearts:1431696899500544002>",
			"Ten":   "<:ten_of_hearts:1431696920429985842>",
			"Jack":  "<:jack_of_hearts:1431696890864599271>",
			"Queen": "<:queen_of_hearts:1431696904206422046>",
			"King":  "<:king_of_hearts:1431696895906021529>",
		},
		"Spades": {
			"Ace":   "<:ace_of_spades:1431696867661713418>",
			"Two":   "<:two_of_spades:1431696930114633879>",
			"Three": "<:three_of_spades:1431696926704926840>",
			"Four":  "<:four_of_spades:1431696886460452916>",
			"Five":  "<:five_of_spades:1431696881653776595>",
			"Six":   "<:six_of_spades:1431696915682168882>",
			"Seven": "<:seven_of_spades:1431696910934081567>",
			"Eight": "<:eight_of_spades:1431696876603703529>",
			"Nine":  "<:nine_of_spades:1431696901320867982>",
			"Ten":   "<:ten_of_spades:1431696921923289178>",
			"Jack":  "<:jack_of_spades:1431696892135473292>",
			"Queen": "<:queen_of_spades:1431696905334947892>",
			"King":  "<:king_of_spades:1431696896770052188>",
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
	sb.WriteString(strings.Join(cards, " "))
	switch {
	case hand.IsBlackjack():
		sb.WriteString(" (blackjack)")
	case hand.IsBusted():
		sb.WriteString(" (busted)")
	case hand.IsSurrendered():
		sb.WriteString(" (surrendered)")
	default:
		sb.WriteString(" (value: ")
		sb.WriteString(fmt.Sprintf("%d", handValue(hand, hidden)))
		sb.WriteString(")")
	}

	return sb.String()
}
