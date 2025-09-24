package slots

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payout struct {
	Win        []Slot `json:"win"`
	OneCoin    int    `json:"1_coin"`
	TwoCoins   int    `json:"2_coins"`
	ThreeCoins int    `json:"3_coins"`
}

type PayoutTable struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GuildID string             `json:"guild_id"`
	Name    string             `json:"name"`
	Payouts []Payout           `json:"payouts"`
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
	// write lookup table to the DB
	return payoutTable
}

func readPayoutTableFromFile(guildID string) *PayoutTable {
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	payoutTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configFileName := filepath.Join(configDir, "slots", "lookuptable", payoutTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default payout table",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return nil
	}

	payoutTable := &PayoutTable{
		GuildID: guildID,
		Name:    LOOKUP_TABLE_NAME,
	}
	err = json.Unmarshal(bytes, &payoutTable.Payouts)
	if err != nil {
		slog.Error("failed to unmarshal payout table",
			slog.String("guildID", guildID),
			slog.String("file", configFileName),
			slog.String("data", string(bytes)),
			slog.Any("error", err),
		)
		return nil
	}

	slog.Info("create new payout table",
		slog.String("guildID", payoutTable.GuildID),
		slog.String("theme", payoutTable.Name),
	)

	return payoutTable
}
