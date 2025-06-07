package payday

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PAYDAY_COLLECTION         = "paydays"
	PAYDAY_ACCOUNT_COLLECTION = "payday_accounts"
)

// readPayday loads payday information for the guild from the database.
func readPayday(guildID string) *Payday {
	filter := bson.M{
		"guild_id": guildID,
	}
	var payday *Payday
	err := db.FindOne(PAYDAY_COLLECTION, filter, &payday)
	if err != nil {
		slog.Debug("payday not found in the database",
			slog.String("guildID", guildID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read payday from the database",
		slog.String("guildID", payday.GuildID),
	)

	return payday
}

// writePayday saves the payday information for the guild into the database.
func writePayday(payday *Payday) error {
	filter := bson.M{"guild_id": payday.GuildID}
	if err := db.UpdateOrInsert(PAYDAY_COLLECTION, filter, payday); err != nil {
		slog.Error("error writing payday",
			slog.String("guildID", payday.GuildID),
			slog.Any("error", err),
		)
	}

	err := db.UpdateOrInsert(PAYDAY_COLLECTION, filter, payday)
	if err != nil {
		slog.Error("unable to save payday to the database",
			slog.String("guildID", payday.GuildID),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("save payday to the database",
		slog.String("guildID", payday.GuildID),
	)
	return nil
}

// readAccount loads payday information for a given account in the guild from the database.
func readAccount(payday *Payday, accountID string) *Account {
	filter := bson.M{"guild_id": payday.GuildID, "member_id": accountID}
	var account *Account
	err := db.FindOne(PAYDAY_ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		slog.Debug("payday account not found in the database",
			slog.String("guildID", payday.GuildID),
			slog.String("memberID", accountID),
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read payday account from the database",
		slog.String("guildID", payday.GuildID),
		slog.String("memberID", accountID),
	)
	account.GuildID = payday.GuildID

	return account
}

// writeAccount saves the payday information for a given account in the guild into the database.
func writeAccount(account *Account) error {
	var filter bson.M
	if account.ID != primitive.NilObjectID {
		filter = bson.M{"_id": account.ID}
	} else {
		filter = bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}
	}
	err := db.UpdateOrInsert(PAYDAY_ACCOUNT_COLLECTION, filter, account)
	if err != nil {
		slog.Debug("unable to write payday account to the database",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("filter", filter),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("wrote payday account to the database",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
	)
	return nil
}
