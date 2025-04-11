package bank

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BANK_COLLECTION    = "banks"
	ACCOUNT_COLLECTION = "bank_accounts"
)

// Resets the monthly balances for all accounts in all banks.
func ResetMonthlyBalances() {
	filter := bson.M{}
	update := bson.M{"monthly_balance": 0}
	err := db.UpdateMany(ACCOUNT_COLLECTION, filter, update)
	if err != nil {
		sslog.Error("unable to reset monthly balances for all accounts",
			slog.Any("error", err),
		)
	}
}

// readBank gets the bank from the database and returns the value, if it exists, or returns nil if the
// bank does not exist in the database.
func readBank(guildID string) *Bank {
	filter := bson.M{"guild_id": guildID}
	var bank Bank
	err := db.FindOne(BANK_COLLECTION, filter, &bank)
	if err != nil {
		sslog.Debug("bank not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	sslog.Debug("read bank from the database",
		slog.String("guildID", guildID),
	)
	return &bank
}

// writeBank creates or updates the bank data in the database being used by the Discord bot.
func writeBank(bank *Bank) error {
	filter := bson.M{"guild_id": bank.GuildID}
	err := db.UpdateOrInsert(BANK_COLLECTION, filter, bank)
	if err != nil {
		sslog.Error("unable to save bank to the database",
			slog.String("guildID", bank.GuildID),
			slog.Any("error", err),
		)
		return err
	}
	sslog.Debug("save bank to the database",
		slog.String("guildID", bank.GuildID),
	)

	return nil
}

// Get all the matching accounts for the given bank.
func readAccounts(guildID string, filter interface{}, sortBy interface{}, limit int64) []*Account {
	var accounts []*Account
	err := db.FindMany(ACCOUNT_COLLECTION, filter, &accounts, sortBy, limit)
	if err != nil {
		sslog.Error("unable to read accounts from the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	sslog.Debug("read accounts from the database",
		slog.String("guildID", guildID),
		slog.Int("count", len(accounts)),
	)

	return accounts
}

// readAccount reads the account from the database and returns the value, if it exists, or returns nil if the
// account does not exist in the database
func readAccount(guildID string, memberID string) *Account {
	filter := bson.M{"guild_id": guildID, "member_id": memberID}
	var account Account
	err := db.FindOne(ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		sslog.Debug("account not found in the database",
			slog.String("guildID", guildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		return nil
	}
	sslog.Debug("read account from the database",
		slog.String("guildID", guildID),
		slog.String("memberID", memberID),
	)

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
	err := db.UpdateOrInsert(ACCOUNT_COLLECTION, filter, account)
	if err != nil {
		sslog.Error("unable to save bank account to the database",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
		return err
	}
	sslog.Debug("save bank account to the database",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
	)

	return nil
}
