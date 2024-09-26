package bank

import (
	"github.com/rbrabson/dgame/database"
	"github.com/rbrabson/dgame/guild"
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

// Bank is the repository for all accounts for a given guild (server).
type Bank struct {
	GuildID  string              `json:"_id" bson:"_id"`
	Accounts map[string]*Account `json:"-" bson:"-"`
}

// Init initializes the banking system with the database used for reading and writing account data
func Init(database database.Client) {
	db = database
}

// New creates a new bank for the specified build
func New(guild guild.Guild) *Bank {
	bank := &Bank{
		GuildID:  guild.ID,
		Accounts: make(map[string]*Account),
	}
	banks[bank.GuildID] = bank
	return bank
}

// GetBank returns the bank for the specified build. If the bank does not exist, then one is created.
func GetBank(guild guild.Guild) *Bank {
	bank := banks[guild.ID]
	if bank == nil {
		bank = New(guild)
	}
	return bank
}
