package bank

import (
	"github.com/rbrabson/dgame/database"
	"github.com/rbrabson/dgame/guild"
	log "github.com/sirupsen/logrus"
)

const (
	BANK_COLLECTION = "banks"
)

var (
	// Database to which to read and write bank accounts
	db    database.Client
	banks = make(map[string]*Bank)
)

var (
	// List of bank accounts in a given guild. The acconts are in a slice, which allows for
	// easy sorting of the accounts based on different criteria (for instance, the account balance).
	bankAccountList = make(map[string][]*Account)
)

// A Bank is the repository for all bank accounts for a given guild (server).
type Bank struct {
	GuildID   string              `json:"_id" bson:"_id"`
	ChannelID string              `json:"channel_id" bson:"channel_id"`
	Accounts  map[string]*Account `json:"-" bson:"-"`
}

// Init initializes the banking system with the database used for reading and writing account data
func Init(database database.Client) {
	db = database
}

// NewBank creates a new bank for the specified build
func NewBank(guild *guild.Guild) *Bank {
	log.Trace("--> bank.New")
	defer log.Trace("<-- bank.New")

	bank := &Bank{
		GuildID:  guild.ID,
		Accounts: make(map[string]*Account),
	}
	banks[bank.GuildID] = bank
	log.WithField("guild", bank.GuildID).Info("create new bank")

	bank.Write()

	return bank
}

// GetBank returns the bank for the specified build. If the bank does not exist, then one is created.
func GetBank(guild *guild.Guild) *Bank {
	log.Trace("--> bank.GetBank")
	defer log.Trace("<-- bank.GetBank")

	bank := banks[guild.ID]
	if bank == nil {
		bank = NewBank(guild)
	}

	log.WithField("guild", bank.GuildID).Trace("return bank")
	return bank
}

// Write creates or updates the bank for a guild in the database being used by the Discord bot.
func (bank *Bank) Write() error {
	log.Trace("--> bank.Bank.Write")
	defer log.Trace("<-- bank.Bank.Write")

	db.Write(bank.GuildID, BANK_COLLECTION, bank.GuildID, bank)
	log.WithField("guild", bank.GuildID).Debug("save bank to the database")
	return nil
}
