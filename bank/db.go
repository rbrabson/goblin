package bank

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BANK_COLLECTION    = "banks"
	ACCOUNT_COLLECTION = "bank_accounts"
)

// Resets the monthly balances for all accounts in all banks.
func ResetMonthlyBalances() {
	log.Trace("--> bank.ResetMonthlyBalances")
	defer log.Trace("<-- bank.ResetMonthlyBalances")

	filter := bson.M{}
	update := bson.M{"monthly_balance": 0}
	err := db.UpdateMany(ACCOUNT_COLLECTION, filter, update)
	if err != nil {
		log.WithError(err).Error("unable to reset monthly balances for all accounts")
	}
}

// readBank gets the bank from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database.
func readBank(guildID string) *Bank {
	log.Trace("--> bank.readBank")
	defer log.Trace("<-- bank.readBank")

	filter := bson.M{"guild_id": guildID}
	var bank Bank
	err := db.FindOne(BANK_COLLECTION, filter, &bank)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Debug("bank not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": bank.GuildID}).Debug("read bank from the database")
	return &bank
}

// writeBank creates or updates the bank data in the database being used by the Discord bot.
func writeBank(bank *Bank) error {
	log.Trace("--> bank.writeBank")
	defer log.Trace("<-- bank.writeBank")

	filter := bson.M{"guild_id": bank.GuildID}
	err := db.UpdateOrInsert(BANK_COLLECTION, filter, bank)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID}).Error("unable to save bank to the database")
		return err
	}
	log.WithFields(log.Fields{"guild": bank.GuildID}).Debug("save bank to the database")

	return nil
}

// Get all the matching accounts for the given bank.
func readAccounts(bank *Bank, filter interface{}, sortBy interface{}, limit int64) []*Account {
	log.Trace("--> bank.readAccounts")
	defer log.Trace("<-- bank.readAccounts")

	var accounts []*Account
	err := db.FindMany(ACCOUNT_COLLECTION, filter, &accounts, sortBy, limit)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID}).Error("unable to read accounts from the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": bank.GuildID, "count": len(accounts)}).Debug("read accounts from the database")

	return accounts
}

// readAccount reads the account from the database and returns the value, if it exists, or returns nil if the
// account does not exist in the database
func readAccount(bank *Bank, memberID string) *Account {
	log.Trace("--> bank.readAccount")
	defer log.Trace("<-- bank.readAccount")

	filter := bson.M{"guild_id": bank.GuildID, "member_id": memberID}
	var account Account
	err := db.FindOne(ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		log.WithFields(log.Fields{"guild": bank.GuildID, "member": memberID}).Debug("account not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Debug("read account from the database")

	return &account
}

// writeAccount creates or updates the member data in the database being used by the Discord bot.
func writeAccount(account *Account) error {
	log.Trace("--> bank.writeAccount")
	defer log.Trace("<-- bank.writeAccount")

	var filter bson.D
	if account.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: account.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: account.GuildID}, {Key: "member_id", Value: account.MemberID}}
	}
	err := db.UpdateOrInsert(ACCOUNT_COLLECTION, filter, account)
	if err != nil {
		log.WithFields(log.Fields{"account": account, "error": err}).Error("unable to save bank account to the database")
		return err
	}
	log.WithFields(log.Fields{"account": account}).Debug("save bank account to the database")

	return nil
}
