package payday

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Account is a user on the server that can a payday every 23 hours
type Account struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID    string             `json:"guild_id" bson:"guild_id"`
	MemberID   string             `json:"member_id" bson:"member_id"`
	NextPayday time.Time          `json:"next_payday" bson:"next_payday"`
}

// newAccount creates new payday information for a server/guild
func newAccount(payday *Payday, memberID string) *Account {
	account := &Account{
		MemberID: memberID,
		GuildID:  payday.GuildID,
	}
	writeAccount(account)

	return account
}

// getNextPayday returns the next payday for the user.
func (a *Account) getNextPayday() time.Time {
	return a.NextPayday
}

// setNextPayday sets the next payday for the user.
func (a *Account) setNextPayday(nextPayday time.Time) {
	a.NextPayday = nextPayday
	err := writeAccount(a)
	if err != nil {
		log.WithFields(log.Fields{"account": a, "error": err}).Error("unable to save account to the database")
		return
	}
	log.WithFields(log.Fields{"account": a}).Debug("set next payday")
}

// String returns a string representation of the Account.
func (account *Account) String() string {
	return fmt.Sprintf("PaydayAccount{ID=%s, GuildID=%s, MemberID=%s, NextPayday=%s}",
		account.ID.Hex(),
		account.GuildID,
		account.MemberID,
		account.NextPayday,
	)
}
