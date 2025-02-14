package bank

import (
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ACCOUNT_COLLECTION = "bank_accounts"
	STARTING_BALANCE   = 0
)

// An Account represents the "bank" account for a given user. This keeps track of the
// in-game currency for the given member of a guild (server).
type Account struct {
	MemberID           string     `json:"_id" bson:"_id"`
	Balance            uint       `json:"balance" bson:"balance"`
	CreatedAt          time.Time  `json:"created_at" bson:"created_at"`
	LifetimeDeposits   uint       `json:"lifetime_deposits" bson:"lifetime_deposits"`
	LifetimeWithdrawls uint       `json:"lifetime_withdrawls" bson:"lifetime_withdrawls"`
	guildID            string     `json:"-" bson:"-"`
	mutex              sync.Mutex `json:"-" bson:"-"`
}

// NewAccount creates a new bank account for a member in the guild (server).
func NewAccount(bank *Bank, memberID string) *Account {
	log.Trace("--> bank.NewAccount")
	defer log.Trace("<-- bank.NewAccount")

	account := &Account{
		MemberID:  memberID,
		Balance:   STARTING_BALANCE,
		CreatedAt: time.Now(),
		guildID:   bank.GuildID,
	}
	bank.Accounts[account.MemberID] = account
	account.Write()
	log.WithFields(log.Fields{"guild": bank.GuildID, "member": memberID}).Info("created new bank account")

	return account
}

// GetAccount returns a bank account for a member in the guild (server). If one doesnt' exist,
// then it is created.
func GetAccount(bank *Bank, memberID string) *Account {
	log.Trace("--> bank.GetAccount")
	defer log.Trace("<-- bank.GetAccount")

	account := bank.Accounts[memberID]
	if account == nil {
		account = NewAccount(bank, memberID)
	}

	return account
}

// Deposit adds the amount to the balance of the account.
func (account *Account) Deposit(amt uint) error {
	log.Trace("--> bank.Account.Deposit")
	defer log.Trace("<-- bank.Account.Deposit")

	account.Balance += amt
	account.LifetimeDeposits += amt
	return account.Write()
}

// SetBalance sets the account's balance to the specified amount. This is typically used
// by an admin to correct an error in the system.
func (account *Account) SetBalance(balance uint) error {
	log.Trace("--> bank.Account.SetBalance")
	defer log.Trace("<-- bank.Account.SetBalance")

	account.mutex.Lock()
	account.Balance = balance
	defer account.mutex.Unlock()

	return account.Write()
}

// Withdraw deducts the amount from the balance of the account
func (account *Account) Withdraw(amt uint) error {
	log.Trace("--> bank.Account.Withdraw")
	defer log.Trace("<-- bank.Account.Withdraw")

	if amt > account.Balance {
		log.WithFields(log.Fields{"guild": account.guildID, "member": account.MemberID, "balance": account.Balance, "amount": amt}).Warn("insufficient funds for withdrawl")
		return ErrInsufficentFunds
	}
	account.mutex.Lock()
	account.Balance -= amt
	account.LifetimeWithdrawls -= amt
	account.mutex.Unlock()

	return account.Write()
}

// Write creates or updates the member data in the database being used by the Discord bot.
func (account *Account) Write() error {
	log.Trace("--> bank.Account.Write")
	defer log.Trace("<-- bank.Account.Write")

	db.Write(account.guildID, ACCOUNT_COLLECTION, account.MemberID, account)
	log.WithFields(log.Fields{"guild": account.guildID, "id": account.MemberID}).Info("save bank account to the database")
	return nil
}

// *** Bank Account Sorting Functions *** //

var (
	// Sort the member accounts based on the balance
	balance = func(a1, a2 *Account) bool {
		return a1.Balance < a2.Balance
	}
)

// By is the type of a "less" function that defines the ordering of its Account arguments.
type By func(a1, a2 *Account) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(accounts []*Account) {
	as := &accountSorter{
		accounts: accounts,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(as)
}

// accountSorter joins a By function and a slice of MemberAccounts to be sorted.
type accountSorter struct {
	accounts []*Account
	by       func(a1, a2 *Account) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *accountSorter) Len() int {
	return len(s.accounts)
}

// Swap is part of sort.Interface.
func (s *accountSorter) Swap(i, j int) {
	s.accounts[i], s.accounts[j] = s.accounts[j], s.accounts[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *accountSorter) Less(i, j int) bool {
	return s.by(s.accounts[i], s.accounts[j])
}
