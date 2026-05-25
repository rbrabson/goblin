package bank

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// An Account represents the "bank" account for a given user. This keeps track of the
// in-game currency for the given member of a guild (server).
type Account struct {
	ID              bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID         string        `json:"guild_id" bson:"guild_id"`
	MemberID        string        `json:"member_id" bson:"member_id"`
	CreatedAt       time.Time     `json:"created_at" bson:"created_at"`
	CurrentBalance  int           `json:"current_balance" bson:"current_balance"`
	MonthlyBalance  int           `json:"monthly_balance" bson:"monthly_balance"`
	LifetimeBalance int           `json:"lifetime_balance" bson:"lifetime_balance"`
}

// GetAccount gets the bank account for the given member.
func GetAccount(guildID string, memberID string) *Account {
	account := readAccount(guildID, memberID)
	if account == nil {
		account = newAccount(guildID, memberID)
		if err := writeAccount(account); err != nil {
			slog.Error("error writing account",
				slog.String("guildID", guildID),
				slog.String("memberID", memberID),
				slog.Any("error", err),
			)
		}
	}

	return account
}

// newAccount creates a new bank account for a member in the guild (server).
func newAccount(guildID string, memberID string) *Account {
	bank := GetBank(guildID)
	account := &Account{
		GuildID:         guildID,
		MemberID:        memberID,
		CurrentBalance:  bank.DefaultBalance,
		LifetimeBalance: bank.DefaultBalance,
		CreatedAt:       time.Now(),
	}
	slog.Info("created new bank account",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
	)

	return account
}

// GetAccounts returns a list of all accounts for the given bank
func GetAccounts(filter interface{}, sortBy interface{}, limit int64) []*Account {
	return readAccounts(filter, sortBy, limit)
}

// Deposit adds the amount to the balance of the account.
func (account *Account) Deposit(amt int) error {
	bank := GetBank(account.GuildID)
	bank.lockBank()
	defer bank.unlockBank()

	account.Refresh()

	account.CurrentBalance += amt
	account.MonthlyBalance += amt
	account.LifetimeBalance += amt

	err := writeAccount(account)
	slog.Info("deposit into account",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
		slog.Int("balance", account.CurrentBalance),
		slog.Int("amount", amt),
	)
	return err
}

// DepositToCurrentOnly adds the amount to the balance of the account.
func (account *Account) DepositToCurrentOnly(amt int) error {
	bank := GetBank(account.GuildID)
	bank.lockBank()
	defer bank.unlockBank()

	account.Refresh()

	account.CurrentBalance += amt

	err := writeAccount(account)
	slog.Info("deposit into current account",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
		slog.Int("balance", account.CurrentBalance),
		slog.Int("amount", amt),
	)
	return err
}

// Withdraw deducts the amount from the balance of the account
func (account *Account) Withdraw(amt int) error {
	return account.withdraw(amt, true)
}

// WithdrawFromCurrentOnly deducts the amount from the current balance of the account. This
// is useful for transactions that should not affect the monthly or lifetime balance.
func (account *Account) WithdrawFromCurrentOnly(amt int) error {
	return account.withdraw(amt, false)
}

// withdraw deducts the amount from the balance of the account. If updateTotals is true, it also updates the monthly
// and lifetime balances. If false, it only updates the current balance.
func (account *Account) withdraw(amt int, updateTotals bool) error {
	bank := GetBank(account.GuildID)
	bank.lockBank()
	defer bank.unlockBank()

	account.Refresh()

	if amt > account.CurrentBalance {
		slog.Warn("insufficient funds for withdrawal",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Int("balance", account.CurrentBalance),
			slog.Int("amount", amt),
		)
		return ErrInsufficientFunds
	}

	account.CurrentBalance -= amt
	if updateTotals {
		account.MonthlyBalance -= amt
		account.LifetimeBalance -= amt
	}

	err := writeAccount(account)
	slog.Info("withdraw from account",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
		slog.Int("balance", account.CurrentBalance),
		slog.Int("amount", amt),
	)
	return err
}

// SetBalance sets the account's balance to the specified amount. An admin typically uses this
// to correct an error in the system, increasing a user's balance. It cannot be used
// to decrease the user's balance.
func (account *Account) SetBalance(balance int) error {
	bank := GetBank(account.GuildID)
	bank.lockBank()
	defer bank.unlockBank()

	account.Refresh()
	account.CurrentBalance = balance

	if balance > account.LifetimeBalance {
		account.LifetimeBalance = balance
	}

	if balance > account.MonthlyBalance {
		account.MonthlyBalance = balance
	}

	err := writeAccount(account)
	slog.Info("set account balance",
		slog.String("guildID", account.GuildID),
		slog.String("memberID", account.MemberID),
		slog.Int("balance", account.CurrentBalance),
	)
	return err
}

// GetBalance returns the current balance of the account.
func (account *Account) GetBalance() int {
	return account.CurrentBalance
}

// Refresh updates the account's balances from the database. This is useful if the account has been modified by another
// process, and you want to ensure that you have the most up-to-date information before performing an operation on the
// account.
func (account *Account) Refresh() {
	currentAccount := readAccount(account.GuildID, account.MemberID)
	if currentAccount != nil {
		account.CurrentBalance = currentAccount.CurrentBalance
		account.MonthlyBalance = currentAccount.MonthlyBalance
		account.LifetimeBalance = currentAccount.LifetimeBalance
	} else {
		slog.Warn("account not found in database",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
		)
	}
}

// GetLifetimeRanking returns the lifetime ranking of the account in the guild (server). The ranking is based on the
// lifetime balance of the account, with the highest balance being ranked first.
func (account *Account) GetLifetimeRanking() int {
	filter := bson.M{
		"guild_id":         account.GuildID,
		"lifetime_balance": bson.M{"$gt": account.LifetimeBalance},
	}

	rank, _ := db.Count(accountCollection, filter)
	rank++
	slog.Debug("lifetime ranking",
		slog.String("guildID", account.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank),
	)
	return rank
}

// GetMonthlyRanking returns the monthly global ranking on the server for a given player. The ranking is based on the
// monthly balance of the account, with the highest balance being ranked first.
func (account *Account) GetMonthlyRanking() int {
	filter := bson.M{
		"guild_id":        account.GuildID,
		"monthly_balance": bson.M{"$gt": account.MonthlyBalance},
	}
	rank, _ := db.Count(accountCollection, filter)
	rank++
	slog.Debug("monthly ranking",
		slog.String("guildID", account.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank))
	return rank
}

// GetCurrentRanking returns the current ranking of the account in the guild (server). The ranking is based on the
// current balance of the account, with the highest balance being ranked first.
func (account *Account) GetCurrentRanking() int {
	filter := bson.M{
		"guild_id":        account.GuildID,
		"current_balance": bson.M{"$gt": account.CurrentBalance},
	}
	rank, _ := db.Count(accountCollection, filter)
	rank++
	slog.Debug("current ranking",
		slog.String("guildID", account.GuildID),
		slog.String("account", account.MemberID),
		slog.Int("rank", rank),
	)

	return rank
}

// String returns a string representation of the account.
func (account *Account) String() string {
	return fmt.Sprintf("Account{ID: %s, GuildID: %s, MemberID: %s, CurrentBalance: %d, MonthlyBalance: %d, LifetimeBalance: %d}",
		account.ID.Hex(),
		account.GuildID,
		account.MemberID,
		account.CurrentBalance,
		account.MonthlyBalance,
		account.LifetimeBalance,
	)
}
