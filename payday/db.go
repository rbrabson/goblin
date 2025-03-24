package payday

import (
	log "github.com/sirupsen/logrus"
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
		log.WithField("guild", guildID).Debug("payday not found in the database")
		return nil
	}
	log.WithField("guild", payday.GuildID).Debug("read payday from the database")

	return payday
}

// writePayday saves the payday information for the guild into the database.
func writePayday(payday *Payday) error {
	filter := bson.M{"guild_id": payday.GuildID}
	db.UpdateOrInsert(PAYDAY_COLLECTION, filter, payday)

	err := db.UpdateOrInsert(PAYDAY_COLLECTION, filter, payday)
	if err != nil {
		log.WithField("guild", payday.GuildID).Error("unable to save payday to the database")
		return err
	}
	log.WithField("guild", payday.GuildID).Debug("save payday to the database")
	return nil
}

// readAccount loads payday information for a given account in the guild from the database.
func readAccount(payday *Payday, accountID string) *Account {
	filter := bson.M{"guild_id": payday.GuildID, "member_id": accountID}
	var account *Account
	err := db.FindOne(PAYDAY_ACCOUNT_COLLECTION, filter, &account)
	if err != nil {
		log.WithFields(log.Fields{"guild": payday.GuildID, "member": accountID}).Debug("payday account not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Debug("read payday account from the database")
	account.GuildID = payday.GuildID

	return account
}

// writeAccount saves the payday information for a given account in the guild into the database.
func writeAccount(account *Account) error {
	var filter bson.M
	if account.ID != primitive.NilObjectID {
		log.WithFields(log.Fields{"account": account}).Debug("writing account with ID")
		filter = bson.M{"_id": account.ID}
	} else {
		log.WithFields(log.Fields{"account": account}).Debug("writing account without ID")
		filter = bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}
	}
	err := db.UpdateOrInsert(PAYDAY_ACCOUNT_COLLECTION, filter, account)
	if err != nil {
		log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID}).Debug("unable to write payday account to the database")
		return err
	}

	return nil
}
