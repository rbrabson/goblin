package bank

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BankCollection    = "banks"
	AccountCollection = "bank_accounts"
)

// ResetMonthlyBalances resets the monthly balances for all accounts in all banks.
func ResetMonthlyBalances() {
	filter := bson.M{}
	update := bson.M{"monthly_balance": 0}
	err := db.UpdateMany(AccountCollection, filter, update)
	if err != nil {
		slog.Error("unable to reset monthly balances for all accounts", "error", err)
	}
}

// readBank gets the bank from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database.
func readBank(guildID string) *Bank {
	filter := bson.M{"guild_id": guildID}
	var bank Bank
	err := db.FindOne(BankCollection, filter, &bank)
	if err != nil {
		slog.Debug("bank not found in the database", "guildID", guildID, "error", err)
		return nil
	}
	slog.Debug("read bank from the database", "guildID", guildID)
	return &bank
}

// writeBank creates or updates the bank data in the database being used by the Discord bot.
func writeBank(bank *Bank) error {
	filter := bson.M{"guild_id": bank.GuildID}
	err := db.UpdateOrInsert(BankCollection, filter, bank)
	if err != nil {
		slog.Error("unable to save bank to the database", "guildID", bank.GuildID, "error", err)
		return err
	}
	slog.Debug("save bank to the database", "guildID", bank.GuildID)

	return nil
}

// Get all the matching accounts for the given bank.
func readAccounts(guildID string, filter interface{}, sortBy interface{}, limit int64) []*Account {
	var accounts []*Account
	err := db.FindMany(AccountCollection, filter, &accounts, sortBy, limit)
	if err != nil {
		slog.Error("unable to read accounts from the database", "guildID", guildID, "error", err)
		return nil
	}
	slog.Debug("read accounts from the database", "guildID", guildID, "count", len(accounts))

	return accounts
}

// readAccount reads the account from the database and returns the value, if it exists, or returns nil if the
// account does not exist in the database
func readAccount(guildID string, memberID string) *Account {
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	var account Account
	err := db.FindOne(AccountCollection, filter, &account)
	if err != nil {
		slog.Debug("account not found in the database", "guildID", guildID, "memberID", memberID, "error", err)
		return nil
	}
	slog.Debug("read account from the database", "guildID", guildID, "memberID", memberID)

	return &account
}

// writeAccount creates or updates the member data in the database being used by the Discord bot.
func writeAccount(account *Account) error {
	var filter bson.D
	if account.ID != primitive.NilObjectID {
		filter = bson.D{{Key: "_id", Value: account.ID}}
	} else {
		filter = bson.D{{Key: "guild_id", Value: account.GuildID}, {Key: "member_id", Value: account.MemberID}}
	}
	err := db.UpdateOrInsert(AccountCollection, filter, account)
	if err != nil {
		slog.Error("unable to save bank account to the database", "guildID", account.GuildID, "memberID", account.MemberID, "error", err)
		return err
	}
	slog.Debug("save bank account to the database", "guildID", account.GuildID, "memberID", account.MemberID)

	return nil
}
