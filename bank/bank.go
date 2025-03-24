package bank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Default values when a new bank is created for a previously unknown guild.
const (
	DEFAULT_BANK_NAME = "Treasury"
	DEFAULT_CURRENCY  = "Coins"
	DEFAULT_BALANCE   = 20000
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
		log.WithField("file", configFileName).Error("failed to read default bank config")
		return getDefaultBank(guildID)
	}

	bank := &Bank{}
	err = json.Unmarshal(bytes, bank)
	if err != nil {
		log.WithField("file", configFileName).Error("failed to unmarshal default bank config")
		return getDefaultBank(guildID)
	}
	bank.GuildID = guildID

	writeBank(bank)
	log.WithField("guild", bank.GuildID).Info("create new bank")

	return bank
}

// getDefaultBank returns a default bank for the given guild.
// This is used when no default bank config file is found, or when
// the default bank config file is invalid.
func getDefaultBank(guildID string) *Bank {
	bank := &Bank{
		GuildID:        guildID,
		Name:           DEFAULT_BANK_NAME,
		Currency:       DEFAULT_CURRENCY,
		DefaultBalance: DEFAULT_BALANCE,
	}
	writeBank(bank)
	log.WithField("guild", bank.GuildID).Info("create new bank")

	return bank
}

// SetDefaultBalance sets the default balance for the bank.
func (b *Bank) SetDefaultBalance(balance int) {
	if balance != b.DefaultBalance {
		b.DefaultBalance = balance
		writeBank(b)
		log.WithFields(log.Fields{"guild": b.GuildID, "balance": b.DefaultBalance}).Info("set default balance")
	}
}

// SetName sets the name of the bank.
func (b *Bank) SetName(name string) {
	if name != b.Name {
		b.Name = name
		writeBank(b)
		log.WithFields(log.Fields{"guild": b.GuildID, "name": b.Name}).Info("set bank name")
	}
}

// SetCurrency sets the currency used by the bank.
func (b *Bank) SetCurrency(currency string) {
	if currency != b.Currency {
		b.Currency = currency
		writeBank(b)
		log.WithFields(log.Fields{"guild": b.GuildID, "currency": b.Currency}).Info("set currency")
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
