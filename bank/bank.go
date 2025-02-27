package bank

import (
	"fmt"

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
	log.Trace("--> bank.GetBank")
	defer log.Trace("<-- bank.GetBank")

	bank := readBank(guildID)
	if bank == nil {
		bank = newBank(guildID)
	}

	return bank
}

// GetAcconts returns a list of all accounts for the given bank
func (b *Bank) GetAccounts(filter interface{}, sortBy interface{}, limit int64) []*Account {
	log.Trace("--> bank.Bank.GetAccounts")
	defer log.Trace("<-- bank.Bank.GetAccounts")

	return readAccounts(b, filter, sortBy, limit)
}

// GetAccount gets the bank account for the given member. If the account doesn't
// exist, then nil is returned.
func (b *Bank) GetAccount(memberID string) *Account {
	log.Trace("--> bank.Bank.getAccount")
	defer log.Trace("<-- bank.Bank.getAccount")

	account := readAccount(b, memberID)
	if account == nil {
		account = newAccount(b, memberID)
	}

	return account
}

// newBank creates a new bank for the given guild.
func newBank(guildID string) *Bank {
	log.Trace("--> bank.newBank")
	defer log.Trace("<-- bank.newBank")

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
	log.Trace("--> bank.Bank.SetDefaultBalance")
	defer log.Trace("<-- bank.Bank.SetDefaultBalance")

	if balance != b.DefaultBalance {
		b.DefaultBalance = balance
		writeBank(b)
		log.WithFields(log.Fields{"guild": b.GuildID, "balance": b.DefaultBalance}).Info("set default balance")
	}
}

// SetName sets the name of the bank.
func (b *Bank) SetName(name string) {
	log.Trace("--> bank.Bank.SetName")
	defer log.Trace("<-- bank.Bank.SetName")

	if name != b.Name {
		b.Name = name
		writeBank(b)
		log.WithFields(log.Fields{"guild": b.GuildID, "name": b.Name}).Info("set bank name")
	}
}

// SetCurrency sets the currency used by the bank.
func (b *Bank) SetCurrency(currency string) {
	log.Trace("--> bank.Bank.SetCurrency")
	defer log.Trace("<-- bank.Bank.SetCurrency")

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
