package bank

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/rbrabson/goblin/discord"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Default values when a new bank is created for a previously unknown guild.
const (
	DefaultBankName = "Treasury"
	DefaultCurrency = "Coins"
	DefaultBalance  = 20000
)

var (
	bankLock = sync.Mutex{}
	banks    = make(map[string]*Bank)
)

// A Bank is the repository for all bank accounts for a given guild (server).
type Bank struct {
	ID             bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID        string        `json:"guild_id" bson:"guild_id"`
	Name           string        `json:"bank_name" bson:"bank_name"`
	Currency       string        `json:"currency" bson:"currency"`
	DefaultBalance int           `json:"default_balance" bson:"default_balance"`
	lock           *sync.Mutex   `bson:"-"`
}

// GetBank returns the bank for the specified guild. If the bank does not exist, then one is created.
func GetBank(guildID string) *Bank {
	bankLock.Lock()
	defer bankLock.Unlock()

	bank := banks[guildID]

	if bank == nil {
		bank = readBank(guildID)
		if bank == nil {
			bank = readBankFromFile(guildID)
		}
		bank.lock = &sync.Mutex{}
	}

	return bank
}

// readBankFromFile creates a new bank for the given guild.
func readBankFromFile(guildID string) *Bank {
	configTheme := os.Getenv("DISCORD_BANK_THEME")
	configFileName := filepath.Join(discord.ConfigDir, "bank", "config", configTheme+".json")
	bytes, err := os.ReadFile(configFileName)
	if err != nil {
		slog.Error("failed to read default bank config", "error", err)
		return getDefaultBank(guildID)
	}

	bank := &Bank{}
	if err := json.Unmarshal(bytes, bank); err != nil {
		slog.Error("failed to unmarshal default bank config", "file", configFileName, "error", err)
		bank = getDefaultBank(guildID)
	}
	bank.GuildID = guildID
	if err := writeBank(bank); err != nil {
		slog.Error("error writing bank", "guildID", guildID, "error", err)
	}

	slog.Info("create new bank", "guildID", bank.GuildID)

	return bank
}

// getDefaultBank returns a default bank for the given guild.
// This is used when no default bank config file is found or when
// the default bank config file is invalid.
func getDefaultBank(guildID string) *Bank {
	bank := &Bank{
		GuildID:        guildID,
		Name:           DefaultBankName,
		Currency:       DefaultCurrency,
		DefaultBalance: DefaultBalance,
	}
	slog.Info("create new bank", "guildID", bank.GuildID)

	return bank
}

// SetDefaultBalance sets the default balance for the bank.
func (b *Bank) SetDefaultBalance(balance int) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if balance != b.DefaultBalance {
		b.DefaultBalance = balance
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank", "guildID", b.GuildID, "error", err)
		}
		slog.Info("set default balance", "guildID", b.GuildID, "balance", b.DefaultBalance)
	}
}

// SetName sets the name of the bank.
func (b *Bank) SetName(name string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if name != b.Name {
		b.Name = name
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank", "guildID", b.GuildID, "error", err)
		}
		slog.Info("set bank name", "name", b.Name, "guildID", b.GuildID)
	}
}

// SetCurrency sets the currency used by the bank.
func (b *Bank) SetCurrency(currency string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if currency != b.Currency {
		b.Currency = currency
		if err := writeBank(b); err != nil {
			slog.Error("error writing bank", "guildID", b.GuildID, "error", err)
		}
		slog.Info("set currency", "guildID", b.GuildID, "currency", b.Currency)
	}
}

// lockBank and unlockBank are used to lock and unlock the bank.
func (b *Bank) lockBank() {
	b.lock.Lock()
}

// unlockBank is used to unlock the bank.
func (b *Bank) unlockBank() {
	b.lock.Unlock()
}

// String returns a string representation of the Bank.
func (b *Bank) String() string {
	b.lock.Lock()
	defer b.lock.Unlock()

	return fmt.Sprintf("Bank{ID: %s, GuildID: %s, Name: %s, Currency: %s, DefaultBalance: %d}",
		b.ID.Hex(),
		b.GuildID,
		b.Name,
		b.Currency,
		b.DefaultBalance,
	)
}
