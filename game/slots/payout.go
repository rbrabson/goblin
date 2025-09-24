package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PAYOUT_FILE_NAME = "payout"
)

type Payout struct {
	Win    []Slot `json:"win" bson:"win"`
	Bet100 int    `json:"100" bson:"100"`
	Bet200 int    `json:"200" bson:"200"`
	Bet300 int    `json:"300" bson:"300"`
}

type PayoutAmount struct {
	Win []string    `json:"win" bson:"win"`
	Bet map[int]int `json:"bet" bson:"bet"`
}

type PayoutTable struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id"`
	Payouts []PayoutAmount     `json:"payouts"`
}

func GetPayoutTable(guildID string) *PayoutTable {
	// TODO: try to read from the DB
	return newPayoutTable(guildID)
}

func newPayoutTable(guildID string) *PayoutTable {
	payoutTable := readPayoutTableFromFile(guildID)
	if payoutTable == nil {
		slog.Error("failed to create lookup table",
			slog.String("guildID", guildID),
		)
		return nil
	}
	// TODO: write lookup table to the DB
	return payoutTable
}

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
		payoutAmount := PayoutAmount{
			Win: make([]string, 0, len(payout.Win)),
			Bet: map[int]int{
				100: payout.Bet100,
				200: payout.Bet200,
				300: payout.Bet300,
			},
		}
		for _, slot := range payout.Win {
			payoutAmount.Win = append(payoutAmount.Win, string(slot))
		}
		payoutTable.Payouts = append(payoutTable.Payouts, payoutAmount)
	}

	slog.Info("create new payout table",
		slog.String("guildID", payoutTable.GuildID),
	)

	return payoutTable
}

func (pt *PayoutTable) GetPayoutAmount(bet int, spin []string) int {
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
			if !slices.Contains(winningSymbols, spin[i]) {
				match = false
				break
			}
		}
		if match {
			amount, ok := payout.Bet[bet]
			if !ok {
				slog.Error("no payout for bet amount",
					slog.String("guildID", pt.GuildID),
					slog.Int("bet", bet),
					slog.Any("win", spin),
				)
				return 0
			}
			return amount
		}
	}

	return 0
}
