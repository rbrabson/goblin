package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PAYOUT_FILE_NAME = "payout"
)

// Payout defines a winning combination and the payout amounts for different bets.
type Payout struct {
	Win    []Slot `json:"win" bson:"win"`
	Bet    int    `json:"bet" bson:"bet"`
	Payout int    `json:"payout" bson:"payout"`
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
	sb.WriteString(", Payout: " + strconv.Itoa(p.Payout))
	sb.WriteString("]")
	sb.WriteString("}")

	return sb.String()
}

// PayoutAmount defines a winning combination and the payout amounts for different bets.
type PayoutAmount struct {
	Win    []string `json:"win" bson:"win"`
	Bet    int      `json:"bet" bson:"bet"`
	Payout int      `json:"payout" bson:"payout"`
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
	sb.WriteString(", Payout: " + strconv.Itoa(p.Payout))
	sb.WriteString("]")
	sb.WriteString("}")

	return sb.String()
}

// PayoutTable defines a table of payouts for a specific guild.
type PayoutTable struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id"`
	Payouts []PayoutAmount     `json:"payouts"`
}

// String returns a string representation of the PayoutTable.
func (pt *PayoutTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("PayoutTable{")
	sb.WriteString("ID: " + pt.ID.Hex())
	sb.WriteString(", GuildID: " + pt.GuildID)
	sb.WriteString(", Payouts: [")
	for _, payout := range pt.Payouts {
		sb.WriteString(", " + payout.String())
	}
	sb.WriteString("]")
	sb.WriteString("}")
	return sb.String()
}

// GetPayoutTable retrieves the payout table for a specific guild.
func GetPayoutTable(guildID string) *PayoutTable {
	pt := newPayoutTable(guildID)
	slices.SortFunc(pt.Payouts, func(a, b PayoutAmount) int {
		if a.Bet != b.Bet {
			return a.Bet - b.Bet
		}
		return b.Payout - a.Payout
	})
	return pt
}

// newPayoutTable creates a new payout table for a specific guild by reading from a file.
func newPayoutTable(guildID string) *PayoutTable {
	payoutTable := readPayoutTableFromFile(guildID)
	if payoutTable == nil {
		slog.Error("failed to create lookup table",
			slog.String("guildID", guildID),
		)
		return nil
	}
	return payoutTable
}

// readPayoutTableFromFile reads the payout table from a JSON file.
func readPayoutTableFromFile(guildID string) *PayoutTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "slots", "payout", PAYOUT_FILE_NAME+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default payout table",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	payouts := &[]Payout{}
	err = json.Unmarshal(bytes, payouts)
	if err != nil {
		slog.Error("failed to unmarshal payout table",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}

	payoutTable := &PayoutTable{
		GuildID: guildID,
		Payouts: make([]PayoutAmount, 0, len(*payouts)),
	}

	for _, payout := range *payouts {
		slog.Debug("loaded payout",
			slog.String("guildID", guildID),
			slog.Any("payout", payout),
		)

		payoutAmount := PayoutAmount{
			Win:    make([]string, 0, len(payout.Win)),
			Bet:    payout.Bet,
			Payout: payout.Payout,
		}
		for _, slot := range payout.Win {
			payoutAmount.Win = append(payoutAmount.Win, string(slot))
		}
		payoutTable.Payouts = append(payoutTable.Payouts, payoutAmount)
	}

	slog.Debug("create new payout table",
		slog.String("guildID", payoutTable.GuildID),
	)

	return payoutTable
}

// GetPayoutAmount returns the payout amount for a given bet and spin result.
func (pt *PayoutTable) GetPayoutAmount(bet int, spin []Symbol) int {
	for _, payout := range pt.Payouts {
		if len(payout.Win) != len(spin) {
			slog.Warn("payout win length does not match spin length",
				slog.String("guildID", pt.GuildID),
				slog.Int("bet", bet),
				slog.Any("win", payout.Win),
				slog.Any("spin", spin),
			)
			continue
		}
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
				slog.String("guildID", pt.GuildID),
				slog.Int("bet", bet),
				slog.Any("win", payout.Win),
				slog.Any("spin", spin),
				slog.Int("payoutBet", payout.Bet),
				slog.Int("payoutAmount", payout.Payout),
			)
			amount := payout.Payout * (bet / payout.Bet)
			return amount
		}
	}

	return 0
}
