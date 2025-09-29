package payday

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Account is a user on the server that can a payday every 23 hours
type Account struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID         string             `json:"guild_id" bson:"guild_id"`
	MemberID        string             `json:"member_id" bson:"member_id"`
	NextPayday      time.Time          `json:"next_payday" bson:"next_payday"`
	CurrentStreak   int                `json:"current_streak" bson:"current_streak"`
	MaxStreak       int                `json:"max_streak" bson:"max_streak"`
	TotalPaydays    int                `json:"total_paydays" bson:"total_paydays"`
	TotalAmountPaid int                `json:"total_amount_paid" bson:"total_amount_paid"`
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
func (a *Account) setNextPayday(minWait time.Duration) {
	a.NextPayday = time.Now().Add(minWait)

	// Save the account to the database.
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
		slog.Int("paydayStreak", a.CurrentStreak),
		slog.Int("maxStreak", a.MaxStreak),
		slog.Time("nextPayday", a.NextPayday),
	)
}

// getPayAmmount returns the amount of credits the user will receive on their next payday.
func (a *Account) getPayAmmount() int {
	payday := GetPayday(a.GuildID)
	a.updateStreak(payday.PaydayFrequency)
	basePay := payday.Amount
	bonus := payday.StreakPerDayBonus
	streakReset := payday.MaxStreak
	streak := a.CurrentStreak

	var pay int
	if streak != 0 && streakReset != 0 && bonus != 0.0 {
		multiplier := (streak - 1) % streakReset
		pay = basePay + (bonus * multiplier)
		slog.Error("**********",
			slog.String("guildID", a.GuildID),
			slog.String("memberID", a.MemberID),
			slog.Int("basePay", basePay),
			slog.Int("bonus", bonus),
			slog.Int("streak", streak),
			slog.Int("streakReset", streakReset),
			slog.Int("multiplier", multiplier),
			slog.Int("pay", pay),
		)
	} else {
		pay = basePay
	}

	return pay
}

// updateStreak updates the user's current streak based on their last payday.
func (a *Account) updateStreak(minWait time.Duration) {
	if a.NextPayday.After(time.Now()) {
		return
	}

	previousPayday := a.NextPayday.Add(-minWait)
	if time.Since(previousPayday) > (2 * 24 * time.Hour) {
		a.CurrentStreak = 1
	} else {
		a.CurrentStreak++
	}
	a.MaxStreak = max(a.MaxStreak, a.CurrentStreak)
	a.TotalPaydays++
	a.TotalPaydays = max(a.TotalPaydays, a.MaxStreak) // To handle adding TotalPaydays to existing accounts
}

// String returns a string representation of the Account.
func (a *Account) String() string {
	return fmt.Sprintf("PaydayAccount{ID=%s, GuildID=%s, MemberID=%s, CurrentStreak=%d, MaxStreak=%d, NextPayday=%s}",
		a.ID.Hex(),
		a.GuildID,
		a.MemberID,
		a.CurrentStreak,
		a.MaxStreak,
		a.NextPayday,
	)
}
