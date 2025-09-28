package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	PAYOUT_FILE_NAME      = "payout"
	TwoConsecutiveSymbols = "[two consecutive non-Spell symbols]"
	AnyOrderRWB           = "[any order AQ/Archer, GW/Wizard, BK/Barbarian]"
)

// Payout defines a winning combination and the payout amounts for different bets.
type Payout struct {
	Win     []Slot  `json:"win" bson:"win"`
	Bet     int     `json:"bet" bson:"bet"`
	Payout  float64 `json:"payout" bson:"payout"`
	Message string  `json:"message" bson:"message"`
}

// String returns a string representation of the Payout.
func (p *Payout) String() string {
	sb := strings.Builder{}
	sb.WriteString("Payout{")
	sb.WriteString("Win: [")
	for i, slot := range p.Win {
		sb.WriteString(slot.String())
		if i < len(p.Win)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	sb.WriteString(", Payouts: [")
	sb.WriteString("Bet: " + strconv.Itoa(p.Bet))
	sb.WriteString(", Payout: " + strconv.FormatFloat(p.Payout, 'f', -1, 64))
	sb.WriteString("]")
	sb.WriteString("}")

	return sb.String()
}

// PayoutAmount defines a winning combination and the payout amounts for different bets.
type PayoutAmount struct {
	Win     []string `json:"win" bson:"win"`
	Bet     int      `json:"bet" bson:"bet"`
	Payout  float64  `json:"payout" bson:"payout"`
	Message string   `json:"message" bson:"message"`
}

// String returns a string representation of the PayoutAmount.
func (p *PayoutAmount) String() string {
	sb := strings.Builder{}
	sb.WriteString("PayoutAmount{")
	sb.WriteString("Win: [")
	for _, slot := range p.Win {
		sb.WriteString(slot)
	}
	sb.WriteString("]")
	sb.WriteString(", Bet: " + strconv.Itoa(p.Bet))
	sb.WriteString(", Payout: " + strconv.FormatFloat(p.Payout, 'f', -1, 64))
	sb.WriteString("]")
	sb.WriteString("}")

	return sb.String()
}

// PayoutTable defines a table of payouts for a specific guild.
type PayoutTable []PayoutAmount

// String returns a string representation of the PayoutTable.
func (pt PayoutTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("PayoutTable{")
	sb.WriteString(", Payouts: [")
	for _, payout := range pt {
		sb.WriteString(", " + payout.String())
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}

// GetPayoutTable retrieves the payout table for a specific guild.
func GetPayoutTable() PayoutTable {
	pt := newPayoutTable()
	slices.SortFunc(pt, func(a, b PayoutAmount) int {
		if a.Bet != b.Bet {
			return a.Bet - b.Bet
		}
		comparision := b.Payout - a.Payout
		if comparision < 0 {
			return -1
		}
		if comparision > 0 {
			return 1
		}
		return 0

	})
	return pt
}

// newPayoutTable creates a new payout table for a specific guild by reading from a file.
func newPayoutTable() PayoutTable {
	payoutTable := readPayoutTableFromFile()
	return payoutTable
}

// readPayoutTableFromFile reads the payout table from a JSON file.
func readPayoutTableFromFile() PayoutTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "payout", PAYOUT_FILE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default payout table",
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	payouts := &[]Payout{}
	err = json.Unmarshal(bytes, payouts)
	if err != nil {
		slog.Error("failed to unmarshal payout table",
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}

	payoutTable := make(PayoutTable, 0, len(*payouts))

	for _, payout := range *payouts {
		payoutAmount := PayoutAmount{
			Win:     make([]string, 0, len(payout.Win)),
			Bet:     payout.Bet,
			Payout:  payout.Payout,
			Message: payout.Message,
		}
		for _, slot := range payout.Win {
			payoutAmount.Win = append(payoutAmount.Win, string(slot))
		}
		payoutTable = append(payoutTable, payoutAmount)
	}

	slog.Debug("create new payout table")

	return payoutTable
}

// GetPayoutAmount returns the payout amount for a given bet and spin result.
func (pt PayoutTable) GetPayoutAmount(bet int, spin []Symbol) (int, string) {
	for _, payout := range pt {
		match := true
		for i := range payout.Win {
			winningSymbols := strings.Split(payout.Win[i], " or ")
			if !slices.Contains(winningSymbols, spin[i].Name) {
				match = false
				break
			}
		}
		if match {
			slog.Debug("found matching payout",
				slog.Int("bet", bet),
				slog.Any("win", payout.Win),
				slog.Int("bet", payout.Bet),
				slog.Float64("payout", payout.Payout),
			)
			amount := payout.Payout * (float64(bet) / float64(payout.Bet))
			return int(amount), payout.Message
		}
	}

	// Check for two consecutive non-blank symbols (not "Spell")
	if (spin[0].Name != "Spell" && spin[0].Name == spin[1].Name && spin[1].Name != spin[2].Name) ||
		(spin[0].Name != spin[1].Name && spin[1].Name != "Spell" && spin[1].Name == spin[2].Name) {
		var payoutAmount PayoutAmount
		for _, p := range pt {
			if len(p.Win) == 1 && p.Win[0] == TwoConsecutiveSymbols {
				payoutAmount = p
				break
			}
		}

		payout := int(payoutAmount.Payout * float64(bet))
		slog.Debug("found matching payout for two consecutive non-blank symbols",
			slog.Int("bet", bet),
			slog.Int("payout", payout),
		)
		return payout, payoutAmount.Message
	}

	slog.Debug("no matching payout found",
		slog.Int("bet", bet),
	)
	return 0, ""
}
