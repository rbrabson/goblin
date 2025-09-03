package bank

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Default values when a new bank is created for a previously unknown guild.
const (
	DefaultBankName = "Treasury"
	DefaultCurrency = "Coins"
	DefaultBalance  = 20000
)

// A Bank is the repository for all bank accounts for a given guild (server).
type Bank struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID        string             `json:"guild_id" bson:"guild_id"`
	Name           string             `json:"bank_name" bson:"bank_name"`
	Currency       string             `json:"currency" bson:"currency"`
	DefaultBalance int                `json:"default_balance" bson:"default_balance"`
}

// GetBank returns the bank for the specified build. If the bank does not exist, then one is created.
func GetBank(guildID string) *Bank {
	bank := readBank(guildID)
	if bank == nil {
		bank = readBankFromFile(guildID)
	}

	return bank
}

// readBankFromFile creates a new bank for the given guild.
func readBankFromFile(guildID string) *Bank {
	configTheme := os.Getenv("DISCORD_DEFAULT_THEME")
	configDir := os.Getenv("DISCORD_CONFIG_DIR")
	configFileName := filepath.Join(configDir, "bank", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default bank config",
			slog.Any("error", err),
		)
		return getDefaultBank(guildID)
	}

	bank := &Bank{}
	err = json.Unmarshal(bytes, bank)
	if err != nil {
		slog.Error("failed to unmarshal default bank config",
			slog.String("file", configFileName),
			slog.Any("error", err),
		)
		return getDefaultBank(guildID)
	}
	bank.GuildID = guildID

	if err := writeBank(bank); err != nil {
		slog.Error("error writing bank",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
	}
	slog.Info("create new bank",
		slog.String("guildID", bank.GuildID),
	)

	return bank
}

// getDefaultBank returns a default bank for the given guild.
// This is used when no default bank config file is found, or when
// the default bank config file is invalid.
func getDefaultBank(guildID string) *Bank {
	bank := &Bank{
		GuildID:        guildID,
		Name:           DefaultBankName,
		Currency:       DefaultCurrency,
		DefaultBalance: DefaultBalance,
	}
	if err := writeBank(bank); err != nil {
		slog.Error("error writing bank",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
	}
	slog.Info("create new bank",
		slog.String("guildID", bank.GuildID),
	)

	return bank
}

// SetDefaultBalance sets the default balance for the bank.
func (b *Bank) SetDefaultBalance(balance int) {
	if balance != b.DefaultBalance {
		b.DefaultBalance = balance
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank",
				slog.String("guildID", b.GuildID),
				slog.Any("error", err),
			)
		}
		slog.Info("set default balance",
			slog.String("guildID", b.GuildID),
			slog.Int("balance", b.DefaultBalance),
		)
	}
}

// SetName sets the name of the bank.
func (b *Bank) SetName(name string) {
	if name != b.Name {
		b.Name = name
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank",
				slog.String("guildID", b.GuildID),
				slog.Any("error", err),
			)
		}
		slog.Info("set bank name",
			slog.String("name", b.Name),
			slog.String("guildID", b.GuildID),
		)
	}
}

// SetCurrency sets the currency used by the bank.
func (b *Bank) SetCurrency(currency string) {
	if currency != b.Currency {
		b.Currency = currency
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank",
				slog.String("guildID", b.GuildID),
				slog.Any("error", err),
			)
		}
		slog.Info("set currency",
			slog.String("guildID", b.GuildID),
			slog.String("currency", b.Currency),
		)
	}
}

// String returns a string representation of the Bank.
func (b *Bank) String() string {
	return fmt.Sprintf("Bank{Bank{ID: %s, GuildID: %s, Name: %s, Currency: %s, DefaultBalance: %d}",
		b.ID.Hex(),
		b.GuildID,
		b.Name,
		b.Currency,
		b.DefaultBalance,
	)
}
