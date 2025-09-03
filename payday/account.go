package payday

import (
	"fmt"
	"log/slog"
	"time"

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
	if err := writeAccount(account); err != nil {
		slog.Error("error writing account",
			slog.Any("error", err),
		)
	}

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
		slog.Error("unable to save account to the database",
			slog.String("guildID", a.GuildID),
			slog.String("memberID", a.MemberID),
			slog.Any("error", err),
		)
		return
	}
	slog.Debug("set next payday",
		slog.String("guildID", a.GuildID),
		slog.String("memberID", a.MemberID),
		slog.Time("nextPayday", a.NextPayday),
	)
}

// String returns a string representation of the Account.
func (a *Account) String() string {
	return fmt.Sprintf("PaydayAccount{ID=%s, GuildID=%s, MemberID=%s, NextPayday=%s}",
		a.ID.Hex(),
		a.GuildID,
		a.MemberID,
		a.NextPayday,
	)
}
