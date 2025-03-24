package bank

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	log "github.com/sirupsen/logrus"
)

// An Account represents the "bank" account for a given user. This keeps track of the
// in-game currency for the given member of a guild (server).
type Account struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	GuildID         string             `json:"guild_id" bson:"guild_id"`
	MemberID        string             `json:"member_id" bson:"member_id"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	CurrentBalance  int                `json:"current_balance" bson:"current_balance"`
	MonthlyBalance  int                `json:"monthly_balance" bson:"monthly_balance"`
	LifetimeBalance int                `json:"lifetime_balance" bson:"lifetime_balance"`
}

// GetAccount gets the bank account for the given member. If the account doesn't
// exist, then nil is returned.
func GetAccount(guildID string, memberID string) *Account {
	account := readAccount(guildID, memberID)
	if account == nil {
		account = newAccount(guildID, memberID)
	}

	return account
}

// GetAcconts returns a list of all accounts for the given bank
func GetAccounts(guildID string, filter interface{}, sortBy interface{}, limit int64) []*Account {
	return readAccounts(guildID, filter, sortBy, limit)
}

// Deposit adds the amount to the balance of the account.
func (account *Account) Deposit(amt int) error {
	account.CurrentBalance += amt
	account.MonthlyBalance += amt
	account.LifetimeBalance += amt

	err := writeAccount(account)
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance, "amount": amt}).Info("deposit into account")
	return err
}

// DepositToCurrentOnly adds the amount to the balance of the account.
func (account *Account) DepositToCurrentOnly(amt int) error {
	account.CurrentBalance += amt

	err := writeAccount(account)
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance, "amount": amt}).Info("deposit into account")
	return err
}

// Withdraw deducts the amount from the balance of the account
func (account *Account) Withdraw(amt int) error {
	if amt > account.CurrentBalance {
		log.WithFields(log.Fields{"guild": account.GuildID, "member": account.GuildID, "balance": account.CurrentBalance, "amount": amt}).Warn("insufficient funds for withdrawl")
		return ErrInsufficentFunds
	}
	account.CurrentBalance -= amt
	account.MonthlyBalance -= amt
	account.LifetimeBalance -= amt

	err := writeAccount(account)
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance, "amount": amt}).Info("withdraw from account")
	return err
}

// WithdrawFromCurrentOnly deducts the amount from the current balance of the account. This
// is useful for transactions that should not affect the monthly or lifetime balance.
func (account *Account) WithdrawFromCurrentOnly(amount int) error {
	if amount > account.CurrentBalance {
		log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance, "amount": amount}).Warn("insufficient funds for withdrawal")
		return ErrInsufficentFunds
	}
	account.CurrentBalance -= amount

	err := writeAccount(account)
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance, "amount": amount}).Info("withdraw from account")
	return err
}

// SetBalance sets the account's balance to the specified amount. This is typically used
// by an admin to correct an error in the system.
func (account *Account) SetBalance(balance int) error {
	account.CurrentBalance = balance

	if balance > account.LifetimeBalance {
		account.LifetimeBalance = balance
	}

	if balance > account.MonthlyBalance {
		account.MonthlyBalance = balance
	}

	err := writeAccount(account)
	log.WithFields(log.Fields{"guild": account.GuildID, "member": account.MemberID, "balance": account.CurrentBalance}).Info("set account balance")
	return err
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
	writeAccount(account)
	log.WithFields(log.Fields{"guild": bank.GuildID, "member": memberID}).Info("created new bank account")

	return account
}

// String returns a string representation of the account.
func (account *Account) String() string {
	return fmt.Sprintf("Account{ID: %s, GuildID: %s, MemberID: %s, Balance: %d, MonthlyBalance: %d, LifetimeBalance: %d}",
		account.ID.Hex(),
		account.GuildID,
		account.MemberID,
		account.CurrentBalance,
		account.MonthlyBalance,
		account.LifetimeBalance,
	)
}
