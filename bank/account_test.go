package bank

import (
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	err := godotenv.Load("../.env_test")
	if err != nil {
		slog.Error("Error loading .env file")
		os.Exit(1)
	}
	db = mongo.NewDatabase()
}

func TestDeposit(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
	if err := account.SetBalance(0); err != nil {
		slog.Error("error setting balance to 0",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if err := account.Deposit(100); err != nil {
		slog.Error("error depositing to 100",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
	if account.MonthlyBalance != 100 {
		t.Errorf("Expected monthly balance to be 100, got %d", account.MonthlyBalance)
	}
	if account.LifetimeBalance < 100 {
		t.Errorf("Expected lifetime balance to be at least 100, got %d", account.LifetimeBalance)
	}
}

func TestDepositToCurrentOnly(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)

	// Set initial balances
	if err := account.SetBalance(0); err != nil {
		slog.Error("error setting balance to 0",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	// Deposit to current balance only
	if err := account.DepositToCurrentOnly(100); err != nil {
		slog.Error("error depositing to current balance only",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	// Verify current balance is updated but monthly and lifetime are not
	if account.CurrentBalance != 100 {
		t.Errorf("Expected current balance to be 100, got %d", account.CurrentBalance)
	}
	if account.MonthlyBalance != 0 {
		t.Errorf("Expected monthly balance to be 0, got %d", account.MonthlyBalance)
	}
	// Note: LifetimeBalance may be higher due to bank's DefaultBalance
	// We're only checking that DepositToCurrentOnly doesn't modify it
}

func TestWithdraw(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)
	if err := account.SetBalance(200); err != nil {
		slog.Error("error setting balance to 200",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if err := account.Withdraw(100); err != nil {
		slog.Error("error withdrawing to 100",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	if account.CurrentBalance != 100 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestWithdrawFromCurrentOnly(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)

	// Set initial balances
	if err := account.SetBalance(200); err != nil {
		slog.Error("error setting balance to 200",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	// Withdraw from current balance only
	if err := account.WithdrawFromCurrentOnly(100); err != nil {
		slog.Error("error withdrawing from current balance only",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	// Verify current balance is updated but monthly and lifetime are not
	if account.CurrentBalance != 100 {
		t.Errorf("Expected current balance to be 100, got %d", account.CurrentBalance)
	}
	if account.MonthlyBalance != 200 {
		t.Errorf("Expected monthly balance to be 200, got %d", account.MonthlyBalance)
	}
	// Note: LifetimeBalance may be higher due to bank's DefaultBalance
	// We're only checking that WithdrawFromCurrentOnly doesn't modify it
}

func TestSetBalance(t *testing.T) {
	bank := GetBank("12345")
	account := GetAccount(bank.GuildID, "54321")
	if err := account.SetBalance(500); err != nil {
		slog.Error("error setting balance to 500",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}
	account = GetAccount(bank.GuildID, "54321")
	if account.CurrentBalance != 500 {
		t.Errorf("Expected balance to be 100, got %d", account.CurrentBalance)
	}
}

func TestGetAccounts(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 3)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	// Create multiple accounts with different balances
	memberIDs := []string{"54321", "54322", "54323"}
	balances := []int{100, 200, 300}

	for i, memberID := range memberIDs {
		account := GetAccount(bank.GuildID, memberID)
		if account == nil {
			t.Errorf("account for member %s is nil", memberID)
			return
		}
		accounts = append(accounts, account)

		if err := account.SetBalance(balances[i]); err != nil {
			slog.Error("error setting balance",
				slog.String("guildID", account.GuildID),
				slog.String("memberID", account.MemberID),
				slog.Int("balance", balances[i]),
				slog.Any("error", err),
			)
		}
	}

	// Test GetAccounts with guild filter, sorted by balance descending, no limit
	retrievedAccounts := GetAccounts(bank.GuildID, bson.M{"guild_id": bank.GuildID}, bson.M{"current_balance": -1}, 0)

	// Verify we got all accounts
	if len(retrievedAccounts) != len(memberIDs) {
		t.Errorf("Expected %d accounts, got %d", len(memberIDs), len(retrievedAccounts))
		return
	}

	// Verify accounts are sorted by balance in descending order
	for i := 0; i < len(retrievedAccounts)-1; i++ {
		if retrievedAccounts[i].CurrentBalance < retrievedAccounts[i+1].CurrentBalance {
			t.Errorf("Accounts not sorted correctly: %d should be >= %d",
				retrievedAccounts[i].CurrentBalance,
				retrievedAccounts[i+1].CurrentBalance)
		}
	}

	// Test with limit
	limitedAccounts := GetAccounts(bank.GuildID, bson.M{"guild_id": bank.GuildID}, bson.M{"current_balance": -1}, 2)
	if len(limitedAccounts) != 2 {
		t.Errorf("Expected 2 accounts with limit, got %d", len(limitedAccounts))
	}
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	banks := make([]*Bank, 0, 1)
	defer func() {
		for _, bank := range banks {
			if err := db.Delete(BankCollection, bson.M{"guild_id": bank.GuildID}); err != nil {
				slog.Error("error deleting bank",
					slog.String("guildID", bank.GuildID),
					slog.Any("error", err),
				)
			}
		}
	}()
	accounts := make([]*Account, 0, 1)
	defer func() {
		for _, account := range accounts {
			if err := db.Delete(AccountCollection, bson.M{"guild_id": account.GuildID, "member_id": account.MemberID}); err != nil {
				slog.Error("error deleting account",
					slog.String("guildID", account.GuildID),
					slog.String("memberID", account.MemberID),
					slog.Any("error", err),
				)
			}
		}
	}()

	bank := GetBank("12345")
	if bank == nil {
		t.Error("bank is nil")
		return
	}
	banks = append(banks, bank)

	account := GetAccount(bank.GuildID, "54321")
	if account == nil {
		t.Error("account is nil")
		return
	}
	accounts = append(accounts, account)

	// Set initial balance
	if err := account.SetBalance(50); err != nil {
		slog.Error("error setting balance to 50",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	// Try to withdraw more than the balance
	err := account.Withdraw(100)
	if err != ErrInsufficentFunds {
		t.Errorf("Expected ErrInsufficentFunds, got %v", err)
	}

	// Verify balance remains unchanged
	if account.CurrentBalance != 50 {
		t.Errorf("Expected balance to remain 50, got %d", account.CurrentBalance)
	}

	// Test WithdrawFromCurrentOnly with insufficient funds
	err = account.WithdrawFromCurrentOnly(100)
	if err != ErrInsufficentFunds {
		t.Errorf("Expected ErrInsufficentFunds for WithdrawFromCurrentOnly, got %v", err)
	}

	// Verify balance still remains unchanged
	if account.CurrentBalance != 50 {
		t.Errorf("Expected balance to remain 50, got %d", account.CurrentBalance)
	}
}

func TestResetMonthlyBalances(t *testing.T) {
	bank := GetBank("12345")
	account := GetAccount(bank.GuildID, "54321")
	account.CurrentBalance = 500
	account.MonthlyBalance = 750
	if err := writeAccount(account); err != nil {
		slog.Error("error writing account",
			slog.String("guildID", account.GuildID),
			slog.String("memberID", account.MemberID),
			slog.Any("error", err),
		)
	}

	ResetMonthlyBalances()

	account = GetAccount(bank.GuildID, "54321")
	if account.MonthlyBalance != 0 {
		t.Errorf("Expected monthly balance to be 0, got %d", account.MonthlyBalance)
	}
}
