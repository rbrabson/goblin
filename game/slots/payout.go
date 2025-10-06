package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	rslots "github.com/rbrabson/slots"
)

const (
	PAYOUT_FILE_NAME = "payout"
)

// GetPayoutTable retrieves the payout table for a specific guild.
func GetPayoutTable() rslots.PayoutTable {
	pt := newPayoutTable()
	slices.SortFunc(pt, func(a, b rslots.PayoutAmount) int {
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
func newPayoutTable() rslots.PayoutTable {
	payoutTable := readPayoutTableFromFile()
	return payoutTable
}

// readPayoutTableFromFile reads the payout table from a JSON file.
func readPayoutTableFromFile() rslots.PayoutTable {
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

	var payouts rslots.PayoutTable
	err = json.Unmarshal(bytes, &payouts)
	if err != nil {
		slog.Error("failed to unmarshal payout table",
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}

	slog.Debug("create new payout table")

	return payouts
}
